package db

import (
	"fmt"
	"sync"

	"github.com/redis/go-redis/v9"

	"github.com/k8scat/wechat-openai/config"
)

var (
	rdbInit sync.Once
	rdb     *redis.Client
)

func GetRedisClient() *redis.Client {
	rdbInit.Do(func() {
		cfg := config.GetConfig()
		rdb = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
			Password: cfg.Redis.Password,
			DB:       0,
		})
	})
	return rdb
}
