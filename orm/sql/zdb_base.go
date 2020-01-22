package sql

import (
	"database/sql"
	"sync"

	"github.com/jeckbjy/gsk/orm/driver"
)

var gEngineMux sync.Mutex
var gEngineMap = make(map[string]Engine)

func init() {
	AddEngine(&mysqlDB{})
	AddEngine(&postgresDB{})
}

func AddEngine(e Engine) {
	gEngineMux.Lock()
	gEngineMap[e.Name()] = e
	gEngineMux.Unlock()
}

func GetEngine(name string) Engine {
	gEngineMux.Lock()
	e := gEngineMap[name]
	gEngineMux.Unlock()
	return e
}

// 抽象不同数据库接口
//
// null处理: https://github.com/guregu/null
// switch database:mysql虽然可以使用USE db实现切换,但最好不要做这样的操作,因为connection是有pool的
// pg没有切换数据库这样的功能
// 只能通过重新创建新的连接的方式支持多数据库切换
type Engine interface {
	Clone() Engine
	Bind(db *sql.DB)
	Close() error
	Name() string
	Current() (string, error)
	CreateDatabase(name string) error
	DropDatabase(name string) error
	CreateTable(name string, columns []driver.Column) error
	DropTable(name string) error
	Indexes(table string) ([]*driver.Index, error)
	CreateIndex(table string, index *driver.Index) error
	DropIndex(table string, name string) error
	Insert(table string, doc interface{}, opts *driver.InsertOptions) (*driver.InsertResult, error)
	Delete(table string, filter driver.Cond, opts *driver.DeleteOptions) (*driver.DeleteResult, error)
	Update(table string, filter driver.Cond, update interface{}, opts *driver.UpdateOptions) (*driver.UpdateResult, error)
	Query(table string, filter driver.Cond, opts *driver.QueryOptions) (driver.QueryResult, error)
}

type baseDB struct {
	db *sql.DB
}

func (d *baseDB) Bind(db *sql.DB) {
	d.db = db
}

func (d *baseDB) Exec(query string) error {
	_, err := d.db.Exec(query)
	return err
}

func (d *baseDB) Open(driver string, dsn string) error {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return err
	}

	d.db = db
	return nil
}

func (d *baseDB) Close() error {
	return d.db.Close()
}

func (d *baseDB) Ping() error {
	return d.db.Ping()
}

func (d *baseDB) CreateDatabase(name string) error {
	return d.Exec("CREATE DATABASE IF NOT EXISTS " + name)
}

func (d *baseDB) DropDatabase(name string) error {
	return d.Exec("DROP database IF EXISTS " + name)
}

func (d *baseDB) CreateTable(name string, columns []driver.Column) error {
	builder := sqlBuilder{}
	builder.Write(spBlank, "CREATE TABLE %s (", name)
	for i, f := range columns {
		if i != 0 {
			builder.Write(spNone, ",")
		}

		builder.Write(spBlank, "%s %s", f.Name, f.Type)
		// NOT NULL
		if f.NotNull {
			builder.Write(spBlank, "NOT NULL")
		}

		if f.AutoIncrement {
			builder.Write(spBlank, "AUTO_INCREMENT")
		}

		if len(f.Default) != 0 {
			// TODO:如果类型是字符串,还需要添加''
			// f.Default还可能是一个函数,比如GETDATE()
			builder.Write(spBlank, "DEFAULT %s", f.Default)
		}

		// 不支持check约束
	}

	// TODO: PRIMARY KEY
	builder.Write(spBlank, ")")
	return d.Exec(builder.String())
}

func (d *baseDB) DropTable(name string) error {
	return d.Exec("DROP TABLE " + name)
}

