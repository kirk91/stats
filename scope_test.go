package stats

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestScopeObtainCounter(t *testing.T) {
	scope := newScope("listener.foo.", new(Store))
	counter := scope.Counter("counter")
	counter1 := scope.Counter("counter")

	assert.Equal(t, counter.Name(), "listener.foo.counter")
	assert.Equal(t, counter1.Name(), "listener.foo.counter")
	assert.Equal(t, len(scope.Counters()), 0)

	counter.Inc()
	counter1.Inc()
	assert.Equal(t, len(scope.Counters()), 1)
}

func TestScopeObtainGauge(t *testing.T) {
	scope := newScope("listener.foo.", new(Store))
	gauge := scope.Gauge("gauge")
	gauge1 := scope.Gauge("gauge")

	assert.Equal(t, gauge.Name(), "listener.foo.gauge")
	assert.Equal(t, gauge1.Name(), "listener.foo.gauge")
	assert.Equal(t, len(scope.Gauges()), 0)

	gauge.Inc()
	gauge1.Inc()
	assert.Equal(t, len(scope.Gauges()), 1)

	gauge.Dec()
	gauge1.Dec()
	assert.Equal(t, len(scope.Gauges()), 1)
}

func TestScopeObtainHistogram(t *testing.T) {
	store := NewStore(NewStoreOption().WithFlushInterval(time.Minute))
	scope := newScope("listener.foo.", store)
	histogram := scope.Histogram("histogram")
	histogram1 := scope.Histogram("histogram")
	assert.Equal(t, histogram.Name(), "listener.foo.histogram")
	assert.Equal(t, histogram1.Name(), "listener.foo.histogram")
	assert.Equal(t, len(scope.Histograms()), 0)

	histogram.Record(uint64(1))
	assert.Equal(t, len(scope.Histograms()), 1)
}

func TestNewScopeWithChildren(t *testing.T) {
	store := NewStore(NewStoreOption().WithFlushInterval(time.Minute))
	scope := newScope("I.am.a.father.", store)
	children := scope.loadChildren()
	assert.Equal(t, 0, len(children))

	child1 := scope.NewChild("child1")
	assert.Equal(t, len(child1.loadChildren()), 0)
	assert.Equal(t, "I.am.a.father.child1.", child1.Name())
	children = scope.loadChildren()
	assert.Equal(t, 1, len(children))

	child2 := scope.NewChild("child2")
	assert.Equal(t, 0, len(child1.loadChildren()))
	assert.Equal(t, "I.am.a.father.child2.", child2.Name())
	children = scope.loadChildren()
	assert.Equal(t, 2, len(children))
}
