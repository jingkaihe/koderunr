package main

import "testing"

func TestNewRandID(t *testing.T) {
	id := NewRandID(10)

	if len(id) != 10 {
		t.Fatalf("The length of %s != 10", id)
	}
}
