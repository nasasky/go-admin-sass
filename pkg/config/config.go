package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v2"
)

// AppConfig 全局配置实例
var AppConfig *Config

// Config 应用配置结构
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	JWT      JWTConfig      `yaml:"jwt"`
	MongoDB  MongoDBConfig  `yaml:"mongodb"`
	Log      LogConfig      `yaml:"log"`
	Security SecurityConfig `yaml:"security"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         string        `yaml:"port" env:"SERVER_PORT" default:"8801"`
	Mode         string        `yaml:"mode" env:"GIN_MODE" default:"debug"`
	ReadTimeout  time.Duration `yaml:"read_timeout" default:"30s"`
	WriteTimeout time.Duration `yaml:"write_timeout" default:"30s"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver          string        `yaml:"driver" default:"mysql"`
	DSN             string        `yaml:"dsn" env:"MYSQL_DSN"`
	MaxIdleConns    int           `yaml:"max_idle_conns" default:"10"`
	MaxOpenConns    int           `yaml:"max_open_conns" default:"100"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" default:"1h"`
	LogLevel        string        `yaml:"log_level" default:"info"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Addr         string        `yaml:"addr" env:"REDIS_ADDR" default:"localhost:6379"`
	Password     string        `yaml:"password" env:"REDIS_PASSWORD"`
	DB           int           `yaml:"db" env:"REDIS_DB" default:"0"`
	PoolSize     int           `yaml:"pool_size" default:"10"`
	DialTimeout  time.Duration `yaml:"dial_timeout" default:"5s"`
	ReadTimeout  time.Duration `yaml:"read_timeout" default:"3s"`
	WriteTimeout time.Duration `yaml:"write_timeout" default:"3s"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	SigningKey      string        `yaml:"signing_key" env:"JWT_SIGNING_KEY"`
	Expiry          time.Duration `yaml:"expiry" default:"24h"`
	RefreshExpiry   time.Duration `yaml:"refresh_expiry" default:"168h"` // 7天
	Issuer          string        `yaml:"issuer" default:"nasa-go-admin"`
	EnableBlacklist bool          `yaml:"enable_blacklist" default:"true"`
}

// MongoDBConfig MongoDB配置
type MongoDBConfig struct {
	Databases map[string]MongoDatabase `yaml:"databases"`
}

