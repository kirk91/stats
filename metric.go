package stats

import (
	"sync/atomic"
)

// Metric is a general interface for stats.
type Metric interface {
	Name() string
	TagExtractedName() string
	Tags() []*Tag
	IsUsed() bool
}

var (
	_ Metric = new(Gauge)
	_ Metric = new(Histogram)
	_ Metric = new(Counter)
)

type metric struct {
	name             string
	tagExtractedName string
	tags             []*Tag
	isUsed           int32
}

func newMetric(name, tagExtractedName string, tags []*Tag) metric {
	return metric{
		name:             name,
		tagExtractedName: tagExtractedName,
		tags:             tags,
	}
}

func (m *metric) Name() string {
	return m.name
}

func (m *metric) TagExtractedName() string {
	return m.tagExtractedName
}

func (m *metric) Tags() []*Tag {
	return m.tags
}

func (m *metric) IsUsed() bool {
	return atomic.LoadInt32(&m.isUsed) == 1
}

func (m *metric) markUsed() {
	atomic.StoreInt32(&m.isUsed, 1)
}

// Gauge is a Metric that represents a single numerical value that can
// arbitrarily go up and down.
type Gauge struct {
	metric
	val uint64
}

func NewGauge(name, tagExtractedName string, tags []*Tag) *Gauge {
	return &Gauge{
		metric: newMetric(name, tagExtractedName, tags),
	}
}

// Set sets the gauge to an arbitrary value.
func (g *Gauge) Set(val uint64) {
	atomic.StoreUint64(&g.val, val)
	g.markUsed()
}

// Add adds the given value to the Gauge.
func (g *Gauge) Add(amount uint64) {
	atomic.AddUint64(&g.val, amount)
	g.markUsed()
}

// Sub subtracts the given value from the Gauge.
func (g *Gauge) Sub(amount uint64) {
	atomic.AddUint64(&g.val, ^uint64(amount-1))
	g.markUsed()
}

// Inc increments the Gauge by 1.
func (g *Gauge) Inc() {
	g.Add(1)
}

// Dec decrements the Gauge by 1.
func (g *Gauge) Dec() {
	g.Sub(1)
}

// Value returns the Gauge value.
func (g *Gauge) Value() uint64 {
	return atomic.LoadUint64(&g.val)
}

// Counter is a Metric that represents a single numerical value
// that only ever goes up. Each increment is added both to a global
// counter as well as periodic counter.
type Counter struct {
	metric

	val         uint64
	pendingIncr uint64
	intervalVal uint64
}

func NewCounter(name, tagExtractedName string, tags []*Tag) *Counter {
	return &Counter{metric: newMetric(name, tagExtractedName, tags)}
}

// Add adds the given value to the counter.
func (c *Counter) Add(amount uint64) {
	atomic.AddUint64(&c.val, amount)
	atomic.AddUint64(&c.pendingIncr, amount)
	c.markUsed()
}

// Inc icrements the counter by 1.
func (c *Counter) Inc() {
	c.Add(1)
}

// Latch returns the periodic counter value and clears it.
func (c *Counter) Latch() uint64 {
	val := atomic.SwapUint64(&c.pendingIncr, 0)
	c.intervalVal = val
	return val
}

// IntervalValue returns the lastest periodic counter value.
func (c *Counter) IntervalValue() uint64 {
	return c.intervalVal
}

// Value returns the global counter value.
func (c *Counter) Value() uint64 {
	return atomic.LoadUint64(&c.val)
}
