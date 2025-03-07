package mongo

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	.go.mongodb.org/mongo-driver/mongo"
	.go.mongodb.org/mongo-driver/mongo/options"
)

func TestNewMongoClient(t *testing.T) {
	db, err := MongoClient("mongodb://localhost:27017/").client()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Disconnect(context.Background())

	if err = db.Ping(context.Background(), nil); err != nil {
		t.Fatal(err)
	}

}

func TestNewMongoClientError(t *testing.T) {
	_, err := MongoClient("invalid:uri")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDisconnect(t *testing.T) {
	db, err := MongoClient("mongodb://localhost:27017/").client()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = db.Disconnect(context.Background())
		if err != nil {
			t.Fatal(err)
		}
	}()

}

func TestInsertRecord(t *testing.T) {
	db, err := MongoClient("mongodb://localhost:27017/").client()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = db.Disconnect(context.Background())
		if err != nil {
			t.Fatal(err)
		}
	}()

	_, err = db.InsertRecord("testDB", "testCollection", bson.M{"id": 1, "name": "John"})
	if err != nil {
		t.Fatal(err)
	}

}

func TestInsertRecordError(t *testing.T) {
	db, err := MongoClient("invalid:uri").client()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFindAllRecords(t *testing.T) {
	db, err := MongoClient("mongodb://localhost:27017/").client()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = db.Disconnect(context.Background())
		if err != nil {
			t.Fatal(err)
		}
	}()

	records, err := db.FindAllRecords("testDB", "testCollection")
	if err != nil {
		t.Fatal(err)
	}

	if len(records) == 0 {
		t.Fatal("expected records")
	}

}

func TestFindAllRecordsError(t *testing.T) {
	db, err := MongoClient("invalid:uri").client()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFindRecordByID(t *testing.T) {
	db, err := MongoClient("mongodb://localhost:27017/").client()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = db.Disconnect(context.Background())
		if err != nil {
			t.Fatal(err)
		}
	}()

	record, err := db.FindRecordByID("testDB", "testCollection", "1")
	if err != nil {
		t.Fatal(err)
	}

	if record == nil {
		t.Fatal("expected record")
	}

}

func TestFindRecordByIDError(t *testing.T) {
	db, err := MongoClient("invalid:uri").client()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFindRecordsWithProjection(t *testing.T) {
	db, err := MongoClient("mongodb://localhost:27017/").client()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = db.Disconnect(context.Background())
		if err != nil {
			t.Fatal(err)
		}
	}()

 records, err := db.FindRecordsWithProjection("testDB", "testCollection", bson.M{"id": 1}, bson.M{"name": 0})
	if err != nil {
		t.Fatal(err)
	}

	if len(records) == 0 {
		t.Fatal("expected records")
	}

}

func TestFindRecordsWithProjectionError(t *testing.T) {
	db, err := MongoClient("invalid:uri").client()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUpdateRecord(t *testing.T) {
	db, err := MongoClient("mongodb://localhost:27017/").client()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = db.Disconnect(context.Background())
		if err != nil {
			t.Fatal(err)
		}
	}()

	_, err = db.InsertRecord("testDB", "testCollection", bson.M{"id": 1, "name": "John"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.UpdateRecord("testDB", "testCollection", bson.M{"id": 1}, bson.M{"$set": bson.M{"name": "Jane"}})
	if err != nil {
		t.Fatal(err)
	}

}

func TestUpdateRecordError(t *testing.T) {
	db, err := MongoClient("invalid:uri").client()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDeleteRecord(t *testing.T) {
	db, err := MongoClient("mongodb://localhost:27017/").client()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = db.Disconnect(context.Background())
		if err != nil {
			t.Fatal(err)
		}
	}()

	_, err = db.InsertRecord("testDB", "testCollection", bson.M{"id": 1, "name": "John"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.DeleteRecord("testDB", "testCollection", bson.M{"id": 1})
	if err == nil {
		t.Fatal("expected error")
	}
}