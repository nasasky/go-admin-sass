package mongodb

import (
	"context"
	"log"
	"nasa-go-admin/pkg/config"
	"time"

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
}

func GetCollection(dbName, collectionKey string) *mongo.Collection {
	cfg := config.GetConfig()

	dbConfig, exists := cfg.MongoDB.Databases[dbName]
	if !exists {
		log.Fatalf("Database %s not found in config", dbName)
	}
	collectionName, exists := dbConfig.Collections[collectionKey]
	if !exists {
		log.Fatalf("Collection %s not found in database %s config", collectionKey, dbName)
	}
	client, exists := clients[dbName]
	if !exists {
		log.Fatalf("MongoDB client for database %s not initialized", dbName)
	}
	return client.Database(dbName).Collection(collectionName)
}
