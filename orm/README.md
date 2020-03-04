# database orm

用于抽象database操作，数据库需要支持常见sql和mongodb，以及常见的kv数据库,功能上需要支持CRUD，索引操作
api接口上跟mongodb更相似，该库并不能取代原生的db驱动，不能操作复杂命令(如Aggregate),不能连表操作,
但是能满足大部分常见单表操作

## 参考

- https://github.com/jinzhu/gorm
- https://github.com/go-gormigrate/gormigrate
- https://github.com/gostor/awesome-go-storage  
- https://github.com/timshannon/bolthold  
- https://github.com/abronan/valkeyrie  
- https://github.com/tidwall/buntdb  

## 注意

- 尚未仔细实现测试细节
