package stats

import (
	"sync"
	"sync/atomic"
)

// Scope is a grouping of stats.
type Scope struct {
	prefix   string
	store    *Store
	refCount uint

	childLock sync.RWMutex
	children  atomic.Value // map[string]*Scope

	gaugesLock     sync.Mutex
	gauges         atomic.Value // map[string]*Gauge, key is the metric's name without prefix
	countersLock   sync.Mutex
	counters       atomic.Value // map[string]*Counter
	histogramsLock sync.Mutex
	histograms     atomic.Value // map[string]*Histogram
}

func newScope(name string, store *Store) *Scope {
	s := &Scope{
		prefix: name,
		store:  store,
	}
	s.children.Store(make(map[string]*Scope))
	s.gauges.Store(make(map[string]*Gauge))
	s.counters.Store(make(map[string]*Counter))
	s.histograms.Store(make(map[string]*Histogram))
	return s
}

// NewChild returns a new child scope
func (scope *Scope) NewChild(name string) *Scope {
	children := scope.loadChildren()
	if child, ok := children[name]; ok {
		return child
	}

	scope.childLock.Lock()
	child := scope.newChildLocked(name)
	scope.childLock.Unlock()
	return child
}

func (scope *Scope) newChildLocked(name string) *Scope {
	children := scope.loadChildren()
	if child, ok := children[name]; ok {
		return child
	}

	tmp := make(map[string]*Scope, len(children))
	child := scope.store.CreateScope(scope.prefix + name)
	tmp[name] = child
	for name, child := range children {
		tmp[name] = child
	}
	scope.updateChildren(tmp)
	return child
}

func (scope *Scope) loadChildren() map[string]*Scope {
	return scope.children.Load().(map[string]*Scope)
}

func (scope *Scope) updateChildren(children map[string]*Scope) {
	scope.children.Store(children)
}

// Name returns the name of the scope
func (scope *Scope) Name() string {
	return scope.prefix
}

func (scope *Scope) loadGauges() map[string]*Gauge {
	return scope.gauges.Load().(map[string]*Gauge)
}

func (scope *Scope) updateGauges(v map[string]*Gauge) {
	scope.gauges.Store(v)
}

func (scope *Scope) loadCounters() map[string]*Counter {
	return scope.counters.Load().(map[string]*Counter)
}

func (scope *Scope) updateCounters(v map[string]*Counter) {
	scope.counters.Store(v)
}

func (scope *Scope) loadHistograms() map[string]*Histogram {
	return scope.histograms.Load().(map[string]*Histogram)
}

func (scope *Scope) updateHistograms(v map[string]*Histogram) {
	scope.histograms.Store(v)
}

// Gauge returns a gauge within the scope namespace.
func (scope *Scope) Gauge(name string) *Gauge {
	// TODO(kik91): sanitize name
	gs := scope.loadGauges()
	if g, ok := gs[name]; ok {
		return g
	}

	scope.gaugesLock.Lock()
	g := scope.gaugeLocked(name)
	scope.gaugesLock.Unlock()
	return g
}

func (scope *Scope) gaugeLocked(name string) *Gauge {
	gs := scope.loadGauges()
	if g, ok := gs[name]; ok {
		return g
	}

	tmp := make(map[string]*Gauge, len(gs))
	for name, g := range gs {
		tmp[name] = g
	}
	finalName := scope.prefix + name
	extractedName, tags := scope.store.getTagsForName(finalName)
	g := NewGauge(finalName, extractedName, tags)
	tmp[name] = g
	scope.updateGauges(tmp)
	return g
}

// Counter returns a counter within the scope namespace.
func (scope *Scope) Counter(name string) *Counter {
	// TODO(kik91): sanitize name
	cs := scope.loadCounters()
	if c, ok := cs[name]; ok {
		return c
	}

	scope.countersLock.Lock()
	c := scope.counterLocked(name)
	scope.countersLock.Unlock()
	return c
}

func (scope *Scope) counterLocked(name string) *Counter {
	cs := scope.loadCounters()
	if c, ok := cs[name]; ok {
		return c
	}

	tmp := make(map[string]*Counter, len(cs))
	for name, c := range cs {
		tmp[name] = c
	}
	finalName := scope.prefix + name
	extractedName, tags := scope.store.getTagsForName(finalName)
	c := NewCounter(finalName, extractedName, tags)
	tmp[name] = c
	scope.updateCounters(tmp)
	return c
}

// Histogram returns a histogram within the scope namespace.
func (scope *Scope) Histogram(name string) *Histogram {
	// TODO(kik91): sanitize name
	hs := scope.loadHistograms()
	if h, ok := hs[name]; ok {
		return h
	}

	scope.histogramsLock.Lock()
	h := scope.histogramLocked(name)
	scope.histogramsLock.Unlock()
	return h
}

func (scope *Scope) histogramLocked(name string) *Histogram {
	hs := scope.loadHistograms()
	if h, ok := hs[name]; ok {
		return h
	}

	tmp := make(map[string]*Histogram, len(hs))
	for name, h := range hs {
		tmp[name] = h
	}
	finalName := scope.prefix + name
	extractedName, tags := scope.store.getTagsForName(finalName)
	h := NewHistogram(scope.store, finalName, extractedName, tags)
	tmp[name] = h
	scope.updateHistograms(tmp)
	return h
}

// Counters returns all known counters within the scope namespace.
func (scope *Scope) Counters() []*Counter {
	cs := scope.loadCounters()
	ret := make([]*Counter, 0, len(cs))
	for _, c := range cs {
		if c.IsUsed() {
			ret = append(ret, c)
		}
	}
	return ret
}

// Gauges returns all known gagues within the scope namespace.
func (scope *Scope) Gauges() []*Gauge {
	gs := scope.loadGauges()
	ret := make([]*Gauge, 0, len(gs))
	for _, g := range gs {
		if g.IsUsed() {
			ret = append(ret, g)
		}
	}
	return ret
}

// Histograms returns all known histograms within the scope namespace.
func (scope *Scope) Histograms() []*Histogram {
	hs := scope.loadHistograms()
	ret := make([]*Histogram, 0, len(hs))
	for _, h := range hs {
		if h.IsUsed() {
			ret = append(ret, h)
		}
	}
	return ret
}
