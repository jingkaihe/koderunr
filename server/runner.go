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

// WaitCtx is the context for the container wait
type WaitCtx struct {
	context.Context
	Cancel context.CancelFunc
}

func newWaitCtx(r *Runner) WaitCtx {
	ctx := context.WithValue(context.Background(), "close", r.closeNotifier)
	ctx = context.WithValue(ctx, "succeed", make(chan struct{}))

	wctx := WaitCtx{}
	wctx.Context, wctx.Cancel = context.WithTimeout(ctx, time.Duration(r.Timeout)*time.Second)

	return wctx
}

// ChSucceed delivers the message that the context's been finished successfully
func (w WaitCtx) ChSucceed() chan struct{} {
	return w.Value("succeed").(chan struct{})
}

// ChClose deliver the message that the context's forced to be closed
func (w WaitCtx) ChClose() <-chan bool {
	return w.Value("close").(<-chan bool)
}

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

	go func(r *Runner, uuid string) {
		err = r.attachContainer(stdoutWriter, stdinReader)
		if err != nil {
			r.logger.Errorf("Container %s cannot be attached - %v", r.shortContainerID(), err)
		}
	}(r, uuid)

	// Start running the container
	err = r.startContainer()
	if err != nil {
		r.logger.Errorf("Container %s cannot be started - %v", r.shortContainerID(), err)
		return
	}
	defer func() {
		r.logger.Infof("Removing container %s", r.containerID)
		DockerClient.ContainerRemove(context.Background(), r.containerID, types.ContainerRemoveOptions{
			Force: true,
		})
		r.logger.Infof("Container %s removed successfully", r.containerID)
	}()

	r.waitContainer(output, newWaitCtx(r))
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
		io.Copy(stdoutWriter, reader)
		if err != nil {
			if err != io.EOF {
				r.logger.Error(err)
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

func (r *Runner) shortContainerID() string {
	return r.containerID[:7]
}

func (r *Runner) waitContainer(output messages, wctx WaitCtx) {
	defer wctx.Cancel()

	go func() {
		_, err := DockerClient.ContainerWait(wctx, r.containerID)
		if err == nil {
			wctx.ChSucceed() <- struct{}{}
		}
	}()

	select {
	case <-wctx.ChSucceed():
		r.logger.Infof("Container %s is executed successfully", r.shortContainerID())
	case <-wctx.ChClose():
		DockerClient.ContainerStop(context.Background(), r.containerID, nil)
		r.logger.Infof("Container %s is stopped since the streamming has been halted", r.shortContainerID())
	case <-wctx.Done():
		switch wctx.Err() {
		case context.DeadlineExceeded:
			msg := fmt.Sprintf("Container %s is terminated caused by %d sec timeout\n", r.shortContainerID(), r.Timeout)
			r.logger.Error(msg)
			output <- msg
		default:
			r.logger.Error(wctx.Err())
		}
	}
}
