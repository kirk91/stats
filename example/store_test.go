package example

import (
	"github.com/kirk91/stats"
	"github.com/kirk91/stats/sink/statsd"
)

const (
	appid = "arch.samaritan"
	ezone = "wg1"
)

func ExampleStore() {
	// NOTE: If the underlying sink supports tag feature, you could
	// add it depends on demand. Otherwise, you should ignore the
	// following section.
	defaultTags := map[string]string{"ezone": ezone}
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
	store := stats.NewStore(stats.NewStoreOption())
	store.SetTagOption(
		stats.NewTagOption().
			WithDefaultTags(defaultTags).
			WithTagExtractStrategies(tagExtractStrategies...))

	// statsd
	statsdAddr := "127.0.0.1:8125"
	statsdSink := statsd.NewSink(statsdAddr, appid)
	store.AddSink(statsdSink)

	scope := store.CreateScope("listener")
	scope.Counter("accept").Inc()

	//Output:
}
