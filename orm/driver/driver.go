// 定义数据库操作接口
package driver

import (
	"fmt"
	"sort"
	"sync"
)

var (
	driversMu sync.RWMutex
	drivers   = make(map[string]Driver)
)

func Find(name string) (Driver, error) {
	driversMu.RLock()
	driveri, ok := drivers[name]
	driversMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("db: unknown driver %q (forgotten import?)", name)
	}

	return driveri, nil
}

// Register makes a database driver available by the provided name.
// If Register is called twice with the same name or if driver is nil,
// it panics.
func Register(name string, driver Driver) {
	driversMu.Lock()
	defer driversMu.Unlock()
	if driver == nil {
		panic("sql: Register driver is nil")
	}
	if _, dup := drivers[name]; dup {
		panic("sql: Register called twice for driver " + name)
	}
	drivers[name] = driver
}

func unregisterAllDrivers() {
	driversMu.Lock()
	defer driversMu.Unlock()
	// For tests.
	drivers = make(map[string]Driver)
}

// Drivers returns a sorted list of the names of the registered drivers.
func Drivers() []string {
	driversMu.RLock()
	defer driversMu.RUnlock()
	var list []string
	for name := range drivers {
		list = append(list, name)
	}
	sort.Strings(list)
	return list
}

type Driver interface {
	Name() string
	Open(opts *OpenOptions) error
	Close() error
	Ping() error
	Database(name string) (Database, error)
	Drop(name string) error
}

type Database interface {
	Indexes(table string) ([]Index, error)
	CreateIndex(table string, keys interface{}, opts *IndexOptions) error
	DropIndex(table string, name string) error

	CreateTable(table string, schema interface{}) error
	DropTable(table string) error

	Insert(table string, doc interface{}, opts *InsertOptions) (*InsertResult, error)
	Delete(table string, filter Cond, opts *DeleteOptions) (*DeleteResult, error)
	Update(table string, filter Cond, update interface{}, opts *UpdateOptions) (*UpdateResult, error)
	Query(table string, filter Cond, opts *QueryOptions) (QueryResult, error)
	//Aggregate(table string, filter Cond) error
}

// Scan和Decode的区别是,Scan对应原生sql中的Scan,而Decode则是反射解析到struct中
type Cursor interface {
	//ID() int64
	Close() error
	Next() bool
	Scan(dest ...interface{}) error
	Decode(model interface{}) error
}
