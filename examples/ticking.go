package main

import (
	"fmt"
	"time"
)

func main() {
	c := time.Tick(1 * time.Second)
	i := 0
	for range c {
		fmt.Printf("Bazinga %d!\n", i)
		if i == 3 {
			break
		}
		i++
	}
}
