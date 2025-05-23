package utils

import (
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"k8s.io/klog/v2"
)

func GetOrSetCache[T any](cache *ristretto.Cache[string, any], cacheKey string, ttl time.Duration, queryFunc func() (T, error)) (T, error) {
	var zero T

	// 如果未设置 TTL 参数，说明不需要缓存，则直接执行查询方法
	if ttl <= 0 {
		return queryFunc()
	}
	// 检查缓存是否命中
	if v, found := cache.Get(cacheKey); found {
		klog.V(8).Infof("k8m cache hit cacheKey= %s", cacheKey)
		return v.(T), nil
	}

	// 缓存未命中，执行查询方法
	result, err := queryFunc()
	if err != nil {
		return zero, err
	}

	// 设置缓存并返回结果
	cache.SetWithTTL(cacheKey, result, 100, ttl)
	cache.Wait()

	return result, nil
}

// ClearCacheByKey 清空指定key的缓存
func ClearCacheByKey(cache *ristretto.Cache[string, any], cacheKey string) {
	if cache == nil || cacheKey == "" {
		return
	}
	cache.Del(cacheKey)
	cache.Wait()
	klog.V(5).Infof("cache cleared for key: %s", cacheKey)
}
