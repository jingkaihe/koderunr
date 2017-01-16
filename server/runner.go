package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	dcli "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
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
func (rnr *Runner) Run(r io.Reader, w io.Writer, conn redis.Conn, uuid string) {
	Runnerthrottle <- struct{}{}
	defer func() { <-Runnerthrottle }()

	err := rnr.createContainer(uuid)
	if err != nil {
		rnr.logger.Errorf("Container %s cannot be created - %v", uuid, err)
		return
	}

	hijackResp, err := DockerClient.ContainerAttach(context.Background(), rnr.containerID, types.ContainerAttachOptions{
		Stdin:  true,
		Stdout: true,
		Stderr: true,
		Stream: true,
	})

	if err != nil {
		rnr.logger.Errorf("Container %s cannot be attached - %v", rnr.shortContainerID(), err)
		return
	}

	go pipeIn(hijackResp.Conn, r, rnr.logger)
	go pipeOut(hijackResp.Reader, w, rnr.logger)

	// Start running the container
	err = rnr.startContainer()
	if err != nil {
		rnr.logger.Errorf("Container %s cannot be started - %v", rnr.shortContainerID(), err)
		return
	}
	defer func() {
		rnr.logger.Infof("Removing container %s", rnr.containerID)
		DockerClient.ContainerRemove(context.Background(), rnr.containerID, types.ContainerRemoveOptions{
			Force: true,
		})
		rnr.logger.Infof("Container %s removed successfully", rnr.containerID)
	}()

	rnr.waitContainer(w, newWaitCtx(rnr))
}

func pipeIn(stdin net.Conn, r io.Reader, logger *logrus.Logger) {
	io.Copy(stdin, r)
}

func pipeOut(r *bufio.Reader, w io.Writer, logger *logrus.Logger) {
	if _, err := stdcopy.StdCopy(w, w, r); err != nil {
		logger.Error(err)
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

func (rnr *Runner) image() string {
	selectedVersion := rnr.Version
	availableVersions := (*appConfig.Languages)[rnr.Lang].Versions

	if selectedVersion == "" {
		if len(availableVersions) > 0 {
			selectedVersion = availableVersions[0]
		} else {
			selectedVersion = "latest"
		}
	}
	return fmt.Sprintf("%s:%s", imageMapper[rnr.Lang], selectedVersion)
}

func (rnr *Runner) createContainer(uuid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	cmd := []string{rnr.Source, uuid}
	lang := (*appConfig.Languages)[rnr.Lang]

	ctr, err := DockerClient.ContainerCreate(ctx,
		&container.Config{
			Cmd:             cmd,
			Image:           rnr.image(),
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

	rnr.containerID = ctr.ID
	return nil
}

func (rnr *Runner) startContainer() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	return DockerClient.ContainerStart(ctx, rnr.containerID, types.ContainerStartOptions{})

}

func (rnr *Runner) shortContainerID() string {
	return rnr.containerID[:7]
}

func (rnr *Runner) waitContainer(w io.Writer, wctx WaitCtx) {
	defer wctx.Cancel()

	go func() {
		_, err := DockerClient.ContainerWait(wctx, rnr.containerID)
		if err == nil {
			wctx.ChSucceed() <- struct{}{}
		}
	}()

	select {
	case <-wctx.ChSucceed():
		rnr.logger.Infof("Container %s is executed successfully", rnr.shortContainerID())
	case <-wctx.ChClose():
		DockerClient.ContainerStop(context.Background(), rnr.containerID, nil)
		rnr.logger.Infof("Container %s is stopped since the streamming has been halted", rnr.shortContainerID())
	case <-wctx.Done():
		switch wctx.Err() {
		case context.DeadlineExceeded:
			msg := fmt.Sprintf("Container %s is terminated caused by %d sec timeout\n", rnr.shortContainerID(), rnr.Timeout)
			rnr.logger.Error(msg)
			fmt.Fprintf(w, "%s\n", msg)
		default:
			rnr.logger.Error(wctx.Err())
		}
	}
}
