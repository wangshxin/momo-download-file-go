package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type Config struct {
	Listen       string
	// 直播文件存储位置
	BaseDir      string
	// 临时生成的M3U8文件目录
	TempDir      string
	// 用于播放的服务器
	PlayListHost []string
	// 下载临时生成的M3U8文件的前缀，用于NGINX解析寻找临时生成的M3U8
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
