package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
	"github.com/garyburd/redigo/redis"
)

// DockerClient for running code
var DockerClient *docker.Client

// DockerAPIVersion is the API version connect to docker
const DockerAPIVersion = "1.23"

// Runner runs the code
type Runner struct {
	Lang          string `json:"lang"`
	Source        string `json:"source"`
	Version       string `json:"version"`
	Timeout       int    `json:"timeout"` // How long is the code going to run
	closeNotifier <-chan bool
	logger        *logrus.Logger
	containerID   string
}

// Runnerthrottle Limit the max throttle for runner
var Runnerthrottle chan struct{}

// FetchCode get the code from Redis Server according to the UUID
func FetchCode(uuid string, redisConn redis.Conn) (r *Runner, err error) {
	value, err := redis.Bytes(redisConn.Do("GET", uuid+"#run"))

	if err != nil {
		return
	}

	r = &Runner{}
	err = json.Unmarshal(value, r)
	return
}

// Run the code in the container
func (r *Runner) Run(output messages, conn redis.Conn, uuid string) {
	Runnerthrottle <- struct{}{}
	defer func() { <-Runnerthrottle }()

	err := r.createContainer(uuid)
	if err != nil {
		r.logger.Errorf("Container %s cannot be created - %v", uuid, err)
		return
	}

	stdoutReader, stdoutWriter := io.Pipe()
	stdinReader, stdinWriter := io.Pipe()

	if err != nil {
		r.logger.Error(err)
	}

	defer stdinWriter.Close()
	defer stdoutWriter.Close()

	go pipeStdin(conn, uuid, stdinWriter, r.logger)
	go pipeStdout(stdoutReader, output, r.logger)

	// Start running the container
	err = r.startContainer()
	if err != nil {
		r.logger.Errorf("Container %s cannot be started - %v", uuid, err)
		return
	}
	defer DockerClient.RemoveContainer(docker.RemoveContainerOptions{ID: r.containerID, Force: true})

	successChan := make(chan struct{})
	errorChan := make(chan error)

	go r.waitContainer(successChan, errorChan)

	go func(r *Runner, uuid string) {
		err = r.attachContainer(stdoutWriter, stdinReader)
		if err != nil {
			r.logger.Errorf("Container %s cannot be attached - %v", uuid, err)
		}
	}(r, uuid)

	select {
	case <-r.closeNotifier:
		DockerClient.StopContainer(r.containerID, 0)
		r.logger.Infof("Container %s is stopped since the streamming has been halted", uuid)
	case <-successChan:
		r.logger.Infof("Container %s is executed successfully", uuid)
	case err := <-errorChan:
		r.logger.Errorf("Container %s failed caused by - %v", uuid, err)
	case <-time.After(time.Duration(r.Timeout) * time.Second):
		msg := fmt.Sprintf("Container %s is terminated caused by 15 sec timeout\n", uuid)
		r.logger.Error(msg)
		output <- msg
	}
}

func pipeStdin(conn redis.Conn, uuid string, stdin *io.PipeWriter, logger *logrus.Logger) {
	psc := redis.PubSubConn{Conn: conn}
	psc.Subscribe(uuid + "#stdin")

	defer func() {
		psc.Unsubscribe(uuid + "#stdin")
		psc.Close()
		conn.Close()
	}()

StdinSubscriptionLoop:
	for {
		switch n := psc.Receive().(type) {
		case redis.Message:
			stdinData := strconv.QuoteToASCII(string(n.Data))
			logger.Infof("Message: %s %s", n.Channel, stdinData)
			stdin.Write(n.Data)
		case error:
			break StdinSubscriptionLoop
		}
	}
	logger.Info("Stdin subscription closed")
}

func pipeStdout(stdout *io.PipeReader, output messages, logger *logrus.Logger) {
	buffer := make([]byte, 512)
	for {
		n, err := stdout.Read(buffer)
		if err != nil {
			if err != io.EOF {
				stdout.Close()
				logger.Error(err)
			}

			close(output)
			break
		}

		data := buffer[0:n]
		output <- string(data)

		// Clear the buffer
		for i := 0; i < n; i++ {
			buffer[i] = 0
		}
	}
}

// NewDockerClient create a new docker client
func NewDockerClient() (*docker.Client, error) {
	dockerHost := os.Getenv("DOCKER_HOST")

	// If DOCKER_HOST exists in env, using docker-machine
	if dockerHost != "" {
		return docker.NewVersionedClientFromEnv(DockerAPIVersion)
	}

	// Otherwise using sock connection
	//TODO: Deal with the TLS case (even though you are not using it for now)
	endpoint := "unix:///var/run/docker.sock"
	return docker.NewVersionedClient(endpoint, DockerAPIVersion)
}

var imageMapper = map[string]string{
	"swift":  "koderunr-swift",
	"ruby":   "koderunr-ruby",
	"python": "koderunr-python",
	"go":     "koderunr-go",
	"c":      "koderunr-c",
	"elixir": "koderunr-erl",
}

func (r *Runner) image() string {
	return imageMapper[r.Lang]
}

func (r *Runner) createContainer(uuid string) error {
	cmd := []string{r.Source, uuid}

	if r.Version != "" {
		cmd = append(cmd, r.Version)
	}
	container, err := DockerClient.CreateContainer(docker.CreateContainerOptions{
		Name: uuid,
		Config: &docker.Config{
			Image:           r.image(),
			NetworkDisabled: true,
			OpenStdin:       true,
			Cmd:             cmd,
			KernelMemory:    1024 * 1024 * 4,
			PidsLimit:       100,
		},
	})

	if err != nil {
		return err
	}

	r.containerID = container.ID
	return nil
}

func (r *Runner) startContainer() error {
	return DockerClient.StartContainer(r.containerID, &docker.HostConfig{
		CPUQuota:   20000,
		MemorySwap: -1,
		Privileged: false,
		CapDrop:    []string{"all"},
		Memory:     80 * 1024 * 1024, // so the memory swap will be the same size
	})
}

func (r *Runner) attachContainer(stdoutWriter *io.PipeWriter, stdinReader *io.PipeReader) error {
	return DockerClient.AttachToContainer(docker.AttachToContainerOptions{
		Container:    r.containerID,
		Stdin:        true,
		Stdout:       true,
		Stderr:       true,
		Stream:       true,
		OutputStream: stdoutWriter,
		ErrorStream:  stdoutWriter,
		InputStream:  stdinReader,
	})
}

func (r *Runner) waitContainer(successChan chan<- struct{}, errorChan chan<- error) {
	_, err := DockerClient.WaitContainer(r.containerID)
	if err != nil {
		r.logger.Error(err)
		errorChan <- err
		return
	}
	successChan <- struct{}{}
}
