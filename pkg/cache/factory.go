package cache

import (
	"time"
)

type CacheConfig struct {
	CacheType   string        `json:"cacheType"`   // 缓存类型: memory / redis
	RedisHost   string        `json:"redisHost"`   // Redis 地址
	RedisPwd    string        `json:"redisPwd"`    // Redis 密码
	RedisDB     int           `json:"redisDB"`     // Redis 数据库索引
	ExpireTime  time.Duration `json:"expireTime"`  // 默认过期时间
	Cleanup     time.Duration `json:"cleanup"`     // 清理间隔 (仅 memory)
	PersistFile string        `json:"persistFile"` // 持久化文件路径 (仅 memory)
}

// Cache 定义缓存接口
type Cache interface {
	Get(key string) (interface{}, bool)                         // 获取缓存
	Set(key string, value interface{}, ttl time.Duration) error // 设置缓存
	Delete(key string) error                                    // 删除缓存
	Keys(pattern string) []string                               // 获取符合 pattern 的 keys
	TTL(key string) time.Duration                               // 获取剩余 TTL
}

var CacheManager Cache

// NewCache 创建缓存实例
func NewCache(config *CacheConfig) {
	if len(config.CacheType) == 0 {
		config.CacheType = "memory"
	}
	if config.ExpireTime == 0 {
		config.ExpireTime = 30
	}
	if config.Cleanup == 0 {
		config.ExpireTime = 60
	}
	switch config.CacheType {
	case "redis":
		CacheManager = NewRedisCache(config.RedisHost, config.RedisPwd, config.RedisDB)
	default:
		CacheManager = NewMemoryCache(config.ExpireTime, config.Cleanup, config.PersistFile)
	}
}
