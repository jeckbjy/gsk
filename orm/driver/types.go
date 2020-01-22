package driver

import (
	"context"
	"errors"
)

var (
	ErrInvalidIndexName = errors.New("invalid index name")
)

// https://docs.mongodb.com/manual/reference/operator/query-comparison/
// Query embedded document: $elemMatch vs. Dot Notation
type Token int

const (
	TOK_NULL Token = iota
	TOK_EQ
	TOK_NE
	TOK_GT
	TOK_GTE
	TOK_LT
	TOK_LTE
	TOK_IN
	TOK_NIN
	// logical Query Operators
	TOK_AND
	TOK_OR
	TOK_NOR
	TOK_NOT
)

type FieldType int

// 支持的数据类型
const (
	FTChar FieldType = 0
)

// 列信息,用于CreateTable
type Column struct {
	Name          string // 字段名
	Type          string // 类型
	Default       string // 默认值, 不需要双引号
	NotNull       bool   // 是否非空
	AutoIncrement bool   // 是否自增
	PrimaryKey    bool   // 是否是主键
	Size          int    // 类型大小
}

// 简单数据表模型,不支持外键等操作
type Model struct {
	Name    string
	Columns []Column
	Indexes []Index
}

func (m *Model) Parse(model interface{}) error {
	return nil
}

type OpenOptions struct {
	URI string
	//MaxIdleConns int           // 默认2
	//MaxOpenConns int           // 默认0,不限制
	//MaxLifetime  time.Duration // 默认0,不过期
}

type InsertOptions struct {
	Context context.Context
}

type DeleteOptions struct {
	One bool
}

type UpdateOptions struct {
	One    bool
	Upsert bool
}

type QueryOptions struct {
	One        bool
	Skip       int64
	Limit      int64
	Sort       map[string]int //1:ascending, -1:descending
	Projection map[string]int //1:include 0:exclude(有些不支持),nil代表全部
	Context    context.Context
}

/**
 * All Results
 */
type InsertResult struct {
	InsertedIDs []interface{}
}

type DeleteResult struct {
	DeletedCount int64
}

// UpdateResult is a result of an update operation.
//
// UpsertedID will be a Go type that corresponds to a BSON type.
type UpdateResult struct {
	// The number of documents that matched the filter.
	MatchedCount int64
	// The number of documents that were modified.
	ModifiedCount int64
	// The number of documents that were upserted.
	UpsertedCount int64
	// The identifier of the inserted document if an upsert took place.
	UpsertedID interface{}
}

// 查询结果,可以调用Decode自动反射结果,也可以调用Cursor,手动遍历结果
type QueryResult interface {
	Cursor() Cursor
	Decode(result interface{}) error
}
