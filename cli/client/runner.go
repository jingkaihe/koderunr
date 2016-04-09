package client

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
)

// TestEndPoint is the endpoint for testing
const TestEndPoint = "http://127.0.0.1:8080/"

// Runner contains the code to be run
type Runner struct {
	lang    string
	source  string
	version string
	uuid    string
}

var extToLang = map[string]string{
	".rb":  "ruby",
	".py":  "python",
	".ex":  "elixir",
	".iex": "elixir",
	".c":   "c",
	".cc":  "c",
	".go":  "go",
}

// NewRunner create a new runner
func NewRunner(version, fName string) (r *Runner, err error) {
	ext := path.Ext(fName)
	lang := extToLang[ext]

	if lang == "" {
		err = fmt.Errorf("%s extension is not supported", ext)
		return
	}

	ctx, err := ioutil.ReadFile(fName)
	r = &Runner{
		lang:    lang,
		source:  string(ctx),
		version: version,
	}
	return
}

// FetchUUID fetch the UUID from the API endpoint
func (r *Runner) FetchUUID(endpoint string) error {
	params := url.Values{"lang": {r.lang}, "source": {string(r.source)}}
	if r.version != "" {
		params["version"] = []string{r.version}
	}
	resp, err := http.PostForm(endpoint+"register/", params)
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

// Run execute the runner
func (r *Runner) Run(endpoint string) error {
	go r.fetchStdin(endpoint)

	// TODO: Build the URI in a classy way
	resp, err := http.Get(endpoint + "run/?uuid=" + r.uuid)
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

func (r *Runner) fetchStdin(endpoint string) error {
	reader := bufio.NewReader(os.Stdin)

	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		params := url.Values{"uuid": {r.uuid}, "input": {text}}

		resp, err := http.PostForm(endpoint+"stdin/", params)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
	}

	return nil
}
