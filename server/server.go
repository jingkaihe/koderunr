package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
)

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/octet-stream")

	source := r.FormValue("source")
	langExt := r.FormValue("lang")

	cmd := exec.Command("docker", "run", "-i", "koderunr", langExt, source)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	go func() {
		if _, err := io.Copy(w, stdout); err != nil {
			fmt.Fprintf(os.Stdout, "Error: %v", err)
		}
	}()

	go func() {
		if _, err := io.Copy(w, stderr); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v", err)
		}
	}()

	cmd.Start()
	cmd.Wait()
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
