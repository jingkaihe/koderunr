package client

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
	ext     string
	source  string
	version string
	uuid    string
}

// NewRunner create a new runner
func NewRunner(version, fName string) (r *Runner, err error) {
	ext := path.Ext(fName)

	ctx, err := ioutil.ReadFile(fName)

	r = &Runner{
		ext:     ext,
		source:  string(ctx),
		version: version,
	}
	return
}

// FetchUUID fetch the UUID from the API endpoint
func (r *Runner) FetchUUID(endpoint string) error {
	params := url.Values{"ext": {r.ext}, "source": {string(r.source)}}
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
