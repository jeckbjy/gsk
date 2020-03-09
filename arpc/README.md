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
- selector优化,更加丰富的策略以及失败处理
