package db

import (
	"database/sql"
	"log"
	"nasa-go-admin/pkg/config"
	"nasa-go-admin/pkg/monitoring"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var Dao *gorm.DB

func Init() {
	// 获取配置
	cfg := config.GetConfig()

	// 获取数据库DSN，优先使用环境变量
	dsn := os.Getenv("Mysql")
	if dsn == "" && cfg.Database.DSN != "" {
		dsn = cfg.Database.DSN
	}
	if dsn == "" {
		log.Fatalf("数据库连接字符串未配置，请设置环境变量 Mysql 或配置文件中的 database.dsn")
	}

	// 创建日志文件夹
	logDir := "gormlog"
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create log directory: %v", err)
	}

	// 创建日志文件，文件名包含日期
	logFile := filepath.Join(logDir, time.Now().Format("2006-01-02")+".log")
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	// 根据配置设置日志级别
	var logLevel logger.LogLevel
	switch cfg.Database.LogLevel {
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

	dbLogger := logger.New(
		log.New(file, "\r\n", log.LstdFlags), // 将日志输出设置为文件
		logger.Config{
			SlowThreshold:             time.Second,
			Colorful:                  false,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      false,
			LogLevel:                  logLevel,
		},
	)

	openDb, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger:                                   dbLogger,
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		log.Fatalf("db connection error is %s", err.Error())
	}

	dbCon, err := openDb.DB()
	if err != nil {
		log.Fatalf("openDb.DB error is  %s", err.Error())
	}

	// 使用配置中的连接池参数
	maxOpenConns := cfg.Database.MaxOpenConns
	maxIdleConns := cfg.Database.MaxIdleConns

	// 根据环境变量动态调整（保持向后兼容）
	if envMaxOpen := os.Getenv("DB_MAX_OPEN_CONNS"); envMaxOpen != "" {
		if parsed, err := strconv.Atoi(envMaxOpen); err == nil && parsed > 0 {
			maxOpenConns = parsed
		}
	}

	dbCon.SetMaxIdleConns(maxIdleConns)                    // 空闲连接数
	dbCon.SetMaxOpenConns(maxOpenConns)                    // 最大连接数
	dbCon.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime) // 连接最大生命周期
	dbCon.SetConnMaxIdleTime(30 * time.Minute)             // 空闲连接最大生命周期

	log.Printf("数据库连接池配置 - MaxOpen: %d, MaxIdle: %d, MaxLifetime: %v",
		maxOpenConns, maxIdleConns, cfg.Database.ConnMaxLifetime)
	Dao = openDb

	// 启动数据库连接池监控（降低频率避免过多日志）
	go startDBMonitoring(dbCon)
}

// 启动数据库连接池监控
func startDBMonitoring(dbCon *sql.DB) {
	ticker := time.NewTicker(60 * time.Second) // 每60秒更新一次，减少日志频率
	defer ticker.Stop()

	for range ticker.C {
		stats := dbCon.Stats()

		// 只在连接使用异常时记录日志
		poolUsageRate := float64(stats.OpenConnections) / float64(stats.MaxOpenConnections)
		if poolUsageRate > 0.7 || stats.InUse > 10 || stats.WaitCount > 0 {
			log.Printf("数据库连接池监控 - 打开: %d/%d (%.1f%%), 使用中: %d, 空闲: %d, 等待: %d",
				stats.OpenConnections, stats.MaxOpenConnections, poolUsageRate*100,
				stats.InUse, stats.Idle, stats.WaitCount)
		}

		monitoring.UpdateDBConnections(stats.InUse)
		monitoring.SaveDatabaseMetric(stats.InUse, stats.Idle, stats.MaxOpenConnections)
	}
}
