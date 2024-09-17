package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

func LoadConfig() RedisConfig {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	redisDB, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		log.Fatalf("Invalid REDIS_DB value: %v", err)
	}

	return RedisConfig{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       redisDB,
	}
}
