package config

import (
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v2"
)

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type MongoDBConfig struct {
	URI         string            `yaml:"uri"`
	Collections map[string]string `yaml:"collections"`
}

type Config struct {
	MongoDB struct {
		Databases map[string]MongoDBConfig `yaml:"databases"`
	} `yaml:"mongodb"`
}

var AppConfig Config

func InitConfig() {
	data, err := ioutil.ReadFile("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	err = yaml.Unmarshal(data, &AppConfig)
	if err != nil {
		log.Fatalf("Failed to unmarshal config file: %v", err)
	}
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


