package main

import (
	"os"

	"github.com/jaxi/koderunr/cli/commands"
)

func main() {
	args := os.Args[1:]

	cli := commands.NewCLI(
		"kode",
		"0.0.1 Beta",
		"Kode - Running code without install the programming language!",
	)

	cli.Cmds = map[string]commands.Command{
		"run": commands.Run{},
	}

	cli.Exec(args)
}
