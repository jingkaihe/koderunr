package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"

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

	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
	}
	defer stdin.Close()

	stdoutReader, stdoutWriter := io.Pipe()

	defer stdoutWriter.Close()

	cmd.Stdout = stdoutWriter
	cmd.Stderr = stdoutWriter

	go func() {
		psc := redis.PubSubConn{Conn: conn}
		psc.Subscribe(uuid + "#stdin")
		defer psc.Close()

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
		fmt.Println("Done")
	}()

	// Doing the streaming
	go func() {
		buffer := make([]byte, 512)
		for {
			n, err := stdoutReader.Read(buffer)
			if err != nil {
				if err != io.EOF {
					stdoutReader.Close()
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				}
				break
			}

			data := buffer[0:n]
			output <- string(data)

			for i := 0; i < n; i++ {
				buffer[i] = 0
			}
		}
	}()

	cmd.Run()
}
