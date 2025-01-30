package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

// FindAllRecords retrieves all records from a collection.
func (m *MongoClient) FindAllRecords(database, collection string) ([]map[string]interface{}, error) {
	coll := m.client.Database(database).Collection(collection)
	cursor, err := coll.Find(context.TODO(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var results []map[string]interface{}
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}
	return results, nil
}

// FindRecordByID retrieves a single record by ID from a collection.
func (m *MongoClient) FindRecordByID(database, collection, id string) (map[string]interface{}, error) {
	coll := m.client.Database(database).Collection(collection)
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err = coll.FindOne(context.TODO(), bson.M{"_id": objectID}).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// FindRecordsWithProjection retrieves records from a collection with a specified projection.
func (m *MongoClient) FindRecordsWithProjection(database, collection string, filter, projection interface{}) ([]map[string]interface{}, error) {
	coll := m.client.Database(database).Collection(collection)
	opts := options.Find().SetProjection(projection)
	cursor, err := coll.Find(context.TODO(), filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var results []map[string]interface{}
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}
	return results, nil
}

// UpdateRecord updates a record in the specified collection using a filter.
func (m *MongoClient) UpdateRecord(database, collection string, filter, update interface{}) (*mongo.UpdateResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	coll := m.client.Database(database).Collection(collection)
	result, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	return result, nil
}
