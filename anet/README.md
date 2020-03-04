# anet asynchronous network

- 异步网络库,设计上类似netty,mina,支持FilterChain设计,默认使用tcp,支持websocket扩展
- nio使用epoll,kqueue替代goroutine,减少内存使用

## 其他参考库

- [easygo](https://github.com/mailru/easygo)
- [gev](https://github.com/Allenxuxu/gev)
- [websocket](https://www.freecodecamp.org/news/million-websockets-and-go-cc58418460bb/)
- [goselect](https://github.com/creack/goselect)