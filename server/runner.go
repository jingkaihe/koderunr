package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/garyburd/redigo/redis"
)

// Runner runs the code
type Runner struct {
	Ext     string `json:"ext"`
	Source  string `json:"source"`
	Version string `json:"version"`
}

// Run the code in the container
func (r *Runner) Run(output messages, conn redis.Conn, uuid string) {
	execArgs := []string{"run", "-i", "koderunr", r.Ext, r.Source}
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

	cmd.Run()
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

		for i := 0; i < n; i++ {
			buffer[i] = 0
		}
	}
}
