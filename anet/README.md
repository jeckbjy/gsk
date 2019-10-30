# anet asynchronous network
- 异步网络库,设计上类似netty,mina,支持FilterChain设计,默认使用tcp,支持websocket扩展
- nio是一个试验性功能,用于使用epoll,kqueue替代goroutine,减少内存使用