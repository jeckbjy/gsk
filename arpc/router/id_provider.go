package router

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/jeckbjy/gsk/arpc"
)

func NewIDProvider() arpc.IDProvider {
	p := &IDProvider{nameDict: make(map[string]int), idDict: make(map[int]string)}
	return p
}

type IDProvider struct {
	mux      sync.RWMutex
	nameDict map[string]int
	idDict   map[int]string
	bindName bool
}

func (p *IDProvider) SetBindName(flag bool) {
	p.mux.Lock()
	p.bindName = flag
	p.mux.Unlock()
}

func (p *IDProvider) Register(name string, id int) error {
	var err error
	p.mux.Lock()
	defer p.mux.Unlock()
	if oid, ok := p.nameDict[name]; ok {
		return fmt.Errorf("duplicate register, name %+v, new_id %+v, old_id %+v", name, id, oid)
	}
	if oname, ok := p.idDict[id]; ok {
		return fmt.Errorf("duplicate register,id %+v, new_name %+v, old_name %+v", id, name, oname)
	}

	p.idDict[id] = name
	p.nameDict[name] = id
	return err
}

func (p *IDProvider) GetID(name string) int {
	p.mux.RLock()
	id := p.nameDict[name]
	p.mux.RUnlock()
	return id
}

func (p *IDProvider) GetName(id int) string {
	p.mux.Lock()
	name := p.idDict[id]
	p.mux.RUnlock()
	return name
}

func (p *IDProvider) Fill(pkg arpc.Packet, msg interface{}) error {
	if m, ok := msg.(arpc.MessageID); ok {
		pkg.SetMsgID(m.MsgID())
		return nil
	}

	var err error
	var name string
	t := reflect.TypeOf(msg).Elem()
	if t.Kind() == reflect.Ptr {
		name = t.Elem().Name()
	} else {
		name = t.Name()
	}

	p.mux.RLock()
	if id, ok := p.nameDict[name]; ok {
		pkg.SetMsgID(id)
	} else if p.bindName {
		pkg.SetName(name)
	} else {
		err = arpc.ErrNotFoundID
	}
	p.mux.RUnlock()

	return err
}
