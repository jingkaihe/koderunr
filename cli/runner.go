package main

import (
	"flag"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
)

var endpoint string

func init() {
	flag.StringVar(&endpoint, "Endpoint", "http://127.0.0.1:8080/", "The endpoint of the API that will run the code")
}

// Runner contains the code to be run
type Runner struct {
	ext    string
	source string
	uuid   string
}

func newRunner(fName string) (r *Runner, err error) {
	ext := path.Ext(fName)

	ctx, err := ioutil.ReadFile(fName)

	r = &Runner{
		ext:    ext,
		source: string(ctx),
	}
	return
}

func (r *Runner) fetchUUID() error {
	resp, err := http.PostForm(endpoint+"register/",
		url.Values{"ext": {r.ext}, "source": {string(r.source)}})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	r.uuid = string(body)
	return nil
}

func (r *Runner) run() error {
	// TODO: Build the URI in a classy way
	resp, err := http.Get(endpoint + "run?uuid=" + r.uuid)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if _, err := io.Copy(os.Stdout, resp.Body); err != nil {
		if err != io.EOF {
			return err
		}
	}

	return nil
}
