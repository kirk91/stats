package stats

// MetricsSnapshot represents the metrics snapshot in a particular time.
type MetricsSnapshot interface {
	// Gauges returns all known guages.
	Gauges() []*Gauge
	// Counters returns all known counters.
	Counters() []*Counter
	// Histograms returns all known histograms.
	Histograms() []*Histogram
}

var _ MetricsSnapshot = new(metricsSnapshot)

type metricsSnapshot struct {
	gauges     []*Gauge
	counters   []*Counter
	histograms []*Histogram
}

func newMetricsSnapshot(gauges []*Gauge, counters []*Counter, histograms []*Histogram) *metricsSnapshot {
	snap := &metricsSnapshot{
		gauges:     gauges,
		counters:   counters,
		histograms: histograms,
	}
	// refresh counter interval value.
	for _, counter := range snap.counters {
		counter.Latch()
	}
	// refresh histogram interval stattistic.
	for _, histogram := range snap.histograms {
		histogram.refreshIntervalStatistic()
	}
	return snap
}

func (snap *metricsSnapshot) Gauges() []*Gauge {
	return snap.gauges
}

func (snap *metricsSnapshot) Counters() []*Counter {
	return snap.counters
}

func (snap *metricsSnapshot) Histograms() []*Histogram {
	return snap.histograms
}

// Sink is a sink for stats. Each Sink is responsible for writing stats
// to a backing store.
type Sink interface {
	// Flush flushes periodic metrics to the backing store.
	Flush(MetricsSnapshot) error
	// WriteHistogram writes a single histogram sample to the backing store directly.
	// It is only used to be compatible with the TSDB which doesn't support
	// pre-aggreated histogram data.
	WriteHistogramSample(h *Histogram, val uint64) error
}
