package cache

import (
	"context"
	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/store"
	"github.com/eko/gocache/store/go_cache/v4"
	gocache "github.com/patrickmn/go-cache"
	"strings"
	"time"
)

type Manager[T any] struct {
	cache *cache.Cache[T]
}

var (
	imageCacheManager *Manager[string]
)

func init() {
	client := gocache.New(5*time.Minute, 5*time.Minute)
	imageCacheManager = &Manager[string]{
		cache: cache.New[string](go_cache.NewGoCache(client)),
	}
}

func ImageCacheManager() *Manager[string] {
	return imageCacheManager
}

func (m *Manager[T]) SetWithExpiration(key string, value T, expir time.Duration) error {
	timeout, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	return m.cache.Set(timeout, key, value, store.WithExpiration(expir))
}

func (m *Manager[T]) GetValue(key string) (value T, err error) {
	timeout, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	const errorMessage = "value not found"
	value, err = m.cache.Get(timeout, key)
	if err != nil && strings.Contains(err.Error(), errorMessage) {
		err = nil
		return
	}
	return
}
