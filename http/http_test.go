package http

import (
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kirk91/stats"
	"github.com/stretchr/testify/assert"
)

func TestPlainHandler(t *testing.T) {
	store := stats.NewStore(nil)
	scope := store.CreateScope("plain")
	defer store.DeleteScope(scope)

	counter1 := scope.Counter("counter1")
	counter1.Inc()
	gauge1 := scope.Gauge("gauge1")
	gauge1.Inc()

	ts := httptest.NewServer(Handler(store))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, http.StatusOK, res.StatusCode)
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	assert.Contains(t, string(data), fmt.Sprintf("%s: %d", counter1.Name(), counter1.Value()))
	assert.Contains(t, string(data), fmt.Sprintf("%s: %d", gauge1.Name(), gauge1.Value()))
}

func TestPrometheusHandler(t *testing.T) {
	store := stats.NewStore(nil)
	scope := store.CreateScope("prometheus")
	defer store.DeleteScope(scope)

	counter1 := scope.Counter("counter1")
	counter1.Inc()
	gauge1 := scope.Gauge("gauge1")
	gauge1.Inc()

	ts := httptest.NewServer(PrometheusHandler(store, "myapp"))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, http.StatusOK, res.StatusCode)
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	body := string(b)
	assert.Contains(t, body, "# TYPE myapp_prometheus_gauge1 gauge\n")
	assert.Contains(t, body, "myapp_prometheus_gauge1{} 1\n")
	assert.Contains(t, body, "# TYPE myapp_prometheus_counter1 counter\n")
	assert.Contains(t, body, "myapp_prometheus_counter1{} 1\n")
}

func TestGzipAccpeted(t *testing.T) {
	tests := []struct {
		acceptEncoding string
		expected       bool
	}{
		{"", false},
		{"*", false},
		{"identity", false},
		{"gzip", true},
		{"deflate, gzip;q=1.0, *;q=0.5", true},
	}

	for i, test := range tests {
		name := fmt.Sprintf("case %d", i+1)
		t.Run(name, func(t *testing.T) {
			header := http.Header{}
			header.Set("Accept-Encoding", test.acceptEncoding)
			assert.Equal(
				t,
				test.expected,
				gzipAccepted(header),
			)
		})
	}
}

func TestHandlerWrite(t *testing.T) {
	t.Run("identity", func(t *testing.T) {
		rw := httptest.NewRecorder()
		req := &http.Request{}
		b := []byte("hello, world")

		h := &handler{}
		h.write(rw, req, b)
		assert.Equal(t, b, rw.Body.Bytes())
	})

	t.Run("gzip", func(t *testing.T) {
		rw := httptest.NewRecorder()
		req := &http.Request{
			Header: http.Header{},
		}
		req.Header.Set("Accept-Encoding", "gzip")
		b := []byte("hello, world")

		h := &handler{}
		h.write(rw, req, b)

		// decode response
		gz, err := gzip.NewReader(rw.Body)
		assert.NoError(t, err)
		actual, err := ioutil.ReadAll(gz)
		assert.NoError(t, err)
		assert.Equal(t, b, actual)
	})
}
