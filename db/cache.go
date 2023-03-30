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
		if config.GetConfig().Redis.Host != "" {
			c = NewRedisCache()
		} else {
			c = NewMemCache()
		}
	})
	return c
}
