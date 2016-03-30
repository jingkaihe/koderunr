package main

import (
	"fmt"
	"os"
)

// examples:
//   $ kode run main.go
//   $ kode run hello.rb
func main() {
	args := os.Args[1:]
	op := args[0]

	if op == "run" {
		fName := args[1]

		runner, err := newRunner(fName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		err = runner.fetchUUID()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		runner.run()
	}
}
