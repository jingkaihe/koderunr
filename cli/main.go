package main

import (
	"os"

	"github.com/jaxi/koderunr/cli/commands"
)

// Version is the version of the command line tool
// Passed in by main.Version
var Version string

func main() {
	args := os.Args[1:]

	cli := commands.NewCLI(
		"kode",
		Version,
		"Kode - Running code without install the programming language!",
	)

	cli.Cmds = map[string]commands.Command{
		"run":       commands.Run{},
		"share":     commands.Share{},
		"languages": commands.Langs{},
	}

	cli.Exec(args)
}
