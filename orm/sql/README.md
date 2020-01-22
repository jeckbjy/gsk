# sql的orm实现

## 需要支持的一些特性
- 自动创建Model
    - 自动创建Table
    - 自动创建Index
    - 自动Migrate(仅限新增),
- 多数据库支持
    - 有一些场景我们希望根据项目分离到不同的数据中操作,以实现隔离,减少相互干扰
- 功能简单,语义明确
    - 使用过gorm的同学都会有一个感受,不使用debug的模式下,完全不知道他翻译成了什么
- API安全
    - gorm使用的是链式调用方式,对于一个新手而言,他很容易用错且致命的问题,调用了Find之后,后边的条件都不会生效。  
比如: `select * from orders where id=1 limit 1` 这样一条简单的sql语句,
如果代码中写成:db.Find(order).Where("id=?", orderId).Limit(1),这里看着还比较符合sql的顺序,
但是最终却翻译成了`select* from orders`,直接忽略了后边的限制条件,而且不会报错,
这种操作很致命还很难发现,因此不建议使用gorm这样的库,不如直接使用原生sql

## 限制
- 连接字符串仅支持URI的连接方式,比如:mysql://user:pass@localhost/dbname,
因为这样方便统一使用,具体使用方式请见[dburl](https://github.com/xo/dburl)

## 安装依赖包

- 官方所有驱动:https://github.com/golang/go/wiki/SQLDrivers
- mysql:  go get github.com/go-sql-driver/mysql
- pg:     go get github.com/lib/pq

## docker相关命令
- 查看镜像: docker container list
- 停止镜像: docker stop [container-id]

## mysql
- 启动: docker run -p 3306:3306 --name mysql -e MYSQL_ROOT_PASSWORD=123456 -d mysql

## postgres
- 启动:docker run -p 5432:5432  --name pg -e POSTGRES_PASSWORD=123456 -d postgres
- 连接:psql -U postgres -W -h localhost -d postgres -p 5432
- 安装:https://stackoverflow.com/questions/44654216/correct-way-to-install-psql-without-full-postgres-on-macos
- 常用命令:
    - 查看数据: SELECT datname FROM pg_database;
    - 切换数据库: \c test;
    
## 其他
- [DSN](https://pear.php.net/manual/en/package.database.db.intro-dsn.php)