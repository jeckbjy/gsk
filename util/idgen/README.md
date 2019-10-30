# ID生成器
这里提供了几种ID生成方法

## [xid](https://github.com/rs/xid)
类似mongodb的产生方式,零配置，ID有序,二进制占用12个bytes,字符串占用20个char

## [tid]
基于时间戳和自增序列的id生成器,类似于snowflake,vesta-id-generator,生成的ID是一个uint64的整数,
支持uint64,支持秒,毫秒两种模式,集群需要配置nodeId才能避免重复,毫秒模式需要注意时间回调会导致ID重复而报错
id粗略有序

## [sid]
shortid,生成的ID是字符串格式,无序并随机,不可以反解析,相比较teris-io/shortid,这里支持了base32,所有字段都参与随机
参考:
[shortid](https://github.com/teris-io/shortid)
[go-shortid](https://github.com/skahack/go-shortid)

## 其他算法
- [luhn](https://www.geeksforgeeks.org/luhn-algorithm/) 模10校验算法
