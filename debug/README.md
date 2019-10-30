# debug
用于服务器调试，分为四个部分，log,metrics,tracing,panic

## alog 异步日志输出
- 异步操作
- 输出支持插件扩展，默认支持terminal,simple file
- 支持格式化

## metrics 指标监控,需要支持常见的tsdb,比如Prometheus,InfluxDB,OpenTSDB等
https://github.com/prometheus/client_golang
https://github.com/uber-go/tally
https://github.com/rcrowley/go-metrics

## tracing 调用链路追踪
常见的库:OpenTracer,zipkin,Jaeger

https://github.com/DataDog/dd-trace-go/blob/v1/ddtrace/opentracer/span.go  
https://github.com/jaegertracing/jaeger-client-go

## panics 异常拦截并上报  
