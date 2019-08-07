package stats

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsAlNum(t *testing.T) {
	for _, r := range "0123456789" {
		assert.Equal(t, true, isAlNum(r))
	}
	for _, r := range "abcdefghijklmnopqrstuvwxyz" {
		assert.Equal(t, true, isAlNum(r))
	}
	for _, r := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		assert.Equal(t, true, isAlNum(r))
	}

	for _, r := range "_\\()/-.$^" {
		assert.Equal(t, false, isAlNum(r))
	}
}

func TestExtractRegexPrefix(t *testing.T) {
	extractRegexPrefix := func(regex string) string {
		e, err := NewTagExtractor("foo", regex, "")
		if err != nil {
			t.Fatal(err)
		}
		return e.prefixToken()
	}

	assert.Equal(t, "prefix", extractRegexPrefix("^prefix\\.foo"))
	assert.Equal(t, "prefix_optional", extractRegexPrefix("^prefix_optional(?:\\.)"))
	assert.Equal(t, "prefix", extractRegexPrefix("^prefix$"))
	assert.Equal(t, "", extractRegexPrefix("(prefix)"))
	assert.Equal(t, "", extractRegexPrefix("^(prefix)"))
	assert.Equal(t, "", extractRegexPrefix("^prefix(foo)"))
	assert.Equal(t, "", extractRegexPrefix("prefix(foo)"))
}

func TestRemoveCharacters(t *testing.T) {
	s := "hello, kimmy & hello, jimmy"
	removeChars := new(IntervalSet)
	removeChars.Insert(0, 7)   // [0, 6)
	removeChars.Insert(14, 21) // [14, 21)
	assert.Equal(t, "kimmy & jimmy", removeCharacters(s, removeChars))
}

func TestExtractTag(t *testing.T) {
	regex := "^service\\.((.*?)\\.)"
	e, err := NewTagExtractor("service_name", regex, "service")
	if err != nil {
		t.Fatal(err)
	}

	// substr mismatch
	assert.False(t, e.extract("broker.zk.avaialbe", nil, nil))

	tags := make([]*Tag, 0, 1)
	removeChars := new(IntervalSet)
	name := "service.corvus_111.downstream_cx_total"
	e.extract(name, &tags, removeChars)
	tagExtractedName := removeCharacters(name, removeChars)
	assert.Equal(t, "service.downstream_cx_total", tagExtractedName)
	assert.Equal(t, 1, len(tags))
	assert.Equal(t, tags[0].Name, "service_name")
	assert.Equal(t, tags[0].Value, "corvus_111")
}

func TestProduceTags(t *testing.T) {
	defaultTags := []*Tag{
		&Tag{Name: "tag1", Value: "hhm"},
		&Tag{Name: "tag2", Value: "wm"},
	}
	p := NewTagProducer(defaultTags...)

	regex := "^service\\.((.*?)\\.)"
	e1, err := NewTagExtractor("service_name", regex, "service")
	if err != nil {
		t.Fatal(err)
	}
	p.AddExtractor(e1)

	regex = "^service(?:\\.).*?\\.redis\\.((.*?)\\.)"
	e2, err := NewTagExtractor("redis_cmd", regex, ".redis.")
	if err != nil {
		t.Fatal(err)
	}
	p.AddExtractor(e2)

	name := "service.corvus_111.downstream_cx_total"
	tagExtractedName, tags := p.Produce(name)
	assert.Equal(t, "service.downstream_cx_total", tagExtractedName)
	assert.Equal(t, 3, len(tags))
	assert.Equal(t, tags[2].Name, "service_name")
	assert.Equal(t, tags[2].Value, "corvus_111")

	name = "service.corvus_111.redis.get.total"
	tagExtractedName, tags = p.Produce(name)
	assert.Equal(t, "service.redis.total", tagExtractedName)
	assert.Equal(t, 4, len(tags))
	assert.Equal(t, tags[3].Name, "redis_cmd")
	assert.Equal(t, tags[3].Value, "get")
}
