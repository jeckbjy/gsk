# broker
消息队列抽象接口设计  
常用的消息队列支持一下几种模式
- 点对点模式:一个消息只能由一个消费者处理,可以有一个或多个消费者竞争处理
- Pub/Sub模式:一个消息可以被所有的消费者处理
- 复合模式:先是Pub/Sub，然后再分组，这样可以增加并发处理能力
- 有些支持消息路由
- 有些消息队列支持事务,但是大部分是不支持的
- API设计上如何支持多个订阅?是需要每个订阅一个goroutine么


- 应用场景:https://www.alibabacloud.com/help/zh/doc-detail/112010.htm?spm=a2c63.p38356.879954.22.18fefc6curWPID
- https://www.alibabacloud.com/help/zh/doc-detail/29532.htm?spm=a2c63.p38356.b99.2.5b906513XS0aJR
- https://www.cnblogs.com/hzmark/p/orderly_message.html