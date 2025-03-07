package mongo

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"github.com/stretchr/testify/assert"
)

func TestNewMongoClient(t *testing.T) {
	mongoURI := "mongodb://localhost:27017/test"
	client, err := NewMongoClient(mongoURI)
	if assert.NoError(t, err) {
		assert.Equal(t, mongoURI, client.client.Options().URI())
	}
}

func TestDisconnect(t *testing.T) {
	mongoURI := "mongodb://localhost:27017/test"
	client, err := NewMongoClient(mongoURI)
	if assert.NoError(t, err) {
		err = client.Disconnect()
		assert.NoError(t, err)

		_, err = client.Disconnect()
		assert.Error(t, err)
	}
}

func TestInsertRecord(t *testing.T) {
	mongoURI := "mongodb://localhost:27017/test"
	client, err := NewMongoClient(mongoURI)
	if assert.NoError(t, err) {
		record := map[string]interface{}{
			"name":  "John Doe",
			"age":   30,
			"email": "john@example.com",
		}

		result, err := client.InsertRecord("test", "records", record)
		if assert.NoError(t, err) {
			assert.Equal(t, bson.M{"name": "John Doe", "age": 30, "email": "john@example.com"}, result.Document().Doc())
		}
	}
}

func TestFindAllRecords(t *testing.T) {
	mongoURI := "mongodb://localhost:27017/test"
	client, err := NewMongoClient(mongoURI)
	if assert.NoError(t, err) {
		record := map[string]interface{}{
			"name":  "John Doe",
			"age":   30,
			"email": "john@example.com",
		}

		_, err = client.InsertRecord("test", "records", record)
		assert.NoError(t, err)

		results, err := client.FindAllRecords("test", "records")
		if assert.NoError(t, err) {
			assert.Equal(t, map[string]interface{}{
				"name":  "John Doe",
				"age":   30,
				"email": "john@example.com",
			}, results[0])
		}
	}
}

func TestFindRecordByID(t *testing.T) {
	mongoURI := "mongodb://localhost:27017/test"
	client, err := NewMongoClient(mongoURI)
	if assert.NoError(t, err) {
		record := map[string]interface{}{
			"name":  "John Doe",
			"age":   30,
			"email": "john@example.com",
		}

		_, err = client.InsertRecord("test", "records", record)
		assert.NoError(t, err)

		result, err := client.FindRecordByID("test", "records", "1")
		if assert.NoError(t, err) {
			assert.Equal(t, map[string]interface{}{
				"name":  "John Doe",
				"age":   30,
				"email": "john@example.com",
			}, result)
		}
	}
}

func TestFindRecordsWithProjection(t *testing.T) {
	mongoURI := "mongodb://localhost:27017/test"
	client, err := NewMongoClient(mongoURI)
	if assert.NoError(t, err) {
		record := map[string]interface{}{
			"name":  "John Doe",
			"age":   30,
			"email": "john@example.com",
		}

		_, err = client.InsertRecord("test", "records", record)
		assert.NoError(t, err)

		result, err := client.FindRecordsWithProjection("test", "records", bson.M{"name": "John Doe"}, bson.M{"_id": 0})
		if assert.NoError(t, err) {
			assert.Equal(t, map[string]interface{}{
				"name":  "John Doe",
				"age":   30,
				"email": "john@example.com",
			}, result[0])
		}
	}
}

func TestUpdateRecord(t *testing.T) {
	mongoURI := "mongodb://localhost:27017/test"
	client, err := NewMongoClient(mongoURI)
	if assert.NoError(t, err) {
		record := map[string]interface{}{
			"name":  "John Doe",
			"age":   30,
			"email": "john@example.com",
		}

		_, err = client.InsertRecord("test", "records", record)
		assert.NoError(t, err)

		update := bson.M{"$set": map[string]interface{}{
			"name": "Jane Doe",
			"age":  31,
		}}

		result, err := client.UpdateRecord("test", "records", bson.M{"name": "John Doe"}, update)
		if assert.NoError(t, err) {
			assert.Equal(t, bson.M{"name": "Jane Doe", "age": 31}, result.modifiedCount)
		}
	}
}