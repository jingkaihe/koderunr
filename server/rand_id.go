package main

import (
	"math/rand"
	"time"
)

var randRunes = []rune("abcdefghijklmnopqrstuvwxyz01234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ")
var randLen = len(randRunes)

// NewRandID create a string with a fix length equal to n
func NewRandID(n int) string {
	rand.Seed(time.Now().UTC().UnixNano())

	s := make([]rune, n)
	for i := 0; i < n; i++ {
		s[i] = randRunes[(rand.Int()+randLen)%randLen]
	}

	return string(s)
}
