package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"time"

	dcli "github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/container"
	"github.com/docker/engine-api/types/network"
	"github.com/garyburd/redigo/redis"
	// "github.com/docker/engine-api/types/strslice"
	"golang.org/x/net/context"
)

// Runner runs the code
type Runner struct {
	Lang          string `json:"lang"`
	Source        string `json:"source"`
	Version       string `json:"version"`
	Timeout       int    `json:"timeout"` // How long is the code going to run
	closeNotifier <-chan bool
	client        *dcli.Client
}

// Runnerthrottle Limit the max throttle for runner
var Runnerthrottle chan struct{}

// Run the code in the container
func (r *Runner) Run(output messages, redisConn redis.Conn, uuid string) {
	Runnerthrottle <- struct{}{}
	defer func() { <-Runnerthrottle }()

	var err error
	r.client, err = NewDockerClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}

	resp, err := r.createContainer(uuid)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}

	err = r.startContainer(resp.ID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
	defer r.client.ContainerRemove(
		context.Background(),
		resp.ID,
		types.ContainerRemoveOptions{Force: true},
	)

	attachResponse, err := r.attachContainer(resp.ID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
	defer func() {
		attachResponse.CloseWrite()
		attachResponse.Close()
	}()

	go pipeStdout(attachResponse.Reader, output)
	go pipeStdin(redisConn, uuid, attachResponse.Conn)

	done := make(chan error)
	go r.waitContainer(uuid, done)

	select {
	case <-r.closeNotifier:
		fmt.Fprintf(os.Stdout, "Container %s is stopped since the streamming has been halted\n", uuid)
	case err := <-done:
		if err == nil {
			fmt.Fprintf(os.Stdout, "Container %s is executed successfully\n", uuid)
		} else {
			msg := fmt.Sprintf("Container %s failed - %v\n", uuid, err)
			output <- msg
			close(output)

			fmt.Fprintf(os.Stderr, msg)
		}
	}
}

func pipeStdin(redisConn redis.Conn, uuid string, stdin io.WriteCloser) {
	psc := redis.PubSubConn{Conn: redisConn}
	psc.Subscribe(uuid + "#stdin")

	defer func() {
		psc.Unsubscribe(uuid + "#stdin")
		psc.Close()
		redisConn.Close()
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

func pipeStdout(stdout *bufio.Reader, output messages) {
	buffer := make([]byte, 512)
	for {
		n, err := stdout.Read(buffer)
		if err != nil {
			if err != io.EOF {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
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

// NewDockerClient creates a new docker client using Unix sock connection
func NewDockerClient() (*dcli.Client, error) {
	defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.23"}
	return dcli.NewClient("unix:///var/run/docker.sock", "v1.23", nil, defaultHeaders)
}

func (r *Runner) createContainer(uuid string) (types.ContainerCreateResponse, error) {
	cmd := []string{r.Source, uuid}

	if r.Version != "" {
		cmd = append(cmd, r.Version)
	}
	config := container.Config{
		Image:           r.image(),
		NetworkDisabled: true,
		OpenStdin:       true,
		Cmd:             cmd,
	}

	resource := container.Resources{
		CPUQuota:     40000,
		Memory:       50 * 1024 * 1024,
		PidsLimit:    50,
		KernelMemory: 4 * 1024 * 1024,
	}

	hostConfig := container.HostConfig{Resources: resource}

	networkConfig := network.NetworkingConfig{}
	return r.client.ContainerCreate(
		context.Background(),
		&config,
		&hostConfig,
		&networkConfig,
		uuid,
	)
}

func (r *Runner) startContainer(uuid string) error {
	return r.client.ContainerStart(context.Background(), uuid)
}

func (r *Runner) attachContainer(uuid string) (types.HijackedResponse, error) {
	return r.client.ContainerAttach(
		context.Background(),
		uuid,
		types.ContainerAttachOptions{
			Stream: true,
			Stdin:  true,
			Stdout: true,
			Stderr: true,
		},
	)
}

func (r *Runner) waitContainer(uuid string, done chan<- error) {
	timeout := time.Duration(r.Timeout) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	sig, err := r.client.ContainerWait(ctx, uuid)
	if err != nil {
		done <- err
		return
	}

	if sig != 0 {
		done <- fmt.Errorf("Runner exit with code %d", sig)
		return
	}

	done <- nil
}
