package main

import (
	"flag"

	"github.com/k8scat/wechat-openai/config"
	"github.com/k8scat/wechat-openai/log"
	"github.com/k8scat/wechat-openai/router"
)

func init() {
	flag.StringVar(&config.CfgFile, "config", "config.yml", "config file")
	flag.Parse()
}

func main() {
	defer log.Sync()

	if err := router.Run(); err != nil {
		panic(err)
	}
}
