package statsd

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kirk91/stats"
)

type statsdServer struct {
	l      net.PacketConn
	closed bool
	buf    bytes.Buffer
}

func newStatsdServer(t *testing.T) *statsdServer {
	l, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("new udp listener failed: %v", err)
	}

	s := &statsdServer{l: l}

	go func() {
		b := make([]byte, 1024)
		for {
			if s.closed {
				return
			}
			n, _, err := l.ReadFrom(b)
			if n == 0 {
				continue
			}
			assert.NoError(t, err)
			s.buf.Write(b[:n])
		}
	}()

	return s
}

func (s *statsdServer) Addr() string {
	return s.l.LocalAddr().String()
}

func (s *statsdServer) Close() {
	s.closed = true
	s.l.Close()
}

func (s *statsdServer) Reset() {
	s.buf.Reset()
}

func (s *statsdServer) Content() string {
	return string(s.buf.Bytes())
}

func getHostname() string {
	s, _ := os.Hostname()
	return strings.Replace(s, ".", "_", -1)
}

func TestFlushCounters(t *testing.T) {
	ss := newStatsdServer(t)
	defer ss.Close()

	prefix := "arch.samaritan"
	s := NewSink(ss.Addr(), prefix)

	c := stats.NewCounter("foo", "", nil)
	c.Inc()
	c.Latch()
	cs := []*stats.Counter{c}

	cli, err := s.getClient()
	assert.NoError(t, err)
	s.flushCounters(cli, cs)
	time.Sleep(time.Millisecond * 200)
	assert.Contains(t, ss.Content(), fmt.Sprintf("%s.%s.foo:1|c", prefix, getHostname()))
}

func TestFlushGauges(t *testing.T) {
	ss := newStatsdServer(t)
	defer ss.Close()

	prefix := "arch.samaritan"
	s := NewSink(ss.Addr(), prefix)

	g := stats.NewGauge("bar", "", nil)
	g.Inc()
	g.Inc()
	g.Dec()
	gs := []*stats.Gauge{g}

	cli, err := s.getClient()
	assert.NoError(t, err)
	s.flushGauges(cli, gs)
	time.Sleep(time.Millisecond * 200)
	assert.Contains(t, ss.Content(), fmt.Sprintf("%s.%s.bar:1|g", prefix, getHostname()))
}

func TestWriteHistogramSample(t *testing.T) {
	ss := newStatsdServer(t)
	defer ss.Close()

	prefix := "arch.samaritan"
	s := NewSink(ss.Addr(), prefix)

	h := stats.NewHistogram(nil, "haha", "", nil)
	s.WriteHistogramSample(h, uint64(10))
	time.Sleep(time.Millisecond * 200)
	assert.Equal(t, ss.Content(), fmt.Sprintf("%s.%s.haha:10|ms\n", prefix, getHostname()))
}
