package db

import (
	"testing"
	"time"

	_ "github.com/jeckbjy/gsk/db/sql"
)

func TestClient(t *testing.T) {
	type Foo struct {
		ID        string    `bson:"_id" json:"_id"`
		Name      string    `bson:"name" json:"name"`
		CreatedAt time.Time `bson:"created_at" json:"created_at"`
	}

	client, err := New("sql")
	if err != nil {
		t.Fatal(err)
	}

	db, err := client.Database("test")
	if err != nil {
		t.Fatal(err)
	}

	table := "foo"
	if _, err := db.Insert(table, &Foo{ID: "a", Name: "a", CreatedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}

	// test query
	if query, err := db.QueryOne(table, Eq("_id", "a")); err == nil {
		result := &Foo{}
		if err := query.Decode(result); err == nil {
			t.Log(result)
		} else {
			t.Fatal(err)
		}
	} else {
		t.Fatal(err)
	}

	if _, err := db.Delete(table, Eq("_id", "a")); err != nil {
		t.Fatal(err)
	}

	if err := client.Drop("test"); err != nil {
		t.Fatal(err)
	}
}
