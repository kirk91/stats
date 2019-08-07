package stats

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithDefaultTagsSetsValue(t *testing.T) {
	t.Parallel()
	opt := NewTagOption()

	newTags := map[string]string{"a": "1"}
	assert.NotEqual(t, newTags, opt.DefaultTags)
	opt.WithDefaultTags(newTags)
	assert.Equal(t, newTags, opt.DefaultTags)
}

func TestDefaultTagsLengthOK(t *testing.T) {
	t.Parallel()
	opt := NewTagOption().WithDefaultTags(map[string]string{"a": "1"})

	assert.Equal(t, len(opt.DefaultTags), len(opt.defaultTags()))
}
