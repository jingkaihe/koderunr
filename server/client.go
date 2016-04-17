package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/garyburd/redigo/redis"
)

type messages chan string

// Client is a proxy struct registered for running
type Client struct {
	runner *Runner
	output messages   // output from runner
	conn   redis.Conn // redis connection
	uuid   string
}

// NewClient creates new client
func NewClient(r *Runner, conn redis.Conn, uuid string) *Client {
	return &Client{
		output: make(messages),
		runner: r,
		conn:   conn,
		uuid:   uuid,
	}
}

// Run kicks start the container
func (cli *Client) Run() {
	cli.runner.Run(cli.output, cli.conn, cli.uuid)
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
			msg = cli.sseFormat(msg)
		}

		fmt.Fprint(w, msg)
		f.Flush()
	}
}

// To make event source comfort.
// From http://www.html5rocks.com/en/tutorials/eventsource/basics/
// If your message is longer, you can break it up by using multiple "data:" lines.
// Two or more consecutive lines beginning with "data:" will be treated as a single
// piece of data, meaning only one message event will be fired. Each line should
// end in a single "\n" (except for the last, which should end with two). The result
// passed to your message handler is a single string concatenated by newline characters.
func (cli *Client) sseFormat(msg string) string {
	lines := strings.Split(msg, "\n")
	for i, line := range lines {
		lines[i] = "data: " + line
	}
	return strings.Join(lines, "\n") + "\n\n"
}
