package tracing

import (
	"time"
)

type StartSpanOptions struct {
	Parent SpanContext
	Tags   map[string]interface{}
	SpanID uint64
}

type FinishOptions struct {
	Duration time.Duration
	Error    error
}

type FinishOption func(options *FinishOptions)
type StartSpanOption func(options *StartSpanOptions)

func WithError(err error) FinishOption {
	return func(options *FinishOptions) {
		options.Error = err
	}
}

func WithParent(parent SpanContext) StartSpanOption {
	return func(options *StartSpanOptions) {
		options.Parent = parent
	}
}

func WithParentSpan(parent Span) StartSpanOption {
	return func(options *StartSpanOptions) {
		options.Parent = parent.Context()
	}
}

func WithTags(tags map[string]interface{}) StartSpanOption {
	return func(options *StartSpanOptions) {
		options.Tags = tags
	}
}

func WithSpanID(spanID uint64) StartSpanOption {
	return func(options *StartSpanOptions) {
		options.SpanID = spanID
	}
}
