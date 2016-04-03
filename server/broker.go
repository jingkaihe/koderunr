package main

import (
	"fmt"
	"os"
)

// Broker is the centralised component for messages passing around
type Broker struct {
	clients map[string]*Client
	joining chan string
	leaving chan string
}

// NewBroker creates a new broker
func NewBroker() *Broker {
	return &Broker{
		clients: make(map[string]*Client),
		joining: make(chan string),
		leaving: make(chan string),
	}
}

// Start runs the broker
func (b *Broker) Start() {
	for {
		select {
		case uuid := <-b.joining:
			b.clients[uuid] = &Client{}
			fmt.Fprintf(os.Stdin, "%s has joined running group", uuid)
		case uuid := <-b.leaving:
			delete(b.clients, uuid)
			fmt.Fprintf(os.Stdin, "%s has left running group", uuid)
		}
	}
}
