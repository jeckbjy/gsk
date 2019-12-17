// asynchronous network,support tcp,kcp,websocket, etc.
package anet

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"

	"github.com/jeckbjy/gsk/util/buffer"
)

var defaultFunc atomic.Value

func Default() NewTranFunc {
	return defaultFunc.Load().(NewTranFunc)
}

func SetDefault(fn NewTranFunc) {
	defaultFunc.Store(fn)
}

type NewTranFunc func() Tran

var (
	ErrHasOpened = errors.New("conn has opened")
	ErrHasClosed = errors.New("conn has closed")
	//ErrNotDialer = errors.New("is not dialer")
)

type Status int

const (
	CONNECTING = Status(iota)
	OPEN
	CLOSING
	CLOSED
)

// Tran 创建Conn,可以是tcp,websocket等协议
// 不同的Tran可以配置不同的FilterChain
// 配置信息只能初始化时创建,非线程安全
type Tran interface {
	String() string
	GetChain() FilterChain
	SetChain(chain FilterChain)
	AddFilters(filters ...Filter)
	Dial(addr string, opts ...DialOption) (Conn, error)
	Listen(addr string, opts ...ListenOption) (Listener, error)
	Close() error
}

// Conn 异步收发消息
type Conn interface {
	Tran() Tran                      // Transport
	Tag() string                     // 额外标识类型
	Get(key string) interface{}      // 获取自定义数据
	Set(key string, val interface{}) // 设置自定义数据
	Status() Status                  // 当前状态
	IsActive() bool                  // 是否已经建立好连接
	IsDial() bool                    // 是否通过Dial建立的连接
	LocalAddr() string               // 本地地址
	RemoteAddr() string              // 远程地址
	ReadLocker() sync.Locker         // 读数据锁,通常都在读协程中处理,并不需要加锁
	Read() *buffer.Buffer            // 异步读缓存,线程安全
	Write(data *buffer.Buffer) error // 异步写数据,线程安全
	Send(msg interface{}) error      // 异步发消息,会调用HandleWrite,没有连接成功时也可以发送,当连接成功后会自动发送缓存数据
	Close() error                    // 调用后将不再接收任何读写操作,并等待所有发送完成后再安全关闭
}

type Listener interface {
	Close() error
	Addr() net.Addr
}

// Filter 用于链式处理Conn各种回调
// InBound: 从前向后执行,包括Read,Open,Error
// OutBound:从后向前执行,包括Write,Close
type Filter interface {
	Name() string
	HandleRead(ctx FilterCtx) error
	HandleWrite(ctx FilterCtx) error
	HandleOpen(ctx FilterCtx) error
	HandleClose(ctx FilterCtx) error
	HandleError(ctx FilterCtx) error
}

// FilterCtx Filter上下文，默认会自动调用Next,如需终止，需要主动调用Abort
type FilterCtx interface {
	Conn() Conn               // Socket Connection
	Data() interface{}        // 获取数据
	SetData(data interface{}) // 设置数据
	Error() error             // 错误信息
	SetError(err error)       // 设置错误信息
	IsAbort() bool            // 是否已经强制终止
	Abort()                   // 终止调用
	Next() error              // 调用下一个
	Jump(index int) error     // 跳转到指定位置,可以是负索引
	JumpBy(name string) error // 通过名字跳转
	Clone() FilterCtx         // 拷贝当前状态,可用于转移到其他协程中继续执行
	Call() error              // 开始执行,执行完成后会释放FilterCtx
}

// FilterChain 管理Filter,并链式调用所有Filter
// Filter分为Inbound和Outbound
// InBound: 从前向后执行,包括Read,Open,Error
// OutBound:从后向前执行,包括Write,Close
type FilterChain interface {
	Len() int                   // 长度
	Front() Filter              // 第一个
	Back() Filter               // 最后一个
	Get(index int) Filter       // 通过索引获取filter
	Index(name string) int      // 通过名字查询索引
	AddFirst(filters ...Filter) // 在前边插入
	AddLast(filters ...Filter)  // 在末尾插入
	HandleOpen(conn Conn)
	HandleClose(conn Conn)
	HandleRead(conn Conn, msg interface{})
	HandleWrite(conn Conn, msg interface{})
	HandleError(conn Conn, err error)
}
