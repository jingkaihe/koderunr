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
	lang       string
	source     string
	version    string
	uuid       string
	endpoint   string
	httpClient http.Client
}

var extToLang = map[string]string{
	".rb":    "ruby",
	".py":    "python",
	".swift": "swift",
	".ex":    "elixir",
	".iex":   "elixir",
	".c":     "c",
	".cc":    "c",
	".go":    "go",
}

// NewRunner create a new runner
func NewRunner(version, fName, endpoint string) (r *Runner, err error) {
	ext := path.Ext(fName)
	lang := extToLang[ext]

	if lang == "" {
		err = fmt.Errorf("%s extension is not supported", ext)
		return
	}

	ctx, err := ioutil.ReadFile(fName)
	if err != nil {
		return
	}

	client := NewHTTPClient(60, 60)

	r = &Runner{
		lang:       lang,
		source:     string(ctx),
		version:    version,
		endpoint:   endpoint,
		httpClient: client,
	}

	return
}

// FetchUUID fetch the UUID from the API endpoint
func (r *Runner) FetchUUID() error {
	params := url.Values{"lang": {r.lang}, "source": {string(r.source)}}
	if r.version != "" {
		params["version"] = []string{r.version}
	}

	resp, err := r.httpClient.PostForm(r.endpoint+"api/register/", params)
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

// Share the code
func (r *Runner) Share() (string, error) {
	params := url.Values{"lang": {r.lang}, "source": {string(r.source)}}
	if r.version != "" {
		params["version"] = []string{r.version}
	}

	resp, err := r.httpClient.PostForm(r.endpoint+"api/save/", params)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	shareURL := fmt.Sprintf("%s#%s", r.endpoint, string(body))
	return shareURL, nil
}

// Run execute the runner
func (r *Runner) Run() error {
	go r.fetchStdin()

	// TODO: Build the URI in a classy way
	resp, err := r.httpClient.Get(r.endpoint + "api/run/?uuid=" + r.uuid)
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

func (r *Runner) fetchStdin() error {
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

		resp, err := r.httpClient.PostForm(r.endpoint+"api/stdin/", params)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
	}

	return nil
}
