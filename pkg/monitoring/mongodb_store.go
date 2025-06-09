package monitoring

import (
	"context"
	"log"
	"nasa-go-admin/mongodb"
	"nasa-go-admin/pkg/goroutinepool"
	"nasa-go-admin/utils"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// HTTPMetric HTTP请求指标（简化版）
type HTTPMetric struct {
	Timestamp  string  `bson:"timestamp"`
	Method     string  `bson:"method"`
	Endpoint   string  `bson:"endpoint"`
	StatusCode int     `bson:"status_code"`
	Duration   float64 `bson:"duration"`
	UserAgent  string  `bson:"user_agent,omitempty"`
	ClientIP   string  `bson:"client_ip,omitempty"`
	UserID     string  `bson:"user_id,omitempty"`
}

// BusinessMetric 业务指标（简化版）
type BusinessMetric struct {
	Timestamp  string `bson:"timestamp"`
	MetricType string `bson:"metric_type"`
	Count      int64  `bson:"count"`
	UserID     string `bson:"user_id,omitempty"`
}

// DatabaseMetric 数据库指标（简化版）
type DatabaseMetric struct {
	Timestamp        string `bson:"timestamp"`
	ConnectionsInUse int    `bson:"connections_in_use"`
	ConnectionsIdle  int    `bson:"connections_idle"`
	MaxOpenConns     int    `bson:"max_open_conns"`
}

// getDatabaseName 根据接口路径获取合适的数据库名称
func getDatabaseName(path string) string {
	// 只有管理后台接口才存储到admin_log_db
	if strings.HasPrefix(path, "/api/admin/") {
		return "admin_log_db"
	}
	// 应用接口存储到app_log_db
	if strings.HasPrefix(path, "/api/app/") {
		return "app_log_db"
	}
	// 其他接口存储到default_log_db
	return "default_log_db"
}

// SaveHTTPMetric 保存HTTP指标到MongoDB
func SaveHTTPMetric(c *gin.Context, duration float64) {
	metric := HTTPMetric{
		Timestamp:  utils.GetCurrentTimeForMongo(),
		Method:     c.Request.Method,
		Endpoint:   c.FullPath(),
		StatusCode: c.Writer.Status(),
		Duration:   duration,
		UserAgent:  c.GetHeader("User-Agent"),
		ClientIP:   c.ClientIP(),
	}

	// 尝试获取用户ID
	if userInfo, exists := c.Get("userInfo"); exists {
		if user, ok := userInfo.(map[string]interface{}); ok {
			if id, exists := user["id"]; exists {
				metric.UserID = strconv.Itoa(int(id.(float64)))
			}
		}
	}

	// 根据接口路径选择数据库
	databaseName := getDatabaseName(c.Request.URL.Path)

	// 使用goroutine池来避免goroutine泄漏
	goroutinepool.Submit(func() error {
		collection := mongodb.GetCollection(databaseName, "logs")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := collection.InsertOne(ctx, metric)
		if err != nil {
			log.Printf("保存HTTP指标到MongoDB失败: %v", err)
		}
		return err
	})
}

// SaveBusinessMetric 保存业务指标到MongoDB (存储到专门的业务指标数据库)
func SaveBusinessMetric(metricType string, userID string) {
	metric := BusinessMetric{
		Timestamp:  utils.GetCurrentTimeForMongo(),
		MetricType: metricType,
		Count:      1,
		UserID:     userID,
	}

	goroutinepool.Submit(func() error {
		collection := mongodb.GetCollection("business_metrics_db", "metrics")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := collection.InsertOne(ctx, metric)
		if err != nil {
			log.Printf("保存业务指标到MongoDB失败: %v", err)
		}
		return err
	})
}

// SaveDatabaseMetric 保存数据库指标到MongoDB (存储到专门的系统指标数据库)
func SaveDatabaseMetric(connectionsInUse, connectionsIdle, maxOpenConns int) {
	metric := DatabaseMetric{
		Timestamp:        utils.GetCurrentTimeForMongo(),
		ConnectionsInUse: connectionsInUse,
		ConnectionsIdle:  connectionsIdle,
		MaxOpenConns:     maxOpenConns,
	}

	goroutinepool.Submit(func() error {
		collection := mongodb.GetCollection("system_metrics_db", "metrics")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := collection.InsertOne(ctx, metric)
		if err != nil {
			log.Printf("保存数据库指标到MongoDB失败: %v", err)
		}
		return err
	})
}

// GetMonitoringStats 获取监控统计数据
func GetMonitoringStats(timeRange string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 计算时间范围
	now := time.Now()
	var startTime time.Time
	switch timeRange {
	case "1h":
		startTime = now.Add(-time.Hour)
	case "24h":
		startTime = now.Add(-24 * time.Hour)
	case "7d":
		startTime = now.Add(-7 * 24 * time.Hour)
	default:
		startTime = now.Add(-time.Hour)
	}

	// 由于时间字段现在是字符串，我们使用字符串比较
	startTimeStr := startTime.Format("2006-01-02 15:04:05")
	nowTimeStr := now.Format("2006-01-02 15:04:05")

	filter := bson.M{
		"timestamp": bson.M{
			"$gte": startTimeStr,
			"$lte": nowTimeStr,
		},
	}

	// 从管理后台日志数据库统计HTTP请求
	adminCollection := mongodb.GetCollection("admin_log_db", "logs")
	httpFilter := filter
	httpFilter["method"] = bson.M{"$exists": true}
	httpCount, _ := adminCollection.CountDocuments(ctx, httpFilter)

	// 从业务指标数据库统计用户登录
	businessCollection := mongodb.GetCollection("business_metrics_db", "metrics")
	loginFilter := filter
	loginFilter["metric_type"] = "user_login"
	loginCount, _ := businessCollection.CountDocuments(ctx, loginFilter)

	// 统计用户注册
	registerFilter := filter
	registerFilter["metric_type"] = "user_register"
	registerCount, _ := businessCollection.CountDocuments(ctx, registerFilter)

	// 从系统指标数据库获取最新的数据库指标
	systemCollection := mongodb.GetCollection("system_metrics_db", "metrics")
	dbFilter := bson.M{"connections_in_use": bson.M{"$exists": true}}
	var latestDbMetric DatabaseMetric
	err := systemCollection.FindOne(ctx, dbFilter, nil).Decode(&latestDbMetric)
	if err != nil {
		log.Printf("获取数据库指标失败: %v", err)
	}

	return map[string]interface{}{
		"timeRange": timeRange,
		"timestamp": now,
		"stats": map[string]interface{}{
			"admin_http_requests": httpCount,
			"user_logins":         loginCount,
			"user_registers":      registerCount,
			"db_connections":      latestDbMetric.ConnectionsInUse,
			"db_max_conns":        latestDbMetric.MaxOpenConns,
		},
	}, nil
}

// GetRecentHTTPRequests 获取最近的HTTP请求 (只从管理后台日志数据库获取)
func GetRecentHTTPRequests(limit int64) ([]HTTPMetric, error) {
	collection := mongodb.GetCollection("admin_log_db", "logs")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"method": bson.M{"$exists": true}}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var metrics []HTTPMetric
	if err = cursor.All(ctx, &metrics); err != nil {
		return nil, err
	}

	// 只返回最新的指定数量
	if int64(len(metrics)) > limit {
		metrics = metrics[:limit]
	}

	return metrics, nil
}
