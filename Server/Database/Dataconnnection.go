package database

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func DatabaseInstance() *mongo.Client {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("warning could not vload the .env file")
	}
	MongoDb := os.Getenv("MONGO_URI")
	if MongoDb == "" {
		log.Fatal("warning unable to load the Mongodb uri")
	}
	clientoptions := options.Client().ApplyURI(MongoDb)
	client, err := mongo.Connect(clientoptions)
	if err != nil {
		return nil
	}
	return client
}

var Client *mongo.Client = DatabaseInstance()

func OpenCollection(CollectionName string) *mongo.Collection {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("warning an error occured while loading the connection", err)

	}
	databaseName := os.Getenv("DATABASE_NAME")
	Collection := Client.Database(databaseName).Collection(CollectionName)
	if Collection == nil {
		return nil
	}
	return Collection

}
