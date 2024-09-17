package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"nasa-go-admin/config"
	"time"
)

var rdb *redis.Client

func InitRedis(config config.RedisConfig) {
	rdb = redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})
}

func StoreToken(userID string, token string, expiration time.Duration) error {
	ctx := context.Background()
	err := rdb.Set(ctx, userID, token, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to store token: %v", err)
	}
	return nil
}

func GetToken(userID string) (string, error) {
	ctx := context.Background()
	token, err := rdb.Get(ctx, userID).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("token not found")
	} else if err != nil {
		return "", fmt.Errorf("failed to get token: %v", err)
	}
	return token, nil
}
