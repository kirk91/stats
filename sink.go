package stats

// Source provides cached access to a particular store's stats.
type Source interface {
	// CachedCounters returns all known counters. Will use cached values
	// if already accessed and ClearCache hasn't been called since.
	CachedCounters() []*Counter
	// CachedGauges returns all known gauges. Will use cached values if
	// already accessed and ClearCache hasn't been called since.
	CachedGauges() []*Gauge
	// CachedHistograms returns all known histograms. Will use cached values
	// if already accessed and ClearCache hasn't been called since.
	CachedHistograms() []*Histogram

	// ClearCache resets the cache so that any future calls to get cached
	// metrics will refresh the set.
	ClearCache()
}

type source struct {
	rawCounters   func() []*Counter
	rawGauges     func() []*Gauge
	rawHistograms func() []*Histogram

	counters   []*Counter
	gauges     []*Gauge
	histograms []*Histogram
}

func (src *source) CachedCounters() []*Counter {
	if src.counters == nil {
		if src.rawCounters == nil {
			return nil
		}
		src.counters = src.rawCounters()
		for _, counter := range src.counters {
			counter.Latch()
		}
	}
	return src.counters
}

func (src *source) CachedGauges() []*Gauge {
	if src.gauges == nil {
		if src.rawGauges == nil {
			return nil
		}
		src.gauges = src.rawGauges()
	}
	return src.gauges
}

func (src *source) CachedHistograms() []*Histogram {
	if src.histograms == nil {
		if src.rawHistograms == nil {
			return nil
		}
		src.histograms = src.rawHistograms()
		for _, histogram := range src.histograms {
			histogram.refreshIntervalStatistic()
		}
	}
	return src.histograms
}

func (src *source) ClearCache() {
	src.counters = nil
	src.gauges = nil
	src.histograms = nil
}

// Sink is a sink for stats. Each Sink is responsible for writing stats
// to a backing store.
type Sink interface {
	// Flush flushes periodic stats to the backing store.
	Flush(Source) error
	// WriteHistogram writes a single histogram sample to the
	// backing store directly.
	WriteHistogramSample(h *Histogram, val uint64) error
}
