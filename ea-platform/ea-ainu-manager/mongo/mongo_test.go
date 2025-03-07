package mongo

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestMongoClient_NewMongoClient(t *testing.T) {
	mclient, err := NewMongoClient("mongodb://localhost:27017/")
	if err != nil {
		t.Fatal(err)
	}

	if mclient.client == nil {
		t.Fatal("expected client to be non-nil")
	}
}

func TestMongoClient_Disconnect(t *testing.T) {
	mclient, err := NewMongoClient("mongodb://localhost:27017/")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err = mclient.Disconnect()
		if err != nil {
			t.Fatal(err)
		}
	}()

	if err := mclient.Disconnect(); err == nil {
		t.Fatal("expected error")
	}
}

func TestMongoClient_InsertRecord(t *testing.T) {
	mclient, err := NewMongoClient("mongodb://localhost:27017/")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err = mclient.Disconnect()
		if err != nil {
			t.Fatal(err)
		}
	}()

	_, err = mclient.InsertRecord("testdb", "testcollection", map[string]interface{}{
		"key": "value",
	})
	if err != nil {
		t.Fatal(err)
	}

	collection := mclient.client.Database("testdb").Collection("testcollection")
	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		t.Fatal(err)
	}
	defer cursor.Close(context.TODO())

	var record map[string]interface{}
	if err = cursor.All(context.TODO(), &record); err != nil {
		t.Fatal(err)
	}

	if record["key"] != "value" {
		t.Fatal("expected record to match")
	}
}

func TestMongoClient_FindAllRecords(t *testing.T) {
	mclient, err := NewMongoClient("mongodb://localhost:27017/")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err = mclient.Disconnect()
		if err != nil {
			t.Fatal(err)
		}
	}()

	records, err := mclient.FindAllRecords("testdb", "testcollection")
	if err != nil {
		t.Fatal(err)
	}

	collection := mclient.client.Database("testdb").Collection("testcollection")
	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		t.Fatal(err)
	}
	defer cursor.Close(context.TODO())

	var record map[string]interface{}
	for cursor.Next(context.TODO()) {
		err = cursor.Decode(&record)
		if err != nil {
			t.Fatal(err)
		}

		if record["key"] != "value" {
			t.Fatal("expected record to match")
		}
	}

	if len(records) != 1 {
		t.Fatal("expected one record, got", len(records))
	}
}

func TestMongoClient_FindRecordByID(t *testing.T) {
	mclient, err := NewMongoClient("mongodb://localhost:27017/")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err = mclient.Disconnect()
		if err != nil {
			t.Fatal(err)
		}
	}()

	_, err = mclient.InsertRecord("testdb", "testcollection", map[string]interface{}{
		"key":    "value",
		"id":     "1234",
	})
	if err != nil {
		t.Fatal(err)
	}

	record, err := mclient.FindRecordByID("testdb", "testcollection", "1234")
	if err != nil {
		t.Fatal(err)
	}

	if record["key"] != "value" {
		t.Fatal("expected record to match")
	}
}

func TestMongoClient_FindRecordsWithProjection(t *testing.T) {
	mclient, err := NewMongoClient("mongodb://localhost:27017/")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err = mclient.Disconnect()
		if err != nil {
			t.Fatal(err)
		}
	}()

	_, err = mclient.InsertRecord("testdb", "testcollection", map[string]interface{}{
		"key": "value",
	})
	if err != nil {
		t.Fatal(err)
	}

	record, err := mclient.FindRecordsWithProjection("testdb", "testcollection", bson.M{"id": "1234"}, bson.M{"key": 1})
	if err != nil {
		t.Fatal(err)
	}

	if record[0]["key"] != "value" {
		t.Fatal("expected record to match")
	}
}

func TestMongoClient_UpdateRecord(t *testing.T) {
	mclient, err := NewMongoClient("mongodb://localhost:27017/")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err = mclient.Disconnect()
		if err != nil {
			t.Fatal(err)
		}
	}()

	_, err = mclient.InsertRecord("testdb", "testcollection", map[string]interface{}{
		"key": "value",
	})
	if err != nil {
		t.Fatal(err)
	}

	record, err := mclient.FindRecordByID("testdb", "testcollection", "1234")
	if err == nil && record["key"] != "value" {
		t.Fatal("expected record to match")
	}
	if err != nil {
		t.Fatal(err)
	}

	mclient.UpdateRecord("testdb", "testcollection", bson.M{"id": "1234"}, map[string]interface{}{
		"$set": map[string]interface{}{
			"key": "new_value",
		},
	)

	record, err = mclient.FindRecordByID("testdb", "testcollection", "1234")
	if err == nil && record["key"] != "new_value" {
		t.Fatal("expected record to match")
	}
}