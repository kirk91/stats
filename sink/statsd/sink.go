package statsd

import (
	"time"

	"github.com/kirk91/statsd"
	"github.com/pkg/errors"

	"github.com/kirk91/stats"
)

var _ stats.Sink = new(sink)

type sink struct {
	address string
	prefix  string
	_client *statsd.Client
}

// New returns a new sink for statsd.
func New(address string, prefix string) *sink {
	return &sink{address: address, prefix: prefix}
}

func (s *sink) getClient() (*statsd.Client, error) {
	var err error
	if s._client == nil {
		s._client, err = statsd.New("udp", s.address, statsd.Prefix(s.prefix))
		if err != nil {
			err = errors.Wrap(err, "error creating statsd client")
		}
	}
	return s._client, err
}

// Flush sends cached metrics from source to sink.
func (s *sink) Flush(snapshot stats.MetricsSnapshot) error {
	cli, err := s.getClient()
	if err != nil {
		return errors.Wrap(err, "error getting client")
	}
	s.flushCounters(cli, snapshot.Counters())
	s.flushGauges(cli, snapshot.Gauges())
	return nil
}

func (s *sink) flushCounters(cli *statsd.Client, cs []*stats.Counter) {
	for _, c := range cs {
		val := c.IntervalValue()
		cli.CountUint64fWithHost(val, c.Name())
	}
}

func (s *sink) flushGauges(cli *statsd.Client, gs []*stats.Gauge) {
	for _, g := range gs {
		cli.GaugeUint64fWithHost(g.Value(), g.Name())
	}
}

func (s *sink) WriteHistogramSample(h *stats.Histogram, val uint64) error {
	cli, err := s.getClient()
	if err != nil {
		return errors.Wrap(err, "error getting client")
	}
	cli.TimingfWithHost(time.Duration(val)*time.Millisecond, h.Name())
	return nil
}
