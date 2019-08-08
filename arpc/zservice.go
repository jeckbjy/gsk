// asynchronous rpc
package arpc

import (
	"errors"
	"github.com/jeckbjy/micro/anet"
	"github.com/jeckbjy/micro/codec"
	"github.com/jeckbjy/micro/util/buffer"
)

var (
	ErrTimeout   = errors.New("rpc timeout")
	ErrNoHandler = errors.New("no handler")
	ErrNoNode    = errors.New("no node found")
	ErrNoConn    = errors.New("no conn found")
	ErrRetryStop = errors.New("retry stop")
)

// PacketCreator 新建一个Packet
type PacketCreator func() IPacket

// IPacket 用于服务器间通信,相当于IRequest和IResponse
// ID:表示唯一数值型ID,Name表示消息唯一名字,这两个值通常用于游戏中使用ID查找处理函数
// Service:RPC服务名,用于服务查询,Method:RPC方法名,用于通过函数名查询消息处理回调
// SeqID:用于异步RPC全局唯一ID,通常用于查询异步处理回调函数
// ID,Name,Method并不需要都存在,使用任意一个都可以查找到回调处理函数,但运行效率和注册方便程度上有所不同
// ID的方式节省空间，速度快，但需要额外的注册
// Name,Method的方式可以通过反射的方式，在注册消息处理的时候，自动解析出函数名和消息名
type IPacket interface {
	SeqID() string           // 异步RPC全局sequence id
	Reply() bool             // 标识是否是RPC应答请求
	ID() int                 // 消息ID,要求是正整数
	Name() string            // 消息名
	Method() string          // RPC方法名
	Service() string         // RPC服务名
	Head() map[string]string // 其他消息头信息
	Body() interface{}       // 解析后的消息体
	Error() string           // 错误信息
	Codec() codec.ICodec     // 消息体编解码
	Bytes() *buffer.Buffer   // 原始数据
	Value(key string) string // 查询数据
	// set操作
	SetSeqID(string)
	SetReply(bool)
	SetID(int)
	SetName(string)
	SetMethod(string)
	SetService(string)
	SetBody(interface{})
	SetError(err string)
	SetCodec(codec.ICodec)
	SetBytes(*buffer.Buffer)
	SetHead(map[string]string)
	SetValue(key string, value string)

	// 解析消息体
	ParseBody(msg interface{}) error

	// 消息包编解码
	Encode(b *buffer.Buffer) error
	Decode(b *buffer.Buffer) error
}

// 消息通信上约定一个Request必须至少返回一个Response,多于一个Rsp的需要手动send消息
// 这样限制在实现Proxy时,可以做一个队列，等到上一个消息处理完再发送下一个消息
// 如果服务器没有返回Response,则服务器会将Request的头信息作为默认返回
type IContext interface {
	Reset()                     // 重置数据
	Request() IPacket           // 消息请求
	Response() IPacket          // 消息应答
	Conn() anet.IConn           // 原始Socket
	Send(msg interface{}) error // 方便直接调用
	Handler() Handler           // 返回Handler
	SetHandler(h Handler)       // 设置Handler
}

// Handler 消息回调处理函数
type Handler func(ctx IContext) error

// Middleware 中间件
type Middleware func(Handler) Handler

// IRouter 消息处理路由,需要支持服务器响应和客户端RPC响应两种形式:
// 1:服务端消息响应,这个比较简单,一般服务器启动时注册好消息回调就可以了,
// 	 可以通过消息ID,消息名,或者调用方法名查找到对应的消息回调
// 2:客户端RPC响应,这需要支持两种模式:
//   a:调用时传入的是一个函数,也就是简单的异步调用
//   b:调用时传入的是一个类指针和Future,这
type IRouter interface {
	// Register:可用于注册消息回调,也可以用于注册消息解析,或者注册RPC回调
	// 1:注册函数,即消息回调,同时也会自动反射解析需要的消息名
	// 3:注册服务,反射所有符合条件的函数,用REGISTER_SRV标识
	// 支持的函数原型有:
	// 原型1:func(ctx IContext)
	// 原型2:func(ctx IContext, req *Request) error
	// 原型3:func(ctx IContext, req *Request, rsp *Response) error
	// 原型1调用不需要反射,效率最高,但是则需要额外手动注册消息解析
	// 原型2,3使用了反射调用,应该会有额外调用消耗,可以通过方法名或者消息名找到消息回调
	RegisterSrv(srv interface{}, opts ...MiscOption) error
	// 注册消息,需要消息原型,消息名(通过反射获得),消息ID(可选)
	RegisterMsg(msg interface{}, opts ...MiscOption) error
	// 注册RPC,同步调用rsp是一个消息指针,异步调用是一个回调函数,原型是func(ctx IContext, rsp *XXResponse)
	RegisterRpc(rsp interface{}, opts *RegisterRpcOptions) error
	// Find 根据消息包,查询消息回调,可能是服务端响应,也可能是客户端RPC响应
	Find(pkg IPacket) (Handler, error)
	// 通过名字查询ID,返回0表示没有找到
	FindID(name string) int
}

// IExecutor 消息执行器,用于设定消息处理线程模型,几种常见的处理模型:
// 单线程:所有消息在一个消息队列中处理
// 多线程:每个消息都在单独的协程中处理
// Hash:根据uid等hash到不同的消息队列中处理
// 同时IExecutor还需要支持middleware功能,便于注入
type IExecutor interface {
	// Use 添加中间件,不能动态修改,只能初始化时设置
	Use(middleware ...Middleware)
	// 执行消息
	Handle(ctx IContext)
}

// IServer 服务器端
type IServer interface {
	Options() ServerOptions
	Init(opts ...ServerOption)
	Start() error
	Stop() error
	Run() error
}

// IClient 异步调用RPC服务
// 需要能监听多个服务
// 需要支持各种LoadBalance算法,比如RoundRobin,Random,Weight等
// 需要支持消息失败重传
// 需要支持异步RPC调用
// Send和Call的区别,Send仅仅发送一条消息,不关心结果,Call则是发起RPC调用,通常是需要返回消息的
type IClient interface {
	Options() ClientOptions
	Init(opts ...ClientOption)
	// 发送一条消息,不关心返回结果
	Send(service string, msg interface{}, opts ...MiscOption) error
	// 异步RPC调用,也可以通过IFuture实现同步效果,调用有几种常见形式:
	// 1:异步调用,传入callback函数,函数原型是:func(ctx IContext, rsp Response)
	// 2:同步调用,传入消息和一个Future,外部使用Future的Wait阻塞调用
	// 3:连续多次调用Call,最后同时阻塞,一个失败则全失败,需要能够同时支持多个服务并发调用,
	//   可以二次封装次函数,批次调用RPC,最终调用一个Callback
	//   比如:
	//   f := arpc.NewFuture()
	//   rsp1 := &Message1{}
	//   c.Call("srv.a", req1, rsp1)
	//   rsp2 := &Message2{}
	//   c.Call("srv.b",req2, rsp2)
	//   f.Wait() 或者起一个goroutine中调用Wait()
	Call(service string, req interface{}, rsp interface{}, opts ...MiscOption) error
}

// IFuture 用于异步RPC调用时,阻塞当前调用
// 需要能够支持同时多个,其中任意一次调用失败则全失败
type IFuture interface {
	Add(delta int) error
	Done() error
	Fail(err error) error
	Wait() error
}
