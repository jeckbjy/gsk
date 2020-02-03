# Go Server Kit(GO服务器开发全家桶)

## 目标

开发本库的一个终极目标是要使其能够快速搭建一个新的后端项目,它需要提供后端开发常见的底层功能,比如异步socket,异步rpc,服务器注册发现,线程模型,配置文件,定时器,测试框架,orm,cli,log,bi等功能,因为它最终面向的是业务开发,因此本库聚合很多开源框架,它不是针对的某一特定领域的库,同时该库也不该包含具体的业务逻辑,他的定位应该是通用的服务器开发中间件。

## 简介

受[go-micro](https://github.com/micro/go-micro)启发，本库想实现一个高效,灵活,丰富的异步RPC游戏微服务框架,使用场景上更倾向于高性能。

- 与go-micro最大的不同是,网络底层的改造,这里的网络底层更类似Netty。
  - 默认的网络是tcp,也更倾向于支持长连接服务器
  - 所有的底层操作都默认是异步操作(Read,Write),上层RPC调用也是异步操作,可以通过Future转为同步调用
  - 更加灵活的FilterChain设计,类似netty
  - 更加灵活的线程模型,通常也需要定制,比如根据玩家UID哈希到不同的消息队列中
  - 更加灵活的编码协议,消息体可以是json,protobuf等编码格式,消息头使用私有协议,rpc通常使用URL来查找回调函数,而游戏通常只使用一个消息ID查找消息回调

- 与gRPC的不同
  - 这里照搬了go-micro的一些设计,提供了更多微服务必备的组件,比如服务发现,消息队列等
  - 协议上的差异,gRPC使用的是http2,这里使用的是tcp

## 设计准则

- 异步网络,高性能
  - 默认提供基于原生tcp+读写goroutine的实现方案
  - 尝试性提供nio方案,更加高效,更节省内存
- 接口设计,易扩展
- 开箱即用,易上手
  - 每个接口都会提供一个默认实现,但不一定最高效,需要通过Plugin来提供高效的实现
- 轻度依赖,易集成
  - 原则上尽量只依赖标准库,对于一些小的第三方库,直接集成在代码中,比较庞大的第三方库尽量放到plugin中实现

## 核心模块

- **anet** 异步网络底层(asynchronous network),参考netty
- **arpc** 异步RPC框架(asynchronous remote procedure call),主要功能,私有协议定义,消息粘包处理,Handler注册,消息路由,RPC调用等功能
- **apm**  性能监控(Application Performance Management),主要功能:log,bi,熔断器,链路追踪等库的封装
- **registry** 服务注册与发现,生产环境可以使用etcd,consul,zookeeper等
- **selector** 客户端Load Balance
- **exec** 业务线程控制,根据不同的业务场景使用不同的线程模型
- **frame** 消息粘包处理
- **broker** 消息队列接口(TODO)
- **store** kv存储
- **orm** 封装数据库CRUD操作,仅限于单表操作,不支持join,aggregate等复杂操作
- **util** 收集了一些常用的辅助库,比如buffer,cache,errors,str,idgen,dsn,定时器等常用功能

## 示例代码

```go
package gsk

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/jeckbjy/gsk/arpc"
)

type echoReq struct {
	Text string `json:"text"`
}

type echoRsp struct {
	Text string `json:"text"`
}

func TestRPC(t *testing.T) {
	// 启动服务器
	name := "echo"
	srv := New(name)
	// 注册回调: register callback
	if err := srv.Register(func(ctx arpc.Context, req *echoReq, rsp *echoRsp) error {
		log.Printf("[server] recv msg,%+v\n", req.Text)
		rsp.Text = fmt.Sprintf("%s world", req.Text)
		return nil
	}); err != nil {
		log.Fatal(err)
	}

	go srv.Run()

	log.Printf("wait for server start")
	time.Sleep(time.Millisecond * 20)

	// 同步RPC调用 synchronous call
	log.Printf("[client] try call sync")
	rsp := &echoRsp{}
	if err := srv.Call(name, &echoReq{Text: "sync hello"}, rsp); err != nil {
		t.Fatal(err)
	} else {
		log.Printf("[client] recv response,%s", rsp.Text)
	}

	// 异步RPC调用 asynchronous call
	log.Printf("[client] try call async")
	err := srv.Call(name, &echoReq{Text: "async hello"}, func(rsp *echoRsp) error {
		log.Printf("[client] recv response,%s", rsp.Text)
		return nil
	})

	if err != nil {
		t.Fatal(err)
	} else {
		log.Printf("[client] async call ok")
	}

	time.Sleep(time.Second * 2)
	srv.Exit()
	t.Log("finish")
}

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
- [gommon](https://github.com/labstack/gommon)
- [dateparse](https://github.com/araddon/dateparse)

## 其他资料

- [Functional Options Pattern in Go](https://halls-of-valhalla.org/beta/articles/functional-options-pattern-in-go,54/)
- [Pattern](https://www.jianshu.com/p/5a3a09894bb5)
- [GoPatterns](https://books.studygolang.com/go-patterns/)
- [Go推荐的工程结构](https://github.com/golang-standards/project-layout)