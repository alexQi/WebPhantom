package cache

import (
	"context"
	"github.com/go-redis/redis"
	"time"
)

// RedisCache 实现 Cache 接口
type RedisCache struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisCache 初始化 Redis 连接
func NewRedisCache(addr, password string, db int) *RedisCache {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &RedisCache{client: rdb, ctx: context.Background()}
}

// Get 获取缓存
func (r *RedisCache) Get(key string) (interface{}, bool) {
	val, err := r.client.Get(key).Result()
	if err == redis.Nil {
		return nil, false
	}
	return val, err == nil
}

// Set 设置缓存
func (r *RedisCache) Set(key string, value interface{}, ttl time.Duration) error {
	return r.client.Set(key, value, ttl).Err()
}

// Delete 删除缓存
func (r *RedisCache) Delete(key string) error {
	return r.client.Del(key).Err()
}

// Keys 获取符合 pattern 的 keys
func (r *RedisCache) Keys(pattern string) []string {
	keys, _ := r.client.Keys(pattern).Result()
	return keys
}

// TTL 获取剩余 TTL
func (r *RedisCache) TTL(key string) time.Duration {
	ttl, err := r.client.TTL(key).Result()
	if err != nil {
		return -1
	}
	return ttl
}
