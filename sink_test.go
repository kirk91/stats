package stats

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSourceCachedCounters(t *testing.T) {
	counters := []*Counter{new(Counter)}
	src := &source{
		rawCounters: func() []*Counter {
			return counters
		},
	}

	cachedCounters := src.CachedCounters()
	assert.Equal(t, len(cachedCounters), 1)

	counter := new(Counter)
	counter.Add(2)
	counters = append(counters, counter) // update original counters
	cachedCounters = src.CachedCounters()
	assert.Equal(t, len(cachedCounters), 1) // cached counters not change
	assert.Equal(t, 0, int(counter.IntervalValue()))

	src.ClearCache() // clear cache
	cachedCounters = src.CachedCounters()
	assert.Equal(t, len(cachedCounters), 2)
	assert.Equal(t, 2, int(counter.IntervalValue()))
}

func TestSourceCachedGauges(t *testing.T) {
	gauges := []*Gauge{new(Gauge)}
	src := &source{
		rawGauges: func() []*Gauge {
			return gauges
		},
	}

	cachedGagues := src.CachedGauges()
	assert.Equal(t, len(cachedGagues), 1)

	gauges = append(gauges, new(Gauge))
	cachedGagues = src.CachedGauges()
	assert.Equal(t, len(cachedGagues), 1)

	src.ClearCache()
	cachedGagues = src.CachedGauges()
	assert.Equal(t, len(cachedGagues), 2)
}

func TestSourceCachedHistograms(t *testing.T) {
	histograms := []*Histogram{NewHistogram(nil, "", "", nil)}
	src := &source{
		rawHistograms: func() []*Histogram {
			return histograms
		},
	}

	cachedHistograms := src.CachedHistograms()
	assert.Equal(t, len(cachedHistograms), 1)

	histograms = append(histograms, NewHistogram(nil, "", "", nil))
	cachedHistograms = src.CachedHistograms()
	assert.Equal(t, len(cachedHistograms), 1)

	src.ClearCache()
	cachedHistograms = src.CachedHistograms()
	assert.Equal(t, len(cachedHistograms), 2)
}
