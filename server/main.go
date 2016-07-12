package main

import "flag"

var servingStatic bool
var runnerThrottleNum int

func init() {
	flag.BoolVar(&servingStatic, "static", false, "if using Go server hosting static files")
	flag.IntVar(&runnerThrottleNum, "runner_throttle", 4, "Limit the max throttle for the runners")
	flag.Parse()

	var err error
	DockerClient, err = NewDockerClient()
	if err != nil {
		panic(err)
	}
}

func main() {
	Runnerthrottle = make(chan struct{}, runnerThrottleNum)

	s := NewServer(16, servingStatic)
	s.Serve("/api/", 8080)
}
