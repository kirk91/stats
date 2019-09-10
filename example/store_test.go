package example

import (
	"context"

	"github.com/kirk91/stats"
	"github.com/kirk91/stats/sink/statsd"
)

const (
	appid  = "samaritan"
	region = "aisa-ch"
	zone   = "hz"
)

func ExampleStore() {
	store := stats.NewStore(stats.NewStoreOption())

	// NOTE: If the underlying sink supports tag feature, you could
	// add it depends on demand. Otherwise, you should ignore the
	// following section.
	defaultTags := map[string]string{
		"region": region,
		"zone":   zone,
	}
	tagExtractStrategies := []stats.TagExtractStrategy{
		{
			// service.(<service_name>.)*
			Name:  "service_name",
			Regex: "^service\\.((.*?)\\.)",
		},
		{
			// service.[<service_name>.]redis.(<redis_cmd>.)<base_stat>
			Name:  "redis_cmd",
			Regex: "^service(?:\\.).*?\\.redis\\.((.*?)\\.)",
		},
	}
	tagOption := stats.NewTagOption().WithDefaultTags(defaultTags).
		WithTagExtractStrategies(tagExtractStrategies...)
	store.SetTagOption(tagOption)

	// add statsd sink
	statsdAddr := "127.0.0.1:8125"
	statsdSink := statsd.NewSink(statsdAddr, appid)
	store.AddSink(statsdSink)

	// start the background routine which flushes the metrics periodically.
	ctx, cancel := context.WithCancel(context.TODO())
	loopDone := make(chan struct{})
	go func() {
		store.FlushingLoop(ctx)
		close(loopDone)
	}()
	defer func() {
		cancel()
		<-loopDone
	}()

	// create a stats group, the metrics which under it will sharing the same prefix.
	scope := store.CreateScope("listener")
	scope.Counter("conn_create").Inc()           // listener.conn_create
	scope.Gauge("conn_active").Inc()             // listener.conn_active
	scope.Histogram("conn_length_sec").Record(1) // listener.conn_length_sec

	//Output:
}
