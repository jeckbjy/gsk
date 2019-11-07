# 定时任务
cron应该支持两种模式,单机和分布式,并且这两种模式是递进关系
- 单机模式[standalone]
    - 需要支持配置解析,配置应该支持两种模式,标准的cron格式和类似google的HumanFriendly格式
    - 需要支持任务调度
- 分布式模式[cluster模式]
    - 需要一个中心调度器,用于任务调度，节点注册
    - 需要支持任务状态管理,失败后可重新调度(如果允许的话)
    - 需要支持任务分片,一个复杂任务可以通过制定分片数hash到不同的机器上执行
    - worker节点注册时,需要支持筛选任务,比如某些节点只支持类型为A的任务,某些节点只支持类型为B类型的任务,再复杂一点,支持正则匹配??

## 参考
- https://cloud.google.com/appengine/docs/standard/java/config/cronref#start-time  
- https://github.com/shyiko/skedule
- https://github.com/singchia/go-timer
- Quartz
