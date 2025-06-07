package db

import (
	"database/sql"
	"log"
	"nasa-go-admin/pkg/monitoring"
	"os"
	"path/filepath"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var Dao *gorm.DB

func Init() {
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

	dbLogger := logger.New(
		log.New(file, "\r\n", log.LstdFlags), // 将日志输出设置为文件
		logger.Config{
			SlowThreshold:             time.Second,
			Colorful:                  false,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      false,
			LogLevel:                  logger.Info,
		},
	)
	openDb, err := gorm.Open(mysql.Open(os.Getenv("Mysql")), &gorm.Config{
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
	// 优化连接池配置
	dbCon.SetMaxIdleConns(20)                  // 增加空闲连接数
	dbCon.SetMaxOpenConns(100)                 // 增加最大连接数
	dbCon.SetConnMaxLifetime(time.Hour)        // 连接最大生命周期
	dbCon.SetConnMaxIdleTime(30 * time.Minute) // 空闲连接最大生命周期
	Dao = openDb

	// 启动数据库连接池监控
	go startDBMonitoring(dbCon)
}

// 启动数据库连接池监控
func startDBMonitoring(dbCon *sql.DB) {
	ticker := time.NewTicker(30 * time.Second) // 每30秒更新一次
	defer ticker.Stop()

	for range ticker.C {
		stats := dbCon.Stats()
		monitoring.UpdateDBConnections(stats.InUse)
		monitoring.SaveDatabaseMetric(stats.InUse, stats.Idle, stats.MaxOpenConnections)
	}
}
