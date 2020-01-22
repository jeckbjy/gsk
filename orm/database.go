package orm

import (
	"errors"

	"github.com/jeckbjy/gsk/orm/driver"
)

var (
	ErrInvalidIndexKey = errors.New("invalid index key")
	ErrNotSupport      = errors.New("not support")
)

const (
	Asc  = driver.Asc  // 升序,默认
	Desc = driver.Desc // 降序
)

type Order = driver.Order
type IndexKey = driver.IndexKey
type Index = driver.Index
type Cond = driver.Cond

type InsertResult = driver.InsertResult
type DeleteResult = driver.DeleteResult
type UpdateResult = driver.UpdateResult
type QueryResult = driver.QueryResult

type Database struct {
	database driver.Database
}

func (d *Database) Indexes(table string) ([]*Index, error) {
	return d.database.Indexes(table)
}

// 创建索引,简单情况只需要传入列名即可,支持以下形式
// d.CreateIndex("test", "uid")
// d.CreateIndex("test", []string{"uid", "create_time"})
// d.CreateIndex("test", []orm.IndexKey{{Name:"uid", Order:orm.Desc}})
func (d *Database) CreateIndex(table string, keys interface{}, opts ...IndexOption) error {
	index := Index{}
	for _, fn := range opts {
		fn(&index)
	}

	switch key := keys.(type) {
	case string:
		index.Keys = append(index.Keys, IndexKey{Name: key, Order: Asc})
	case []string:
		for _, k := range key {
			index.Keys = append(index.Keys, IndexKey{Name: k, Order: Asc})
		}
	case []IndexKey:
		index.Keys = key
	default:
		return ErrNotSupport
	}

	if len(index.Keys) == 0 {
		return ErrInvalidIndexKey
	}

	index.GenerateName()

	return d.database.CreateIndex(table, &index)
}

func (d *Database) DropIndex(table string, name string) error {
	return d.database.DropIndex(table, name)
}

// CreateTable 会自动创建缺失的表,列,和索引,但不会删除和改变列类型
func (d *Database) CreateTable(model interface{}) error {
	dm := driver.Model{}
	if err := dm.Parse(model); err != nil {
		return err
	}
	if err := d.database.CreateTable(dm.Name, dm.Columns); err != nil {
		return err
	}
	// 创建索引
	return nil
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
