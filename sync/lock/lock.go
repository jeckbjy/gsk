package lock

// Lock is a distributed locking interface
type Lock interface {
	// Acquire a lock with given id
	Acquire(id string, opts ...AcquireOption) error
	// Release the lock with given id
	Release(id string) error
}
