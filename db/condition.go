package db

import "github.com/jeckbjy/gsk/db/driver"

func Eq(key string, value interface{}) Cond {
	return &expr{op: driver.TOK_EQ, key: key, val: value}
}

func Gt(key string, value interface{}) Cond {
	return &expr{driver.TOK_GT, key, value}
}

func Gte(key string, value interface{}) Cond {
	return &expr{driver.TOK_GTE, key, value}
}

func Lt(key string, value interface{}) Cond {
	return &expr{driver.TOK_LT, key, value}
}

func Lte(key string, value interface{}) Cond {
	return &expr{driver.TOK_LTE, key, value}
}

func Ne(key string, value interface{}) Cond {
	return &expr{driver.TOK_NE, key, value}
}

func In(key string, value interface{}) Cond {
	return &expr{driver.TOK_IN, key, value}
}

func Nin(key string, value interface{}) Cond {
	return &expr{driver.TOK_NIN, key, value}
}

func Not(cond Cond) Cond {
	return &unary{driver.TOK_NOT, cond}
}

func And(conds ...Cond) Cond {
	return &list{driver.TOK_AND, conds}
}

func Or(conds ...Cond) Cond {
	return &list{driver.TOK_OR, conds}
}

func Nor(conds ...Cond) Cond {
	return &list{driver.TOK_NOR, conds}
}

type Token = driver.Token

type expr struct {
	op  Token
	key string
	val interface{}
}

func (e *expr) Operator() Token {
	return e.op
}

func (e *expr) Key() string {
	return e.key
}
func (e *expr) Value() interface{} {
	return e.val
}

type unary struct {
	op Token
	x  Cond
}

func (u *unary) Operator() Token {
	return u.op
}

func (u *unary) X() Cond {
	return u.x
}

type binary struct {
	op Token
	x  Cond
	y  Cond
}

func (b *binary) Operator() Token {
	return b.op
}

func (b *binary) X() Cond {
	return b.x
}

func (b *binary) Y() Cond {
	return b.y
}

type list struct {
	op    Token
	conds []Cond
}

func (l *list) Operator() Token {
	return l.op
}

func (l *list) List() []Cond {
	return l.conds
}
