package stats

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricsSnapshot(t *testing.T) {
	t.Run("gauges", func(t *testing.T) {
		gauge1 := new(Gauge)
		gauge1.Set(1)

		gauges := []*Gauge{gauge1}
		snapshot := newMetricsSnapshot(gauges, nil, nil)
		assert.Len(t, snapshot.Gauges(), 1)
	})

	t.Run("counters", func(t *testing.T) {
		counter1 := new(Counter)
		counter1.Inc()

		counters := []*Counter{counter1}
		snapshot := newMetricsSnapshot(nil, counters, nil)
		assert.Len(t, snapshot.Counters(), 1)
		assert.EqualValues(t, 1, snapshot.Counters()[0].IntervalValue())

		// interval value not change when counter update
		counter1.Inc()
		assert.EqualValues(t, 1, snapshot.Counters()[0].IntervalValue())
	})

	t.Run("histograms", func(t *testing.T) {
		histogram1 := NewHistogram(nil, "", "", nil)
		histogram1.Record(1)

		histograms := []*Histogram{histogram1}
		snapshot := newMetricsSnapshot(nil, nil, histograms)
		assert.Len(t, snapshot.Histograms(), 1)
		assert.EqualValues(t, 1, snapshot.Histograms()[0].IntervalStatistics().SampleCount())

		// interval statistic not change when histogram update
		histogram1.Record(1)
		assert.EqualValues(t, 1, snapshot.Histograms()[0].IntervalStatistics().SampleCount())
	})
}
