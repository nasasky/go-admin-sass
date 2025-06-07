package redis

import (
	"context"
	"errors"
	"fmt"
	"log"
	"nasa-go-admin/config"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	rdb         *redis.Client
	initOnce    sync.Once
	initialized bool
	initErr     error
	ErrNil      = errors.New("redis: nil")
)

// InitRedis 初始化 Redis 客户端
func InitRedis(config config.RedisConfig) error {
	initOnce.Do(func() {
		log.Printf("Initializing Redis client with address: %s, DB: %d", config.Addr, config.DB)

		rdb = redis.NewClient(&redis.Options{
			Addr:         config.Addr,
			Password:     config.Password,
			DB:           config.DB,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
			PoolSize:     10,
		})

		// 测试连接
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := rdb.Ping(ctx).Err(); err != nil {
			initErr = fmt.Errorf("failed to connect to Redis at %s: %w", config.Addr, err)
			log.Printf("ERROR: %v", initErr)
			return
		}

		initialized = true
		log.Printf("Successfully connected to Redis at %s, DB: %d", config.Addr, config.DB)
	})

	return initErr
}

// GetClient 获取 Redis 客户端实例
func GetClient() *redis.Client {
	if !initialized && initErr == nil {
		// 尝试使用默认配置初始化
		cfg := config.RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		}

		log.Print("Redis client not initialized, attempting with default configuration")
		if err := InitRedis(cfg); err != nil {
			log.Printf("ERROR: Failed to initialize Redis with default config: %v", err)
		}
	}

	if rdb == nil {
		log.Print("WARNING: Redis client is nil, some features may not work")
	}

	return rdb
}

// IsConnected 检查 Redis 是否已连接
func IsConnected() bool {
	if rdb == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	return rdb.Ping(ctx).Err() == nil
}

// CloseRedis 关闭 Redis 连接
func CloseRedis() error {
	if rdb != nil {
		log.Print("Closing Redis connection")
		return rdb.Close()
	}
	return nil
}

// 存储用户信息，包括 token 和其他字段
func StoreUserInfo(userID string, userInfo map[string]interface{}, expiration time.Duration) error {
	key := "user_info:" + userID // 添加前缀
	ctx := context.Background()
	// 使用 HMSET 存储用户信息
	err := rdb.HMSet(ctx, key, userInfo).Err()
	if err != nil {
		return fmt.Errorf("failed to store user info: %v", err)
	}
	// 设置过期时间
	err = rdb.Expire(ctx, key, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set expiration for user info: %v", err)
	}
	return nil
}

// 获取用户信息
func GetUserInfo(userID string) (map[string]string, error) {
	key := "user_info:" + userID // 添加前缀
	ctx := context.Background()
	// 使用 HGETALL 获取用户信息
	userInfo, err := rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %v", err)
	}
	if len(userInfo) == 0 {
		return nil, fmt.Errorf("user info not found")
	}
	return userInfo, nil
}

// 删除用户信息
func DeleteUserInfo(userID string) error {
	key := "user_info:" + userID // 添加前缀
	ctx := context.Background()
	err := rdb.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete user info: %v", err)
	}
	return nil
}

func StoreToken(userID string, token string, expiration time.Duration) error {
	key := "token:" + userID // 添加前缀
	ctx := context.Background()
	err := rdb.Set(ctx, key, token, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to store token: %v", err)
	}
	return nil
}

func GetToken(userID string) (string, error) {
	key := "token:" + userID // 添加前缀
	ctx := context.Background()
	token, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("token not found")
	} else if err != nil {
		return "", fmt.Errorf("failed to get token: %v", err)
	}
	return token, nil
}

func DeleteToken(userID string) error {
	key := "token:" + userID // 添加前缀
	ctx := context.Background()
	err := rdb.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete token: %v", err)
	}
	return nil
}

func DeleteKey(key string) error {
	ctx := context.Background()
	err := rdb.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete key: %v", err)
	}
	return nil
}