// 创建索引,TODO:主键当作索引处理
func (d *baseDB) CreateIndex(table string, index *driver.Index) error {
	// CREATE [UNIQUE] INDEX index_name ON table_name (column_name) [USING BTREE];
	builder := sqlBuilder{}
	if index.Unique {
		builder.Write(spNone, "CREATE UNIQUE INDEX %s ON %s (", index.Name, table)
	} else {
		builder.Write(spNone, "CREATE INDEX %s ON %s (", index.Name, table)
	}

	for i, key := range index.Keys {
		if i != 0 {
			builder.Write(spNone, ",")
		}

		if key.Order == driver.Asc {
			builder.Write(spBlank, "%s %s", key.Name, "ASC")
		} else {
			builder.Write(spBlank, "%s %s", key.Name, "DESC")
		}
	}

	builder.Write(spNone, ")")
	return d.Exec(builder.String())
}

func (d *baseDB) DropIndex(table string, name string) error {
	builder := sqlBuilder{}
	builder.Write(spNone, "DROP INDEX IF EXISTS %s ON %s", name, table)
	return d.Exec(builder.String())
}

func (d *baseDB) Insert(table string, doc interface{}, opts *driver.InsertOptions) (*driver.InsertResult, error) {
	// INSERT INTO table_name (col1, col2,...) VALUES (val1, val2,...)
	result := &driver.InsertResult{}
	// TODO:ToMap,ToList 变成有序?
	records := make([]map[string]interface{}, 0)
	values := make([]interface{}, 0)

	for _, row := range records {
		keyb := sqlBuilder{}
		valb := sqlBuilder{}
		values := values[:0]
		for k, v := range row {
			keyb.Write(spComma, k)
			valb.Write(spComma, "?")
			values = append(values, v)
		}
		//
		builder := sqlBuilder{}
		builder.Write(spNone, "INSERT INTO %s (%s) VALUES (%s)", table, keyb.String(), valb.String())
		res, err := d.db.ExecContext(opts.Context, builder.String(), values...)
		if err != nil {
			return nil, err
		}

		lastId, err := res.LastInsertId()
		if err != nil {
			return nil, err
		}

		result.InsertedIDs = append(result.InsertedIDs, lastId)
	}

	return result, nil
}

func (d *baseDB) Delete(table string, filter driver.Cond, opts *driver.DeleteOptions) (*driver.DeleteResult, error) {
	where, err := toWhere(filter)
	if err != nil {
		return nil, err
	}

	b := sqlBuilder{}
	b.Write(spBlank, "DELETE FROM %s", table)
	b.Write(spBlank, where)
	if opts.One {
	}
	b.Write(spBlank, "LIMIT 1")
	res, err := d.db.Exec(b.String())
	if err != nil {
		return nil, err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	return &driver.DeleteResult{DeletedCount: count}, nil
}

func (d *baseDB) Update(table string, filter driver.Cond, update interface{}, opts *driver.UpdateOptions) (*driver.UpdateResult, error) {
	// UPDATE table_name SET 列名称 = 新值 WHERE 列名称 = 某值 [LIMIT 1]
	where, err := toWhere(filter)
	if err != nil {
		return nil, err
	}

	builder := sqlBuilder{}
	builder.Write(spBlank, "UPDATE %s SET VALUES", table)
	builder.Write(spBlank, where)
	if opts.One {
		builder.Write(spBlank, "LIMIT 1")
	}

	res, err := d.db.Exec(builder.String())
	if err != nil {
		return nil, err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	return &driver.UpdateResult{MatchedCount: count}, nil
}

func (d *baseDB) Query(table string, filter driver.Cond, opts *driver.QueryOptions) (driver.QueryResult, error) {
	where, err := toWhere(filter)
	if err != nil {
		return nil, err
	}

	builder := sqlBuilder{}
	//builder.Write(spBlank, "SELECT %s FROM %s")
	builder.Write(spBlank, where)
	if opts.One {
		builder.Write(spBlank, "LIMIT 1")
	}

	rows, err := d.db.QueryContext(opts.Context, builder.String())
	if err != nil {
		return nil, err
	}

	result := newQueryResult(rows)

	return result, nil
}
