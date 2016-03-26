package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
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

		// TODO: Deal with the HTTP timeout
		resp, err := http.PostForm("http://127.0.0.1:8080/",
			url.Values{"lang": {ext}, "source": {string(ctx)}})

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		if _, err := io.Copy(os.Stdout, resp.Body); err != nil {
			if err != io.EOF {
				fmt.Fprintf(os.Stdout, "Error: %v", err)
			}
		}
	}
}
