package http

import (
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

	ts := httptest.NewServer(PrometheusHandler(store))
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
	assert.Contains(t, body, "# TYPE samaritan_prometheus_gauge1 gauge\n")
	assert.Contains(t, body, "samaritan_prometheus_gauge1{} 1\n")
	assert.Contains(t, body, "# TYPE samaritan_prometheus_counter1 counter\n")
	assert.Contains(t, body, "samaritan_prometheus_counter1{} 1\n")
}
