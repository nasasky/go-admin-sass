package monitoring

import (
	"context"
	"log"
	"nasa-go-admin/mongodb"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// HTTPMetric HTTP请求指标（简化版）
type HTTPMetric struct {
	Timestamp  time.Time `bson:"timestamp"`
	Method     string    `bson:"method"`
	Endpoint   string    `bson:"endpoint"`
	StatusCode int       `bson:"status_code"`
	Duration   float64   `bson:"duration"`
	UserAgent  string    `bson:"user_agent,omitempty"`
	ClientIP   string    `bson:"client_ip,omitempty"`
	UserID     string    `bson:"user_id,omitempty"`
}

// BusinessMetric 业务指标（简化版）
type BusinessMetric struct {
	Timestamp  time.Time `bson:"timestamp"`
	MetricType string    `bson:"metric_type"`
	Count      int64     `bson:"count"`
	UserID     string    `bson:"user_id,omitempty"`
}

// DatabaseMetric 数据库指标（简化版）
type DatabaseMetric struct {
	Timestamp        time.Time `bson:"timestamp"`
	ConnectionsInUse int       `bson:"connections_in_use"`
	ConnectionsIdle  int       `bson:"connections_idle"`
	MaxOpenConns     int       `bson:"max_open_conns"`
}

// SaveHTTPMetric 保存HTTP指标到MongoDB
func SaveHTTPMetric(c *gin.Context, duration float64) {
	metric := HTTPMetric{
		Timestamp:  time.Now(),
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

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("保存HTTP指标失败: %v", r)
			}
		}()

		collection := mongodb.GetCollection("admin_log_db", "logs")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := collection.InsertOne(ctx, metric)
		if err != nil {
			log.Printf("保存HTTP指标到MongoDB失败: %v", err)
		}
	}()
}

// SaveBusinessMetric 保存业务指标到MongoDB
func SaveBusinessMetric(metricType string, userID string) {
	metric := BusinessMetric{
		Timestamp:  time.Now(),
		MetricType: metricType,
		Count:      1,
		UserID:     userID,
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("保存业务指标失败: %v", r)
			}
		}()

		collection := mongodb.GetCollection("admin_log_db", "logs")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := collection.InsertOne(ctx, metric)
		if err != nil {
			log.Printf("保存业务指标到MongoDB失败: %v", err)
		}
	}()
}

// SaveDatabaseMetric 保存数据库指标到MongoDB
func SaveDatabaseMetric(connectionsInUse, connectionsIdle, maxOpenConns int) {
	metric := DatabaseMetric{
		Timestamp:        time.Now(),
		ConnectionsInUse: connectionsInUse,
		ConnectionsIdle:  connectionsIdle,
		MaxOpenConns:     maxOpenConns,
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("保存数据库指标失败: %v", r)
			}
		}()

		collection := mongodb.GetCollection("admin_log_db", "logs")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := collection.InsertOne(ctx, metric)
		if err != nil {
			log.Printf("保存数据库指标到MongoDB失败: %v", err)
		}
	}()
}

// GetMonitoringStats 获取监控统计数据
func GetMonitoringStats(timeRange string) (map[string]interface{}, error) {
	collection := mongodb.GetCollection("admin_log_db", "logs")
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

	filter := bson.M{
		"timestamp": bson.M{
			"$gte": startTime,
			"$lte": now,
		},
	}

	// 统计HTTP请求
	httpFilter := filter
	httpFilter["method"] = bson.M{"$exists": true}
	httpCount, _ := collection.CountDocuments(ctx, httpFilter)

	// 统计用户登录
	loginFilter := filter
	loginFilter["metric_type"] = "user_login"
	loginCount, _ := collection.CountDocuments(ctx, loginFilter)

	// 统计用户注册
	registerFilter := filter
	registerFilter["metric_type"] = "user_register"
	registerCount, _ := collection.CountDocuments(ctx, registerFilter)

	// 获取最新的数据库指标
	dbFilter := bson.M{"connections_in_use": bson.M{"$exists": true}}
	var latestDbMetric DatabaseMetric
	err := collection.FindOne(ctx, dbFilter, nil).Decode(&latestDbMetric)
	if err != nil {
		log.Printf("获取数据库指标失败: %v", err)
	}

	return map[string]interface{}{
		"timeRange": timeRange,
		"timestamp": now,
		"stats": map[string]interface{}{
			"http_requests":  httpCount,
			"user_logins":    loginCount,
			"user_registers": registerCount,
			"db_connections": latestDbMetric.ConnectionsInUse,
			"db_max_conns":   latestDbMetric.MaxOpenConns,
		},
	}, nil
}

// GetRecentHTTPRequests 获取最近的HTTP请求
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
