package lock

import "time"

type Options struct {
	Nodes  []string
	Prefix string
}

type AcquireOptions struct {
	TTL  time.Duration
	Wait time.Duration
}

type Option func(o *Options)
type AcquireOption func(o *AcquireOptions)

// Nodes sets the addresses the underlying lock implementation
func Nodes(a ...string) Option {
	return func(o *Options) {
		o.Nodes = a
	}
}

// Prefix sets a prefix to any lock ids used
func Prefix(p string) Option {
	return func(o *Options) {
		o.Prefix = p
	}
}

// TTL sets the lock ttl
func TTL(t time.Duration) AcquireOption {
	return func(o *AcquireOptions) {
		o.TTL = t
	}
}

// Wait sets the wait time
func Wait(t time.Duration) AcquireOption {
	return func(o *AcquireOptions) {
		o.Wait = t
	}
}
