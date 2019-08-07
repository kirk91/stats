package stats

import (
	"sort"
)

// Interval represents a continous span.
type Interval struct {
	Left  int
	Right int
}

func newInterval(left, right int) Interval {
	return Interval{
		Left:  left,
		Right: right,
	}
}

// IntervalSet is a interval set.
type IntervalSet struct {
	data []Interval
}

func newIntervalSet() *IntervalSet {
	return &IntervalSet{}
}

// Insert inserts a new span.
func (iset *IntervalSet) Insert(left, right int) {
	// skip the interval elemenets which are completely smaller than the new value.
	n := len(iset.data)
	i := sort.Search(len(iset.data), func(i int) bool {
		return !(iset.data[i].Right < left)
	})
	if i == n {
		iset.data = append(iset.data, newInterval(left, right))
		return
	}

	intervals := make([]Interval, i, len(iset.data))
	copy(intervals, iset.data[:i])

	if left >= iset.data[i].Left {
		left = iset.data[i].Left
	}

	for ; i < n; i++ {
		if right < iset.data[i].Left {
			intervals = append(intervals, newInterval(left, right))
			intervals = append(intervals, iset.data[i:]...)
			break
		}
		if right >= iset.data[i].Left && right <= iset.data[i].Right {
			intervals = append(intervals, newInterval(left, iset.data[i].Right))
			if i < n-1 {
				intervals = append(intervals, iset.data[i+1:]...)
			}
			break
		}
	}

	if i == n {
		intervals = append(intervals, newInterval(left, right))
	}
	iset.data = intervals
}

// AllIntervals returns all intervals.
func (iset *IntervalSet) AllIntervals() []Interval {
	intervals := make([]Interval, len(iset.data))
	copy(intervals, iset.data)
	return intervals
}
