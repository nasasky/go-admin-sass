package mongodb

import (
	"context"
	"log"
	"nasa-go-admin/config"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var clients = make(map[string]*mongo.Client)

func InitMongoDB() {
	for dbName, dbConfig := range config.AppConfig.MongoDB.Databases {
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
}

func GetCollection(dbName, collectionKey string) *mongo.Collection {
	dbConfig, exists := config.AppConfig.MongoDB.Databases[dbName]
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
