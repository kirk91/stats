package stats

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGauge(t *testing.T) {
	name := "blabla"
	tagExtractedName := "bla"
	tags := []*Tag{{Name: "tag1", Value: "dummy"}}
	g := NewGauge(name, tagExtractedName, tags)
	assert.Equal(t, g.Name(), name)
	assert.Equal(t, g.Tags(), tags)
	assert.Equal(t, g.TagExtractedName(), tagExtractedName)

	g.Set(1)
	assert.Equal(t, g.Value(), uint64(1))

	g.Inc()
	assert.Equal(t, g.Value(), uint64(2))

	g.Add(2)
	assert.Equal(t, g.Value(), uint64(4))

	g.Dec()
	assert.Equal(t, g.Value(), uint64(3))

	g.Sub(2)
	assert.Equal(t, g.Value(), uint64(1))
}

func TestCounter(t *testing.T) {
	name := "mong"
	tagExtractedName := "mo"
	c := NewCounter(name, tagExtractedName, nil)
	assert.Equal(t, c.Name(), name)
	assert.Equal(t, c.TagExtractedName(), tagExtractedName)
	assert.Equal(t, c.Tags(), []*Tag(nil))

	c.Inc()
	c.Add(2)
	assert.Equal(t, uint64(3), c.Latch())
	assert.Equal(t, uint64(3), c.Value())
	assert.Equal(t, uint64(3), c.IntervalValue())
	assert.Equal(t, uint64(0), c.Latch())
	assert.Equal(t, uint64(0), c.IntervalValue())

	c.Add(2)
	assert.Equal(t, uint64(2), c.Latch())
	assert.Equal(t, uint64(5), c.Value())
	assert.Equal(t, uint64(2), c.IntervalValue())
}
