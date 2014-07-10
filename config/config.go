package config

import (
	"encoding/json"
	"os"
)

type ServerDefinition struct {
	URL           string `json:"url"`
	DataDirectory string `json:"data_directory"`
	Hostname      string `json:"hostname"`
	Port          int    `json:"port"`
}

func ParseConfig(path string) ([]ServerDefinition, error) {
	reader, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	var config []ServerDefinition
	decoder := json.NewDecoder(reader)
	if err = decoder.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}
