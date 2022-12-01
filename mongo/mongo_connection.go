package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"permission-service-go/mongo/registry"
)

const database = "permission-service"

var (
	Client   = createMongoClient()
	Database = createMongoDatabase()
)

func createMongoClient() *mongo.Client {
	mongoOptions := options.Client().ApplyURI(getMongoUri()).SetRegistry(registry.MongoRegistry)
	client, err := mongo.Connect(context.Background(), mongoOptions)

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
