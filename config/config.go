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
	OpenAI struct {
		BaseURL string `yaml:"base_url"`
		Key     string `yaml:"key"`
	} `yaml:"openai"`

	Wechat struct {
		AppID          string `yaml:"app_id"`
		AppSecret      string `yaml:"app_secret"`
		Token          string `yaml:"token"`
		EncodingAESKey string `yaml:"encoding_aes_key"`
	} `yaml:"wechat"`

	App struct {
		BaseURL string `yaml:"base_url"`
	}

	Storage string `yaml:"storage"`
	Redis   struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Password string `yaml:"password"`
	}
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
