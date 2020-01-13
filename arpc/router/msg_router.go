package router

import (
	"fmt"
	"net/http"
	"reflect"
	"sync"

	"github.com/jeckbjy/gsk/arpc"
)

type _MsgInfo struct {
	Handler arpc.Handler
	Extra   interface{}
}

type MsgRouter struct {
	mux  sync.RWMutex         //
	list []*_MsgInfo          // ID列表
	dict map[string]*_MsgInfo // (name/method)=>MsgInfo
}

func (r *MsgRouter) Init() {
	r.dict = make(map[string]*_MsgInfo)
}

func (r *MsgRouter) Handle(ctx arpc.Context) arpc.Handler {
	r.mux.RLock()
	defer r.mux.RUnlock()

	pkg := ctx.Message()
	info := r.find(pkg)
	if info != nil {
		ctx.SetData(info.Extra)
		return info.Handler
	}

	return nil
}

func (r *MsgRouter) find(pkg arpc.Packet) *_MsgInfo {
	if pkg.MsgID() != 0 {
		index := toIndex(pkg.MsgID())
		if index < len(r.list) {
			return r.list[index]
		}
	}

	if r.dict == nil {
		return nil
	}
	if pkg.Name() != "" {
		return r.dict[pkg.Name()]
	} else if pkg.Method() != "" {
		return r.dict[pkg.Method()]
	}

	return nil
}

func (r *MsgRouter) Register(cb interface{}, o *arpc.MiscOptions) error {
	v := reflect.ValueOf(cb)
	handler, err := toHandler(&v, cb)
	if err != nil {
		return err
	}

	r.mux.Lock()
	defer r.mux.Unlock()

	info := &_MsgInfo{Handler: handler, Extra: o.Extra}

	// ID和Method可以共存,因为可以支持多种协议,比如tcp为了高效可以使用ID,http为了简单方便可以使用Method
	// 但Name则不是必须的,只有在测试的环境下才需要使用,因为使用Name完全可以用ID的方式代替
	// 但ID的方式有个缺点,需要外部指定唯一ID,名字可以反射获得
	if o.ID != 0 {
		if !arpc.IsValidID(o.ID) {
			return arpc.ErrInvalidID
		}

		index := toIndex(o.ID)
		if index >= len(r.list) {
			list := make([]*_MsgInfo, index+1)
			copy(list, r.list)
			r.list = list
		}

		if r.list[index] != nil {
			return fmt.Errorf("duplicate msgid=%+v", o.ID)
		}

		r.list[index] = info
	}

	if len(o.Method) != 0 {
		r.dict[o.Method] = info
	}

	if o.ID == 0 && len(o.Method) == 0 {
		// 都没有指定则默认使用name
		t := v.Type()
		if t.NumIn() > 1 {
			name := t.In(1).Elem().Name()
			r.dict[name] = info
		}
	}

	return err
}

func toHandler(v *reflect.Value, cb interface{}) (arpc.Handler, error) {
	// func(ctx Context) error
	if handler, ok := cb.(arpc.Handler); ok {
		return handler, nil
	}

	t := v.Type()

	if t.Kind() != reflect.Func {
		return nil, arpc.ErrNotSupport
	}

	if t.NumIn() != 3 || t.NumOut() != 1 || !isContext(t.In(0)) || !isError(t.Out(0)) {
		return nil, arpc.ErrInvalidHandler
	}

	// TODO: 是否需要支持:func(ctx Context, req *XXRequest) error
	// func(ctx Context, req *XXRequest, rsp *XXResponse) error
	handler := func(ctx arpc.Context) error {
		req := ctx.Message()
		if req.Body() != nil {
			if err := arpc.DecodeBody(req, reflect.New(t.In(0).Elem()).Interface()); err != nil {
				return err
			}
		}

		// auto create response
		rsp := ctx.Response()
		if rsp == nil {
			msg := reflect.New(t.In(2))
			rsp = arpc.NewPacket()
			rsp.SetSeqID(req.SeqID())
			rsp.SetBody(rsp)
			_ = arpc.GetIDProvider().Fill(rsp, msg)
		}

		in := []reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(req.Body()), reflect.ValueOf(rsp.Body())}
		out := v.Call(in)
		if !out[0].IsNil() {
			// TODO: 解析返回错误
			err := out[0].Interface().(error)
			rsp.SetStatus(http.StatusInternalServerError, err.Error())
		}

		return ctx.Send(rsp)
	}

	return handler, nil
}

// 将ID转换成索引
func toIndex(id int) int {
	return id - arpc.IDMin
}
