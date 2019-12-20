# apm(Application Performance Management)
用于服务器调试,bug追踪等服务质量监控,log,metrics,tracing,breaker,bi,alert

## alog 异步日志输出
- 异步操作
- 输出支持插件扩展，默认支持terminal,simple file
- 支持格式化

## metrics 指标监控
需要支持常见的tsdb,比如Prometheus,InfluxDB,OpenTSDB等  
https://github.com/prometheus/client_golang  
https://github.com/uber-go/tally  
https://github.com/rcrowley/go-metrics  

## tracing 调用链路追踪
常见的库:OpenTracer,zipkin,Jaeger

https://github.com/DataDog/dd-trace-go/blob/v1/ddtrace/opentracer/span.go    
https://github.com/jaegertracing/jaeger-client-go  

## breaker 熔断器
Circuit Breakers Pattern

## 一些库
- [gopsutil](https://github.com/shirou/gopsutil)
- [ratelimit](https://github.com/uber-go/ratelimit/)
- [gomail](https://github.com/go-gomail/gomail)
- [SendingMail](https://github.com/golang/go/wiki/SendingMail)

## 开源APM
- https://blog.csdn.net/konglongaa/article/details/55807192

## 其他一些资料
- [CurrentMemory](https://golangcode.com/print-the-current-memory-usage/)
- [monitoring](https://scene-si.org/2018/08/06/basic-monitoring-of-go-apps-with-the-runtime-package/)
- [uber-go/ratelimit](https://www.cyhone.com/articles/analysis-of-uber-go-ratelimit/)
- [boomer兼容Locust压力测试](https://github.com/myzhan/boomer)
- [http压力测试](https://github.com/link1st/go-stress-testing)
