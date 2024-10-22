package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type config struct {
	PollRate    uint `yaml:"PollRate"`    // Tezos delegation polling frequency in seconds
	PollMaxSize uint `yaml:"PollMaxSize"` // Max number of delegations retrieved when polling
}

func (c config) String() string {
	return fmt.Sprintf("configuration: \n - PollRate: every %d secondes,\n - PollMaxSize: %d", c.PollRate, c.PollMaxSize)
}

func getConfig() (config, error) {
	var conf config
	fileByte, err := os.ReadFile("config.yaml")
	if err != nil {
		return config{}, fmt.Errorf("could not read config file: %w", err)
	}

	if err := yaml.Unmarshal(fileByte, &conf); err != nil {
		return config{}, fmt.Errorf("could not unmarshall config: %w", err)
	}
	InfoLog.Print(conf.String())
	return conf, nil
}
