# http二次开发

## 客户端
client封装了http操作,可以通过反射自动解析数据,支持retry

## 服务器
服务器基于[echo](https://github.com/labstack/echo),
与echo不同的是,做了一些减法,去除了对其他库的依赖,去除了autocert,去除了log,将来也许可以做进一步简化,
这里希望实现的并不是一个完善的web服务器,而是仅仅希望增强一些标准库,用于Gateway对api转换
