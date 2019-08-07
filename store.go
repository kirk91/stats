package stats

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Store is a storage for all known counters, gauges and histograms.
type Store struct {
	mu            sync.RWMutex
	flushInterval time.Duration
	tp            *TagProducer
	sinks         atomic.Value // []Sink

	scopes map[string]*Scope
	errors chan error

	done chan struct{}
}

// NewStore returns a stats storage.
func NewStore(o *StoreOption) *Store {
	if o == nil {
		o = NewStoreOption()
	}
	store := &Store{
		flushInterval: o.FlushInterval,
		scopes:        make(map[string]*Scope),
		errors:        make(chan error),
	}
	store.sinks.Store(o.Sinks)
	return store
}

// Errors returns chan receiving errors occurred during flushing.
func (store *Store) Errors() <-chan error {
	return store.errors
}

// FlushingLoop flushes stats to remote destinations(defined by Sinks) at an interval, it blocks untils ctx canceled.
// NOTE: this function must be called or metrics won't be sent out.
func (store *Store) FlushingLoop(ctx context.Context) {
	source := &source{
		rawCounters:   store.Counters,
		rawGauges:     store.Gauges,
		rawHistograms: store.Histograms,
	}
	ticker := time.NewTicker(store.flushInterval)
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			sinks := store.Sinks()
			for _, sink := range sinks {
				err := sink.Flush(source)
				store.sendError(err)
			}
			source.ClearCache() // clear cached counters, gauges and histograms
		}
	}
}

func (store *Store) sendError(err error) {
	if err == nil {
		return
	}
	select {
	case store.errors <- err:
	default:
	}
}

// SetTagOption sets tagging option to given one.
func (store *Store) SetTagOption(o *TagOption) error {
	var tp *TagProducer
	if len(o.DefaultTags) > 0 || len(o.TagExtractStrategies) > 0 {
		tp = NewTagProducer(o.defaultTags()...)
		for _, strategy := range o.TagExtractStrategies {
			te, err := NewTagExtractor(strategy.Name, strategy.Regex, strategy.SubStr)
			if err != nil {
				return err
			}
			tp.AddExtractor(te)
		}
	}
	store.tp = tp
	return nil
}

func (store *Store) setTagProducer(tp *TagProducer) {
	store.tp = tp
}

// AddSink adds a sink
func (store *Store) AddSink(sink Sink) {
	store.mu.Lock()
	defer store.mu.Unlock()

	sinks := store.Sinks()
	tmpSinks := make([]Sink, len(sinks), len(sinks)+1)
	copy(tmpSinks, sinks)
	tmpSinks = append(tmpSinks, sink)
	store.sinks.Store(tmpSinks)
}

// Sinks return all known sinks.
func (store *Store) Sinks() []Sink {
	return store.sinks.Load().([]Sink)
}

// CreateScope creates the named Scope.
func (store *Store) CreateScope(name string) *Scope {
	store.mu.Lock()
	defer store.mu.Unlock()

	// fix suffix
	if len(name) > 0 && !strings.HasSuffix(name, ".") {
		name += "."
	}
	scope, ok := store.scopes[name]
	if !ok {
		scope = newScope(name, store)
		store.scopes[name] = scope
	}
	scope.refCount++
	return scope
}

// DeleteScope deletes the named Scope.
func (store *Store) DeleteScope(scope *Scope) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.deleteScopeWithAllChilds(scope)
}

func (store *Store) deleteScopeWithAllChilds(scope *Scope) {
	name := scope.Name()
	scope, ok := store.scopes[name]
	if !ok {
		return
	}
	children := scope.loadChildren()
	for _, child := range children {
		store.deleteScopeWithAllChilds(child)
	}
	if scope.refCount == 1 {
		delete(store.scopes, name)
		return
	}
	scope.refCount--
}

// Scopes returns all known scopes.
func (store *Store) Scopes() []*Scope {
	store.mu.RLock()
	scopes := make([]*Scope, 0, len(store.scopes))
	for _, scope := range store.scopes {
		scopes = append(scopes, scope)
	}
	store.mu.RUnlock()
	return scopes
}

// Counters returns all known counters.
func (store *Store) Counters() []*Counter {
	scopes := store.Scopes()
	cs := make([]*Counter, 0, len(scopes)*2)
	for _, scope := range scopes {
		cs = append(cs, scope.Counters()...)
	}
	return cs
}

// Gauges returns all known gauges.
func (store *Store) Gauges() []*Gauge {
	scopes := store.Scopes()
	gs := make([]*Gauge, 0, len(scopes)*2)
	for _, scope := range scopes {
		gs = append(gs, scope.Gauges()...)
	}
	return gs
}

// Histograms returns all known histograms.
func (store *Store) Histograms() []*Histogram {
	scopes := store.Scopes()
	hs := make([]*Histogram, 0, len(scopes)*2)
	for _, scope := range scopes {
		hs = append(hs, scope.Histograms()...)
	}
	return hs
}

func (store *Store) deliverHistogramSampleToSinks(h *Histogram, val uint64) {
	sinks := store.Sinks()
	for _, sink := range sinks {
		sink.WriteHistogramSample(h, val)
	}
}

func (store *Store) clearAllScopes() {
	store.mu.Lock()
	for name := range store.scopes {
		delete(store.scopes, name)
	}
	store.mu.Unlock()
}

func (store *Store) getTagsForName(name string) (string, []*Tag) {
	if store.tp == nil {
		return name, nil
	}
	return store.tp.Produce(name)
}
