package commands

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/jaxi/koderunr/cli/client"
)

// Run is the command struct for running code
type Run struct {
}

// Help command of the run
func (r Run) Help() string {
	helpText := `
Usage: kode run [options] [filename]

  Auto detect the programming language of the file and run it remoted.
  The result will be displayed onto the terminal asynchronously.

Options:

  -v=<version> Version of the programming language you want to use

  -endpoint=<url> The endpoint that you want the code to be run on

Examples:

  $ kode run main.go
  $ kode run -v=2.3.0 foo.rb
`
	return strings.TrimSpace(helpText)
}

// ShortDescription for the Run command
func (r Run) ShortDescription() string {
	return "Run the code remotely on runner and returns the result asynchronously"
}

// Exec is the command that will execute the Run command
func (r Run) Exec(args []string) int {
	var endpoint string
	var langVersion string

	flag.StringVar(&endpoint, "endpoint", "http://127.0.0.1:8080/", "Endpoint of the API")
	flag.StringVar(&langVersion, "version", "", "Version of the language")
	flag.Parse()

	runner, err := client.NewRunner(langVersion, args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	err = runner.FetchUUID(endpoint)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to fetch UUID - %v\n", err)
		return 1
	}

	err = runner.Run(endpoint)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to execute the code - %v\n", err)
		return 1
	}

	return 0
}
