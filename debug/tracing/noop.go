package tracing

import "time"

type noopTracer struct{}

func (noopTracer) StartSpan(name string, opts ...StartSpanOption) Span {
	return noopSpan{}
}

func (noopTracer) Extract(carrier interface{}) (SpanContext, error) {
	return noopSpanContext{}, nil
}

func (noopTracer) Inject(ctx SpanContext, carrier interface{}) error {
	return nil
}

func (noopTracer) Stop() {
}

type noopSpan struct{}

func (n noopSpan) Context() SpanContext {
	return noopSpanContext{}
}

func (n noopSpan) Tracer() Tracer {
	return nil
}

func (n noopSpan) SetName(name string) {
}

func (n noopSpan) SetTag(key string, value string) {
}

func (n noopSpan) Annotate(time time.Time, value string) {
}

func (n noopSpan) Finish(opts ...FinishOption) {
}

func (n noopSpan) Flush() {
}

//
type noopSpanContext struct{}

func (context noopSpanContext) SpanID() interface{} {
	return 0
}

func (context noopSpanContext) TraceID() interface{} {
	return 0
}
