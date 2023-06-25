package config

import (
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DiscordToken        string `yaml:"discord_token"`
	LadderMode          string `yaml:"ladder_mode"`
	OpenAIKey           string `yaml:"openai_key"`
	MongoDBName         string `yaml:"mongo_db"`
	MongoAdmin          string `yaml:"mongo_admin"`
	MongoPass           string `yaml:"mongo_pass"`
	MongoURI            string `yaml:"mongo_uri"`
	MongoCollectionName string `yaml:"mongo_collection_name"`
}

// function that reads a json file and returns a Config struct
func ReadConfig(path string) (*Config, error) {
	conf := &Config{}
	jsonFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	bytes, _ := io.ReadAll(jsonFile)
	err = yaml.Unmarshal(bytes, conf)
	if err != nil {
		return nil, err
	}

	return conf, nil
}
