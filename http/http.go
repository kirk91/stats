package http

import (
	"net/http"

	"github.com/kirk91/stats"
)

type statsFormat string

const (
	formatPlain      statsFormat = "plain"
	formatPrometheus             = "prometheus"
)

// Handler returns an HTTP handler that shows the metrics by text in the store.
func Handler(store *stats.Store) http.Handler {
	return newHandler(store, formatPlain)
}

// PrometheusHandler returns an HTTP handler that shows the metrics
// by prometheus in the stats store.
func PrometheusHandler(store *stats.Store) http.Handler {
	return newHandler(store, formatPrometheus)
}

type handler struct {
	*stats.Store
	format statsFormat
}

func newHandler(store *stats.Store, format statsFormat) *handler {
	return &handler{
		Store:  store,
		format: format,
	}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: add metrics filter
	formater := newStatsFormater(h.format)
	b := formater.Format(
		h.Store.Gauges(),
		h.Store.Counters(),
		h.Store.Histograms(),
	)
	w.Write(b)
}

func newStatsFormater(format statsFormat) statsFormater {
	switch format {
	case formatPrometheus:
		return newPrometheusStatsFormatter()
	case formatPlain:
		fallthrough
	default:
		return newPlainStatsFormatter()
	}
}
