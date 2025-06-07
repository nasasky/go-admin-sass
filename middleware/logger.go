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

		// 获取用户信息 - 改进错误处理和调试
		var userInfo map[string]interface{}
		var userID string
		var username string = "anonymous"

		// 尝试从多个来源获取用户ID
		if user, exists := c.Get("uid"); exists && user != nil {
			userID = fmt.Sprintf("%v", user)
			log.Printf("[DEBUG] 从context获取到用户ID: %s", userID)
		} else if userIDFromHeader := c.GetHeader("User-ID"); userIDFromHeader != "" {
			userID = userIDFromHeader
			log.Printf("[DEBUG] 从header获取到用户ID: %s", userID)
		} else if userIDFromQuery := c.Query("user_id"); userIDFromQuery != "" {
			userID = userIDFromQuery
			log.Printf("[DEBUG] 从query获取到用户ID: %s", userID)
		} else {
			log.Printf("[DEBUG] 未找到用户ID，请求路径: %s", path)
		}

		// 如果找到用户ID，尝试从Redis获取用户信息
		if userID != "" {
			// 尝试多种Redis键格式
			userInfo = getUserInfoFromRedis(userID)

			if userInfo != nil && len(userInfo) > 0 {
				log.Printf("[DEBUG] 成功从Redis获取用户信息: %+v", userInfo)
				// 获取用户名
				if uname, ok := userInfo["username"].(string); ok && uname != "" {
					username = uname
				}
			} else {
				log.Printf("[DEBUG] Redis中未找到用户 %s 的信息", userID)
				// 创建基本用户信息
				userInfo = map[string]interface{}{
					"user_id": userID,
					"error":   "user info not found in redis",
				}
			}
		} else {
			// 没有用户ID的情况
			userInfo = map[string]interface{}{
				"user_id": nil,
				"note":    "no user authentication",
			}
			log.Printf("[DEBUG] 请求未包含用户认证信息")
		}

		// 获取响应状态码和返回数据
		statusCode := c.Writer.Status()
		responseData := responseBody.body.String()

		// 提取响应中的 code 和 message
		var responseMap map[string]interface{}
		var code interface{} = nil
		var message interface{} = nil
		if responseData != "" {
			if err := json.Unmarshal([]byte(responseData), &responseMap); err == nil {
				if val, ok := responseMap["code"]; ok {
					code = val
				}
				if val, ok := responseMap["message"]; ok {
					message = val
				}
			}
		}

		// 构建完整的日志记录
		logEntry := map[string]interface{}{
			"timestamp":        timestamp,
			"method":           method,
			"path":             path,
			"client_ip":        clientIP,
			"user_id":          userID,
			"username":         username,
			"user_info":        userInfo,
			"request_params":   params,
			"latency_ms":       latency.Milliseconds(),
			"latency":          latency.String(),
			"status_code":      statusCode,
			"response_body":    responseData,
			"response_code":    code,
			"response_message": message,
			"request_size":     len(bodyBytes),
			"response_size":    len(responseData),
			"user_agent":       c.Request.UserAgent(),
			"referer":          c.Request.Referer(),
			"log_type":         logType,
		}

		log.Printf("[DEBUG] 准备保存日志: 用户ID=%s, 用户名=%s, 路径=%s", userID, username, path)

		// 异步保存日志到 MongoDB，避免影响请求性能
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if _, err := collection.InsertOne(ctx, logEntry); err != nil {
				log.Printf("Failed to insert log entry: %v", err)
			} else {
				log.Printf("[DEBUG] 日志保存成功: 用户ID=%s, 路径=%s", userID, path)
			}
		}()
	}
}

