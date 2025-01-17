package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoClient wraps the MongoDB client.
type MongoClient struct {
	client *mongo.Client
}

// NewMongoClient initializes a new MongoDB client.
func NewMongoClient(uri string) (*MongoClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	// Verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return &MongoClient{client: client}, nil
}

// Disconnect closes the MongoDB client connection.
func (m *MongoClient) Disconnect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return m.client.Disconnect(ctx)
}

// InsertRecord inserts a record into a specified collection.
func (m *MongoClient) InsertRecord(database, collection string, record interface{}) (*mongo.InsertOneResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	coll := m.client.Database(database).Collection(collection)
	return coll.InsertOne(ctx, record)
}
