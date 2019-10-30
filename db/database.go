package db

import "github.com/jeckbjy/gsk/db/driver"

type Database struct {
	database driver.Database
}

func (d *Database) Indexes(table string) ([]Index, error) {
	return d.database.Indexes(table)
}

func (d *Database) CreateIndex(table string, keys interface{}, opts ...IndexOption) error {
	o := driver.IndexOptions{}
	for _, fn := range opts {
		fn(&o)
	}
	return d.database.CreateIndex(table, keys, &o)
}

func (d *Database) DropIndex(table string, name string) error {
	return d.database.DropIndex(table, name)
}

func (d *Database) CreateTable(name string, schema interface{}) error {
	return d.database.CreateTable(name, schema)
}

func (d *Database) DropTable(name string) error {
	return d.database.DropTable(name)
}

func (d *Database) Insert(table string, doc interface{}, opts ...InsertOption) (*InsertResult, error) {
	o := driver.InsertOptions{}
	for _, fn := range opts {
		fn(&o)
	}
	return d.database.Insert(table, doc, &o)
}

func (d *Database) Delete(table string, filter Cond, opts ...DeleteOption) (*DeleteResult, error) {
	o := driver.DeleteOptions{}
	for _, fn := range opts {
		fn(&o)
	}

	return d.database.Delete(table, filter, &o)
}

func (d *Database) DeleteOne(table string, filter Cond, opts ...DeleteOption) (*DeleteResult, error) {
	o := driver.DeleteOptions{}
	o.One = true
	for _, fn := range opts {
		fn(&o)
	}

	return d.database.Delete(table, filter, &o)
}

func (d *Database) Update(table string, filter Cond, update interface{}, opts ...UpdateOption) (*UpdateResult, error) {
	o := driver.UpdateOptions{}
	for _, fn := range opts {
		fn(&o)
	}

	return d.database.Update(table, filter, update, &o)
}

func (d *Database) UpdateOne(table string, filter Cond, update interface{}, opts ...UpdateOption) (*UpdateResult, error) {
	o := driver.UpdateOptions{}
	o.One = true
	for _, fn := range opts {
		fn(&o)
	}

	return d.database.Update(table, filter, update, &o)
}

func (d *Database) Query(table string, filter Cond, opts ...QueryOption) error {
	return nil
}

func (d *Database) QueryOne(table string, filter Cond, opts ...QueryOption) (QueryResult, error) {
	o := driver.QueryOptions{}
	o.One = true
	for _, fn := range opts {
		fn(&o)
	}

	return d.database.Query(table, filter, &o)
}
