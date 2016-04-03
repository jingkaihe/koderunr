package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
)

type messages chan string

// Client is a registed node in Broker
type Client struct {
	input  messages // input to runner
	output messages // output from runner
}

// NewClient creates new client
func NewClient() *Client {
	return &Client{
		input:  make(messages),
		output: make(messages),
	}
}

func (cli *Client) Read(msg string) {
	cli.input <- msg
}

// Writing things out
func (cli *Client) Write(w http.ResponseWriter, isEvtSource bool) {
	f, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "The server does not support streaming!", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Accel-Buffering", "no")

	for msg := range cli.output {
		if isEvtSource == true {
			// To make event source comfort.
			// From http://www.html5rocks.com/en/tutorials/eventsource/basics/
			// If your message is longer, you can break it up by using multiple "data:" lines.
			// Two or more consecutive lines beginning with "data:" will be treated as a single
			// piece of data, meaning only one message event will be fired. Each line should
			// end in a single "\n" (except for the last, which should end with two). The result
			// passed to your message handler is a single string concatenated by newline characters.
			lines := strings.Split(msg, "\n")
			for i, line := range lines {
				lines[i] = "data: " + line
			}
			msg = strings.Join(lines, "\n")
			fmt.Fprintf(w, "%s\n\n", msg)
		} else {
			fmt.Fprint(w, msg)
		}

		f.Flush()
	}
}

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
