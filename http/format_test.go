package http

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/kirk91/stats"
	"github.com/stretchr/testify/assert"
)

func TestPlainFormatter(t *testing.T) {
	store := stats.NewStore(nil)
	scope := store.CreateScope("plain-stats")
	defer store.DeleteScope(scope)

	g1 := scope.Gauge("g1")
	g1.Inc()
	c1 := scope.Counter("c1")
	c1.Inc()
	h1 := scope.Histogram("h1")
	h1.Record(1)

	ff := newPlainFormatterFactory()
	f := ff.Create()
	res := f.Format(scope.Gauges(), scope.Counters(), scope.Histograms())

	var expect bytes.Buffer
	expect.WriteString(fmt.Sprintf("%s: %v\n", c1.Name(), c1.Value()))
	expect.WriteString(fmt.Sprintf("%s: %v\n", g1.Name(), g1.Value()))
	expect.WriteString(fmt.Sprintf("%s: %v\n", h1.Name(), h1.Summary()))
	assert.Equal(t, expect.Bytes(), res)
}

func TestFormatGaugeForPrometheus(t *testing.T) {
	g1 := stats.NewGauge("foo.sash", "foo",
		[]*stats.Tag{{Name: "tag1", Value: "sash"}})
	g2 := stats.NewGauge("foo.bos", "foo",
		[]*stats.Tag{{Name: "tag1", Value: "bos"}})
	g3 := stats.NewGauge("bar", "bar", nil)
	g1.Inc()
	g2.Set(2)

	ff := newPrometheusFormatterFactory("myapp")
	f := ff.Create()
	res := f.Format([]*stats.Gauge{g1, g2, g3}, nil, nil)
	expect := `# TYPE myapp_foo gauge
myapp_foo{tag1="sash"} 1
myapp_foo{tag1="bos"} 2
# TYPE myapp_bar gauge
myapp_bar{} 0
`
	assert.Equal(t, expect, string(res))
}

func TestFormatCounterForPrometheus(t *testing.T) {
	c1 := stats.NewCounter("foo.sash", "foo",
		[]*stats.Tag{{Name: "tag1", Value: "sash"}})
	c2 := stats.NewCounter("foo.bos", "foo",
		[]*stats.Tag{{Name: "tag1", Value: "bos"}})
	c3 := stats.NewCounter("bar", "bar", nil)
	c1.Inc()
	c2.Add(2)

	ff := newPrometheusFormatterFactory("myapp")
	f := ff.Create()
	res := f.Format(nil, []*stats.Counter{c1, c2, c3}, nil)
	expect := `# TYPE myapp_foo counter
myapp_foo{tag1="sash"} 1
myapp_foo{tag1="bos"} 2
# TYPE myapp_bar counter
myapp_bar{} 0
`
	assert.Equal(t, expect, string(res))
}

