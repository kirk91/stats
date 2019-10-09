package stats

import (
	"fmt"
	"runtime"
	"strings"
	"sync/atomic"

	hist "github.com/samaritan-proxy/circonusllhist"
)

// the default value is enough is most scenarios.
var (
	defaultSupportedQuantiles = []float64{
		0.0, 0.25, 0.5, 0.9, 0.95, 0.99, 1.0,
	}
	defaultSupportedBuckets = []float64{
		0.5, 1, 5, 10, 25, 50, 100, 250, 500, 1000,
		2500, 5000, 10000, 30000, 60000, 300000, 600000, 1800000, 3600000,
	}
)

// HistogramStatistics holds the computed statistic of a histogram.
type HistogramStatistics struct {
	*hist.Histogram

	sampleCount uint64
	sampleSum   float64

	// quantile releated
	sqs []float64 // supported
	cqs []float64 // computed

	// bucket releated
	sbs []float64 // supported
	cbs []uint64  // computed
}

func newHistogramStatistics(h *hist.Histogram) *HistogramStatistics {
	s := &HistogramStatistics{
		Histogram:   h,
		sampleCount: h.SampleCount(),
		sampleSum:   h.ApproxSum(),
	}

	// quantiles
	s.sqs = defaultSupportedQuantiles
	s.cqs, _ = h.ApproxQuantile(s.sqs)

	// buckets
	s.sbs = defaultSupportedBuckets
	for _, b := range s.sbs {
		s.cbs = append(s.cbs, h.ApproxCountBelow(b))
	}

	return s
}

// SupportedQuantiles returns the supported quantiles.
func (s *HistogramStatistics) SupportedQuantiles() []float64 {
	return s.sqs
}

// ComputedQuantiles returns the computed quantile values during the period.
func (s *HistogramStatistics) ComputedQuantiles() []float64 {
	if len(s.cqs) == 0 {
		return make([]float64, len(s.sqs))
	}
	return s.cqs
}

// SupportedBuckets returns the supported buckets.
func (s *HistogramStatistics) SupportedBuckets() []float64 {
	return s.sbs
}

// ComputedBuckets returns the computed bucket values during the period.
func (s *HistogramStatistics) ComputedBuckets() []uint64 {
	if len(s.sbs) == 0 {
		return make([]uint64, len(s.sbs))
	}
	return s.cbs
}

// SampleCount returns the total number of value during the period.
func (s *HistogramStatistics) SampleCount() uint64 {
	return s.sampleCount
}

// SampleSum returns the sum of all values during the period.
func (s *HistogramStatistics) SampleSum() float64 {
	return s.sampleSum
}

// A Histogram records values one at a time.
type Histogram struct {
	metric
	store *Store

	sampleCount uint64
	rawCount    uint64
	raws        []*hist.Histogram

	itl *hist.Histogram // interval hist
	cum *hist.Histogram // cumulative hist
}

// NewHistogram creates a histogram with given params.
// NOTE: It should only be used in unit tests.
func NewHistogram(store *Store, name, tagExtractedName string, tags []*Tag) *Histogram {
	h := &Histogram{
		store:    store,
		metric:   newMetric(name, tagExtractedName, tags),
		itl:      hist.NewNoLocks(),
		cum:      hist.New(),
		rawCount: uint64(runtime.GOMAXPROCS(0)),
	}
	h.raws = make([]*hist.Histogram, h.rawCount)
	for i := uint64(0); i < h.rawCount; i++ {
		h.raws[i] = hist.New()
	}
	return h
}

// Record records a value to the Histogram.
func (h *Histogram) Record(val uint64) {
	raw := h.raws[atomic.AddUint64(&h.sampleCount, 1)%h.rawCount]
	raw.RecordIntScale(int64(val), 0)
	if h.store != nil {
		h.store.deliverHistogramSampleToSinks(h, val)
	}
	h.markUsed()
}

// RefreshIntervalStatistic refreshs the interval statistics of histogram.
// NOTE: It should only be used in unit tests.
func (h *Histogram) RefreshIntervalStatistics() {
	// merge and reset all raw hists
	merged := hist.NewNoLocks()
	for _, raw := range h.raws {
		merged.Merge(raw)
		// NOTE: Merge and reset is not atomic, maybe some
		// samples would be dropped.
		raw.FullReset()
	}
	h.itl = merged.Copy()
	h.cum.Merge(merged)
}

// IntervalStatistics returns the interval statistics of Histogram.
func (h *Histogram) IntervalStatistics() *HistogramStatistics {
	return newHistogramStatistics(h.itl.Copy())
}

// CumulativeStatistics returns the cumulative statistics of Histogram.
func (h *Histogram) CumulativeStatistics() *HistogramStatistics {
	return newHistogramStatistics(h.cum.Copy())
}

// Summary returns the summary of the histogram.
func (h *Histogram) Summary() string {
	if h.cum.SampleCount() == 0 {
		return "No recorded values"
	}

	itlStat := newHistogramStatistics(h.itl)
	cumStat := newHistogramStatistics(h.cum)
	var summary []string
	for i, q := range cumStat.SupportedQuantiles() {
		summary = append(summary,
			fmt.Sprintf("P%d(%d,%d)", int(100*q),
				int64(itlStat.ComputedQuantiles()[i]),
				int64(cumStat.ComputedQuantiles()[i]),
			),
		)
	}
	return strings.Join(summary, " ")
}
