package service

import (
	"sync"

	"github.com/dgraph-io/ristretto/v2"
)

var cache *ristretto.Cache[string, any]
var once sync.Once

type cacheService struct {
}

func CacheService() *cacheService {
	once.Do(func() {
		if c, err := ristretto.NewCache(&ristretto.Config[string, any]{
			NumCounters: 1e7,     // number of keys to track frequency of (10M).
			MaxCost:     1 << 30, // maximum cost of cache (1GB).
			BufferItems: 64,      // number of keys per Get buffer.
		}); err == nil {
			cache = c
		}

	})
	return &cacheService{}
}
func (c *cacheService) CacheInstance() *ristretto.Cache[string, any] {
	return cache
}
