# Go Server Kit(GO服务器开发全家桶)

## 简介
受[go-micro](https://github.com/micro/go-micro)启发，本库想实现一个高效,灵活,丰富的异步RPC游戏微服务框架,使用场景上更倾向于高性能游戏服务器。  
与go-micro最大的不同是,网络底层的改造,这里的网络底层更类似Netty。

* 默认的网络是tcp,也更倾向于支持长连接服务器
* 所有的底层操作都默认是异步操作(Read,Write),上层RPC调用也是异步操作,可以通过Future转为同步调用
* 更加灵活的FilterChain设计,类似netty
* 更加灵活的线程模型,通常也需要定制,比如根据玩家UID哈希到不同的消息队列中
* 更加灵活的编解格式,rpc通常使用方法名查找回调函数,而游戏服务器通常只使用一个消息ID查找消息回调

与gRPC的不同
* 这里照搬了go-micro的一些设计,提供了更多微服务必备的组件,比如服务发现,消息队列等,
* 协议上的差异,gRPC使用的是http2,这里使用的是tcp

## 特性与设计准则
* 接口设计,易于扩展
* 异步网络,高性能
* 开箱即用,易上手
  - 每个接口都会提供一个默认实现,但不一定最高效
* 轻度依赖,易集成
  - 原则上尽量只依赖标准库,对于一些小的第三方库,直接集成在代码中,比较庞大的第三方库尽量放到plugin中实现,
  目前额外依赖了golang.org/x/net/ipv4

## 核心模块
- **anet** 异步网络底层(asynchronous network),参考netty
- **arpc** 异步RPC框架(asynchronous remote procedure call),使用私有协议,消息路由
- **apm**  性能监控(Application Performance Management)
- **registry** 服务注册与发现,生产环境可以使用etcd,consul,zookeeper等
- **selector** 客户端Load Balance
- **broker** 消息队列接口
- **store** kv存储
- **db** 封装简单orm

## 示例代码

## 依赖
- [x/net](https://golang.org/x/net/ipv4)  
  当无法下载时,可以使用下边的方法:
``` go
mkdir -p $GOPATH/src/golang.org/x
cd $GOPATH/src/golang.org/x
git clone https://github.com/golang/net.git
```

## 集成或参考的第三方库
- [go-micro](https://github.com/micro/go-micro)
- [backoff](https://github.com/cenkalti/backoff)
- [backoff](https://github.com/rfyiamcool/backoff)
- [shortid](https://github.com/teris-io/shortid)
- [xid](https://github.com/rs/xid)
- [hashstructure](https://github.com/mitchellh/hashstructure)
- [go-ssdp](https://github.com/koron/go-ssdp)
- [smudge](https://github.com/clockworksoul/smudge)
- [mergo](https://github.com/imdario/mergo)
- [base58](https://github.com/mr-tron/base58)
- [fsnotify](https://github.com/fsnotify/fsnotify)

## 其他资料
- [Functional Options Pattern in Go](https://halls-of-valhalla.org/beta/articles/functional-options-pattern-in-go,54/)
- [Pattern](https://www.jianshu.com/p/5a3a09894bb5)
- [GoPatterns](https://books.studygolang.com/go-patterns/)