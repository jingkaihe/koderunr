package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"time"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	dcli "github.com/docker/docker/client"
	"github.com/garyburd/redigo/redis"
)

// DockerClient for running code
var DockerClient *dcli.Client

// DockerAPIVersion is the API version connect to docker
const DockerAPIVersion = "1.24"

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
	defer DockerClient.ContainerRemove(context.Background(), r.containerID, types.ContainerRemoveOptions{
		Force: true,
	})

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
		DockerClient.ContainerStop(context.Background(), r.containerID, nil)
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
func NewDockerClient() (*dcli.Client, error) {
	os.Setenv("DOCKER_API_VERSION", DockerAPIVersion)
	return dcli.NewEnvClient()
}

var imageMapper = map[string]string{
	"swift":  "koderunr-swift",
	"ruby":   "koderunr-ruby",
	"python": "koderunr-python",
	"go":     "koderunr-go",
	"c":      "koderunr-c",
	"dotnet": "koderunr-dotnet",
	"fsharp": "koderunr-fsharp",
}

func (r *Runner) image() string {
	selectedVersion := r.Version
	availableVersions := (*appConfig.Languages)[r.Lang].Versions

	if selectedVersion == "" {
		if len(availableVersions) > 0 {
			selectedVersion = availableVersions[0]
		} else {
			selectedVersion = "latest"
		}
	}
	return fmt.Sprintf("%s:%s", imageMapper[r.Lang], selectedVersion)
}

func (r *Runner) createContainer(uuid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	cmd := []string{r.Source, uuid}
	lang := (*appConfig.Languages)[r.Lang]

	ctr, err := DockerClient.ContainerCreate(ctx,
		&container.Config{
			Cmd:             cmd,
			Image:           r.image(),
			OpenStdin:       true,
			AttachStdin:     true,
			AttachStdout:    true,
			AttachStderr:    true,
			NetworkDisabled: true,
		},
		&container.HostConfig{
			Privileged: false,
			CapDrop:    []string{"all"},
			Resources: container.Resources{
				CPUQuota:   lang.GetCPUQuota(),
				MemorySwap: -1,
				Memory:     lang.GetMemory(),
				PidsLimit:  lang.GetPidsLimit(),
			},
		},
		&network.NetworkingConfig{},
		uuid,
	)

	if err != nil {
		return err
	}

	r.containerID = ctr.ID
	return nil
}

func (r *Runner) startContainer() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	return DockerClient.ContainerStart(ctx, r.containerID, types.ContainerStartOptions{})

}

func (r *Runner) attachContainer(stdoutWriter *io.PipeWriter, stdinReader *io.PipeReader) error {
	hijackResp, err := DockerClient.ContainerAttach(context.Background(), r.containerID, types.ContainerAttachOptions{
		Stdin:  true,
		Stdout: true,
		Stderr: true,
		Stream: true,
	})

	if err != nil {
		return err
	}

	go func(reader *bufio.Reader) {
		for {
			io.Copy(stdoutWriter, reader)
			if err != nil {
				if err != io.EOF {
					r.logger.Error(err)
				}
				break
			}
		}
	}(hijackResp.Reader)

	go func(writer net.Conn) {
		scanner := bufio.NewScanner(stdinReader)
		for scanner.Scan() {
			fmt.Fprintf(writer, "%s\n", scanner.Text())
		}
	}(hijackResp.Conn)

	return nil
}

func (r *Runner) waitContainer(successChan chan<- struct{}, errorChan chan<- error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.Timeout)*time.Second)
	defer cancel()

	_, err := DockerClient.ContainerWait(ctx, r.containerID)

	if err != nil {
		r.logger.Error(err)
		errorChan <- err
		return
	}
	successChan <- struct{}{}
}
