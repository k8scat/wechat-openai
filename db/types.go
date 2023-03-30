package db

import "time"

type Cache interface {
	Set(key string, value interface{}, expires time.Duration) error
	Get(key string) (interface{}, error)
	Exists(key string) bool
}
