package main

import (
	"context"
	"fmt"
	"log"
	"nasa-go-admin/mongodb"
	"nasa-go-admin/pkg/config"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB集合自动初始化工具
// 可以在应用启动时调用，确保所需的集合和索引都已创建

// CollectionInfo 集合信息结构
type CollectionInfo struct {
	Database   string
	Collection string
	Indexes    []IndexInfo
}

// IndexInfo 索引信息结构
type IndexInfo struct {
	Keys   bson.D
	Unique bool
	Name   string
}

// 定义需要初始化的集合和索引
var requiredCollections = []CollectionInfo{
	{
		Database:   "notification_log_db",
		Collection: "push_records",
		Indexes: []IndexInfo{
			{Keys: bson.D{{"message_id", 1}}, Unique: true, Name: "message_id_unique"},
			{Keys: bson.D{{"push_time", -1}}, Unique: false, Name: "push_time_desc"},
			{Keys: bson.D{{"message_type", 1}}, Unique: false, Name: "message_type_idx"},
			{Keys: bson.D{{"target", 1}}, Unique: false, Name: "target_idx"},
			{Keys: bson.D{{"success", 1}}, Unique: false, Name: "success_idx"},
			{Keys: bson.D{{"sender_id", 1}}, Unique: false, Name: "sender_id_idx"},
			{Keys: bson.D{{"status", 1}}, Unique: false, Name: "status_idx"},
		},
	},
	{
		Database:   "notification_log_db",
		Collection: "notification_logs",
		Indexes: []IndexInfo{
			{Keys: bson.D{{"message_id", 1}}, Unique: false, Name: "message_id_idx"},
			{Keys: bson.D{{"timestamp", -1}}, Unique: false, Name: "timestamp_desc"},
			{Keys: bson.D{{"event_type", 1}}, Unique: false, Name: "event_type_idx"},
			{Keys: bson.D{{"user_id", 1}}, Unique: false, Name: "user_id_idx"},
		},
	},
	{
		Database:   "notification_log_db",
		Collection: "admin_user_receive_records",
		Indexes: []IndexInfo{
			{Keys: bson.D{{"message_id", 1}}, Unique: false, Name: "message_id_idx"},
			{Keys: bson.D{{"user_id", 1}}, Unique: false, Name: "user_id_idx"},
			{Keys: bson.D{{"created_at", -1}}, Unique: false, Name: "created_at_desc"},
			{Keys: bson.D{{"message_id", 1}, {"user_id", 1}}, Unique: true, Name: "message_user_unique"},
			{Keys: bson.D{{"is_received", 1}}, Unique: false, Name: "is_received_idx"},
			{Keys: bson.D{{"is_read", 1}}, Unique: false, Name: "is_read_idx"},
			{Keys: bson.D{{"is_confirmed", 1}}, Unique: false, Name: "is_confirmed_idx"},
			{Keys: bson.D{{"delivery_status", 1}}, Unique: false, Name: "delivery_status_idx"},
			{Keys: bson.D{{"push_channel", 1}}, Unique: false, Name: "push_channel_idx"},
			{Keys: bson.D{{"username", 1}}, Unique: false, Name: "username_idx"},
		},
	},
	{
		Database:   "notification_log_db",
		Collection: "admin_user_online_status",
		Indexes: []IndexInfo{
			{Keys: bson.D{{"user_id", 1}}, Unique: true, Name: "user_id_unique"},
			{Keys: bson.D{{"is_online", 1}}, Unique: false, Name: "is_online_idx"},
			{Keys: bson.D{{"last_seen", -1}}, Unique: false, Name: "last_seen_desc"},
			{Keys: bson.D{{"username", 1}}, Unique: false, Name: "username_idx"},
		},
	},
}

// AutoInitMongoDB 自动初始化MongoDB集合和索引
func AutoInitMongoDB() error {
	log.Printf("🔧 开始自动初始化MongoDB集合和索引...")

	// 初始化配置
	if err := config.InitConfig(); err != nil {
		return fmt.Errorf("初始化配置失败: %v", err)
	}

	// 初始化MongoDB连接
	mongodb.InitMongoDB()

	var totalCollections, totalIndexes int

	for _, collInfo := range requiredCollections {
		log.Printf("📦 处理数据库: %s, 集合: %s", collInfo.Database, collInfo.Collection)

		// 获取集合
		collection := mongodb.GetCollection(collInfo.Database, collInfo.Collection)
		if collection == nil {
			log.Printf("❌ 无法获取集合: %s.%s", collInfo.Database, collInfo.Collection)
			continue
		}

		// 创建集合 (如果不存在，MongoDB会自动创建)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// 检查集合是否存在
		exists, err := collectionExists(collection.Database(), collInfo.Collection)
		if err != nil {
			log.Printf("❌ 检查集合是否存在失败: %v", err)
			continue
		}

		if !exists {
			log.Printf("✅ 集合 %s.%s 不存在，将在第一次写入时自动创建", collInfo.Database, collInfo.Collection)
		} else {
			log.Printf("✅ 集合 %s.%s 已存在", collInfo.Database, collInfo.Collection)
		}
		totalCollections++

		// 创建索引
		for _, indexInfo := range collInfo.Indexes {
			indexModel := mongo.IndexModel{
				Keys:    indexInfo.Keys,
				Options: options.Index().SetUnique(indexInfo.Unique).SetName(indexInfo.Name),
			}

			_, err := collection.Indexes().CreateOne(ctx, indexModel)
			if err != nil {
				// 如果索引已存在，跳过错误
				if mongo.IsDuplicateKeyError(err) ||
					containsString(err.Error(), "already exists") ||
					containsString(err.Error(), "IndexKeySpecsConflict") {
					log.Printf("⚠️ 索引 %s 已存在，跳过创建", indexInfo.Name)
				} else {
					log.Printf("❌ 创建索引 %s 失败: %v", indexInfo.Name, err)
					continue
				}
			} else {
				log.Printf("✅ 创建索引 %s 成功", indexInfo.Name)
			}
			totalIndexes++
		}
	}

	log.Printf("🎉 MongoDB自动初始化完成! 处理了 %d 个集合，%d 个索引", totalCollections, totalIndexes)
	return nil
}

// collectionExists 检查集合是否存在
func collectionExists(database *mongo.Database, collectionName string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collections, err := database.ListCollectionNames(ctx, bson.M{"name": collectionName})
	if err != nil {
		return false, err
	}

	return len(collections) > 0, nil
}

// containsString 检查字符串是否包含子字符串
func containsString(str, substr string) bool {
	return len(str) >= len(substr) &&
		(str == substr ||
			(len(str) > len(substr) &&
				(str[:len(substr)] == substr ||
					str[len(str)-len(substr):] == substr ||
					findSubstring(str, substr))))
}

// findSubstring 查找子字符串
func findSubstring(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func init() {
	// 可以在需要时调用 AutoInitMongoDB()
}
