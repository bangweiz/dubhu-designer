package db

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// ConnectDB connects to MongoDB, verifies connectivity via Ping, and returns the Database handle.
// Returns an error if the connection or ping fails.
func ConnectDB() (*mongo.Database, error) {
	uri := "mongodb://localhost:27017/?directConnection=true"
	clientOptions := options.Client().ApplyURI(uri)

	client, err := mongo.Connect(clientOptions)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("could not ping MongoDB: %w", err)
	}

	fmt.Println("Successfully connected to MongoDB!")
	return client.Database("dubhu"), nil
}
