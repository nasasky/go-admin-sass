package admin_model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SystemMetrics 系统指标数据
type SystemMetrics struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Timestamp   string             `bson:"timestamp" json:"timestamp"`
	MetricType  string             `bson:"metric_type" json:"metric_type"` // http_request, db_connection, user_action, etc.
	MetricName  string             `bson:"metric_name" json:"metric_name"` // 具体指标名称
	Value       float64            `bson:"value" json:"value"`
	Labels      map[string]string  `bson:"labels,omitempty" json:"labels,omitempty"` // 标签信息
	Description string             `bson:"description,omitempty" json:"description,omitempty"`
	CreatedAt   string             `bson:"created_at" json:"created_at"`
}

// HTTPMetrics HTTP请求指标
type HTTPMetrics struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Timestamp  string             `bson:"timestamp" json:"timestamp"`
	Method     string             `bson:"method" json:"method"`
	Endpoint   string             `bson:"endpoint" json:"endpoint"`
	StatusCode int                `bson:"status_code" json:"status_code"`
	Duration   float64            `bson:"duration" json:"duration"` // 响应时间(秒)
	UserAgent  string             `bson:"user_agent,omitempty" json:"user_agent,omitempty"`
	ClientIP   string             `bson:"client_ip,omitempty" json:"client_ip,omitempty"`
	UserID     int                `bson:"user_id,omitempty" json:"user_id,omitempty"`
	CreatedAt  string             `bson:"created_at" json:"created_at"`
}

// DatabaseMetrics 数据库指标
type DatabaseMetrics struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Timestamp        time.Time          `bson:"timestamp" json:"timestamp"`
	ConnectionsInUse int                `bson:"connections_in_use" json:"connections_in_use"`
	ConnectionsIdle  int                `bson:"connections_idle" json:"connections_idle"`
	MaxOpenConns     int                `bson:"max_open_conns" json:"max_open_conns"`
	QueryCount       int64              `bson:"query_count" json:"query_count"`
	SlowQueryCount   int64              `bson:"slow_query_count" json:"slow_query_count"`
	AvgQueryDuration float64            `bson:"avg_query_duration" json:"avg_query_duration"`
	CreatedAt        time.Time          `bson:"created_at" json:"created_at"`
}

// BusinessMetrics 业务指标
type BusinessMetrics struct {
	ID             primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	Timestamp      time.Time              `bson:"timestamp" json:"timestamp"`
	MetricType     string                 `bson:"metric_type" json:"metric_type"` // user_login, user_register, order_create, etc.
	Count          int64                  `bson:"count" json:"count"`
	TotalCount     int64                  `bson:"total_count" json:"total_count"`   // 累计总数
	HourlyCount    int64                  `bson:"hourly_count" json:"hourly_count"` // 小时内统计
	DailyCount     int64                  `bson:"daily_count" json:"daily_count"`   // 日内统计
	AdditionalData map[string]interface{} `bson:"additional_data,omitempty" json:"additional_data,omitempty"`
	CreatedAt      time.Time              `bson:"created_at" json:"created_at"`
}

// ErrorMetrics 错误指标
type ErrorMetrics struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Timestamp  time.Time          `bson:"timestamp" json:"timestamp"`
	ErrorType  string             `bson:"error_type" json:"error_type"` // api_error, db_error, auth_error
	ErrorCode  string             `bson:"error_code" json:"error_code"`
	ErrorMsg   string             `bson:"error_msg" json:"error_msg"`
	Endpoint   string             `bson:"endpoint,omitempty" json:"endpoint,omitempty"`
	UserID     int                `bson:"user_id,omitempty" json:"user_id,omitempty"`
	StackTrace string             `bson:"stack_trace,omitempty" json:"stack_trace,omitempty"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
}

// CacheMetrics 缓存指标
type CacheMetrics struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Timestamp  time.Time          `bson:"timestamp" json:"timestamp"`
	CacheType  string             `bson:"cache_type" json:"cache_type"` // redis, local_cache
	Operation  string             `bson:"operation" json:"operation"`   // get, set, delete
	HitCount   int64              `bson:"hit_count" json:"hit_count"`
	MissCount  int64              `bson:"miss_count" json:"miss_count"`
	HitRate    float64            `bson:"hit_rate" json:"hit_rate"`
	KeyPattern string             `bson:"key_pattern,omitempty" json:"key_pattern,omitempty"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
}

// MonitoringDashboard 监控仪表板配置
type MonitoringDashboard struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description,omitempty" json:"description,omitempty"`
	Widgets     []DashboardWidget  `bson:"widgets" json:"widgets"`
	IsDefault   bool               `bson:"is_default" json:"is_default"`
	CreatedBy   int                `bson:"created_by" json:"created_by"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// DashboardWidget 仪表板组件
type DashboardWidget struct {
	ID         string                 `bson:"id" json:"id"`
	Type       string                 `bson:"type" json:"type"` // chart, metric, table
	Title      string                 `bson:"title" json:"title"`
	Position   WidgetPosition         `bson:"position" json:"position"`
	Size       WidgetSize             `bson:"size" json:"size"`
	Config     map[string]interface{} `bson:"config" json:"config"`
	DataSource string                 `bson:"data_source" json:"data_source"`
	Query      string                 `bson:"query" json:"query"`
}

// WidgetPosition 组件位置
type WidgetPosition struct {
	X int `bson:"x" json:"x"`
	Y int `bson:"y" json:"y"`
}

// WidgetSize 组件大小
type WidgetSize struct {
	Width  int `bson:"width" json:"width"`
	Height int `bson:"height" json:"height"`
}
