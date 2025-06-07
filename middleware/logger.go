package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"strings"
	"time"

	"nasa-go-admin/redis" // 假设 Redis 操作封装在这个包中

	"github.com/gin-gonic/gin"
)

// 通用请求日志中间件
func RequestLogger(logType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过特定路径的日志记录
		if c.Request.URL.Path == "/api/admin/system/log" {
			c.Next()
			return
		}

		// 根据日志类型选择数据库
		var databaseName string
		switch logType {
		case "request_app_log":
			databaseName = "app_log_db"
		case "request_admin_log":
			databaseName = "admin_log_db"
		default:
			databaseName = "default_log_db"
		}

		// 获取 MongoDB 集合
		collection := GetMongoCollection(databaseName, "logs")

		// 开始时间
		start := time.Now()

		// 读取请求体
		bodyBytes, _ := ioutil.ReadAll(c.Request.Body)
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

		// 创建一个缓冲区，用于捕获响应数据
		responseBody := &responseWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = responseBody

		// 处理请求
		c.Next()

		// 结束时间
		end := time.Now()
		latency := end.Sub(start)

		// 请求方法
		method := c.Request.Method

		// 请求路径
		path := c.Request.URL.Path

		// 处理所有参数
		params := extractRequestParams(c, bodyBytes)

		// 客户端IP
		clientIP := c.ClientIP()

		// 格式化时间
		timestamp := start.Format("2006-01-02 15:04:05")

		// 获取用户信息
		user, exists := c.Get("uid")
		var userInfo map[string]string
		if exists {
			// 查询 Redis 中的用户信息
			userID := fmt.Sprintf("%v", user) // 将 uid 转为字符串
			var err error
			userInfo, err = redis.GetUserInfo(userID) // 假设 GetUserInfo 是封装的 Redis 查询方法
			if err != nil {
				log.Printf("Failed to get user info from Redis: %v", err)
			}
		} else {
			userInfo = map[string]string{"error": "user not found"}
		}

		// 获取响应状态码和返回数据
		statusCode := c.Writer.Status()
		responseData := responseBody.body.String()

		// 提取响应中的 code 和 message
		var responseMap map[string]interface{}
		var code interface{} = nil
		var message interface{} = nil
		if err := json.Unmarshal([]byte(responseData), &responseMap); err == nil {
			if val, ok := responseMap["code"]; ok {
				code = val
			}
			if val, ok := responseMap["message"]; ok {
				message = val
			}
		}

		// 日志记录
		logEntry := map[string]interface{}{
			"method":       method,
			"path":         path,
			"clientIP":     clientIP,
			"params":       stringifyParams(params), // 转换为字符串
			"latency":      latency.String(),
			"timestamp":    timestamp,
			"user":         userInfo["username"], // 记录从 Redis 获取的用户信息
			"status":       statusCode,
			"responseBody": responseData, // 记录完整返回参数
			"code":         code,         // 单独记录 code
			"message":      message,      // 单独记录 message
		}

		// 保存日志到 MongoDB
		_, err := collection.InsertOne(context.Background(), logEntry)
		if err != nil {
			log.Printf("Failed to insert log entry: %v", err)
		}
	}
}

// 自定义响应捕获器
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body.Write(b) // 捕获响应数据
	return rw.ResponseWriter.Write(b)
}

// 提取请求参数
func extractRequestParams(c *gin.Context, bodyBytes []byte) map[string]interface{} {
	params := make(map[string]interface{})

	// 处理查询参数 (GET参数)
	if c.Request.URL.RawQuery != "" {
		queryParams, _ := url.ParseQuery(c.Request.URL.RawQuery)
		for key, values := range queryParams {
			if len(values) == 1 {
				params[key] = values[0]
			} else {
				params[key] = values
			}
		}
	}

	// 处理请求体参数 (POST/PUT/DELETE等方法)
	if len(bodyBytes) > 0 {
		contentType := c.GetHeader("Content-Type")

		// 处理JSON格式数据
		if strings.Contains(contentType, "application/json") {
			var jsonParams map[string]interface{}
			if err := json.Unmarshal(bodyBytes, &jsonParams); err == nil {
				for k, v := range jsonParams {
					params[k] = v
				}
			}
		} else if strings.Contains(contentType, "application/x-www-form-urlencoded") {
			// 处理表单数据
			formParams, _ := url.ParseQuery(string(bodyBytes))
			for key, values := range formParams {
				if len(values) == 1 {
					params[key] = values[0]
				} else {
					params[key] = values
				}
			}
		}
	}

	// 敏感信息过滤
	filterSensitiveData(params)

	return params
}

// 过滤敏感信息
func filterSensitiveData(params map[string]interface{}) {
	sensitiveFields := []string{"password", "pwd", "token", "secret", "key"}
	for _, field := range sensitiveFields {
		if params[field] != nil {
			params[field] = "******"
		}
	}
}

func stringifyParams(params map[string]interface{}) string {
	// 将参数序列化为 JSON 字符串
	jsonData, err := json.Marshal(params)
	if err != nil {
		// 如果序列化失败，返回一个提示字符串
		return "failed to stringify params"
	}
	return string(jsonData)
}
