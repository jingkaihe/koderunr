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

type runner struct {
	ext    string
	source string
	uuid   string
}

func newRunner(fName string) (r *runner, err error) {
	ext := path.Ext(fName)

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
//   $ koderunr run main.go
//   $ koderunr run hello.rb
func main() {
	args := os.Args[1:]
	op := args[0]

	if op == "run" {
		fName := args[1]

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
