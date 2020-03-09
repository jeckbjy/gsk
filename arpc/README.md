# arpc asynchronous rpc

异步RPC通信框架,主要功能

- 1:消息的粘包,以及编解码处理,使用私有协议
- 2:消息路由的注册
- 3:服务注册与发现
- 4:客户端RPC调用
  - 消息重传
  - 客户端Load Balancer
  - 同步调用,异步调用支持

## TODO

- 需要提供一种机制,保证底层socket读写失败能通知上层处理异常
- RPC Call时,超时不再是必须参数,需要底层能通过socket断开连接时,主动触发rpc失败回调
- Retry机制梳理
- registry支持Namespace,Zone等信息,Namespace可用于支持多环境,Zone可用于支持多区域,客户端选举时,可优先选举相同区域的,相同区域不存在再选择其他区域,以达到异地多活的效果
- reigstry是否需要支持鉴权,如何支持?
- registry改名为naming service?
- selector优化,更加丰富的策略以及失败处理
- 优雅下线

## 其他资料

- [聊聊微服务的服务注册与发现](http://jm.taobao.org/2018/06/26/%E8%81%8A%E8%81%8A%E5%BE%AE%E6%9C%8D%E5%8A%A1%E7%9A%84%E6%9C%8D%E5%8A%A1%E6%B3%A8%E5%86%8C%E4%B8%8E%E5%8F%91%E7%8E%B0/)
- [bilibili discovery](https://github.com/bilibili/discovery)