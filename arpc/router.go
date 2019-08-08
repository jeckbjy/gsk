package arpc

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/jeckbjy/micro/util/reflects"
	"github.com/jeckbjy/micro/util/times"
)

func NewRouter() IRouter {
	return &Router{
		names:   make(map[string]*rInfo),
		methods: make(map[string]*rInfo),
		replies: make(map[string]*rReply),
	}
}

var (
	ErrInvalidHandler = fmt.Errorf("not handler")
	ErrInvalidSeqID   = fmt.Errorf("invalid sequence id")
	ErrInvalidFuture  = fmt.Errorf("invalid future")
	ErrInvalidService = fmt.Errorf("not find endpoint in service")
	ErrInvalidType    = fmt.Errorf("invalid type")
)

var IDLimit = 4096

// 服务端消息相应信息
type rInfo struct {
	ID      int          // 消息ID
	Name    string       // 消息名
	Method  string       // 方法名
	Handler Handler      // 消息回调
	Req     reflect.Type // 消息请求类型,用于反射创建消息
}

// 客户端异步回调信息
type rReply struct {
	SeqID    string    // 唯一RpcID
	Handler  Handler   // 消息相应,如果传入的是
	Deadline int64     // 超时时间戳
	Check    CheckFunc // 校验是否需要重试
	Retry    RetryFunc // 重试回调函数
}

// IDMap 比较小的ID直接使用数组保存,比较大的id使用map保存,通常不需要map
type rIDMap struct {
	ids []*rInfo
	idm map[int]*rInfo
}

func (m *rIDMap) Get(id int) *rInfo {
	if id < len(m.ids) {
		return m.ids[id]
	} else if m.idm != nil {
		return m.idm[id]
	}

	return nil
}

func (m *rIDMap) Set(id int, info *rInfo) {
	if id < IDLimit {
		if id > len(m.ids) {
			size := int(1.5 * float32(id))
			if size > IDLimit {
				size = IDLimit
			}
			ids := make([]*rInfo, size)
			copy(ids, m.ids)
			m.ids = ids
		}
	} else {
		if m.idm == nil {
			m.idm = make(map[int]*rInfo)
		}
		m.idm[id] = info
	}
}

// Router 消息回调处理,支持几种方式
// 1:消息ID映射,假定ID都比较小,通常不会超过4096,使用数组映射,
//   而且这种模式下,无法通过函数反射获取消息ID,需要额外单独注册
// 2:消息名映射
// 3:方法名映射
// 4:RPC映射
// 查询顺序:rpc->方法名->消息ID->消息名
// TODO:ttl check min heap
// TODO:重新梳理ID,Name,Method重复注册时的映射关系
type Router struct {
	ids     rIDMap             // 消息ID映射
	names   map[string]*rInfo  // 消息名映射
	methods map[string]*rInfo  // 方法名映射
	replies map[string]*rReply // RPC消息应答
	mux     sync.Mutex         // reply锁
}

func (r *Router) FindID(name string) int {
	if info, ok := r.names[name]; ok {
		return info.ID
	}

	return 0
}

func (r *Router) Find(pkg IPacket) (Handler, error) {
	if pkg.SeqID() != "" && pkg.Reply() {
		// process rpc response
		reply := r.findRemoveReply(pkg.SeqID())
		if reply != nil {
			return reply.Handler, nil
		}
		return nil, ErrNoHandler
	}

	// 查询服务端消息回调
	info := r.findInfo(pkg)
	if info != nil && info.Handler != nil {
		return info.Handler, nil
	}

	return nil, ErrNoHandler
}

// findRemoveReply 查找并删除rpc回调
func (r *Router) findRemoveReply(seqID string) *rReply {
	r.mux.Lock()
	defer r.mux.Unlock()
	if reply, ok := r.replies[seqID]; ok {
		delete(r.replies, seqID)
		return reply
	}

	return nil
}

func (r *Router) findInfo(req IPacket) *rInfo {
	if req.Method() != "" {
		if info, ok := r.methods[req.Method()]; ok {
			return info
		}
	}

	if req.ID() != 0 {
		return r.ids.Get(req.ID())
	}

	if req.Name() != "" {
		return r.names[req.Name()]
	}

	return nil
}

