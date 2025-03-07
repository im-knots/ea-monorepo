package mongo

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"github.com/stretchr/testify/assert"
)

func TestNewMongoClient(t *testing.T) {
	mongo_uri := "mongodb://localhost:27017/"
	expected_client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongo_uri))
	assert.NoError(t, err)
	actual_client, err := NewMongoClient(mongo_uri)
	assert.NotNil(t, actual_client)
	assert.Equal(t, expected_client, actual_client.client)
}

func TestNewMongoClient_timeout(t *testing.T) {
	mongo_uri := "mongodb://localhost:27017/"
	expected_err := mongo.ConnectError
	err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongo_uri))
	assert.Error(t, err)
	actual_client, err := NewMongoClient(mongo_uri)
	assert.Nil(t, actual_client)
}

func TestDisconnect(t *testing.T) {
	mongo_uri := "mongodb://localhost:27017/"
	mongo_client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongo_uri))
	assert.NoError(t, err)

	disconnect_err := mongo.DisconnectError
	actual_err := mongo_client.Disconnect()
	assert.Error(t, actual_err)
	assert.Equal(t, disconnect_err, actual_err)
}

func TestUpdateRecord_success(t *testing.T) {
	mongo_uri := "mongodb://localhost:27017/"
	mongo_client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongo_uri))
	assert.NoError(t, err)

	db, collection := "testDB", "testCollection"
 filter := bson.M{"key": "value"}
	update = bson.M{"$set": bson.M{"newKey": "newValue"}}
	expected_result := &mongo.UpdateResult{ModifiedCount: 1}
	actual_result, actual_err := mongo_client.UpdateRecord(db, collection, filter, update)
	assert.NoError(t, actual_err)
	assert.Equal(t, expected_result, actual_result)
}

func TestUpdateRecord_failure(t *testing.T) {
	mongo_uri := "mongodb://localhost:27017/"
	mongo_client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongo_uri))
	assert.NoError(t, err)

	db, collection := "testDB", "testCollection"
	filter := bson.M{"key": "value"}
	update = bson.M{"$inc": bson.M{"newKey": -1}}
	expected_err := mongo.UpdateError
	actual_result, actual_err := mongo_client.UpdateRecord(db, collection, filter, update)
	assert.Error(t, actual_err)
	assert.Equal(t, expected_err, actual_err)
}

func TestUpdateRecord_filter_not_found(t *testing.T) {
	mongo_uri := "mongodb://localhost:27017/"
	mongo_client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongo_uri))
	assert.NoError(t, err)

	db, collection := "testDB", "testCollection"
	filter := bson.M{"key": "value"}
	update = bson.M{"$set": bson.M{"newKey": "newValue"}}
	expected_err := mongo.UpdateError
	actual_result, actual_err := mongo_client.UpdateRecord(db, collection, filter, update)
	assert.Error(t, actual_err)
	assert.Equal(t, expected_err, actual_err)
}