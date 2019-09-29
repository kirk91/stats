package stats

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockSink struct {
	flushCallback         func(Source)
	flushCalled           bool
	writeHistSampleCalled bool
}

func (sink *mockSink) Flush(src Source) error {
	if sink.flushCallback != nil {
		sink.flushCallback(src)
	}
	return nil
}

func (sink *mockSink) WriteHistogramSample(h *Histogram, val uint64) error {
	sink.writeHistSampleCalled = true
	return nil
}

func TestNewScopeWithInvalidTagExtractStrategy(t *testing.T) {
	strategy := TagExtractStrategy{
		Name:  "blabla",
		Regex: `[ ]\K(?<!\d )(?=(?: ?\d){8})(?!(?: ?\d){9})\d[ \d]+\d`,
	}
	err := NewStore(nil).SetTagOption(NewTagOption().WithTagExtractStrategies(strategy))
	assert.Error(t, err)
}

func TestStoreCreateScope(t *testing.T) {
	store := NewStore(
		NewStoreOption().
			WithFlushInterval(time.Second).
			WithSinks(new(mockSink)))

	name := "listener.blabla."
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			store.CreateScope(name)
			wg.Done()
		}()
	}

	wg.Wait()
	assert.Equal(t, len(store.Scopes()), 1)
}

func TestStoreDeleteScope(t *testing.T) {
	store := NewStore(
		NewStoreOption().
			WithFlushInterval(time.Second).
			WithSinks(new(mockSink)))

	name := "listener.blabla"
	scope := store.CreateScope(name)
	store.CreateScope(name)
	store.CreateScope(name)

	store.DeleteScope(scope) // refcount is 3
	assert.Equal(t, len(store.Scopes()), 1)

	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			store.DeleteScope(scope)
			wg.Done()
		}()
	}
	wg.Wait()
	assert.Equal(t, len(store.Scopes()), 0)
}

func TestStoreDeleteScopeWithChildren(t *testing.T) {
	store := NewStore(
		NewStoreOption().
			WithFlushInterval(time.Second).
			WithSinks(new(mockSink)))

	name := "father.blabla"
	scope := store.CreateScope(name)
	store.CreateScope(name)
	store.CreateScope(name)
	assert.Equal(t, len(store.Scopes()), 1)

	child1 := scope.NewChild(".child1")
	child2 := scope.NewChild(".child2")
	child2.NewChild(".child3")
	assert.Equal(t, len(store.Scopes()), 4)

	store.DeleteScope(child1)
	assert.Equal(t, len(store.Scopes()), 3)

	store.DeleteScope(scope)
	assert.Equal(t, len(store.Scopes()), 1)

	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			store.DeleteScope(scope)
			wg.Done()
		}()
	}
	wg.Wait()
	assert.Equal(t, len(store.Scopes()), 0)
}

func TestStoreCounters(t *testing.T) {
	store := NewStore(
		NewStoreOption().
			WithFlushInterval(time.Second).
			WithSinks(new(mockSink)))

	scope1 := store.CreateScope("scope1")
	scope1.Counter("cx_active").Inc()
	scope1.Counter("cx_total").Inc()

	scope2 := store.CreateScope("scope2")
	scope2.Counter("cx_active").Inc()

	cs := store.Counters()
	assert.Equal(t, len(cs), 3)
}

func TestStoreGauges(t *testing.T) {
	store := NewStore(
		NewStoreOption().
			WithFlushInterval(time.Second).
			WithSinks(new(mockSink)))

	scope1 := store.CreateScope("scope1")
	scope1.Gauge("cx_active").Inc()
	scope1.Gauge("cx_total").Inc()

	scope2 := store.CreateScope("scope2")
	scope2.Gauge("cx_active").Inc()

	gs := store.Gauges()
	assert.Equal(t, len(gs), 3)
}

func TestStoreHistograms(t *testing.T) {
	store := NewStore(
		NewStoreOption().
			WithFlushInterval(time.Second).
			WithSinks(new(mockSink)))

	scope1 := store.CreateScope("scope1")
	scope1.Histogram("cx_active")
	scope1.Histogram("cx_total")

	scope2 := store.CreateScope("scope2")
	scope2.Histogram("cx_active")

	hs := store.Histograms()
	assert.Equal(t, len(hs), 0)

	scope1.Histogram("cx_active").Record(uint64(1))
	scope1.Histogram("cx_total").Record(uint64(1))
	scope2.Histogram("cx_active").Record(uint64(1))

	hs = store.Histograms()
	assert.Equal(t, len(hs), 3)
}

func TestStoreDeliverHistogramSampleToSinks(t *testing.T) {
	sink1 := new(mockSink)
	sink2 := new(mockSink)
	store := NewStore(
		NewStoreOption().
			WithFlushInterval(time.Second).
			WithSinks(sink1, sink2))

	scope := store.CreateScope("")
	h := scope.Histogram("haha")
	h.Record(uint64(time.Millisecond))

	assert.True(t, sink1.writeHistSampleCalled)
	assert.True(t, sink2.writeHistSampleCalled)
}

func TestStorePeroidcFlush(t *testing.T) {
	sink1 := new(mockSink)
	sink1.flushCallback = func(src Source) {
		assert.Equal(t, len(src.CachedCounters()), 1)
		assert.Equal(t, len(src.CachedGauges()), 0)
	}

	opt := NewStoreOption().
		WithFlushInterval(time.Millisecond * 100).
		WithSinks(sink1)
	store := NewStore(opt)

	// add gauge
	scope := store.CreateScope("")
	scope.Counter("foo").Inc()

	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(time.Millisecond*150, cancel)
	store.FlushingLoop(ctx)
}

func TestStoreDefaultTags(t *testing.T) {
	store := NewStore(nil)
	store.SetTagOption(
		NewTagOption().
			WithDefaultTags(map[string]string{"tag1": "value1"}))

	extractedName, tags := store.getTagsForName("service.corvus_111")
	assert.Equal(t, extractedName, "service.corvus_111")
	assert.Equal(t, 1, len(tags))
	assert.Equal(t, tags[0].Name, "tag1")
	assert.Equal(t, tags[0].Value, "value1")
}

func TestStoreDynamicTags(t *testing.T) {
	strategies := []TagExtractStrategy{
		TagExtractStrategy{
			Name:  "service_name",
			Regex: "^service\\.((.*?)\\.)",
		},
	}
	store := NewStore(nil)
	store.SetTagOption(NewTagOption().WithTagExtractStrategies(strategies...))

	extractedName, tags := store.getTagsForName("service.corvus_111.conn_total")
	assert.Equal(t, extractedName, "service.conn_total")
	assert.Equal(t, 1, len(tags))
	assert.Equal(t, tags[0].Name, "service_name")
	assert.Equal(t, tags[0].Value, "corvus_111")
}
