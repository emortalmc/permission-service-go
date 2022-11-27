package mongo

import (
	"context"
	"flag"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
)

const database = "permission-service"

var (
	Client   = createMongoClient()
	Database = createMongoDatabase()
)

func createMongoClient() *mongo.Client {
	flag.Parse()
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(getMongoUri()))

	if err != nil {
		panic(err)
	}

	return client
}

func createMongoDatabase() *mongo.Database {
	return Client.Database(database)
}

func getMongoUri() string {
	uri, ok := os.LookupEnv("MONGO_URI")
	if ok {
		return uri
	}
	return "mongodb://localhost:27017"
}
