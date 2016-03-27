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
func (r *Runner) Run(w http.ResponseWriter, isEvtStream bool) {
	f, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "The server does not support streaming!", http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

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

			if isEvtStream == true {
				fmt.Fprintf(w, "data: %s\n", string(data))
			} else {
				w.Write(data)
			}

			f.Flush()

			for i := 0; i < n; i++ {
				buffer[i] = 0
			}
		}
	}()

	cmd.Run()
}
