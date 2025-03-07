package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"

	"github.com/stretchr/testify/assert"
)

func MockMongoClient() *MongoClient {
	return &MongoClient{
		client: &mockClient,
	}
}

type mockClient struct {
	db   mongo.Database
	coll mongo.Collection
	err  error
}

func (mc *mockClient) Database(name string) mongo.Database {
	return &mockDB{
		Name: name,
		Collection: func() (*mockCollection, error) {
			return nil, nil
		},
	}
}

type mockDB struct {
	Name      string
	Collection func() (*mockCollection, error)
	err        error
}

func (db *mockDB) Collection(name string) mongo.Collection {
	if db.Collection == nil {
		db.Collection = func() (*mockCollection, error) {
			return &mockCollection{
				Name:  name,
				Data:  []byte("mock data"),
				Err:   db.err,
				Find: func(filter interface{}) (*mongo.Cursor, error) {
					return &mockCursor{
						Filter: filter,
						Data:    []byte("mock cursor data"),
					}, nil
				},
			}, nil
		}
	}

	return db.Collection(name)
}

type mockCollection struct {
	Name  string
	Data  []byte
Err   error
Find  func(filter interface{}) (*mongo.Cursor, error)
}

func (col *mockCollection) InsertOne(ctx context.Context, record interface{}) (*mongo.InsertOneResult, error) {
	if col.Find == nil {
		return &mockInsertOneResult{
			ID: "mock id",
			Data: map[string]interface{}{
				"name": "mock name",
				"date": time.Now(),
			},
			Err: col.Err,
		}, nil
	}

 filter := bson.M{"name": "mock name"}
	projection := options.FindOne().SetProjection(bson.M{"_id": 0})
	cursor, err := col.Find(ctx, filter, projection)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var result map[string]interface{}
	if err = cursor.All(ctx, &result); err != nil {
		return nil, err
	}

	return &mockInsertOneResult{
		ID: "mock id",
		Data: result,
		Err: col.Err,
	}, nil
}

type mockCursor struct {
	Filter interface{}
	Data    []byte
}

func (c *mockCursor) Close(ctx context.Context) error {
	return c.Err
}

type mockInsertOneResult struct {
	ID   string
	Data map[string]interface{}
	Err  error
}

func TestMongoClient_NewMongoClient(t *testing.T) {
	mongoClient := MockMongoClient()
	assert.NoError(t, mongoClient(clientURI))
}

func TestMongoClient Disconnect(t *testing.T) {
	mongoClient := MockMongoClient()
	mongoClient.client = &mockClient{err: errors.New("test error")}
	assert.Error(t, mongoClient.Disconnect())
}

func TestMongoClient_InsertRecord(t *testing.T) {
	mongoClient := MockMongoClient()
	mongoClient.client = &mockClient{
		db: &mockDB{
			Collection: func() (*mockCollection, error) {
				return &mockCollection{}, nil
			},
			err: errors.New("test error"),
		}}
	assert.NoError(t, mongoClient.InsertRecord("test", "test", map[string]interface{}{}))
}

func TestMongoClient_FindAllRecords(t *testing.T) {
	mongoClient := MockMongoClient()
	mongoClient.client = &mockClient{
		db: &mockDB{
			Collection: func() (*mockCollection, error) {
				return &mockCollection{
					Data: []byte("mock data"),
				}, nil
			},
			err: errors.New("test error"),
		}}
	result := make(map[string]interface{})
	assert.NoError(t, mongoClient.FindAllRecords("test", "test"))
	assert.Equal(t, result, map[string]interface{}{"name": "mock name"})
}

func TestMongoClient_FindRecordByID(t *testing.T) {
	mongoClient := MockMongoClient()
	mongoClient.client = &mockClient{
		db: &mockDB{
			Collection: func() (*mockCollection, error) {
				return &mockCollection{
					Data: []byte("mock data"),
				}, nil
			},
			err: errors.New("test error"),
		}}
	result := make(map[string]interface{})
	assert.NoError(t, mongoClient.FindRecordByID("test", "test", "mock id"))
	assert.Equal(t, result, map[string]interface{}{"name": "mock name"})
}

func TestMongoClient_FindRecordsWithProjection(t *testing.T) {
	mongoClient := MockMongoClient()
	mongoClient.client = &mockClient{
		db: &mockDB{
			Collection: func() (*mockCollection, error) {
				return &mockCollection{
					Data: []byte("mock data"),
				}, nil
			},
			err: errors.New("test error"),
		}}
	result := make(map[string]interface{})
	assert.NoError(t, mongoClient.FindRecordsWithProjection("test", "test", bson.M{"name": "mock name"}, map[string]interface{}{}))
	assert.Equal(t, result, map[string]interface{}{"name": "mock name"})
}

func TestMongoClient_UpdateRecord(t *testing.T) {
	mongoClient := MockMongoClient()
	mongoClient.client = &mockClient{
		db: &mockDB{
			Collection: func() (*mockCollection, error) {
				return &mockCollection{
					Data: []byte("mock data"),
				}, nil
			},
			err: errors.New("test error"),
		}}
	result := make(map[string]interface{})
	assert.NoError(t, mongoClient.UpdateRecord("test", "test", bson.M{"name": "mock name"}, map[string]interface{}{}))
	assert.Equal(t, result, map[string]interface{}{"name": "mock updated name"})
}

func clientURI() string {
	return "mongodb://localhost:27017"
}