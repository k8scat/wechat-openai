package config

import (
	"os"
	"sync"

	"gopkg.in/yaml.v2"
)

var (
	CfgFile string

	initCfg sync.Once
	cfg     *Config
)

type Config struct {
	Openai struct {
		Key string `yaml:"key"`
	} `yaml:"openai"`

	Wechat struct {
		AppID          string `yaml:"app_id"`
		AppSecret      string `yaml:"app_secret"`
		Token          string `yaml:"token"`
		EncodingAESKey string `yaml:"encoding_aes_key"`
	} `yaml:"wechat"`
}

func GetConfig() *Config {
	initCfg.Do(func() {
		b, err := os.ReadFile(CfgFile)
		if err != nil {
			panic(err)
		}
		err = yaml.Unmarshal(b, &cfg)
		if err != nil {
			panic(err)
		}
	})
	return cfg

}
