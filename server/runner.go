package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/garyburd/redigo/redis"
)

// Runner runs the code
type Runner struct {
	Lang          string `json:"lang"`
	Source        string `json:"source"`
	Version       string `json:"version"`
	Timeout       int    `json:"timeout"` // How long is the code going to run
	closeNotifier <-chan bool
}

// Runnerthrottle Limit the max throttle for runner
var Runnerthrottle chan struct{}

// Run the code in the container
func (r *Runner) Run(output messages, conn redis.Conn, uuid string) {
	Runnerthrottle <- struct{}{}
	defer func() { <-Runnerthrottle }()

	execArgs := []string{
		"run",
		"-i",            // run in interactive mode
		"--net", "none", // disables all incoming and outgoing networking
		"--cpu-quota=40000", // a container can use 15% of a CPU resource
		"--pids-limit=200",
		"--memory='50mb'", // use 50mb mem
		"--memory-swap=0",
		"--name", uuid, // Give the runner a name so we can force kill it accordingly
		r.image(),
		r.Source,
		uuid,
	}

	if r.Version != "" {
		execArgs = append(execArgs, r.Version)
	}

	cmd := exec.Command("docker", execArgs...)

	stdoutReader, stdoutWriter := io.Pipe()
	cmd.Stdout = stdoutWriter
	cmd.Stderr = stdoutWriter

	stdinWriter, err := cmd.StdinPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
	}

	defer stdinWriter.Close()
	defer stdoutWriter.Close()

	go pipeStdin(conn, uuid, stdinWriter)
	go pipeStdout(stdoutReader, output)

	cmd.Start()
	// Kill the container
	defer exec.Command("docker", "rm", "-f", uuid).Run()

	done := make(chan error)
	go func(cmd *exec.Cmd) {
		done <- cmd.Wait()
	}(cmd)

	select {
	case <-r.closeNotifier:
		fmt.Fprintf(os.Stdout, "Container %s is stopped since the streamming has been halted\n", uuid)
	case err := <-done:
		if err == nil {
			fmt.Fprintf(os.Stdout, "Container %s is executed successfully\n", uuid)
		} else {
			fmt.Fprintf(os.Stderr, "Container %s failed due to %v\n", uuid, err)
		}

	case <-time.After(time.Duration(r.Timeout) * time.Second):
		msg := fmt.Sprintf("Container %s is terminated caused by 15 sec timeout\n", uuid)
		fmt.Fprintf(os.Stderr, msg)
		output <- msg
	}
}

func pipeStdin(conn redis.Conn, uuid string, stdin io.WriteCloser) {
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
