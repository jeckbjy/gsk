package sql

import (
	"github.com/jeckbjy/gsk/orm/driver"
)

type sqlDB struct {
	driver *sqlDriver
	engine Engine
	dbname string
}

func (d *sqlDB) Close() error {
	return d.engine.Close()
}

func (d *sqlDB) Indexes(table string) ([]*driver.Index, error) {
	return d.engine.Indexes(table)
}

func (d *sqlDB) CreateIndex(table string, index *driver.Index) error {
	if index.Background {
		go d.engine.CreateIndex(table, index)
		return nil
	} else {
		return d.engine.CreateIndex(table, index)
	}
}

func (d *sqlDB) DropIndex(table string, name string) error {
	return d.engine.DropIndex(table, name)
}

func (d *sqlDB) CreateTable(name string, columns []driver.Column) error {
	return d.engine.CreateTable(name, columns)
}

func (d *sqlDB) DropTable(name string) error {
	return d.engine.DropTable(name)
}

func (d *sqlDB) Insert(table string, doc interface{}, opts *driver.InsertOptions) (*driver.InsertResult, error) {
	return d.engine.Insert(table, doc, opts)
}

func (d *sqlDB) Delete(table string, filter driver.Cond, opts *driver.DeleteOptions) (*driver.DeleteResult, error) {
	return d.engine.Delete(table, filter, opts)
}

func (d *sqlDB) Update(table string, filter driver.Cond, update interface{}, opts *driver.UpdateOptions) (*driver.UpdateResult, error) {
	return d.engine.Update(table, filter, update, opts)
}

func (d *sqlDB) Query(table string, filter driver.Cond, opts *driver.QueryOptions) (driver.QueryResult, error) {
	return d.engine.Query(table, filter, opts)
}
