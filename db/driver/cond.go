package driver

type Cond interface {
	Operator() Token
}

// 例如，TOK_EQ
type ExprCond interface {
	Cond
	Key() string
	Value() interface{}
}

type UnaryCond interface {
	Cond
	X() Cond
}

type BinaryCond interface {
	Cond
	X() Cond
	Y() Cond
}

type ListCond interface {
	Cond
	List() []Cond
}