func TestFormatHistogramForPrometheus(t *testing.T) {
	t.Run("no values and tags", func(t *testing.T) {
		h := stats.NewHistogram(nil, "h", "h", nil)
		ff := newPrometheusFormatterFactory("myapp")
		f := ff.Create()
		res := f.Format(nil, nil, []*stats.Histogram{h})
		expect := `# TYPE myapp_h histogram
myapp_h_bucket{le="0.5"} 0
myapp_h_bucket{le="1"} 0
myapp_h_bucket{le="5"} 0
myapp_h_bucket{le="10"} 0
myapp_h_bucket{le="25"} 0
myapp_h_bucket{le="50"} 0
myapp_h_bucket{le="100"} 0
myapp_h_bucket{le="250"} 0
myapp_h_bucket{le="500"} 0
myapp_h_bucket{le="1000"} 0
myapp_h_bucket{le="2500"} 0
myapp_h_bucket{le="5000"} 0
myapp_h_bucket{le="10000"} 0
myapp_h_bucket{le="30000"} 0
myapp_h_bucket{le="60000"} 0
myapp_h_bucket{le="300000"} 0
myapp_h_bucket{le="600000"} 0
myapp_h_bucket{le="1800000"} 0
myapp_h_bucket{le="3600000"} 0
myapp_h_bucket{le="+Inf"} 0
myapp_h_sum{} 0
myapp_h_count{} 0
`
		assert.Equal(t, expect, string(res))
	})

	t.Run("has values and tags", func(t *testing.T) {
		h1 := stats.NewHistogram(nil, "h", "h", []*stats.Tag{{Name: "tag1", Value: "foo"}})
		h2 := stats.NewHistogram(nil, "h", "h", []*stats.Tag{{Name: "tag2", Value: "bar"}})
		h1.Record(800)
		h1.Record(8000)
		h1.RefreshIntervalStatistics()
		h2.Record(50000)
		h2.RefreshIntervalStatistics()

		ff := newPrometheusFormatterFactory("myapp")
		f := ff.Create()
		res := f.Format(nil, nil, []*stats.Histogram{h1, h2})
		expect := `# TYPE myapp_h histogram
myapp_h_bucket{tag1="foo",le="0.5"} 0
myapp_h_bucket{tag1="foo",le="1"} 0
myapp_h_bucket{tag1="foo",le="5"} 0
myapp_h_bucket{tag1="foo",le="10"} 0
myapp_h_bucket{tag1="foo",le="25"} 0
myapp_h_bucket{tag1="foo",le="50"} 0
myapp_h_bucket{tag1="foo",le="100"} 0
myapp_h_bucket{tag1="foo",le="250"} 0
myapp_h_bucket{tag1="foo",le="500"} 0
myapp_h_bucket{tag1="foo",le="1000"} 1
myapp_h_bucket{tag1="foo",le="2500"} 1
myapp_h_bucket{tag1="foo",le="5000"} 1
myapp_h_bucket{tag1="foo",le="10000"} 2
myapp_h_bucket{tag1="foo",le="30000"} 2
myapp_h_bucket{tag1="foo",le="60000"} 2
myapp_h_bucket{tag1="foo",le="300000"} 2
myapp_h_bucket{tag1="foo",le="600000"} 2
myapp_h_bucket{tag1="foo",le="1800000"} 2
myapp_h_bucket{tag1="foo",le="3600000"} 2
myapp_h_bucket{tag1="foo",le="+Inf"} 2
myapp_h_sum{tag1="foo"} 8855
myapp_h_count{tag1="foo"} 2
myapp_h_bucket{tag2="bar",le="0.5"} 0
myapp_h_bucket{tag2="bar",le="1"} 0
myapp_h_bucket{tag2="bar",le="5"} 0
myapp_h_bucket{tag2="bar",le="10"} 0
myapp_h_bucket{tag2="bar",le="25"} 0
myapp_h_bucket{tag2="bar",le="50"} 0
myapp_h_bucket{tag2="bar",le="100"} 0
myapp_h_bucket{tag2="bar",le="250"} 0
myapp_h_bucket{tag2="bar",le="500"} 0
myapp_h_bucket{tag2="bar",le="1000"} 0
myapp_h_bucket{tag2="bar",le="2500"} 0
myapp_h_bucket{tag2="bar",le="5000"} 0
myapp_h_bucket{tag2="bar",le="10000"} 0
myapp_h_bucket{tag2="bar",le="30000"} 0
myapp_h_bucket{tag2="bar",le="60000"} 1
myapp_h_bucket{tag2="bar",le="300000"} 1
myapp_h_bucket{tag2="bar",le="600000"} 1
myapp_h_bucket{tag2="bar",le="1800000"} 1
myapp_h_bucket{tag2="bar",le="3600000"} 1
myapp_h_bucket{tag2="bar",le="+Inf"} 1
myapp_h_sum{tag2="bar"} 50500
myapp_h_count{tag2="bar"} 1
`
		assert.Equal(t, expect, string(res))
	})
}
