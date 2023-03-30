package db

import (
	"time"

	"github.com/silenceper/wechat/v2/cache"
)

var _ Cache = (*memCache)(nil)

type memCache struct {
	cache.Memory
}

func NewMemCache() Cache {
	return &memCache{
		Memory: *cache.NewMemory(),
	}
}

func (m *memCache) Exists(key string) bool {
	return m.Memory.IsExist(key)
}

func (m *memCache) Set(key string, value interface{}, expires time.Duration) error {
	return m.Memory.Set(key, value, expires)
}

func (m *memCache) Get(key string) (interface{}, error) {
	return m.Memory.Get(key), nil
}
