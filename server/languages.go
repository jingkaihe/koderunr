package main

import (
	"encoding/json"
	"io/ioutil"
)

// Language gives the specification of a programming language
type Language struct {
	Versions  []string
	CPUQuota  int64
	Memory    int64
	PidsLimit int64
}

// Languages tells languages specifications
type Languages map[string]Language

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

// GetCPUQuota returns CPUQuota of the given language
func (l *Language) GetCPUQuota() int64 {
	if l.CPUQuota != 0 {
		return l.CPUQuota
	}

	return 20000
}

// GetMemory returns memory of the given language
func (l *Language) GetMemory() int64 {
	if l.Memory != 0 {
		return l.Memory
	}

	return 80 * 1024 * 1024
}

// GetPidsLimit returns processes limitation of the given language
func (l *Language) GetPidsLimit() int64 {
	if l.PidsLimit != 0 {
		return l.PidsLimit
	}

	return 100
}
