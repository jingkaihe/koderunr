package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
)

func streaming(w http.ResponseWriter, pipeReader *io.PipeReader) {
	buffer := make([]byte, 256)
	for {
		n, err := pipeReader.Read(buffer)
		if err != nil {
			if err != io.EOF {
				pipeReader.Close()
				fmt.Fprintf(os.Stderr, "Error: %v", err)
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
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/octet-stream")

	source := r.FormValue("source")
	langExt := r.FormValue("lang")

	cmd := exec.Command("docker", "run", "-i", "koderunr", langExt, source)

	pipeReader, pipeWriter := io.Pipe()
	defer pipeWriter.Close()
	defer pipeWriter.Close()

	cmd.Stdout = pipeWriter
	cmd.Stderr = pipeWriter

	go streaming(w, pipeReader)

	cmd.Run()

	fmt.Println("Request finished")
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
