package database

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func DatabaseInstance() *mongo.Client {

	err := godotenv.Load()
	if err != nil {
		log.Println("warning: could not load the .env file")
	}
	log.Println("ENV TEST:", os.Getenv("MONGO_URI"))

	MongoDb := os.Getenv("MONGO_URI")
	if MongoDb == "" {
		log.Fatal("MONGO_URI not found in environment variables")
	}

	client, err := mongo.Connect(options.Client().ApplyURI(MongoDb))
	if err != nil {
		log.Fatal(err)
	}

	log.Println("MongoDB connected successfully")

	return client
}

var Client *mongo.Client = DatabaseInstance()

func OpenCollection(CollectionName string) *mongo.Collection {

	databaseName := os.Getenv("DATABASE_NAME")

	if databaseName == "" {
		log.Fatal("DATABASE_NAME not set in .env")
	}

	Collection := Client.Database(databaseName).Collection(CollectionName)

	return Collection
}