// MongoDatabase MongoDB数据库配置
type MongoDatabase struct {
	URI         string            `yaml:"uri"`
	Collections map[string]string `yaml:"collections"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `yaml:"level" default:"info"`
	Format     string `yaml:"format" default:"json"`   // json 或 text
	Output     string `yaml:"output" default:"stdout"` // stdout, file, both
	FilePath   string `yaml:"file_path" default:"logs/app.log"`
	MaxSize    int    `yaml:"max_size" default:"100"` // MB
	MaxBackups int    `yaml:"max_backups" default:"7"`
	MaxAge     int    `yaml:"max_age" default:"30"` // days
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	EnableHTTPS     bool     `yaml:"enable_https" default:"false"`
	TLSCertFile     string   `yaml:"tls_cert_file"`
	TLSKeyFile      string   `yaml:"tls_key_file"`
	AllowedOrigins  []string `yaml:"allowed_origins"`
	TrustedProxies  []string `yaml:"trusted_proxies"`
	RateLimit       int      `yaml:"rate_limit" default:"1000"` // 每分钟请求数
	EnableRateLimit bool     `yaml:"enable_rate_limit" default:"true"`
}

// InitConfig 初始化配置
func InitConfig() error {
	// 加载环境变量
	if err := loadEnv(); err != nil {
		log.Printf("Warning: failed to load .env file: %v", err)
	}

	// 创建默认配置
	config := &Config{}
	setDefaults(config)

	// 尝试从配置文件加载
	if err := loadFromFile(config); err != nil {
		log.Printf("Warning: failed to load config file: %v", err)
	}

	// 从环境变量覆盖配置
	if err := loadFromEnv(config); err != nil {
		return fmt.Errorf("failed to load config from environment: %w", err)
	}

	// 验证配置
	if err := validateConfig(config); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	AppConfig = config
	return nil
}

// loadEnv 加载环境变量文件
func loadEnv() error {
	env := os.Getenv("GO_ENV")
	if env == "" {
		env = "development"
	}

	envFiles := []string{
		".env",
		fmt.Sprintf(".env.%s", env),
		".env.local",
	}

	for _, file := range envFiles {
		if _, err := os.Stat(file); err == nil {
			if err := godotenv.Load(file); err != nil {
				return err
			}
		}
	}

	return nil
}

// setDefaults 设置默认值
func setDefaults(config *Config) {
	config.Server.Port = "8801"
	config.Server.Mode = "debug"
	config.Server.ReadTimeout = 30 * time.Second
	config.Server.WriteTimeout = 30 * time.Second

	config.Database.Driver = "mysql"
	config.Database.MaxIdleConns = 10
	config.Database.MaxOpenConns = 100
	config.Database.ConnMaxLifetime = time.Hour
	config.Database.LogLevel = "info"

	config.Redis.Addr = "localhost:6379"
	config.Redis.DB = 0
	config.Redis.PoolSize = 10
	config.Redis.DialTimeout = 5 * time.Second
	config.Redis.ReadTimeout = 3 * time.Second
	config.Redis.WriteTimeout = 3 * time.Second

	config.JWT.Expiry = 24 * time.Hour
	config.JWT.RefreshExpiry = 168 * time.Hour
	config.JWT.Issuer = "nasa-go-admin"
	config.JWT.EnableBlacklist = true

	config.Log.Level = "info"
	config.Log.Format = "json"
	config.Log.Output = "stdout"
	config.Log.FilePath = "logs/app.log"
	config.Log.MaxSize = 100
	config.Log.MaxBackups = 7
	config.Log.MaxAge = 30

	config.Security.EnableHTTPS = false
	config.Security.RateLimit = 1000
	config.Security.EnableRateLimit = true
}

// loadFromFile 从配置文件加载
func loadFromFile(config *Config) error {
	configFile := os.Getenv("CONFIG_FILE")
	if configFile == "" {
		configFile = "config/config.yaml"
	}

	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, config)
}

// loadFromEnv 从环境变量加载
func loadFromEnv(config *Config) error {
	// Server配置
	if port := os.Getenv("SERVER_PORT"); port != "" {
		config.Server.Port = port
	}
	if mode := os.Getenv("GIN_MODE"); mode != "" {
		config.Server.Mode = mode
	}

	// Database配置 - 兼容原有的环境变量名
	if dsn := os.Getenv("Mysql"); dsn != "" {
		config.Database.DSN = dsn
	} else if dsn := os.Getenv("MYSQL_DSN"); dsn != "" {
		config.Database.DSN = dsn
	}

	// Redis配置
	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
		config.Redis.Addr = addr
	}
	if password := os.Getenv("REDIS_PASSWORD"); password != "" {
		config.Redis.Password = password
	}
	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		if db, err := strconv.Atoi(dbStr); err == nil {
			config.Redis.DB = db
		}
	}

	// JWT配置
	if signingKey := os.Getenv("JWT_SIGNING_KEY"); signingKey != "" {
		config.JWT.SigningKey = signingKey
	}

	return nil
}

// validateConfig 验证配置
func validateConfig(config *Config) error {
	// 验证必需的配置项
	if config.Database.DSN == "" {
		return fmt.Errorf("database DSN is required")
	}

	if config.JWT.SigningKey == "" {
		return fmt.Errorf("JWT signing key is required")
	}

	// 验证端口号
	if _, err := strconv.Atoi(strings.TrimPrefix(config.Server.Port, ":")); err != nil {
		return fmt.Errorf("invalid server port: %s", config.Server.Port)
	}

	// 验证模式
	validModes := []string{"debug", "release", "test"}
	modeValid := false
	for _, mode := range validModes {
		if config.Server.Mode == mode {
			modeValid = true
			break
		}
	}
	if !modeValid {
		return fmt.Errorf("invalid server mode: %s", config.Server.Mode)
	}

	return nil
}

// GetConfig 获取配置实例
func GetConfig() *Config {
	if AppConfig == nil {
		log.Fatal("config not initialized, call InitConfig() first")
	}
	return AppConfig
}

// IsProduction 判断是否为生产环境
func IsProduction() bool {
	return AppConfig != nil && AppConfig.Server.Mode == "release"
}

// IsDevelopment 判断是否为开发环境
func IsDevelopment() bool {
	return AppConfig != nil && AppConfig.Server.Mode == "debug"
}
