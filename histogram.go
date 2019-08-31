package stats

import (
	"runtime"
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

	sampleCount  uint64
	rawHistCount uint64
	rawHists     []*hist.Histogram

	itlHist *hist.Histogram // interval hist
	cumHist *hist.Histogram // cumulative hist
}

func NewHistogram(store *Store, name, tagExtractedName string, tags []*Tag) *Histogram {
	h := &Histogram{
		store:        store,
		metric:       newMetric(name, tagExtractedName, tags),
		itlHist:      hist.NewNoLocks(),
		cumHist:      hist.New(),
		rawHistCount: uint64(runtime.GOMAXPROCS(0)),
	}
	h.rawHists = make([]*hist.Histogram, h.rawHistCount)
	for i := uint64(0); i < h.rawHistCount; i++ {
		h.rawHists[i] = hist.New()
	}
	return h
}

// Record records a value to the Histogram.
func (h *Histogram) Record(val uint64) {
	rawHist := h.rawHists[atomic.AddUint64(&h.sampleCount, 1)%h.rawHistCount]
	rawHist.RecordIntScale(int64(val), 0)
	if h.store != nil {
		h.store.deliverHistogramSampleToSinks(h, val)
	}
	h.markUsed()
}

func (h *Histogram) refreshIntervalStatistic() {
	// merge all raw hists
	merged := hist.NewNoLocks()
	for _, rawHist := range h.rawHists {
		merged.Merge(rawHist)
	}
	h.itlHist = merged.Copy()
	h.cumHist.Merge(merged)
}

// IntervalStatistics returns the interval statistics of Histogram.
func (h *Histogram) IntervalStatistics() *HistogramStatistics {
	return newHistogramStatistics(h.itlHist)
}
