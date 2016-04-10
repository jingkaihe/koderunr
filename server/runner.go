package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/fsouza/go-dockerclient"
	"github.com/garyburd/redigo/redis"
)

// DockerClient for running code
var DockerClient *docker.Client

// Runner runs the code
type Runner struct {
	Lang    string `json:"lang"`
	Source  string `json:"source"`
	Version string `json:"version"`
	Timeout int    `json:"timeout"` // How long is the code going to run
}

// Runnerthrottle Limit the max throttle for runner
var Runnerthrottle chan struct{}

// Run the code in the container
func (r *Runner) Run(output messages, conn redis.Conn, uuid string) {
	Runnerthrottle <- struct{}{}
	defer func() { <-Runnerthrottle }()

	container, err := r.createContainer(uuid)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Container %s cannot be created - %v\n", uuid, err)
		return
	}

	stdoutReader, stdoutWriter := io.Pipe()
	stdinReader, stdinWriter := io.Pipe()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
	}

	defer stdinWriter.Close()
	defer stdoutWriter.Close()

	go pipeStdin(conn, uuid, stdinWriter)
	go pipeStdout(stdoutReader, output)

	// Start running the container
	err = r.startContainer(container.ID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Container %s cannot be started - %v\n", uuid, err)
		return
	}
	defer DockerClient.RemoveContainer(docker.RemoveContainerOptions{ID: container.ID, Force: true})

	successChan := make(chan struct{})
	errorChan := make(chan error)

	go func() {
		_, err := DockerClient.WaitContainer(container.ID)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			errorChan <- err
			return
		}
		successChan <- struct{}{}
	}()

	go func() {
		err = r.attachContainer(container.ID, stdoutWriter, stdinReader)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Container %s cannot be attached - %v\n", uuid, err)
		}
	}()

	select {
	case <-successChan:
		fmt.Fprintf(os.Stdout, "Container %s is executed successfully\n", uuid)
	case err := <-errorChan:
		fmt.Fprintf(os.Stdout, "Container %s failed caused by - %v\n", uuid, err)
	case <-time.After(time.Duration(r.Timeout) * time.Second):
		msg := fmt.Sprintf("Container %s is terminated caused by 15 sec timeout\n", uuid)
		fmt.Fprintf(os.Stderr, msg)
		output <- msg
	}
}

func pipeStdin(conn redis.Conn, uuid string, stdin *io.PipeWriter) {
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
			fmt.Printf("Message: %s %s\n", n.Channel, n.Data)
			stdin.Write(n.Data)
		case error:
			break StdinSubscriptionLoop
		}
	}
	fmt.Println("Stdin subscription closed")
}

func pipeStdout(stdout *io.PipeReader, output messages) {
	buffer := make([]byte, 512)
	for {
		n, err := stdout.Read(buffer)
		if err != nil {
			if err != io.EOF {
				stdout.Close()
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
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

func getDockerClient() (*docker.Client, error) {
	dockerHost := os.Getenv("DOCKER_HOST")

	// If DOCKER_HOST exists in env, using docker-machine
	if dockerHost != "" {
		return docker.NewClientFromEnv()
	}

	// Otherwise using sock connection
	//TODO: Deal with the TLS case (even though you are not using it for now)
	endpoint := "unix:///var/run/docker.sock"
	return docker.NewClient(endpoint)
}

var imageMapper = map[string]string{
	"ruby":   "koderunr-ruby",
	"python": "koderunr-python",
	"go":     "koderunr-go",
	"c":      "koderunr-c",
	"elixir": "koderunr-erl",
}

func (r *Runner) image() string {
	return imageMapper[r.Lang]
}

func (r *Runner) createContainer(uuid string) (*docker.Container, error) {
	cmd := []string{r.Source, uuid}

	if r.Version != "" {
		cmd = append(cmd, r.Version)
	}
	return DockerClient.CreateContainer(docker.CreateContainerOptions{
		Name: uuid,
		Config: &docker.Config{
			Image:           r.image(),
			NetworkDisabled: true,
			OpenStdin:       true,
			Cmd:             cmd,
		},
	})
}

func (r *Runner) startContainer(containerID string) error {
	return DockerClient.StartContainer(containerID, &docker.HostConfig{
		CPUQuota: 40000,
		Memory:   50 * 1024 * 1024,
	})
}

func (r *Runner) attachContainer(containerID string, stdoutWriter *io.PipeWriter, stdinReader *io.PipeReader) error {
	return DockerClient.AttachToContainer(docker.AttachToContainerOptions{
		Container:    containerID,
		Stdin:        true,
		Stdout:       true,
		Stderr:       true,
		Stream:       true,
		OutputStream: stdoutWriter,
		ErrorStream:  stdoutWriter,
		InputStream:  stdinReader,
	})
}
