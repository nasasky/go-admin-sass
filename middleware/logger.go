package middleware

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

// 设置日志输出文件
func SetupLogFile(logDir string) *os.File {
	// 创建日志文件夹
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create log directory: %v", err)
	}

	// 创建日志文件，文件名包含日期
	logFile := filepath.Join(logDir, time.Now().Format("2006-01-02")+".log")
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	return file
}

// 通用请求日志中间件
func RequestLogger(logDir string) gin.HandlerFunc {
	// 设置日志输出到指定文件
	logFile := SetupLogFile(logDir)
	logger := log.New(logFile, "", log.LstdFlags)

	return func(c *gin.Context) {
		// 开始时间
		start := time.Now()

		// 读取请求体
		bodyBytes, _ := ioutil.ReadAll(c.Request.Body)
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

		// 处理请求
		c.Next()

		// 结束时间
		end := time.Now()
		latency := end.Sub(start)

		// 请求方法
		method := c.Request.Method

		// 请求路径
		path := c.Request.URL.Path

		// 请求参数
		params := c.Request.URL.RawQuery

		// 客户端IP
		clientIP := c.ClientIP()

		// 请求体参数
		requestBody := string(bodyBytes)

		// 日志格式
		logger.Printf("%s %s %s %s %s %s", method, path, clientIP, params, requestBody, latency)
	}
}
