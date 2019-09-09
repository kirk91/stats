package stats

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	hist "github.com/samaritan-proxy/circonusllhist"
)

var (
	defaultSupportedQuantiles = []float64{0.0, 0.25, 0.5, 0.9, 0.95, 0.99, 1.0}
)

// HistogramStatistics holds the computed statistic of a histogram.
type HistogramStatistics struct {
	*hist.Histogram
	mu sync.RWMutex

	// supported quantiles
	qs []float64
	// computed quantile values
	qvals []float64
}

func newHistogramStatistics(h *hist.Histogram) *HistogramStatistics {
	qs := defaultSupportedQuantiles
	qvals, _ := h.ApproxQuantile(qs)
	return &HistogramStatistics{
		Histogram: h,
		qs:        qs,
		qvals:     qvals,
	}
}

// SupportedQuantiles returns the supported quantiles.
func (hs *HistogramStatistics) SupportedQuantiles() []float64 {
	return hs.qs
}

// ComputedQuantiles returns the computed quantile values during the period.
func (hs *HistogramStatistics) ComputedQuantiles() []float64 {
	if len(hs.qvals) == 0 {
		return make([]float64, len(hs.qs))
	}
	return hs.qvals
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

func (h *Histogram) refreshIntervalStatistic() {
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
	return newHistogramStatistics(h.itl)
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
