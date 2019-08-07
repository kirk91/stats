package stats

import "time"

// StoreOption contains options of a Store like flush interval, default tags, etc.
type StoreOption struct {
	FlushInterval time.Duration
	Sinks         []Sink
}

const defaultStoreFlushInterval = time.Second * 5

// NewStoreOption creates a StoreOption with FlushInterval set to 5s.
func NewStoreOption() *StoreOption {
	return &StoreOption{FlushInterval: defaultStoreFlushInterval}
}

// WithFlushInterval returns a StoreOption that sets flush interval for the store.
func (opt *StoreOption) WithFlushInterval(interval time.Duration) *StoreOption {
	opt.FlushInterval = interval
	return opt
}

// WithSinks returns a StoreOption that sets sinks for the sotre.
func (opt *StoreOption) WithSinks(sinks ...Sink) *StoreOption {
	opt.Sinks = sinks
	return opt
}
