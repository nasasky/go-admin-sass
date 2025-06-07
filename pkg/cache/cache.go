package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheManager 缓存管理器
type CacheManager struct {
	redis     *redis.Client
	local     *LocalCache
	enabled   bool
	redisAddr string
}

// LocalCache 本地缓存
type LocalCache struct {
	data map[string]*CacheItem
	mu   sync.RWMutex
}

// CacheItem 缓存项
type CacheItem struct {
	Value     interface{}
	ExpiresAt time.Time
}

// NewCacheManager 创建缓存管理器
func NewCacheManager(redisClient *redis.Client) *CacheManager {
	cm := &CacheManager{
		redis: redisClient,
		local: &LocalCache{
			data: make(map[string]*CacheItem),
		},
		enabled: true,
	}

	// 启动本地缓存清理协程
	go cm.cleanupLocalCache()

	return cm
}

// Get 获取缓存值，优先本地缓存，然后Redis
func (cm *CacheManager) Get(ctx context.Context, key string, dest interface{}) error {
	if !cm.enabled {
		return fmt.Errorf("cache disabled")
	}

	// 1. 尝试本地缓存
	if value, found := cm.getFromLocal(key); found {
		return json.Unmarshal(value, dest)
	}

	// 2. 尝试Redis缓存
	if cm.redis != nil {
		data, err := cm.redis.Get(ctx, key).Bytes()
		if err == nil {
			// 同时存储到本地缓存（较短TTL）
			cm.setToLocal(key, data, 5*time.Minute)
			return json.Unmarshal(data, dest)
		}
	}

	return fmt.Errorf("cache miss")
}

// Set 设置缓存值，同时存储到本地和Redis
func (cm *CacheManager) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if !cm.enabled {
		return nil
	}

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	// 设置本地缓存（较短TTL避免内存过度使用）
	localTTL := ttl
	if localTTL > 10*time.Minute {
		localTTL = 10 * time.Minute
	}
	cm.setToLocal(key, data, localTTL)

	// 异步设置Redis缓存
	if cm.redis != nil {
		go func() {
			cm.redis.Set(context.Background(), key, data, ttl)
		}()
	}

	return nil
}

// Delete 删除缓存
func (cm *CacheManager) Delete(ctx context.Context, key string) error {
	// 删除本地缓存
	cm.deleteFromLocal(key)

	// 删除Redis缓存
	if cm.redis != nil {
		cm.redis.Del(ctx, key)
	}

	return nil
}

// getFromLocal 从本地缓存获取
func (cm *CacheManager) getFromLocal(key string) ([]byte, bool) {
	cm.local.mu.RLock()
	defer cm.local.mu.RUnlock()

	item, exists := cm.local.data[key]
	if !exists || time.Now().After(item.ExpiresAt) {
		return nil, false
	}

	if data, ok := item.Value.([]byte); ok {
		return data, true
	}

	return nil, false
}

// setToLocal 设置本地缓存
func (cm *CacheManager) setToLocal(key string, value []byte, ttl time.Duration) {
	cm.local.mu.Lock()
	defer cm.local.mu.Unlock()

	cm.local.data[key] = &CacheItem{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}
}

// deleteFromLocal 删除本地缓存
func (cm *CacheManager) deleteFromLocal(key string) {
	cm.local.mu.Lock()
	defer cm.local.mu.Unlock()

	delete(cm.local.data, key)
}

// cleanupLocalCache 清理过期的本地缓存
func (cm *CacheManager) cleanupLocalCache() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		cm.local.mu.Lock()
		now := time.Now()
		for key, item := range cm.local.data {
			if now.After(item.ExpiresAt) {
				delete(cm.local.data, key)
			}
		}
		cm.local.mu.Unlock()
	}
}

// GetStats 获取缓存统计信息
func (cm *CacheManager) GetStats() map[string]interface{} {
	cm.local.mu.RLock()
	localItemCount := len(cm.local.data)
	cm.local.mu.RUnlock()

	stats := map[string]interface{}{
		"enabled":         cm.enabled,
		"local_items":     localItemCount,
		"redis_connected": cm.redis != nil,
	}

	if cm.redis != nil {
		if info, err := cm.redis.Info(context.Background(), "stats").Result(); err == nil {
			stats["redis_info"] = info
		}
	}

	return stats
}

// Clear 清空所有缓存
func (cm *CacheManager) Clear(ctx context.Context) error {
	// 清空本地缓存
	cm.local.mu.Lock()
	cm.local.data = make(map[string]*CacheItem)
	cm.local.mu.Unlock()

	// 清空Redis缓存（谨慎使用）
	if cm.redis != nil {
		return cm.redis.FlushDB(ctx).Err()
	}

	return nil
}

// Enable/Disable 启用/禁用缓存
func (cm *CacheManager) Enable() {
	cm.enabled = true
}

func (cm *CacheManager) Disable() {
	cm.enabled = false
}

// 全局缓存管理器实例
var GlobalCache *CacheManager

// InitCache 初始化全局缓存
func InitCache() {
	// 这里可以根据配置决定是否使用Redis
	// 目前先创建一个只有本地缓存的实例
	GlobalCache = NewCacheManager(nil)
}
