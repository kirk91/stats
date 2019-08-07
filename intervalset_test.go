package stats

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntervalSet(t *testing.T) {
	iset := newIntervalSet()

	iset.Insert(3, 8)
	expected := []Interval{{3, 8}}
	assert.Equal(t, iset.AllIntervals(), expected)

	iset.Insert(12, 16)
	expected = []Interval{{3, 8}, {12, 16}}
	assert.Equal(t, iset.AllIntervals(), expected)

	iset.Insert(25, 36)
	expected = []Interval{{3, 8}, {12, 16}, {25, 36}}
	assert.Equal(t, iset.AllIntervals(), expected)

	iset.Insert(18, 20)
	expected = []Interval{{3, 8}, {12, 16}, {18, 20}, {25, 36}}
	assert.Equal(t, iset.AllIntervals(), expected)

	iset.Insert(10, 19)
	expected = []Interval{{3, 8}, {10, 20}, {25, 36}}
	assert.Equal(t, iset.AllIntervals(), expected)

	iset.Insert(15, 37)
	expected = []Interval{{3, 8}, {10, 37}}
	assert.Equal(t, iset.AllIntervals(), expected)
}
