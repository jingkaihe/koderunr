package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/garyburd/redigo/redis"
)

type messages chan string

// Client is a proxy struct registered for running
type Client struct {
	runner       *Runner
	stdoutWriter io.Writer
	stdoutReader io.Reader
	stdinWriter  io.Writer
	stdinReader  io.Reader
	conn         redis.Conn // redis connection
	uuid         string
}

// NewClient creates new client
func NewClient(r *Runner, conn redis.Conn, uuid string) *Client {
	stdoutReader, stdoutWriter := io.Pipe()
	stdinReader, stdinWriter := io.Pipe()
	return &Client{
		stdoutReader: stdoutReader,
		stdoutWriter: stdoutWriter,
		stdinReader:  stdinReader,
		stdinWriter:  stdinWriter,
		runner:       r,
		conn:         conn,
		uuid:         uuid,
	}
}

// Run kicks start the container
func (cli *Client) Run() {
	cli.runner.Run(cli.stdinReader, cli.stdoutWriter, cli.conn, cli.uuid)
}

func (cli *Client) Read() {
	psc := redis.PubSubConn{Conn: cli.conn}
	psc.Subscribe(cli.uuid + "#stdin")

	defer func() {
		psc.Unsubscribe(cli.uuid + "#stdin")
		psc.Close()
		cli.conn.Close()
	}()

StdinSubscriptionLoop:
	for {
		switch n := psc.Receive().(type) {
		case redis.Message:
			stdinData := strconv.QuoteToASCII(string(n.Data))
			cli.runner.logger.Infof("Message: %s %s", n.Channel, stdinData)
			cli.stdinWriter.Write(n.Data)
		case error:
			break StdinSubscriptionLoop
		}
	}
	cli.runner.logger.Info("Stdin subscription closed")
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

	buffer := make([]byte, 512)

	for {
		n, err := cli.stdoutReader.Read(buffer)
		if err != nil {
			break
		}

		msg := string(buffer[0:n])

		if isEvtSource == true {
			msg = cli.sseFormat(msg)
		}

		if _, err := fmt.Fprint(w, msg); err != nil {
			cli.logger().Errorf("Response is not writable for %s\n", msg)
			return
		}
		f.Flush()

		// Clear the buffer
		for i := 0; i < n; i++ {
			buffer[i] = 0
		}
	}

	if isEvtSource == true {
		msg := cli.sseFormat("\n")
		if _, err := fmt.Fprint(w, msg); err != nil {
			cli.logger().Errorf("Response is not writable for %s\n", msg)
			return
		}
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
	// if msg does not contain linebreak, we simply wrote that out
	if !strings.Contains(msg, "\n") {
		return fmt.Sprintf("data: %s\n", msg)
	}

	lines := strings.Split(msg, "\n")
	var b bytes.Buffer

	for i, line := range lines {
		if i == len(lines)-1 && line == "" {
			continue
		}
		fmt.Fprintf(&b, "data: %s\n\n", line)
	}

	return b.String()
}

func (cli *Client) logger() *logrus.Logger {
	return cli.runner.logger
}
