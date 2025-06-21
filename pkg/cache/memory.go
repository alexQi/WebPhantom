package cache

import (
	"encoding/gob"
	"github.com/patrickmn/go-cache"
	"noctua/pkg/logger"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

// MemoryCache 实现 `Cache`，并支持持久化
type MemoryCache struct {
	store       *cache.Cache
	persistFile string
	stopChan    chan struct{}
}

// NewMemoryCache 创建缓存，并自动加载
func NewMemoryCache(defaultExpiration, cleanupInterval time.Duration, persistFile string) *MemoryCache {
	m := &MemoryCache{
		store:       cache.New(defaultExpiration*time.Minute, cleanupInterval*time.Minute),
		persistFile: persistFile,
		stopChan:    make(chan struct{}),
	}

	// 加载缓存
	m.LoadFromFile()
	m.startAutoSave()
	return m
}

// Get 获取缓存
func (m *MemoryCache) Get(key string) (interface{}, bool) {
	return m.store.Get(key)
}

// Set 设置缓存
func (m *MemoryCache) Set(key string, value interface{}, ttl time.Duration) error {
	m.store.Set(key, value, ttl)
	return nil
}

// Delete 删除 Key
func (m *MemoryCache) Delete(key string) error {
	m.store.Delete(key)
	return nil
}

// Keys 获取符合 pattern 的所有 key
func (m *MemoryCache) Keys(pattern string) []string {
	var matchedKeys []string
	re := regexp.MustCompile(pattern)

	for key := range m.store.Items() {
		if re.MatchString(key) {
			matchedKeys = append(matchedKeys, key)
		}
	}
	return matchedKeys
}

// TTL 获取 Key 的剩余 TTL
func (m *MemoryCache) TTL(key string) time.Duration {
	item, found := m.store.Items()[key]
	if !found {
		return -1
	}

	// 计算剩余 TTL
	if item.Expiration > 0 {
		ttl := time.Until(time.Unix(0, item.Expiration).In(time.Local))
		if ttl < 0 {
			return -1
		}
		return ttl
	}
	return cache.NoExpiration
}

// SaveToFile **Gob 持久化**
func (m *MemoryCache) SaveToFile() {
	// 确保目录存在
	dir := filepath.Dir(m.persistFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		logger.Log.Errorf("Failed to create cache directory %s: %v", dir, err)
		return
	}

	file, err := os.Create(m.persistFile)
	if err != nil {
		logger.Log.Errorf("Failed to persist cache to %s: %v", m.persistFile, err)
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			logger.Log.Errorf("Failed to close file %s: %v", file.Name(), err)
		}
	}(file)

	encoder := gob.NewEncoder(file)
	items := m.store.Items()
	if err := encoder.Encode(items); err != nil {
		logger.Log.Errorf("Failed to encode cache to %s: %v", m.persistFile, err)
		return
	}
}

// LoadFromFile **Gob 读取**
func (m *MemoryCache) LoadFromFile() {
	file, err := os.Open(m.persistFile)
	if os.IsNotExist(err) {
		logger.Log.Warnf("Cache file %s does not exist, possibly first run", m.persistFile)
		return
	} else if err != nil {
		logger.Log.Errorf("Failed to read cache file %s: %v", m.persistFile, err)
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			logger.Log.Errorf("Failed to close file %s: %v", file.Name(), err)
		}
	}(file)

	var items map[string]cache.Item
	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(&items); err != nil {
		logger.Log.Errorf("Failed to decode cache from %s: %v", m.persistFile, err)
		return
	}

	for k, v := range items {
		m.store.Set(k, v.Object, time.Duration(v.Expiration))
	}
	logger.Log.Infof("Cache successfully loaded from %s", m.persistFile)
}

// startAutoSave 定时保存
func (m *MemoryCache) startAutoSave() {
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				m.SaveToFile()
			case <-m.stopChan:
				ticker.Stop()
				m.SaveToFile() // 停止前最后保存一次
				logger.Log.Infof("Auto-save stopped, cache saved to %s", m.persistFile)
				return
			}
		}
	}()
}

// StopAutoSave 停止自动保存
func (m *MemoryCache) StopAutoSave() {
	close(m.stopChan)
}
