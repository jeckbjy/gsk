package sql

import (
	"database/sql"
	"errors"

	"github.com/jeckbjy/gsk/db/driver"
)

func init() {
	driver.Register("sql", &sqlDriver{})
}

type sqlDriver struct {
	db *sql.DB
}

func (d *sqlDriver) Name() string {
	return "sql"
}

func (d *sqlDriver) Open(opts *driver.OpenOptions) error {
	db, err := sql.Open(opts.Driver, opts.URI)
	if err != nil {
		return err
	}

	d.db = db
	return nil
}

func (d *sqlDriver) Close() error {
	return d.db.Close()
}

func (d *sqlDriver) Ping() error {
	return d.db.Ping()
}

func (d *sqlDriver) Drop(name string) error {
	_, err := d.db.Exec("DROP database IF EXISTS %s", name)
	return err
}

func (d *sqlDriver) Database(name string) (driver.Database, error) {
	if d.db != nil {
		return &sqlDB{db: d.db}, nil
	}

	return nil, errors.New("null db")
}