func (r *Router) addInfo(o *MiscOptions, info *rInfo) error {
	if info.Req != nil {
		info.Name = info.Req.Name()
	}
	if o.ID != 0 {
		info.ID = 0
	}
	if o.Name != "" {
		info.Name = o.Name
	}
	if o.Method != "" {
		info.Method = o.Method
	}
	// 因为handler的注册和message的注册可能是分开的,所以需要合并老的info
	// 只通过info.Name合并
	// 不会存在method重复的情况
	var old *rInfo
	if info.Name != "" {
		old = r.names[info.Name]
	}

	if old != nil {
		// 两个都注册过
		//if old.Handler != nil && info.Handler != nil {
		//	return fmt.Errorf("duplicate handler:%+v", old.Name)
		//}

		if old.ID != 0 {
			info.ID = old.ID
		}

		if old.Handler != nil {
			info.Handler = old.Handler
		}
	}

	if info.ID != 0 {
		r.ids.Set(info.ID, info)
	}

	if info.Name != "" {
		r.names[info.Name] = info
	}
	if info.Method != "" {
		r.methods[info.Method] = info
	}
	return nil
}

func (r *Router) addReply(reply *rReply) error {
	r.mux.Lock()
	defer r.mux.Unlock()
	r.replies[reply.SeqID] = reply
	return nil
}

func (r *Router) RegisterMsg(msg interface{}, opts ...MiscOption) error {
	o := MiscOptions{}
	o.Init(opts...)

	v := reflect.ValueOf(msg)
	info := &rInfo{Req: v.Type()}
	return r.addInfo(&o, info)
}

func (r *Router) RegisterSrv(srv interface{}, opts ...MiscOption) error {
	o := MiscOptions{}
	o.Init(opts...)

	v := reflect.ValueOf(srv)
	switch v.Kind() {
	case reflect.Ptr:
		return r.addService(&o, v)
	case reflect.Func:
		// 注册消息回调
		v := reflect.ValueOf(srv)
		handler, req := r.extract(v)
		if handler != nil {
			info := &rInfo{Handler: handler, Req: req, Method: reflects.FuncName(v.Pointer())}
			return r.addInfo(&o, info)
		}

		return ErrInvalidHandler
	}

	return nil
}

func (r *Router) RegisterRpc(rsp interface{}, o *RegisterRpcOptions) error {
	if o.SeqID == "" {
		return ErrInvalidSeqID
	}

	var handler Handler
	v := reflect.ValueOf(rsp)
	switch v.Kind() {
	case reflect.Ptr:
		if v.Elem().Kind() != reflect.Struct {
			return fmt.Errorf("bad rpc response type")
		}

		if o.Future == nil {
			return ErrInvalidFuture
		}

		// 传入的是一个结构体指针,暗示这是阻塞调用
		// 需要额外传入一个Future,否则没有意义,外部是没有办法响应消息
		handler = func(ctx IContext) error {
			if err := ctx.Request().ParseBody(rsp); err != nil {
				return err
			}

			// 通知完成
			return o.Future.Done()
		}
	case reflect.Func:
		// 传入的是一个函数回调,暗示这是非阻塞调用,则不需要Future
		// 回调原型是func(ctx IContext, rsp *XXResponse)
		// 校验函数原型
		m := v.Type()
		if m.NumIn() != 2 || !m.In(0).Implements(gTypeCtx) {
			return ErrInvalidHandler
		}

		p1 := m.In(1)
		if p1.Kind() != reflect.Ptr || p1.Elem().Kind() != reflect.Struct {
			return ErrInvalidHandler
		}

		handler = func(ctx IContext) error {
			req := ctx.Request()
			rsp := reflect.New(m.In(1))
			if err := req.Codec().Decode(req.Bytes(), rsp.Interface()); err != nil {
				return err
			}
			// 调用回调
			in := []reflect.Value{reflect.ValueOf(ctx), rsp}
			v.Call(in)
			if o.Future != nil {
				return o.Future.Done()
			}
			return nil
		}
	default:
		return ErrInvalidType
	}

	deadline := times.Now() + int64(o.TTL/time.Millisecond)
	reply := &rReply{SeqID: o.SeqID, Handler: handler, Retry: o.Retry, Deadline: deadline}
	return r.addReply(reply)
}

