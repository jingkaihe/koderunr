package main

import "flag"

var appConfig *Config

func init() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.json", "Configuration for the Koderunr")
	flag.Parse()

	var err error
	appConfig, err = ReadConfigFile(configPath)
	DockerClient, err = NewDockerClient()
	if err != nil {
		panic(err)
	}
}

func main() {
	Runnerthrottle = make(chan struct{}, appConfig.RunnerThrottleNum)

	s := NewServer(16, appConfig.Static)
	s.Serve("/api/", 8080)
}
