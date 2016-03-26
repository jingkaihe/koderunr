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
		if i == 10 {
			break
		}
		i++
	}
}
