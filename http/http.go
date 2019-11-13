package http

import (
	"net/http"

	"github.com/kirk91/stats"
)

// Handler returns an HTTP handler that shows the metrics by text in the store.
func Handler(store *stats.Store) http.Handler {
	ff := newPlainFormatterFactory()
	return newHandler(store, ff)
}

// PrometheusHandler returns an HTTP handler that shows the metrics
// by prometheus in the stats store.
func PrometheusHandler(store *stats.Store, namespace string) http.Handler {
	ff := newPrometheusFormatterFactory(namespace)
	return newHandler(store, ff)
}

type handler struct {
	*stats.Store
	ff formatterFactory
}

func newHandler(store *stats.Store, ff formatterFactory) *handler {
	return &handler{
		Store: store,
		ff:    ff,
	}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: add metrics filter
	formater := h.ff.Create()
	b := formater.Format(
		h.Store.Gauges(),
		h.Store.Counters(),
		h.Store.Histograms(),
	)
	w.Write(b) //nolint:errcheck
}
