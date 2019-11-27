# arpc asynchronous rpc
异步RPC通信框架,主要功能:  
- 1:消息的粘包,以及编解码处理,使用私有协议
- 2:消息路由的注册
- 3:服务注册与发现
- 4:客户端RPC调用
    - 消息重传
    - 客户端Load Balancer
    - 同步调用,异步调用支持
