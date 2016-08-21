package main

import (
	"encoding/json"
	"io/ioutil"
)

// Languages specify the versions of languages
type Languages map[string][]string

// ReadLanuagesFile load language configuration file from JSON into Config struct
func ReadLanuagesFile(path string) (*Languages, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var langs Languages
	err = json.Unmarshal(file, &langs)
	return &langs, err
}
