package statsd

import (
	"time"

	"github.com/kirk91/statsd"
	"github.com/pkg/errors"

	"github.com/kirk91/stats"
)

var _ stats.Sink = new(Sink)

type Sink struct {
	address string
	prefix  string
	_client *statsd.Client
}

// NewSink returns a new Sink for statsd.
func NewSink(address string, prefix string) *Sink {
	return &Sink{address: address, prefix: prefix}
}

func (s *Sink) getClient() (*statsd.Client, error) {
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
func (s *Sink) Flush(source stats.Source) error {
	cli, err := s.getClient()
	if err != nil {
		return errors.Wrap(err, "error getting client")
	}
	s.flushCounters(cli, source.CachedCounters())
	s.flushGauges(cli, source.CachedGauges())
	return nil
}

func (s *Sink) flushCounters(cli *statsd.Client, cs []*stats.Counter) {
	for _, c := range cs {
		val := c.IntervalValue()
		cli.CountUint64fWithHost(val, c.Name())
	}
}

func (s *Sink) flushGauges(cli *statsd.Client, gs []*stats.Gauge) {
	for _, g := range gs {
		cli.GaugeUint64fWithHost(g.Value(), g.Name())
	}
}

func (s *Sink) WriteHistogramSample(h *stats.Histogram, val uint64) error {
	cli, err := s.getClient()
	if err != nil {
		return errors.Wrap(err, "error getting client")
	}
	cli.TimingfWithHost(time.Duration(val)*time.Millisecond, h.Name())
	return nil
}
