package sql

import (
	"database/sql"

	"github.com/jeckbjy/gsk/orm/driver"
)

func newQueryResult(rows *sql.Rows) *sqlQueryResult {
	r := &sqlQueryResult{}
	r.cursor.rows = rows
	return r
}

type sqlQueryResult struct {
	cursor sqlCursor
}

func (r *sqlQueryResult) Cursor() driver.Cursor {
	return &r.cursor
}

func (r *sqlQueryResult) Decode(result interface{}) error {
	return driver.Decode(&r.cursor, result)
}
