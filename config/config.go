package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type Config struct {
	Listen       string
	BaseDir      string
	TempDir      string
	PlayListHost []string
	TempPrefix   string
}

var _globalConfig Config

func GlobalConfig() *Config {
	return &_globalConfig
}

func DoLoadConfigFile(filename string, logger *log.Logger) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	logger.Printf("Load config from %s\n", filename)
	return json.Unmarshal(data, &_globalConfig)
}

func LoadConfigFile(configFileName string, logger *log.Logger) error {
	var filenames []string
	if configFileName != "" {
		filenames = append(filenames, configFileName)
	} else {
		filenames = []string{"config.json", "/Users/wsx/momo-download-file-go"}
	}
	var err error
	for _, filename := range filenames {
		err = DoLoadConfigFile(filename, logger)
		if err == nil {
			return nil
		}
	}
	return err
}
