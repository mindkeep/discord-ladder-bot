package config

import (
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Token      string `yaml:"token"`
	LadderMode string `yaml:"ladder_mode"`
}

// function that reads a json file and returns a Config struct
func ReadConfig(path string) (Config, error) {
	jsonFile, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer jsonFile.Close()

	bytes, _ := io.ReadAll(jsonFile)
	var conf Config
	err = yaml.Unmarshal(bytes, &conf)
	if err != nil {
		return Config{}, err
	}

	return conf, nil
}
