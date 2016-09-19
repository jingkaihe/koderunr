package main

import (
	"encoding/json"
	"io/ioutil"
)

// Config is the configuration for up and running
type Config struct {
	LanguagesFile     string `json:"languages_file"`
	Static            bool   `json:"static"`
	RunnerThrottleNum int    `json:"runner_throttle_num"`
	Port              int    `json:"port"`
	Languages         *Languages
}

// ReadConfigFile load config file from JSON into Config struct
func ReadConfigFile(path string) (*Config, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = json.Unmarshal(file, &cfg)

	if err == nil {
		cfg.Languages, err = ReadLanuagesFile(cfg.LanguagesFile)
	}
	return &cfg, err
}