// 解析Service中符合条件的函数,都没有则认为是注册消息
func (r *Router) addService(o *MiscOptions, v reflect.Value) error {
	var err error
	count := 0
	for i := 0; i < v.NumMethod(); i++ {
		m := v.Type().Method(i)
		if m.PkgPath != "" {
			// unexported
			continue
		}

		handler, req := r.extract(v.Method(i))
		if handler == nil {
			continue
		}

		info := &rInfo{Handler: handler, Req: req, Method: m.Name}
		e := r.addInfo(o, info)
		if e == nil {
			count++
		} else {
			err = e
		}
	}
	if err != nil {
		return err
	}

	if count == 0 {
		return ErrInvalidService
	}

	return nil
}

var (
	gTypeCtx = reflect.TypeOf((*IContext)(nil)).Elem()
	gTypeErr = reflect.TypeOf((*error)(nil)).Elem()
)

// 解析函数:返回Handler和Request类型
// func(ctx IContext) error
// func(ctx IContext, req XRequest) error
// func(ctx IContext, req XRequest, rsp XResponse) error
func (r *Router) extract(v reflect.Value) (handler Handler, reqType reflect.Type) {
	m := v.Type()
	if m.NumIn() < 1 || m.NumIn() > 3 || m.NumOut() != 1 {
		return
	}
	// the first in arg must be IContext
	if !m.In(0).Implements(gTypeCtx) || !m.Out(0).Implements(gTypeErr) {
		return
	}

	switch m.NumIn() {
	case 1:
		handler = v.Interface().(Handler)
	case 2:
		// like: func(ctx IContext, req *XRequest)
		if m.In(1).Kind() != reflect.Ptr {
			return
		}
		reqType = m.In(1).Elem()
		handler = func(ctx IContext) error {
			req := reflect.New(reqType)
			if err := ctx.Request().ParseBody(req); err != nil {
				return err
			}
			in := []reflect.Value{reflect.ValueOf(ctx), req}
			out := v.Call(in)
			return out[0].Interface().(error)
		}
	case 3:
		// like: func(ctx IContext, req *XRequest, rsp *XResponse) error
		if m.In(1).Kind() != reflect.Ptr || m.In(2).Kind() != reflect.Ptr {
			return
		}

		reqType = m.In(1).Elem()
		rspType := m.In(2).Elem()
		handler = func(ctx IContext) error {
			req := reflect.New(reqType)
			if err := ctx.Request().ParseBody(req); err != nil {
				return err
			}

			rsp := reflect.New(rspType)
			ctx.Response().SetBody(rsp)

			in := []reflect.Value{reflect.ValueOf(ctx), req, rsp}
			out := v.Call(in)
			return out[0].Interface().(error)
		}
	}

	return
}

func (r *Router) Run() {
	// 定时检测rpc消息是否超时,100通过配置?
	t := time.NewTicker(time.Millisecond * 100)
	for ; true; <-t.C {
		fmt.Printf("tick")
		now := times.Now()
		expired := make([]*rReply, 0)
		r.mux.Lock()
		// TODO:小顶堆优化
		for k, v := range r.replies {
			if v.Deadline >= now {
				delete(r.replies, k)
				seqID, ttl, err := v.Check()
				if err != nil {
					continue
				}
				v.SeqID = seqID
				v.Deadline = now + int64(ttl/time.Millisecond)
				r.replies[seqID] = v
				expired = append(expired, v)
			}
		}
		r.mux.Unlock()
		// 处理过期消息
		for _, v := range expired {
			err := v.Retry()
			if err != nil {
				r.mux.Lock()
				delete(r.replies, v.SeqID)
				r.mux.Unlock()
			}
		}
	}
}
