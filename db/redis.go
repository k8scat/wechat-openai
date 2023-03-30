package db

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/k8scat/wechat-openai/config"
	"github.com/k8scat/wechat-openai/log"
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

var _ Cache = (*redisCache)(nil)

type redisCache struct {
	ctx context.Context
	redis.Client
}

func NewRedisCache() Cache {
	return &redisCache{
		ctx:    context.Background(),
		Client: *GetRedisClient(),
	}
}

func (r *redisCache) Set(key string, value interface{}, expires time.Duration) error {
	return r.Client.Set(r.ctx, key, value, expires).Err()
}

func (r *redisCache) Get(key string) (interface{}, error) {
	cmd := r.Client.Get(r.ctx, key)
	if cmd.Err() != nil {
		return nil, cmd.Err()
	}
	return cmd.Val(), nil
}

func (r *redisCache) Exists(key string) bool {
	cmd := r.Client.Exists(r.ctx, key)
	if cmd.Err() != nil {
		log.Error("redis exists cmd run failed", zap.Error(cmd.Err()), zap.Stack("stack"))
		return false
	}
	return cmd.Val() > 0
}