// getUserInfoFromRedis 从Redis获取用户信息，尝试多种键格式
func getUserInfoFromRedis(userID string) map[string]interface{} {
	// 尝试的键格式列表
	keyFormats := []string{
		"user_info:" + userID, // 原始格式
		"user:info:" + userID, // 新格式
		userID,                // 直接使用userID
	}

	for _, key := range keyFormats {
		log.Printf("[DEBUG] 尝试Redis键: %s", key)

		// 方法1: 尝试作为Hash获取
		if hashData, err := redis.GetUserInfo(userID); err == nil && len(hashData) > 0 {
			log.Printf("[DEBUG] 通过Hash方式获取到用户信息: %+v", hashData)
			// 转换 map[string]string 到 map[string]interface{}
			userInfo := make(map[string]interface{})
			for k, v := range hashData {
				userInfo[k] = v
			}
			return userInfo
		}

		// 方法2: 尝试作为JSON字符串获取
		if jsonData, err := redis.GetClient().Get(context.Background(), key).Result(); err == nil && jsonData != "" {
			log.Printf("[DEBUG] 通过JSON方式获取到数据: %s", jsonData)
			var userInfo map[string]interface{}
			if err := json.Unmarshal([]byte(jsonData), &userInfo); err == nil {
				log.Printf("[DEBUG] JSON解析成功: %+v", userInfo)
				return userInfo
			} else {
				log.Printf("[DEBUG] JSON解析失败: %v", err)
			}
		}
	}

	log.Printf("[DEBUG] 所有Redis键格式都未找到用户信息")
	return nil
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

// 提取请求参数 - 改进参数提取逻辑
func extractRequestParams(c *gin.Context, bodyBytes []byte) map[string]interface{} {
	params := make(map[string]interface{})

	// 处理查询参数 (GET参数)
	if c.Request.URL.RawQuery != "" {
		queryParams, err := url.ParseQuery(c.Request.URL.RawQuery)
		if err == nil {
			for key, values := range queryParams {
				if len(values) == 1 {
					params[key] = values[0]
				} else {
					params[key] = values
				}
			}
		}
	}

	// 处理路径参数
	for _, param := range c.Params {
		params["path_"+param.Key] = param.Value
	}

	// 处理请求体参数 (POST/PUT/DELETE等方法)
	if len(bodyBytes) > 0 {
		contentType := c.GetHeader("Content-Type")

		// 处理JSON格式数据
		if strings.Contains(contentType, "application/json") {
			var jsonParams map[string]interface{}
			if err := json.Unmarshal(bodyBytes, &jsonParams); err == nil {
				for k, v := range jsonParams {
					params["body_"+k] = v
				}
			} else {
				// 如果JSON解析失败，保存原始数据
				params["raw_body"] = string(bodyBytes)
			}
		} else if strings.Contains(contentType, "application/x-www-form-urlencoded") {
			// 处理表单数据
			formParams, err := url.ParseQuery(string(bodyBytes))
			if err == nil {
				for key, values := range formParams {
					if len(values) == 1 {
						params["form_"+key] = values[0]
					} else {
						params["form_"+key] = values
					}
				}
			}
		} else if strings.Contains(contentType, "multipart/form-data") {
			// 处理文件上传表单
			if err := c.Request.ParseMultipartForm(32 << 20); err == nil { // 32MB max
				if c.Request.MultipartForm != nil {
					for key, values := range c.Request.MultipartForm.Value {
						if len(values) == 1 {
							params["multipart_"+key] = values[0]
						} else {
							params["multipart_"+key] = values
						}
					}
					// 记录文件信息
					if len(c.Request.MultipartForm.File) > 0 {
						fileInfo := make(map[string]interface{})
						for key, files := range c.Request.MultipartForm.File {
							fileNames := make([]string, len(files))
							for i, file := range files {
								fileNames[i] = file.Filename
							}
							fileInfo[key] = fileNames
						}
						params["uploaded_files"] = fileInfo
					}
				}
			}
		} else {
			// 其他类型的请求体，保存原始数据（限制长度）
			bodyStr := string(bodyBytes)
			if len(bodyStr) > 1000 {
				bodyStr = bodyStr[:1000] + "...(truncated)"
			}
			params["raw_body"] = bodyStr
		}
	}

	// 敏感信息过滤
	filterSensitiveData(params)

	return params
}

// 过滤敏感信息 - 改进敏感信息过滤
func filterSensitiveData(params map[string]interface{}) {
	sensitiveFields := []string{
		"password", "pwd", "passwd", "pass",
		"token", "access_token", "refresh_token", "auth_token",
		"secret", "key", "api_key", "private_key",
		"credit_card", "card_number", "cvv", "ssn",
		"id_card", "passport",
	}

	for key, value := range params {
		keyLower := strings.ToLower(key)
		for _, sensitiveField := range sensitiveFields {
			if strings.Contains(keyLower, sensitiveField) {
				// 根据值的类型进行处理
				switch v := value.(type) {
				case string:
					if len(v) > 0 {
						params[key] = "******"
					}
				case []string:
					if len(v) > 0 {
						params[key] = []string{"******"}
					}
				default:
					params[key] = "******"
				}
				break
			}
		}
	}
}

// stringifyParams 已经不需要了，因为我们直接存储map[string]interface{}
// 但为了兼容性保留这个函数
func stringifyParams(params map[string]interface{}) string {
	jsonData, err := json.Marshal(params)
	if err != nil {
		return "failed to stringify params"
	}
	return string(jsonData)
}
