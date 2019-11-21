package http

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/kirk91/stats"
)

const (
	headerContentEncoding = "Content-Encoding"
	headerAccpetEncoding  = "Accept-Encoding"
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
	h.write(w, r, b)
}

var gzipPool = sync.Pool{
	New: func() interface{} {
		return gzip.NewWriter(nil)
	},
}

func (h *handler) write(rw http.ResponseWriter, req *http.Request, b []byte) {
	w := io.Writer(rw)

	// check if accept gzip encoding
	if gzipAccepted(req.Header) {
		gw := gzipPool.Get().(*gzip.Writer)
		defer gzipPool.Put(gw)

		gw.Reset(w)
		defer gw.Close()
		w = gw
		// set content-encoding header
		rw.Header().Set(headerContentEncoding, "gzip")
	}

	w.Write(b) //nolint:errcheck
}

func gzipAccepted(header http.Header) bool {
	val := header.Get(headerAccpetEncoding)
	parts := strings.Split(val, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		// Accept-Encoding: gzip
		// Accept-Encoding: deflate, gzip;q=1.0, *;q=0.5
		if part == "gzip" || strings.HasPrefix(part, "gzip;") {
			return true
		}
	}
	return false
}
