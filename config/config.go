package config

type Config interface {
	Close() error
	Load() error
	Sync() error
	Watch(path ...string) error
}

type Watch interface {
	Next()
	Stop() error
}
