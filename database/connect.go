package database

import (
	"context"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoURI = os.Getenv("MONGO_URI")

var DB *mongo.Database
var Context context.Context

func Connect() *mongo.Client {
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017/"
	}

	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI)) // open connection to mongodb database

	// error check
	if err != nil {
		log.Fatal(err)
	}

	// Get context with timeout after 8 seconds
	ctx := context.Background()

	// eror check
	err = client.Connect(ctx) // connect client
	if err != nil {
		log.Fatal(err)
	}

	// Set global variables for database and context
	DB = client.Database("RealTimeChatDB")
	Context = ctx

	return client
}
