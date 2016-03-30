package main

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
)

const endpoint = "http://127.0.0.1:8080/"

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
