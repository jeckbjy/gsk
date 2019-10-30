package router

import (
	"errors"
	"reflect"

	"github.com/jeckbjy/gsk/arpc"
	"github.com/jeckbjy/gsk/util/reflects"
)

func New() arpc.Router {
	return &Router{names: make(map[string]*_Info), methods: make(map[string]*_Info)}
}

type Router struct {
	ids     []*_Info          // ID
	names   map[string]*_Info // 消息名映射
	methods map[string]*_Info // 方法名映射
}

type _Info struct {
	ID      uint
	Name    string
	Method  string
	Handler arpc.Handler // 消息回调
	Req     reflect.Type // 消息请求类型,用于反射创建消息
}

func (r *Router) Find(pkg arpc.Packet) (arpc.Handler, error) {
	// 查询服务端消息回调
	info, err := r.findInfo(pkg)
	if err != nil {
		return nil, err
	}
	if info != nil && info.Handler != nil {
		return info.Handler, nil
	}

	return nil, arpc.ErrNoHandler
}

func (r *Router) findInfo(req arpc.Packet) (*_Info, error) {
	if req.Method() != "" {
		if info, ok := r.methods[req.Method()]; ok {
			return info, nil
		}
	}

	if req.ID() != 0 {
		if int(req.ID()) >= len(r.ids) {
			return nil, errors.New("bad msg id")
		}

		return r.ids[req.ID()], nil
	}

	if req.Name() != "" {
		return r.names[req.Name()], nil
	}

	return nil, nil
}

//
func (r *Router) Register(srv interface{}) error {
	v := reflect.ValueOf(srv)
	t := v.Type()
	switch t.Kind() {
	case reflect.Ptr:
		//
		var last error
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
			if err := r.addInfo(handler, req, m.Name); err != nil {
				last = err
			} else {
				count++
			}
		}
		if last != nil {
			return last
		}
		if count == 0 {
			return errors.New("invalid service")
		}

		return nil
	case reflect.Func:
		handler, req := r.extract(v)
		if handler != nil {
			return r.addInfo(handler, req, reflects.FuncName(v.Pointer()))
		}

		return ErrInvalidHandler
	default:
		return arpc.ErrNotSupport
	}
}

func (r *Router) addInfo(handler arpc.Handler, req reflect.Type, method string) error {
	info := &_Info{Name: req.Name(), Method: method}
	// 校验注册过?
	r.names[info.Name] = info
	r.methods[info.Method] = info

	return nil
}

// 解析函数:返回Handler和Request类型
// func(ctx IContext) error
// func(ctx IContext, req XRequest) error
// func(ctx IContext, req XRequest, rsp XResponse) error
func (r *Router) extract(v reflect.Value) (handler arpc.Handler, reqType reflect.Type) {
	m := v.Type()
	if m.NumIn() < 1 || m.NumIn() > 3 || m.NumOut() != 1 {
		return
	}
	// the first in arg must be IContext
	//if !m.In(0).Implements(gTypeCtx) || !m.Out(0).Implements(gTypeErr) {
	//	return
	//}

	switch m.NumIn() {
	case 1:
		handler = v.Interface().(arpc.Handler)
	case 2:
		// func(ctx IContext, req *XRequest)
		if m.In(1).Kind() != reflect.Ptr {
			return
		}
		reqType = m.In(1).Elem()
		handler = func(ctx arpc.IContext) error {
			msg := reflect.New(reqType)
			if err := ctx.Request().Parse(msg); err != nil {
				return err
			}
			in := []reflect.Value{reflect.ValueOf(ctx), msg}
			out := v.Call(in)
			return out[0].Interface().(error)
		}
	case 3:
		// func(ctx IContext, req *XRequest, rsp *XResponse) error
		if m.In(1).Kind() != reflect.Ptr || m.In(2).Kind() != reflect.Ptr {
			return
		}

		reqType = m.In(1).Elem()
		rspType := m.In(2).Elem()
		handler = func(ctx arpc.IContext) error {
			msg := reflect.New(reqType)
			if err := ctx.Request().Parse(msg); err != nil {
				return err
			}

			rsp := reflect.New(rspType)
			ctx.Response().SetBody(rsp)

			in := []reflect.Value{reflect.ValueOf(ctx), msg, rsp}
			out := v.Call(in)
			return out[0].Interface().(error)
		}
	}

	return
}
