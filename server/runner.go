package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

// Runner runs the code
type Runner struct {
	Ext     string `json:"ext"`
	Source  string `json:"source"`
	Version string `json:"version"`
}

// Run the code in the container
func (r *Runner) Run(input, output messages) {
	execArgs := []string{"run", "-i", "koderunr", r.Ext, r.Source}
	if r.Version != "" {
		execArgs = append(execArgs, r.Version)
	}

	cmd := exec.Command("docker", execArgs...)

	pipeReader, pipeWriter := io.Pipe()
	defer pipeWriter.Close()

	cmd.Stdout = pipeWriter
	cmd.Stderr = pipeWriter

	// Doing the streaming
	go func() {
		buffer := make([]byte, 512)
		for {
			n, err := pipeReader.Read(buffer)
			if err != nil {
				if err != io.EOF {
					pipeReader.Close()
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
