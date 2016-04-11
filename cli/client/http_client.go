package client

import (
	"net"
	"net/http"
	"time"
)

// NewHTTPClient creates a http client that takes dial and read timeout
// into account
func NewHTTPClient(openTime, readTime int) http.Client {

	transport := http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			return transportDial(network, addr, openTime, readTime)
		},
	}

	return http.Client{
		Transport: &transport,
	}
}

func transportDial(network, addr string, openTime, readTime int) (net.Conn, error) {
	openTimeout := time.Duration(openTime) * time.Second
	readTimeout := time.Duration(readTime) * time.Second

	conn, err := net.DialTimeout(network, addr, openTimeout)
	if err != nil {
		return conn, err
	}

	deadline := time.Now().Add(readTimeout)
	err = conn.SetDeadline(deadline)

	return conn, err
}
