package sql

import (
	"database/sql"
	"fmt"

	"github.com/jeckbjy/gsk/orm/driver"
)

// docker中启动: docker run -p 5432:5432  --name pg -e POSTGRES_PASSWORD=123456 -d postgres
// 驱动:https://github.com/lib/pq
type postgresDB struct {
	baseDB
}

func (d *postgresDB) Name() string {
	return "postgres"
}

func (d *postgresDB) Clone() Engine {
	return &postgresDB{}
}

func (d *postgresDB) Current() (string, error) {
	row := d.db.QueryRow("select current_database()")
	var name sql.NullString
	if err := row.Scan(&name); err != nil {
		return "", err
	}

	return name.String, nil
}

func (d *postgresDB) CreateDatabase(name string) error {
	// https://notathoughtexperiment.me/blog/how-to-do-create-database-dbname-if-not-exists-in-postgres-in-golang/
	query := fmt.Sprintf("SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = '%s');", name)
	row := d.db.QueryRow(query)
	var exists bool
	err := row.Scan(&exists)
	if err != nil {
		return err
	}

	if !exists {
		return d.Exec("CREATE DATABASE " + name)
	}

	return nil
}

func (d *postgresDB) Indexes(table string) ([]*driver.Index, error) {
	return nil, nil
}
