package main

import (
	"flag"

	"github.com/k8scat/wechat-openai/config"
	"github.com/k8scat/wechat-openai/db"
	"github.com/k8scat/wechat-openai/log"
	"github.com/k8scat/wechat-openai/router"
)

var port int

func init() {
	flag.StringVar(&config.CfgFile, "config", "config.yml", "config file")
	flag.IntVar(&port, "port", 8080, "port")
	flag.Parse()
}

func main() {
	defer log.Sync()
	defer db.GetRedisClient().Close()

	if err := router.Run(port); err != nil {
		panic(err)
	}
}
