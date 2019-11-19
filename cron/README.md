# 分布式定时任务设计
- 概述

分布式定时任务需要能够在集群中进行任务调度，任务状态管理，任务故障转移重试恢复，
能够对任务进行分片加速并管理状态,能够进行简单的任务编排(DAG可能有点复杂，可以简单的线性编排)

- 任务调度
    - 需要一个中心调度器,用于任务调度,状态管理,工作节点注册发现
    - 工作节点启动后需要主动注册自己信息,能够处理哪种类型的任务,运行时需要定时上报自己状态,比如cpu,内存,处理任务数量
    
- 任务管理
    - 一个新任务创建后,需要持久化到数据(etcd,zookeeper)
    - 一个任务需要有唯一ID,任务类型,任务当前状态(进行中,恢复中,失败,成功,废弃),任务分为几步作业,当前进行到第几步,任何一步失败则全部失败
    - 一个任务可以分为多个步骤,每个步骤可以sharding分片,每个分片状态都需要进行管理
    
- 任务配置
    - 是否开启
    - 任务描述
    - 任务类型
    - 任务透传参数
    - 任务启动时间
    - 任务是否可重入
    - 最大重试次数
    - 任务过期时间,超过此时间将不再尝试恢复
    
## 一些库
- [cron](https://github.com/robfig/cron)
- [dcron](https://github.com/LibiChai/dcron)

## 参考
- https://cloud.google.com/appengine/docs/standard/java/config/cronref#start-time  
- https://github.com/shyiko/skedule
- https://github.com/singchia/go-timer
- Quartz
