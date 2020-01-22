package sql

import (
	"database/sql"

	"github.com/jeckbjy/gsk/orm/driver"
)

// 实现mysql接口
// mysql -h 127.0.0.1 -P 3306 -u root
// docker run -p 3306:3306 --name mysql -e MYSQL_ROOT_PASSWORD=123456 -d mysql
// 查看: docker container list;
// 删除: docker stop xxxx
// MySQL驱动: github.com/go-sql-driver/mysql
type mysqlDB struct {
	baseDB
}

func (d *mysqlDB) Name() string {
	return "mysql"
}

func (d *mysqlDB) Clone() Engine {
	return &mysqlDB{}
}

func (d *mysqlDB) Current() (string, error) {
	row := d.db.QueryRow("select database()")
	var name sql.NullString
	if err := row.Scan(&name); err != nil {
		return "", err
	}

	return name.String, nil
}

//func (d *mysqlDB) Use(name string) error {
//	return d.Exec("USE " + name)
//}

type MySQLIndex struct {
	Table        string
	NonUnique    int
	KeyName      string
	SeqInIndex   int
	ColumnName   string
	Collation    string         // A升序,B降序
	Cardinality  int            //
	SubPart      sql.NullString //
	Packed       sql.NullBool   //
	Null         string
	IndexType    string
	Comment      sql.NullString
	IndexComment sql.NullString
}

// 查询索引信息
func (d *mysqlDB) Indexes(table string) ([]*driver.Index, error) {
	// https://www.mysqltutorial.org/mysql-index/mysql-show-indexes/
	// https://dev.mysql.com/doc/refman/8.0/en/show-index.html
	rows, err := d.db.Query("SHOW INDEXES FROM " + table)
	if err != nil {
		return nil, err
	}

	indexes := make([]*MySQLIndex, 0)
	for rows.Next() {
		index := &MySQLIndex{}
		err := rows.Scan(
			&index.Table,
			&index.NonUnique,
			&index.KeyName,
			&index.SeqInIndex,
			&index.ColumnName,
			&index.Collation,
			&index.Cardinality,
			&index.SubPart,
			&index.Packed,
			&index.Null,
			&index.IndexType,
			&index.Comment,
			&index.IndexComment,
		)
		if err != nil {
			return nil, err
		}

		indexes = append(indexes, index)
	}

	result := make([]*driver.Index, 0, len(indexes))
	dict := make(map[string]*driver.Index)
	for _, idx := range indexes {
		rindex, ok := dict[idx.KeyName]
		if !ok {
			rindex = &driver.Index{
				Name:   idx.KeyName,
				Unique: idx.NonUnique == 0,
			}
			result = append(result, rindex)
		}
		order := driver.Asc
		if idx.Collation == "B" {
			order = driver.Desc
		}
		rindex.Keys = append(rindex.Keys, driver.IndexKey{Name: idx.ColumnName, Order: order})
	}

	return result, nil
}
