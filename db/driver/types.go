package driver

import "context"

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

type OpenOptions struct {
	Client interface{} // 原生的Client
	Driver string
	URI    string
}

type IndexOptions struct {
	Name       string
	Background bool
	Sparse     bool
	Unique     bool
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
