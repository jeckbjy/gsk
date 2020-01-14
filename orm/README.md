# database orm
用于抽象database操作，数据库需要支持常见sql和mongodb，以及常见的kv数据库,功能上需要支持CRUD，索引操作
api接口上跟mongodb更相似，该库并不能取代原生的db驱动，不能操作复杂命令,不能连表操作,但是能满足大部分常见需求

TODO:目前只是粗略的实现了一下sql,还需要根据具体引擎(mysql,postgres,sqlite)定制,尤其是创建index,schema
实现上可以判断Model是否提供Indexes接口来自动创建index,或者通过tag自动添加

## 参考
- https://github.com/jinzhu/gorm
- https://github.com/gostor/awesome-go-storage  
- https://github.com/timshannon/bolthold  
- https://github.com/abronan/valkeyrie  
- https://github.com/tidwall/buntdb  
