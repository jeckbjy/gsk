package atomicx

// Error is an atomic type-safe wrapper around Value for errors
type Error struct{ v Value }

// errorHolder is non-nil holder for error object.
// atomic.Value panics on saving nil object, so err object needs to be
// wrapped with valid object first.
type errorHolder struct{ err error }

// NewError creates new atomic error object
func NewError(err error) *Error {
	e := &Error{}
	if err != nil {
		e.Store(err)
	}
	return e
}

// Load atomically loads the wrapped error
func (e *Error) Load() error {
	v := e.v.Load()
	if v == nil {
		return nil
	}

	eh := v.(errorHolder)
	return eh.err
}

// Store atomically stores error.
// NOTE: a holder object is allocated on each Store call.
func (e *Error) Store(err error) {
	e.v.Store(errorHolder{err: err})
}
