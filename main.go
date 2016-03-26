package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
)

func isGoProg(args []string) bool {
	if len(args) != 3 {
		return false
	}
	return args[0] == "go" && args[1] == "run"
}

// examples:
//   $ koderunr go run main.go
//   $ koderunr ruby hello.rb
func main() {
	args := os.Args[1:]
	if isGoProg(args) {
		fName := args[2]

		ext := path.Ext(fName)
		if ext != ".go" {
			fmt.Fprintf(os.Stderr, "the File extension %s is not go", ext)
			os.Exit(1)
		}

		ctx, err := ioutil.ReadFile(fName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot open %s: %v\n", fName, err)
			os.Exit(1)
		}

		cmd := exec.Command("docker", "run", "-i", "koderunr", ext, string(ctx))
		stdout, err := cmd.StdoutPipe()
		stderr, err := cmd.StderrPipe()

		go func() {
			if _, err := io.Copy(os.Stdout, stdout); err != nil {
				fmt.Fprintf(os.Stdout, "Error: %v", err)
			}
		}()

		go func() {
			if _, err := io.Copy(os.Stdout, stderr); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v", err)
			}
		}()

		cmd.Start()
		cmd.Wait()
		// fmt.Scanf("Press Anykey to continue")
	}
}
