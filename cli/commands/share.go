package commands

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// Share is a command that will display the code in the CLI
type Share struct{}

// Help give a specific instructions about how to use the share command
func (s Share) Help() string {
	text := `
Usage: kode share [filename]

  Share the code and create a URL (so it can be shown in the browser)

filename:

	The file that contains the source code you want to share

Examples:

  $ kode share main.go
  http://koderunr.tech/#zW0CX1qn02
`
	return strings.TrimSpace(text)
}

// ShortDescription give a brief introduction about what share command does.
func (s Share) ShortDescription() string {
	return "kode share [filename] - Share the code by creating a URL"
}

// Exec fetch the share id and compose the uri
func (s Share) Exec(args []string) int {
	runner, err := createRunnerFromArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	url, err := runner.Share()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to share the code - %v\n", err)
		return 1
	}

	if isOSDarwin() {
		exec.Command("open", url).Run()
	} else {
		fmt.Println(url)
	}

	return 0
}

func isOSDarwin() bool {
	return runtime.GOOS == "darwin"
}
