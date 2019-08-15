// asynchronous network,support tcp,kcp,websocket, etc.
package anet

import (
	"net"

	"github.com/jeckbjy/gsk/util/buffer"
)

var Default CreateTranFunc
var gTranFuncMap = make(map[string]CreateTranFunc)

func Add(name string, fn CreateTranFunc) {
	gTranFuncMap[name] = fn
	Default = fn
}

func New(name string) ITran {
	if fn, ok := gTranFuncMap[name]; ok {
		return fn()
	}

	return nil
}

// NewDefault 新建一个默认的Transport
func NewDefault() ITran {
	return Default()
}

type CreateTranFunc func() ITran

// ITran 创建IConn,可以是tcp,kcp，websocket等协议
// 不同的Tran可以配置不同的FilterChain
type ITran interface {
	String() string
	GetChain() IFilterChain
	SetChain(chain IFilterChain)
	AddFilters(filters ...IFilter)
	Dial(addr string, opts ...DialOption) (IConn, error)
	Listen(addr string, opts ...ListenOption) (IListener, error)
	Close() error
}

type Status int

const (
	DISCONNECTED = Status(iota)
	CONNECTED
	CLOSED
	RECONNECTING
	CONNECTING
)

// IConn 异步收发Socket
type IConn interface {
	Status() Status                  // Socket状态
	LocalAddr() net.Addr             // 本地地址
	RemoteAddr() net.Addr            // 远程地址
	Read() *buffer.Buffer            // 异步读缓存
	Write(data *buffer.Buffer) error // 异步写数据,线程安全
	// 异步发消息,会调用HandleWrite,没有连接成功时也可以发送,当连接成功后会自动发送缓存数据
	Send(msg interface{}) error
	Close() error
}

type IListener interface {
	Close() error
	Addr() net.Addr
}

// IFilter 用于链式处理IConn各种回调
type IFilter interface {
	Name() string
	HandleRead(ctx IFilterCtx)
	HandleWrite(ctx IFilterCtx)
	HandleOpen(ctx IFilterCtx)
	HandleClose(ctx IFilterCtx)
	HandleError(ctx IFilterCtx)
}

// IFilterCtx Filter上下文，默认会自动调用Next,如需终止，需要主动调用Abort
type IFilterCtx interface {
	Conn() IConn              // Socket Connection
	Data() interface{}        // 获取数据
	SetData(data interface{}) // 设置数据
	Error() error             // 错误信息
	SetError(err error)       // 设置错误信息
	IsAbort() bool            // 是否已经强制终止
	Abort(err error)          // 强制终止调用,如果err不为nil则会触发HandleError
	Next()                    // 调用下一个
	Jump(index int) error     // 跳转到指定位置,可以是负索引
	JumpBy(name string) error // 通过名字跳转
	Clone() IFilterCtx        // 拷贝当前状态,可用于转移到其他协程中继续执行
	Call()                    // 从当前位置开始执行
}

// IFilterChain 管理IFilter,并链式调用所有IFilter
// IFilter分为Inbound和Outbound
// InBound: 从前向后执行,包括Read,Open,Error
// OutBound:从后向前执行,包括Write,Close
type IFilterChain interface {
	Len() int                    // 长度
	Front() IFilter              // 第一个
	Back() IFilter               // 最后一个
	Get(index int) IFilter       // 通过索引获取filter
	Index(name string) int       // 通过名字查询索引
	AddFirst(filters ...IFilter) // 在前边插入
	AddLast(filters ...IFilter)  // 在末尾插入
	HandleOpen(conn IConn)
	HandleClose(conn IConn)
	HandleRead(conn IConn, msg interface{})
	HandleWrite(conn IConn, msg interface{})
	HandleError(conn IConn, err error)
}
