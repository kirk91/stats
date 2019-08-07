package stats

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWithFlushIntervalSetsValue(t *testing.T) {
	t.Parallel()

	opt := NewStoreOption()
	oldInterval := opt.FlushInterval
	newInterval := time.Hour
	assert.NotEqual(t, oldInterval, newInterval)

	opt.WithFlushInterval(newInterval)
	assert.Equal(t, newInterval, opt.FlushInterval)
}

// func TestWithSinksSetsValue(t *testing.T) {
// t.Parallel()
// ctl := gomock.NewController(t)
// defer ctl.Finish()

// // mockSink := NewMockSink(ctl)

// opt := NewStoreOption()
// oldSinks := opt.Sinks
// newSinks := []Sink{mockSink}
// assert.NotEqual(t, oldSinks, newSinks)

// opt.WithSinks(newSinks...)
// assert.Equal(t, newSinks, opt.Sinks)
// }
