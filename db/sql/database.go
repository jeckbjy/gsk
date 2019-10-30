package sql

import (
	"database/sql"

	"github.com/jeckbjy/gsk/db/driver"
)

type sqlDB struct {
	db *sql.DB
}

func (d *sqlDB) Indexes(table string) ([]driver.Index, error) {
	// show index?
	panic("implement me")
}

func (d *sqlDB) CreateIndex(table string, keys interface{}, opts *driver.IndexOptions) error {
	indexKeys, err := driver.ParseIndex(keys)
	if err != nil {
		return err
	}

	name := opts.Name
	if name == "" {
		name = driver.ParseIndexName(indexKeys)
	}

	columnsBuilder := sqlBuilder{}
	for _, idx := range indexKeys {
		var order string
		if idx.Order == 1 {
			order = "ASC"
		} else {
			order = "DESC"
		}
		columnsBuilder.Write(',', "%s %s", idx.Key, order)
	}
	// CREATE UNIQUE INDEX index_name ON table_name (column_name)
	columns := columnsBuilder.String()
	builder := sqlBuilder{}
	if opts.Unique {
		builder.Write(0, "CREATE UNIQUE INDEX %s ON %s (%s)", name, table, columns)
	} else {
		builder.Write(0, "CREATE INDEX %s ON %s (%s)", name, table, columns)
	}

	_, execErr := d.db.Exec(builder.String())
	return execErr
}

func (d *sqlDB) DropIndex(table string, name string) error {
	builder := sqlBuilder{}
	builder.Write(0, "DROP INDEX IF EXISTS %s ON %s", name, table)
	_, err := d.db.Exec(builder.String())
	return err
}

func (d *sqlDB) CreateTable(table string, schema interface{}) error {
	panic("implement me")
}

func (d *sqlDB) DropTable(table string) error {
	builder := sqlBuilder{}
	builder.Write(0, "DROP TABLE %s", table)
	_, err := d.db.Exec(builder.String())
	return err
}

func (d *sqlDB) Insert(table string, doc interface{}, opts *driver.InsertOptions) (*driver.InsertResult, error) {
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

func (d *sqlDB) Delete(table string, filter driver.Cond, opts *driver.DeleteOptions) (*driver.DeleteResult, error) {
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

func (d *sqlDB) Update(table string, filter driver.Cond, update interface{}, opts *driver.UpdateOptions) (*driver.UpdateResult, error) {
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

func (d *sqlDB) Query(table string, filter driver.Cond, opts *driver.QueryOptions) (driver.QueryResult, error) {
	where, err := toWhere(filter)
	if err != nil {
		return nil, err
	}

	builder := sqlBuilder{}
	builder.Write(spBlank, "SELECT %s FROM %s")
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
