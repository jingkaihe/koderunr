package commands

import "fmt"

// Command is the interface for a single command
type Command interface {
	Help() string
	Exec(args []string) int
	ShortDescription() string
}

// CLI is a list of command
type CLI struct {
	Cmds    map[string]Command
	App     string
	Version string
	Intro   string
}

// NewCLI Create the new CLI client
func NewCLI(app, version, intro string) *CLI {
	return &CLI{
		App:     app,
		Version: version,
		Intro:   intro,
	}
}

// Brief give the list of short descriptions of the commands
// TODO: Need some align - Looks a bit untidy.
func (cli *CLI) Brief() {
	fmt.Println()
	fmt.Println(cli.Intro)
	fmt.Printf("\nCommands:\n\n")

	fmt.Printf("kode version - The version of the kode client\n\n")
	fmt.Printf("kode help [cmd] - The usage of the given command\n\n")

	for _, cmd := range cli.Cmds {
		fmt.Printf("%s\n\n", cmd.ShortDescription())
	}

}

// Help Give the usage about how to exec a certain command
func (cli *CLI) Help(cmdName string) {
	cmd := cli.Cmds[cmdName]
	if cli == nil {
		fmt.Printf("%s command does not exist.\n", cmdName)
		cli.Brief()
		return
	}

	fmt.Print(cmd.Help())
}

// GetVersion prints out the version of the CLI
func (cli *CLI) GetVersion() {
	fmt.Printf("%s - %s\n", cli.App, cli.Version)
}

// Exec execute the command
func (cli *CLI) Exec(args []string) {
	if len(args) == 0 {
		cli.Brief()
		return
	}

	cmdName := args[0]
	argsLen := len(args) - 1

	switch cmdName {
	case "":
		cli.Brief()
	case "version":
		cli.GetVersion()
	case "help":
		if argsLen > 0 {
			cli.Help(args[1])
		} else {
			cli.Brief()
		}
	default:
		if argsLen > 0 {
			cli.RunCmd(cmdName, args[1:])
		} else {
			cli.Brief()
		}
	}
}

// RunCmd execute the given command
func (cli *CLI) RunCmd(cmdName string, args []string) {
	cmd := cli.Cmds[cmdName]
	if cmd == nil {
		cli.Brief()
		return
	}

	cmd.Exec(args)
}
