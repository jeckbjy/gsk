package store

import "time"

type WriteOptions struct {
	IsDir bool
	TTL   time.Duration
}

type ReadOptions struct {
	// Consistent defines if the behavior of a Get operation is
	// linearizable or not. Linearizability allows us to 'see'
	// objects based on a real-time total order as opposed to
	// an arbitrary order or with stale values ('inconsistent'
	// scenario).
	Consistent bool
}

// LockOptions contains optional request parameters
type LockOptions struct {
	Value     []byte        // Optional, value to associate with the lock
	TTL       time.Duration // Optional, expiration ttl associated with the lock
	RenewLock chan struct{} // Optional, chan used to control and stop the session ttl renewal for the lock
}
