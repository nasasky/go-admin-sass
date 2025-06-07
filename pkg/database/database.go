package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"nasa-go-admin/pkg/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	DB     *gorm.DB
	dbFile *os.File
)

// InitDatabase 初始化数据库连接
func InitDatabase() error {
	cfg := config.GetConfig()

	// 创建日志文件
	logFile, err := createLogFile()
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}
	dbFile = logFile

	// 配置日志
	dbLogger := createLogger(logFile, cfg.Database.LogLevel)

	// 连接数据库
	db, err := gorm.Open(mysql.Open(cfg.Database.DSN), &gorm.Config{
		Logger:                                   dbLogger,
		DisableForeignKeyConstraintWhenMigrating: true,
		PrepareStmt:                              true, // 启用预编译语句缓存
		CreateBatchSize:                          1000, // 批量创建大小
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	DB = db
	log.Printf("Database connected successfully. Pool settings: MaxIdle=%d, MaxOpen=%d, MaxLifetime=%v",
		cfg.Database.MaxIdleConns, cfg.Database.MaxOpenConns, cfg.Database.ConnMaxLifetime)

	return nil
}

// createLogFile 创建日志文件
func createLogFile() (*os.File, error) {
	logDir := "gormlog"
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		return nil, err
	}

	logFile := filepath.Join(logDir, time.Now().Format("2006-01-02")+".log")
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return file, nil
}

// createLogger 创建数据库日志器
func createLogger(file *os.File, level string) logger.Interface {
	var logLevel logger.LogLevel
	switch level {
	case "silent":
		logLevel = logger.Silent
	case "error":
		logLevel = logger.Error
	case "warn":
		logLevel = logger.Warn
	case "info":
		logLevel = logger.Info
	default:
		logLevel = logger.Info
	}

	return logger.New(
		log.New(file, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond, // 慢查询阈值
			Colorful:                  false,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true, // 记录参数化查询
			LogLevel:                  logLevel,
		},
	)
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	if DB == nil {
		log.Fatal("database not initialized, call InitDatabase() first")
	}
	return DB
}

// HealthCheck 数据库健康检查
func HealthCheck() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// GetStats 获取数据库连接池统计信息
func GetStats() map[string]interface{} {
	if DB == nil {
		return map[string]interface{}{
			"error": "database not initialized",
		}
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return map[string]interface{}{
			"error": fmt.Sprintf("failed to get underlying sql.DB: %v", err),
		}
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}
}

// Close 关闭数据库连接
func Close() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	if dbFile != nil {
		dbFile.Close()
	}

	return sqlDB.Close()
}

// WithTransaction 事务处理器
func WithTransaction(fn func(*gorm.DB) error) error {
	tx := DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// AutoMigrate 自动迁移数据库表结构
func AutoMigrate(models ...interface{}) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	for _, model := range models {
		if err := DB.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to migrate model %T: %w", model, err)
		}
	}

	log.Printf("Database migration completed for %d models", len(models))
	return nil
}
