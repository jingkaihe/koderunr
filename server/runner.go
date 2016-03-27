package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
)

// Runner runs the code
type Runner struct {
	Ext    string `json:"ext"`
	Source string `json:"source"`
}

// Run the code in the container
func (r *Runner) Run(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/octet-stream")
	cmd := exec.Command("docker", "run", "-i", "koderunr", r.Ext, r.Source)

	pipeReader, pipeWriter := io.Pipe()
	defer pipeWriter.Close()
	defer pipeWriter.Close()

	cmd.Stdout = pipeWriter
	cmd.Stderr = pipeWriter

	// Doing the streaming
	go func() {
		buffer := make([]byte, 256)
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
			w.Write(data)

			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}

			for i := 0; i < n; i++ {
				buffer[i] = 0
			}
		}
	}()

	cmd.Run()
}
