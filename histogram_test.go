package stats

import (
	"testing"

	hist "github.com/samaritan-proxy/circonusllhist"
	"github.com/stretchr/testify/assert"
)

func TestHistogramStatistic(t *testing.T) {
	hist := hist.New()
	for i := 2; i < 100; i++ {
		hist.RecordIntScale(int64(i), 0)
	}
	s := newHistogramStatistics(hist)
	assert.EqualValues(t, 98, s.SampleCount())
	assert.NotZero(t, s.SampleSum())
	assert.Equal(t, defaultSupportedQuantiles, s.SupportedQuantiles())
	assert.Equal(t, len(defaultSupportedQuantiles), len(s.ComputedQuantiles()))
	assert.Equal(t, defaultSupportedBuckets, s.SupportedBuckets())
	assert.Equal(t, len(s.SupportedBuckets()), len(s.ComputedBuckets()))
}

func TestHistogram(t *testing.T) {
	h := NewHistogram(nil, "foo.bar", "foo", nil)
	assert.Contains(t, h.Summary(), "No recorded values")

	itlStat := h.IntervalStatistics()
	cumStat := h.CumulativeStatistics()
	assert.Equal(t, uint64(0), itlStat.SampleCount())
	assert.Equal(t, uint64(0), cumStat.SampleCount())

	h.Record(3)
	itlStat = h.IntervalStatistics()
	cumStat = h.CumulativeStatistics()
	assert.Equal(t, uint64(0), itlStat.SampleCount())
	assert.Equal(t, uint64(0), cumStat.SampleCount())

	// refresh
	h.refreshIntervalStatistic()
	itlStat = h.IntervalStatistics()
	cumStat = h.CumulativeStatistics()
	assert.Equal(t, uint64(1), itlStat.SampleCount())
	assert.Equal(t, uint64(1), cumStat.SampleCount())

	// refresh twice
	h.refreshIntervalStatistic()
	itlStat = h.IntervalStatistics()
	cumStat = h.CumulativeStatistics()
	assert.Equal(t, uint64(0), itlStat.SampleCount())
	assert.Equal(t, uint64(1), cumStat.SampleCount())

	assert.Contains(t, h.Summary(), "P0")
	assert.Contains(t, h.Summary(), "P25")
	assert.Contains(t, h.Summary(), "P50")
	assert.Contains(t, h.Summary(), "P90")
	assert.Contains(t, h.Summary(), "P95")
	assert.Contains(t, h.Summary(), "P99")
	assert.Contains(t, h.Summary(), "P100")
}
