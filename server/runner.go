package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/garyburd/redigo/redis"
)

// Runner runs the code
type Runner struct {
	Ext     string `json:"ext"`
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

	execArgs := []string{
		"run",
		"-i",            // run in interactive mode
		"--rm",          // automatically remove the container when it exits
		"--net", "none", // disables all incoming and outgoing networking
		"--cpu-quota=15000", // a container can use 15% of a CPU resource
		"--memory='50mb'",   // use 50mb mem
		"--name", uuid,      // Give the runner a name so we can force kill it accordingly
		"koderunr", r.Ext, r.Source}
	if r.Version != "" {
		execArgs = append(execArgs, r.Version)
	}

	cmd := exec.Command("docker", execArgs...)

	stdoutReader, stdoutWriter := io.Pipe()
	cmd.Stdout = stdoutWriter
	cmd.Stderr = stdoutWriter

	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
	}

	defer stdin.Close()
	defer stdoutWriter.Close()

	wg := sync.WaitGroup{}
	wg.Add(2)

	go pipeStdin(conn, uuid, stdin, wg)
	go pipeStdout(stdoutReader, output, wg)

	// Start running the container
	cmd.Start()
	done := make(chan error, 1)

	// Receive message when the job get done
	go func() {
		done <- cmd.Wait()
	}()

	select {
	// gracefully Kill the container when it's being hanging around for too long
	case <-time.After(time.Duration(r.Timeout) * time.Second):
		if err := cmd.Process.Signal(syscall.SIGKILL); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to shotdown %v\n", err)
		} else {
			timeoutMsg := fmt.Sprintf("Running container %s is shutdown since reached the timeout %ds\n", uuid, r.Timeout)
			fmt.Fprintf(os.Stdout, timeoutMsg)
			output <- timeoutMsg
		}
	// when the container has finished
	case err := <-done:
		if err != nil {
			fmt.Printf("Container %s done with error = %v\n", uuid, err)
		} else {
			fmt.Printf("Container %s done gracefully without error\n", uuid)
		}
	}

	// Force kill the container
	exec.Command("docker", "rm", "-f", uuid).Run()
}

func pipeStdin(conn redis.Conn, uuid string, stdin io.WriteCloser, wg sync.WaitGroup) {
	psc := redis.PubSubConn{Conn: conn}
	psc.Subscribe(uuid + "#stdin")

	defer func() {
		psc.Unsubscribe(uuid + "#stdin")
		psc.Close()
		conn.Close()
		wg.Done()
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

func pipeStdout(stdout *io.PipeReader, output messages, wg sync.WaitGroup) {
	defer wg.Done()

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
