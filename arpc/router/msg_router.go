package router

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/jeckbjy/gsk/util/reflects"

	"github.com/jeckbjy/gsk/arpc"
)

type _MsgInfo struct {
	Handler arpc.Handler
	ID      uint
	Name    string
	Method  string
}

type _MsgRouter struct {
	mux     sync.RWMutex
	ids     []*_MsgInfo
	names   map[string]*_MsgInfo
	methods map[string]*_MsgInfo
}

func (r *_MsgRouter) Init() {
	r.names = make(map[string]*_MsgInfo)
	r.methods = make(map[string]*_MsgInfo)
}

func (r *_MsgRouter) Find(pkg arpc.Packet) (arpc.Handler, error) {
	r.mux.RLock()
	var info *_MsgInfo
	if pkg.MsgID() > 0 && int(pkg.MsgID()) < len(r.ids) {
		info = r.ids[pkg.MsgID()]
	}

	if info == nil && pkg.Method() != "" {
		info = r.methods[pkg.Method()]
	}

	if info == nil && pkg.Name() != "" {
		info = r.names[pkg.Name()]
	}
	r.mux.RUnlock()

	if info != nil && info.Handler != nil {
		return info.Handler, nil
	}

	return nil, arpc.ErrNoHandler
}

func (r *_MsgRouter) Register(srv interface{}, o *arpc.RegisterOptions) error {
	v := reflect.ValueOf(srv)
	t := v.Type()
	if t.Kind() != reflect.Func {
		return arpc.ErrNotSupport
	}

	if t.NumIn() < 1 || t.NumIn() > 3 || t.NumOut() != 1 {
		return arpc.ErrInvalidHandler
	}

	if !arpc.IsContext(t.In(0)) || !arpc.IsError(t.Out(0)) {
		return arpc.ErrInvalidHandler
	}

	var handler arpc.Handler
	// 原型
	// func(ctx Context) error
	// func(ctx Context, req *XXRequest) error
	// func(ctx Context, req *XXRequest, rsp *XXResponse) error
	switch t.NumIn() {
	case 1:
		handler = srv.(arpc.Handler)
	case 2:
		if !arpc.IsMessage(t.In(1)) {
			return arpc.ErrInvalidHandler
		}
		handler = func(ctx arpc.Context) error {
			msg := reflect.New(t.In(1).Elem())
			if err := ctx.Request().DecodeBody(msg); err != nil {
				return err
			}
			in := []reflect.Value{reflect.ValueOf(ctx), msg}
			out := v.Call(in)
			return out[0].Interface().(error)
		}
	case 3:
		if !arpc.IsMessage(t.In(1)) && !arpc.IsMessage(t.In(2)) {
			return arpc.ErrInvalidHandler
		}
		handler = func(ctx arpc.Context) error {
			msg := reflect.New(t.In(1).Elem())
			if err := ctx.Request().DecodeBody(msg); err != nil {
				return err
			}
			rsp := reflect.New(t.In(2))
			ctx.Response().SetBody(rsp)
			in := []reflect.Value{reflect.ValueOf(ctx), msg, rsp}
			out := v.Call(in)
			return out[0].Interface().(error)
		}
	}

	name := o.Name
	if len(name) == 0 && t.NumIn() > 1 {
		name = t.In(1).Elem().Name()
	}

	method := o.Method
	if len(method) == 0 {
		method = reflects.FuncName(v.Pointer())
	}

	info := &_MsgInfo{Handler: handler, ID: o.ID, Name: name, Method: method}
	if info.ID == 0 && info.Name == "" && info.Method == "" {
		return ErrInvalidHandler
	}

	r.mux.Lock()
	err := r.add(info)
	r.mux.Unlock()

	return err
}

func (r *_MsgRouter) add(info *_MsgInfo) error {
	if info.ID != 0 {
		if int(info.ID) >= len(r.ids) {
			ids := make([]*_MsgInfo, info.ID+1)
			copy(ids, r.ids)
			r.ids[info.ID] = info
		}

		if r.ids[info.ID] != nil {
			return fmt.Errorf("duplicate register,msgid=%+v", info.ID)
		}
		r.ids[info.ID] = info
	}

	if info.Method != "" {
		if _, ok := r.methods[info.Method]; ok {
			return fmt.Errorf("duplicate register,method=%+v", info.Method)
		}
		r.methods[info.Method] = info
	}

	if info.Name != "" {
		r.names[info.Name] = info
	}

	return nil
}
