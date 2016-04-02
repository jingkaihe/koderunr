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
Usage: kode run [filename] [options]

  Auto detect the programming language of the file and run it remoted.
  The result will be displayed onto the terminal asynchronously.

filename:

	The file that contains the source code you want to run

options:

  -version=<version> Version of the programming language you want to use

  -endpoint=<url> The endpoint that you want the code to be run on

Examples:

  $ kode run main.go
  $ kode run -version=2.3.0 foo.rb
`
	return strings.TrimSpace(helpText)
}

// ShortDescription for the Run command
func (r Run) ShortDescription() string {
	return "Run the code remotely on runner and returns the result asynchronously"
}

// Exec is the command that will execute the Run command
func (r Run) Exec(args []string) int {
	// Parse the version and endpoint from the arguments passed in
	flagargs := args[1:]

	fmt.Println(flagargs)

	runFlagSet := flag.NewFlagSet("run", flag.ExitOnError)
	endpointFlag := runFlagSet.String("endpoint", "http://koderunr.tech/api/", "Endpoint of the API")
	langVersionFlag := runFlagSet.String("version", "", "Version of the language")
	debugFlag := runFlagSet.Bool("debug", false, "Debug mode use local endpoint")

	runFlagSet.Parse(flagargs)

	var endpoint string

	if *debugFlag == true {
		endpoint = client.TestEndPoint
	} else {
		endpoint = *endpointFlag
	}

	// Started running the code
	runner, err := client.NewRunner(*langVersionFlag, args[0])
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
