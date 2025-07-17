package mongodb

import (
	"context"
	"log"
	"nasa-go-admin/pkg/config"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var clients = make(map[string]*mongo.Client)

func InitMongoDB() {
	cfg := config.GetConfig()

	for dbName, dbConfig := range cfg.MongoDB.Databases {
		client, err := mongo.NewClient(options.Client().ApplyURI(dbConfig.URI))
		if err != nil {
			log.Fatalf("Failed to create MongoDB client for %s: %v", dbName, err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err = client.Connect(ctx)
		if err != nil {
			log.Fatalf("Failed to connect to MongoDB for %s: %v", dbName, err)
		}
		clients[dbName] = client
	}
	log.Printf("MongoDB连接已初始化，共 %d 个数据库", len(cfg.MongoDB.Databases))

	// 自动初始化集合和索引
	if err := AutoEnsureCollectionsAndIndexes(); err != nil {
		log.Printf("⚠️ MongoDB集合自动初始化失败: %v", err)
		log.Printf("   请手动运行: mongosh < scripts/setup_mongodb_collections.js")
	} else {
		log.Printf("✅ MongoDB集合和索引自动初始化成功")
	}
}

func GetCollection(dbName, collectionKey string) *mongo.Collection {
	cfg := config.GetConfig()

	dbConfig, exists := cfg.MongoDB.Databases[dbName]
	if !exists {
		log.Printf("⚠️ 数据库配置 %s 不存在，请检查配置文件", dbName)
		// 返回nil而不是直接崩溃，让调用方处理
		return nil
	}
	collectionName, exists := dbConfig.Collections[collectionKey]
	if !exists {
		log.Printf("⚠️ 集合配置 %s 不存在于数据库 %s 中", collectionKey, dbName)
		return nil
	}
	client, exists := clients[dbName]
	if !exists {
		log.Printf("⚠️ MongoDB客户端 %s 未初始化，请检查MongoDB连接", dbName)
		return nil
	}
	return client.Database(dbName).Collection(collectionName)
}

// CollectionIndexInfo 集合索引信息
type CollectionIndexInfo struct {
	DatabaseKey   string
	CollectionKey string
	Indexes       []IndexInfo
}

// IndexInfo 索引信息
type IndexInfo struct {
	Keys   bson.D
	Unique bool
	Name   string
}

// 定义需要确保存在的集合和索引
var requiredCollections = []CollectionIndexInfo{
	{
		DatabaseKey:   "notification_log_db",
		CollectionKey: "push_records",
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
		DatabaseKey:   "notification_log_db",
		CollectionKey: "notification_logs",
		Indexes: []IndexInfo{
			{Keys: bson.D{{"message_id", 1}}, Unique: false, Name: "message_id_idx"},
			{Keys: bson.D{{"timestamp", -1}}, Unique: false, Name: "timestamp_desc"},
			{Keys: bson.D{{"event_type", 1}}, Unique: false, Name: "event_type_idx"},
			{Keys: bson.D{{"user_id", 1}}, Unique: false, Name: "user_id_idx"},
		},
	},
	{
		DatabaseKey:   "notification_log_db",
		CollectionKey: "admin_user_receive_records",
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
		DatabaseKey:   "notification_log_db",
		CollectionKey: "admin_user_online_status",
		Indexes: []IndexInfo{
			{Keys: bson.D{{"user_id", 1}}, Unique: true, Name: "user_id_unique"},
			{Keys: bson.D{{"is_online", 1}}, Unique: false, Name: "is_online_idx"},
			{Keys: bson.D{{"last_seen", -1}}, Unique: false, Name: "last_seen_desc"},
			{Keys: bson.D{{"username", 1}}, Unique: false, Name: "username_idx"},
		},
	},
}

// AutoEnsureCollectionsAndIndexes 自动确保集合和索引存在
func AutoEnsureCollectionsAndIndexes() error {
	cfg := config.GetConfig()

	// 用于跟踪已处理的数据库，避免重复警告
	processedDBs := make(map[string]bool)
	skippedDBs := make(map[string]bool)

	for _, collInfo := range requiredCollections {
		// 获取数据库配置
		dbConfig, exists := cfg.MongoDB.Databases[collInfo.DatabaseKey]
		if !exists {
			// 只在第一次遇到时显示警告
			if !skippedDBs[collInfo.DatabaseKey] {
				log.Printf("⚠️ 数据库配置 %s 不存在，跳过", collInfo.DatabaseKey)
				skippedDBs[collInfo.DatabaseKey] = true
			}
			continue
		}

		// 获取集合名称
		collectionName, exists := dbConfig.Collections[collInfo.CollectionKey]
		if !exists {
			log.Printf("⚠️ 集合配置 %s.%s 不存在，跳过", collInfo.DatabaseKey, collInfo.CollectionKey)
			continue
		}

		// 获取客户端
		client, exists := clients[collInfo.DatabaseKey]
		if !exists {
			log.Printf("⚠️ 数据库客户端 %s 不存在，跳过", collInfo.DatabaseKey)
			continue
		}

		database := client.Database(collInfo.DatabaseKey)
		collection := database.Collection(collectionName)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// 检查集合是否存在
		exists, err := collectionExists(database, collectionName)
		if err != nil {
			log.Printf("⚠️ 检查集合 %s.%s 是否存在失败: %v", collInfo.DatabaseKey, collectionName, err)
			continue
		}

		if !exists {
			// 自动创建集合
			log.Printf("📦 集合 %s.%s 不存在，正在自动创建...", collInfo.DatabaseKey, collectionName)
			if err := createCollection(database, collectionName); err != nil {
				log.Printf("❌ 创建集合 %s.%s 失败: %v", collInfo.DatabaseKey, collectionName, err)
				continue
			}
			log.Printf("✅ 集合 %s.%s 创建成功", collInfo.DatabaseKey, collectionName)
		} else {
			log.Printf("📋 集合 %s.%s 已存在", collInfo.DatabaseKey, collectionName)
		}

		// 创建索引
		successCount := 0
		for _, indexInfo := range collInfo.Indexes {
			indexModel := mongo.IndexModel{
				Keys:    indexInfo.Keys,
				Options: options.Index().SetUnique(indexInfo.Unique).SetName(indexInfo.Name),
			}

			_, err := collection.Indexes().CreateOne(ctx, indexModel)
			if err != nil {
				// 如果索引已存在或其他可忽略的错误，继续
				if mongo.IsDuplicateKeyError(err) ||
					containsString(err.Error(), "already exists") ||
					containsString(err.Error(), "IndexKeySpecsConflict") {
					log.Printf("📋 索引 %s.%s.%s 已存在", collInfo.DatabaseKey, collectionName, indexInfo.Name)
					successCount++
				} else {
					log.Printf("❌ 创建索引 %s.%s.%s 失败: %v", collInfo.DatabaseKey, collectionName, indexInfo.Name, err)
				}
			} else {
				log.Printf("✅ 创建索引 %s.%s.%s 成功", collInfo.DatabaseKey, collectionName, indexInfo.Name)
				successCount++
			}
		}

		log.Printf("📊 集合 %s.%s: 成功处理 %d/%d 个索引",
			collInfo.DatabaseKey, collectionName, successCount, len(collInfo.Indexes))

		// 标记数据库已处理
		processedDBs[collInfo.DatabaseKey] = true
	}

	// 显示处理总结
	if len(processedDBs) > 0 {
		log.Printf("✅ MongoDB集合和索引自动初始化成功")
	}

	return nil
}

// createCollection 创建集合
func createCollection(database *mongo.Database, collectionName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 创建一个空的文档来触发集合创建
	_, err := database.Collection(collectionName).InsertOne(ctx, bson.M{"_created": time.Now()})
	if err != nil {
		return err
	}

	// 删除创建的测试文档
	_, err = database.Collection(collectionName).DeleteOne(ctx, bson.M{"_created": bson.M{"$exists": true}})
	return err
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
	if len(str) < len(substr) {
		return false
	}
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
