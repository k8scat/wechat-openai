package db

import (
	"sync"

	"github.com/k8scat/wechat-openai/config"
)

var (
	c         Cache
	cacheInit sync.Once
)

func GetCache() Cache {
	cacheInit.Do(func() {
		switch config.GetConfig().Storage {
		case "redis":
			c = NewRedisCache()
		case "memory":
			c = NewMemCache()
		default:
			panic("unknown storage")
		}
	})
	return c
}
