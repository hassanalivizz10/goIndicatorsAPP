package config
/*
Config will contain DB connect and Retrieve the DB instance Methods.
*/

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"indicatorsAPP/utils"
	
)

var mongoClient *mongo.Client
var mongoDB *mongo.Database

func ConnectToMongoDB() error {
	mongoURI := utils.GetEnv("MONGO_URI")
	//fmt.Println("mongoURI",mongoURI)
	if mongoURI == "" {
		return fmt.Errorf("MONGO_URI environment variable is not set")
	}

	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	mongoClient = client
	mongoDB = client.Database(utils.GetEnv("MONGO_DB"))

	return nil
}

func GetMongoDB() *mongo.Database {
	return mongoDB
}
