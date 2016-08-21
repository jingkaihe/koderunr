package commands

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/jaxi/koderunr/cli/client"
)

// Langs is the command struct that shows supported languages
type Langs struct {
}

// Help command of the run
func (l Langs) Help() string {
	helpText := `
Usage: kode languages

  Shows the languages that are supported

Examples:

  $ kode languages

  Supported languages:
    Golang
    Ruby
      * 2.3.0
      * 1.9.3
  ...
`
	return strings.TrimSpace(helpText)
}

// ShortDescription for the Run command
func (l Langs) ShortDescription() string {
	return "kode languages - Shows available running languages"
}

// Exec is the command that will show the version of languages
func (l Langs) Exec(args []string) int {
	// Parse the endpoint from the arguments passed in
	flagargs := args

	langsFlagSet := flag.NewFlagSet("langs", flag.ExitOnError)
	endpointFlag := langsFlagSet.String("endpoint", Endpoint+"/api", "Endpoint of the API")

	langsFlagSet.Parse(flagargs)

	// TODO: Build the URI in a classy way
	httpClient := client.NewHTTPClient(60, 60)
	resp, err := httpClient.Get(*endpointFlag + "/langs/")

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	defer resp.Body.Close()

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	fmt.Fprintf(os.Stdout, string(body))

	return 0
}
