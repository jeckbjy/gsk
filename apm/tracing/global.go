package tracing

import "sync"

var (
	mux    sync.RWMutex
	tracer Tracer = &noopTracer{}
)

func SetTracer(t Tracer) {
	mux.Lock()
	defer mux.Unlock()
	tracer.Stop()
	tracer = t
}

func GetTracer() Tracer {
	mux.RLock()
	defer mux.RUnlock()
	return tracer
}

func StartSpan(name string, opts ...StartSpanOption) Span {
	return tracer.StartSpan(name, opts...)
}

func Extract(carrier interface{}) (SpanContext, error) {
	return tracer.Extract(carrier)
}

func Inject(ctx SpanContext, carrier interface{}) error {
	return tracer.Inject(ctx, carrier)
}

func Stop() {
	tracer.Stop()
}
