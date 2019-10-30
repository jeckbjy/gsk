package tracing

import "time"

type Tracer interface {
	StartSpan(name string, opts ...StartSpanOption) Span
	Extract(carrier interface{}) (SpanContext, error)
	Inject(ctx SpanContext, carrier interface{}) error
	Stop()
}

type Span interface {
	Context() SpanContext
	Tracer() Tracer
	SetName(name string)
	SetTag(key string, value string)
	Annotate(time time.Time, value string)
	Finish(opts ...FinishOption)
	Flush()
}

type SpanContext interface {
	SpanID() interface{}
	TraceID() interface{}
}
