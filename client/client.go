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

const endpoint = "http://127.0.0.1:8080/"

func isGoProg(args []string) bool {
	if len(args) != 3 {
		return false
	}
	return args[0] == "go" && args[1] == "run"
}

type runner struct {
	ext    string
	source string
	uuid   string
}

func newRunner(fName string) (r *runner, err error) {

	ext := path.Ext(fName)
	if ext != ".go" {
		err = fmt.Errorf("the File extension %s is not go", ext)
	}

	ctx, err := ioutil.ReadFile(fName)

	r = &runner{
		ext:    ext,
		source: string(ctx),
	}
	return
}

func (r *runner) fetchUUID() error {
	resp, err := http.PostForm(endpoint+"register/",
		url.Values{"ext": {r.ext}, "source": {string(r.source)}})
	defer resp.Body.Close()

	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	r.uuid = string(body)
	return nil
}

func (r *runner) run() error {
	// TODO: Build the URI in a classy way
	resp, err := http.Get(endpoint + "run?uuid=" + r.uuid)
	defer resp.Body.Close()

	if err != nil {
		return err
	}

	if _, err := io.Copy(os.Stdout, resp.Body); err != nil {
		if err != io.EOF {
			return err
		}
	}

	return nil
}

// examples:
//   $ koderunr go run main.go
//   $ koderunr ruby hello.rb
func main() {
	args := os.Args[1:]
	if isGoProg(args) {
		fName := args[2]

		rnr, err := newRunner(fName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		err = rnr.fetchUUID()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		rnr.run()
	}
}
