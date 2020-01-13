# cron 定时任务
区别于timer,cron更注重定时任务格式的解析,因此cron不能用于大规模定时器  
定时任务的格式分为两种：  
一:标准格式，例如:30 * * * * 表示每半个小时执行一次  
二:human-friendly,例如:every monday 09:00  

分布式定时任务是另外一个复杂的话题,在单机cron的基础之上,
还需要对任务进行管理,调度,sharding分拆,恢复,重试,通知等复杂操作

TODO: 目前仅仅是集成了robfig/cron,未来需要扩展cron支持human-friendly模式的解析

## 参考库
- [cron](https://github.com/robfig/cron)
- [dcron](https://github.com/LibiChai/dcron)
- [skedule](https://github.com/shyiko/skedule)
- [dolphinscheduler](https://dolphinscheduler.apache.org/zh-cn/)

## 参考文档
- [Google](https://cloud.google.com/appengine/docs/standard/java/config/cronref#schedule_format)
- [Quartz]