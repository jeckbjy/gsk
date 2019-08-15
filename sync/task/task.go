package task

type Task interface {
}

type Command struct {
	Name string
	Func func() error
}
