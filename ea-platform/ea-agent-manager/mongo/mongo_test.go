package mongo

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mockClient struct {
	db         map[string]*mongo.Database
	collections map[string]*mongo.Collection
}

func (m *mockClient) Database(name string) (*mongo.Database, error) {
	if db, ok := m.db[name]; ok {
		return db, nil
	}
	return nil, mongo.ErrNoDatabase()
}

func (m *mockClient) Collection(name string) (*mongo.Collection, error) {
	if coll, ok := m.collections[name]; ok {
		return coll, nil
	}
	return nil, mongo.ErrCollectionNamespaceNotFound()
}

func (m *mockClient) InsertOne(ctx context.Context, record interface{}) (*mongo.InsertOneResult, error) {
	return &mongo.InsertOneResult{
		InsertedIDs: []string{"ID"},
	}, nil
}

func (m *mockClient) Find(ctx context.Context, filter bson.M) *mongo.Cursor {
	cursor := &mongo.Cursor{
		First: func() bool { return true },
		Next: func() bool { return false },
	}
	m.collections["test"].FindOne(mongo.CursorIterator(cursor))
	return cursor
}

func (m *mockClient) Close(ctx context.Context) error {
	return nil
}

func TestMongoClient_NewMongoClient(t *testing.T) {
	client := &MongoClient{
		client: &mockClient{},
	}
	if err := client.NewMongoClient("mongodb://localhost:27017/"); err != nil {
		t.Errorf("NewMongoClient should not return an error")
	}
}

func TestMongoClient_InsertRecord(t *testing.T) {
	client := &MongoClient{
		client: &mockClient{},
	}
	collection, _ := client.client.Database("test").Collection("test")
	record := map[string]interface{}{"name": "John", "age": 25}
	result, err := client.InsertRecord("test", "test", record)
	if err != nil {
		t.Errorf("InsertRecord should not return an error")
	}
	if result.InsertedIDs[0] != "ID" {
		t.Errorf("InsertRecord should insert ID into test collection")
	}
}

func TestMongoClient_FindAllRecords(t *testing.T) {
	client := &MongoClient{
		client: &mockClient{},
	}
	collection, _ := client.client.Database("test").Collection("test")
	record := map[string]interface{}{"name": "John", "age": 25}
	m.client.db["test"] = collection
	result, err := client.FindAllRecords("test", "test")
	if err != nil {
		t.Errorf("FindAllRecords should not return an error")
	}
	if len(result) != 1 || result[0]["name"] != record["name"] {
		t.Errorf("FindAllRecords should find one record in test collection")
	}
}

func TestMongoClient_FindRecordByID(t *testing.T) {
	client := &MongoClient{
		client: &mockClient{},
	}
	collection, _ := client.client.Database("test").Collection("test")
	record := map[string]interface{}{"name": "John", "age": 25}
	m.client.db["test"] = collection
	result, err := client.FindRecordByID("test", "test", record["id"])
	if err != nil {
		t.Errorf("FindRecordByID should not return an error")
	}
	if result["name"] != record["name"] {
		t.Errorf("FindRecordByID should find one record with id in test collection")
	}
}

func TestMongoClient_FindRecordsWithProjection(t *testing.T) {
	client := &MongoClient{
		client: &mockClient{},
	}
	collection, _ := client.client.Database("test").Collection("test")
	record := map[string]interface{}{"name": "John", "age": 25}
	m.client.db["test"] = collection
	result, err := client.FindRecordsWithProjection("test", "test", record, bson.M{})
	if err != nil {
		t.Errorf("FindRecordsWithProjection should not return an error")
	}
	if len(result) != 1 || result[0]["name"] != record["name"] {
		t.Errorf("FindRecordsWithProjection should find one record in test collection with projection")
	}
}

func TestMongoClient_UpdateRecord(t *testing.T) {
	client := &MongoClient{
		client: &mockClient{},
	}
	collection, _ := client.client.Database("test").Collection("test")
	record := map[string]interface{}{"name": "John", "age": 25}
	m.client.db["test"] = collection
	result, err := client.UpdateRecord("test", "test", record, bson.M{})
	if err != nil {
		t.Errorf("UpdateRecord should not return an error")
	}
	if result.ModifiedCount != 1 {
		t.Errorf("UpdateRecord should update one record in test collection")
	}
}

func TestMongoClient_DeleteRecord(t *testing.T) {
	client := &MongoClient{
		client: &mockClient{},
	}
	collection, _ := client.client.Database("test").Collection("test")
	record := map[string]interface{}{"name": "John", "age": 25}
	m.client.db["test"] = collection
	result, err := client.DeleteRecord("test", "test", record)
	if err != nil {
		t.Errorf("DeleteRecord should not return an error")
	}
	if result.DeletedCount != 1 {
		t.Errorf("DeleteRecord should delete one record from test collection")
	}
}