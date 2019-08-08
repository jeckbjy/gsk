package selector

import (
	"context"
	"github.com/jeckbjy/micro/registry"
)

type Options struct {
	Registry registry.IRegistry
	Strategy Strategy

	// Other options for implementations of the interface
	// can be stored in a context
	Context context.Context
}

type SelectOptions struct {
	Filters  []Filter
	Strategy Strategy

	// Other options for implementations of the interface
	// can be stored in a context
	Context context.Context
}

// Option used to initialise the selector
type Option func(*Options)

// SelectOption used when making a select call
type SelectOption func(*SelectOptions)
