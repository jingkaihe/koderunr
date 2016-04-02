# Koderunr

Koderunner (read code runner) is a tool that can let you run code without installing the language on your machine. Currently heavily under development.

## Install

Currently WIP, need to figure that out.

One of the possible approaches is cross-compiling the source and distributing the binary versions, so you don't need any non-sense dependencies on your machine.


## Web Interface

Need to find a server. Or argubaly host it on GitHub page?

## Examples (proposal)

Suppose you have a golang file called `main.go`, which has the source

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	c := time.Tick(1 * time.Second)
	i := 0
	for now := range c {
		fmt.Printf("%v %d\n", now, i)
		if i == 3 {
			break
		}
		i++
	}
}
```

You can execute the source code by running the command - and the results will be outputted as below with time intervals.

```bash
$ kode run main.go # main.go is the file you have locally
2016-03-28 09:24:14.590078119 +0100 BST 0
2016-03-28 09:24:15.59435734 +0100 BST 1
2016-03-28 09:24:16.590340842 +0100 BST 2
2016-03-28 09:24:17.590881844 +0100 BST 3
```

Also you can specified version of the language you want by

```bash
$ kode run foo.rb -v 2.2.0 # Running foo.rb using ruby 2.2.0
```

## TODO

- [x] ~~Support more languages (e.g. C, python, ruby, Erlang), at the moment only Go is supported.~~ Now supporting Go, C, ruby, python.
- [ ] Running the Docker containers in a proper Docker client rather than using system calls, so we can create-attach-run-kill containers automatically. (Currently tidy up the dead containers manually)
- [x] Serious cli support
- [x] cli configuration
- [x] Binaries distribution
- [x] Web interface that sucks less...
- [ ] Server and Docker containers hosting
- [x] Support programming language versioning
