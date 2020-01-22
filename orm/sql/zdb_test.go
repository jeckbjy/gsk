package sql

import (
	"net/url"
	"testing"

	"github.com/jeckbjy/gsk/orm/driver"
	//_ "github.com/go-sql-driver/mysql"
	//_ "github.com/lib/pq"
)

func openMySQL() *mysqlDB {
	d := &mysqlDB{}
	if err := d.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/"); err != nil {
		panic(err)
	}
	return d
}

func openPostgress() *postgresDB {
	d := &postgresDB{}
	if err := d.Open("postgres", "postgres://postgres:123456@localhost/postgres?sslmode=disable"); err != nil {
		panic(err)
	}
	return d
}

func TestDatabase(t *testing.T) {
	d := openMySQL()
	dbname := "test"
	tblname := "persons"
	if err := d.CreateDatabase(dbname); err != nil {
		t.Fatal(err)
	}

	if current, err := d.Current(); err != nil {
		t.Fatal(err)
	} else {
		t.Log("CurrentDB is", current)
	}

	//if err := d.Use(dbname); err != nil {
	//	t.Fatal(err)
	//}

	if current, err := d.Current(); err != nil {
		t.Fatal(err)
	} else {
		t.Log("CurrentDB is", current)
	}

	err := d.CreateTable(tblname,
		[]driver.Column{
			{Name: "PersonID", Type: "int"},
			{Name: "LastName", Type: "varchar(255)"},
		})
	if err != nil {
		t.Fatal("CreateTable", err)
	} else {
		t.Log("CreateTable")
	}

	if err := d.DropTable(tblname); err != nil {
		t.Fatal(err)
	} else {
		t.Log("DropTable")
	}

	if err := d.DropDatabase(dbname); err != nil {
		t.Fatal(err)
	} else {
		t.Log("drop database")
	}
}

func TestIndex(t *testing.T) {
	d := mysqlDB{}
	if err := d.Open("mysql", "root:@tcp(127.0.0.1:3306)/iap"); err != nil {
		t.Fatal(err)
	}

	indexes, err := d.Indexes("user")
	if err != nil {
		t.Fatal(err)
	} else {
		for _, idx := range indexes {
			t.Log(idx)
		}
	}
}

func TestURL(t *testing.T) {
	urls := []string{
		`root:123456@tcp(127.0.0.1:3306)`,
		`root:123456@tcp(127.0.0.1:3306)/`,
		`mysql://root:123456@tcp(127.0.0.1:3306)/`,
	}
	for _, str := range urls {
		u, err := url.Parse(str)
		if err != nil {
			t.Fatal(u, err)
		} else {
			t.Log(u)
		}
	}
}
