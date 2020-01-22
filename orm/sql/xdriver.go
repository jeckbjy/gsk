package sql

import (
	"database/sql"
	"errors"
	"sync"

	"github.com/jeckbjy/gsk/orm/driver"
	"github.com/jeckbjy/gsk/util/dsn"
)

var (
	ErrNotFoundSQLEngine = errors.New("not found sql engine")
)

func init() {
	driver.Register("sql", &sqlDriver{})
}

// sqlDriver 需要支持同时连接不同数据库,目前每个数据库都会单独Open,
// TODO:可以研究一下是否能复用连接池
// 连接方式仅支持URI连接
type sqlDriver struct {
	client    *sql.DB           // 初始连接
	databases map[string]*sqlDB // 所有数据库
	uri       *dsn.URL          // 原始连接
	current   string            // 当前数据库名
	engine    Engine            // 不同数据库引擎
	mux       sync.Mutex
}

func (d *sqlDriver) Name() string {
	return "sql"
}

func (d *sqlDriver) Open(opts *driver.OpenOptions) error {
	d.mux.Lock()
	defer d.mux.Unlock()
	uri, err := dsn.Parse(opts.URI)
	if err != nil {
		return err
	}

	e := GetEngine(uri.Driver)
	if e == nil {
		return ErrNotFoundSQLEngine
	}

	if db, err := sql.Open(uri.Driver, uri.DSN); err != nil {
		return err
	} else {
		d.client = db
	}
	d.uri = uri
	d.engine = e
	d.databases = make(map[string]*sqlDB)
	return nil
}

func (d *sqlDriver) Close() error {
	d.mux.Lock()
	if d.databases != nil {
		for _, db := range d.databases {
			_ = db.Close()
		}
		d.databases = nil
	}
	client := d.client
	d.client = nil
	d.uri = nil
	d.current = ""
	d.mux.Unlock()
	if client != nil {
		return client.Close()
	}

	return nil
}

func (d *sqlDriver) Ping() error {
	return d.client.Ping()
}

func (d *sqlDriver) Drop(name string) error {
	_, err := d.client.Exec("DROP database IF EXISTS %s", name)
	return err
}

func (d *sqlDriver) Database(name string) (driver.Database, error) {
	var result *sqlDB
	var err error
	d.mux.Lock()
	if d.databases == nil {
		d.databases = make(map[string]*sqlDB)
	}
	if db, ok := d.databases[name]; ok {
		result = db
	} else {
		result, err = d.createDB(name)
		if result != nil {
			d.databases[name] = result
		}
	}
	d.mux.Unlock()
	return result, err
}

func (d *sqlDriver) createDB(name string) (*sqlDB, error) {
	var db *sql.DB
	if name != d.current {
		if err := d.engine.CreateDatabase(name); err != nil {
			return nil, err
		}
		d.uri.Database = name
		if err := d.uri.Build(); err != nil {
			return nil, err
		}
		b, err := sql.Open(d.uri.Driver, d.uri.DSN)
		if err != nil {
			return nil, err
		}
		db = b
	}
	e := d.engine.Clone()
	e.Bind(db)
	return &sqlDB{driver: d, engine: e, dbname: name}, nil
}
