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
	hstats := newHistogramStatistics(hist)
	assert.Equal(t, defaultSupportedQuantiles, hstats.SupportedQuantiles())
	assert.Equal(t, len(defaultSupportedQuantiles), len(hstats.ComputedQuantiles()))
}

func TestHistogram(t *testing.T) {
	h := NewHistogram(nil, "foo.bar", "foo", nil)
	assert.Contains(t, h.Summary(), "No recorded values")

	itlStat := h.IntervalStatistics()
	assert.Equal(t, uint64(0), itlStat.SampleCount())

	h.Record(3)
	itlStat = h.IntervalStatistics()
	assert.Equal(t, uint64(0), itlStat.SampleCount())

	// refresh
	h.refreshIntervalStatistic()
	itlStat = h.IntervalStatistics()
	assert.Equal(t, uint64(1), itlStat.SampleCount())

	// refresh twice
	h.refreshIntervalStatistic()
	itlStat = h.IntervalStatistics()
	assert.Equal(t, uint64(0), itlStat.SampleCount())

	assert.Contains(t, h.Summary(), "P0")
	assert.Contains(t, h.Summary(), "P25")
	assert.Contains(t, h.Summary(), "P50")
	assert.Contains(t, h.Summary(), "P90")
	assert.Contains(t, h.Summary(), "P95")
	assert.Contains(t, h.Summary(), "P99")
	assert.Contains(t, h.Summary(), "P100")
}
