package mongo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMongoClient(t *testing.T) {
	client, err := NewMongoClient("mongodb://localhost:27017/")
	assert.NoError(t, err)
	assert.NotNil(t, client.client)
}

func TestDisconnect(t *testing.T) {
	m := &MongoClient{client: mongo.NewClient(mongo.ClientOptions{})}
	err := m.Disconnect()
	assert.Error(t, err)
}

func TestInsertRecordSuccess(t *testing.T) {
	collectionName := "test_collection"
	client, err := NewMongoClient("mongodb://localhost:27017/")
	if err != nil {
		t.Fatal(err)
	}
	defer client.Disconnect()

	record := map[string]interface{}{
		"name": "John Doe",
		"age":  30,
	}

	res, err := client.InsertRecord("test_database", collectionName, record)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	client.Disconnect()
}

func TestInsertRecordFailure(t *testing.T) {
	collectionName := "test_collection"
	m := &MongoClient{client: mongo.NewClient(mongo.ClientOptions{})}
	err := m.InsertRecord("test_database", collectionName, nil)
	assert.Error(t, err)
}

func TestFindAllRecordsSuccess(t *testing.T) {
	client, err := NewMongoClient("mongodb://localhost:27017/")
	if err != nil {
		t.Fatal(err)
	}
	defer client.Disconnect()

	collectionName := "test_collection"
	record := map[string]interface{}{
		"name": "John Doe",
		"age":  30,
	}

	err = client.InsertRecord("test_database", collectionName, record)
	assert.NoError(t, err)

	res, err := client.FindAllRecords("test_database", collectionName)
	assert.NoError(t, err)
	assert.Len(t, res, 1)

	client.Disconnect()
}

func TestFindAllRecordsFailure(t *testing.T) {
	collectionName := "test_collection"
	m := &MongoClient{client: mongo.NewClient(mongo.ClientOptions{})}
	err := m.FindAllRecords("test_database", collectionName)
	assert.Error(t, err)
}

func TestFindRecordByIDSuccess(t *testing.T) {
	client, err := NewMongoClient("mongodb://localhost:27017/")
	if err != nil {
		t.Fatal(err)
	}
	defer client.Disconnect()

	collectionName := "test_collection"
	record := map[string]interface{}{
		"name": "John Doe",
		"age":  30,
	}

	err = client.InsertRecord("test_database", collectionName, record)
	assert.NoError(t, err)

	res, err := client.FindRecordByID("test_database", collectionName, "test_id")
	assert.NoError(t, err)
	assert.NotNil(t, res)

	client.Disconnect()
}

func TestFindRecordByIDFailure(t *testing.T) {
	collectionName := "test_collection"
	m := &MongoClient{client: mongo.NewClient(mongo.ClientOptions{})}
	err := m.FindRecordByID("test_database", collectionName, "non_existent_id")
	assert.Error(t, err)
}

func TestFindRecordsWithProjectionSuccess(t *testing.T) {
	client, err := NewMongoClient("mongodb://localhost:27017/")
	if err != nil {
		t.Fatal(err)
	}
	defer client.Disconnect()

	collectionName := "test_collection"
	record := map[string]interface{}{
		"name": "John Doe",
		"age":  30,
	}

	err = client.InsertRecord("test_database", collectionName, record)
	assert.NoError(t, err)

	res, err := client.FindRecordsWithProjection("test_database", collectionName, bson.M{"id": "test_id"}, bson.M{"name": 1})
	assert.NoError(t, err)
	assert.Len(t, res, 1)

	client.Disconnect()
}

func TestFindRecordsWithProjectionFailure(t *testing.T) {
	collectionName := "test_collection"
	m := &MongoClient{client: mongo.NewClient(mongo.ClientOptions{})}
	err := m.FindRecordsWithProjection("test_database", collectionName, bson.M{"id": "non_existent_id"}, nil)
	assert.Error(t, err)
}

func TestUpdateRecordSuccess(t *testing.T) {
	client, err := NewMongoClient("mongodb://localhost:27017/")
	if err != nil {
		t.Fatal(err)
	}
	defer client.Disconnect()

	collectionName := "test_collection"
	record := map[string]interface{}{
		"name": "John Doe",
		"age":  30,
	}

	err = client.InsertRecord("test_database", collectionName, record)
	assert.NoError(t, err)

	res, err := client.UpdateRecord("test_database", collectionName, bson.M{"name": ""}, bson.M{"$set": map[string]interface{}{"age": 31}})
	assert.NoError(t, err)
	assert.NotNil(t, res)

	client.Disconnect()
}

func TestUpdateRecordFailure(t *testing.T) {
	collectionName := "test_collection"
	m := &MongoClient{client: mongo.NewClient(mongo.ClientOptions{})}
	err := m.UpdateRecord("test_database", collectionName, bson.M{"name": ""}, nil)
	assert.Error(t, err)
}