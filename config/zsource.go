package config

import (
	"crypto/md5"
	"errors"
	"fmt"
	"time"
)

var (
	// ErrWatcherStopped is returned when source watcher has been stopped
	ErrWatcherStopped = errors.New("watcher stopped")
)

// Source is the source from which config is loaded
type Source interface {
	Read() (*ChangeSet, error)
	Write(*ChangeSet) error
	Watch() (Watcher, error)
	String() string
}

// Watcher watches a source for changes
type Watcher interface {
	Next() (*ChangeSet, error)
	Stop() error
}

// ChangeSet represents a set of changes from a source
type ChangeSet struct {
	Data      []byte    // Raw encoded config data
	Checksum  string    // MD5 checksum of the data
	Format    string    // Encoding format e.g json, yaml, toml, xml
	Source    string    // Source of the config e.g file, consul, etcd
	Timestamp time.Time // Time of loading or update
}

// Sum returns the md5 checksum of the ChangeSet data
func (c *ChangeSet) Sum() string {
	h := md5.New()
	h.Write(c.Data)
	return fmt.Sprintf("%x", h.Sum(nil))
}
