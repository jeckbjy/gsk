package breaker

import "sync"

func NewGroup(cfg *Config) Group {
	g := &_Group{}
	g.Reload(cfg)
	return g
}

type _Group struct {
	mux      sync.RWMutex
	breakers map[string]Breaker
	cfg      *Config
}

func (g *_Group) Get(key string) Breaker {
	g.mux.RLock()
	b := g.breakers[key]
	g.mux.RUnlock()
	if b != nil {
		return b
	}

	g.mux.Lock()
	b = newBreaker(g.cfg)
	g.breakers[key] = b
	g.mux.Unlock()

	return b
}

func (g *_Group) Reload(cfg *Config) {
	if cfg == nil || cfg == g.cfg {
		return
	}
	cfg.fix()
	g.mux.Lock()
	g.cfg = cfg
	g.breakers = make(map[string]Breaker, len(g.breakers))
	g.mux.Unlock()
}

func (g *_Group) Exec(key string, run, fallback func() error) error {
	b := g.Get(key)
	if err := b.Allow(); err != nil {
		return fallback()
	}

	return run()
}
