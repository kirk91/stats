package stats

import (
	"bytes"
	"regexp"
	"strings"
)

// Tag represents a tag for stats.
type Tag struct {
	Name  string
	Value string
}

// TagExtractor is used to extract tags from stats name.
type TagExtractor struct {
	name        string
	re          *regexp.Regexp
	regexPrefix string
	subStr      string
}

// NewTagExtractor creates a tag extractor.
func NewTagExtractor(name, regex, subStr string) (*TagExtractor, error) {
	// TODO(kirk91): check regex
	re, err := regexp.Compile(regex)
	if err != nil {
		return nil, err
	}

	e := &TagExtractor{
		name:   name,
		re:     re,
		subStr: subStr,
	}
	e.regexPrefix = e.extractRegexPrefix(regex)
	return e, nil
}

func (e *TagExtractor) extractRegexPrefix(regex string) string {
	if !strings.HasPrefix(regex, "^") {
		return ""
	}

	n := len(regex)
	for i := 1; i < n; i++ {
		if isAlNum(rune(regex[i])) || regex[i] == '_' {
			continue
		}
		if i == 1 {
			break
		}

		var isLast bool
		if i == n-1 {
			isLast = true
		}
		if (!isLast && isRegexStartsWithDot(regex[i:])) || (isLast && regex[i] == '$') {
			return regex[1:i]
		}
	}
	return ""
}

func isRegexStartsWithDot(regex string) bool {
	return strings.HasPrefix(regex, "\\.") || strings.HasPrefix(regex, "(?:\\.)")
}

func isAlNum(r rune) bool {
	if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
		return true
	}
	return false
}

func (e *TagExtractor) prefixToken() string {
	return e.regexPrefix
}

func (e *TagExtractor) extract(metricName string, tags *[]*Tag, removeChars *IntervalSet) bool {
	if e.subStr != "" && !strings.Contains(metricName, e.subStr) {
		return false
	}

	matches := findStringSubmatch(e.re, metricName)
	if len(matches) <= 1 {
		return false
	}

	// removeSubExpr is the first submatch, It represents the portion of the string to be removed.
	removeSubExpr := matches[1]

	// valueSubExpr is the optional second submatch, It is usually inside the first submatch
	// to allow the expression to strip off extra characters that should be removed from
	// the string but also not necessary in the tag value ("." for example). If there is no
	// second submatch, then the valueSubExpr is the same as removeSubExpr.
	valueSubExpr := removeSubExpr
	if len(matches) > 2 {
		valueSubExpr = matches[2]
	}

	tag := &Tag{
		Name:  e.name,
		Value: valueSubExpr.val,
	}
	*tags = append(*tags, tag)

	removeChars.Insert(removeSubExpr.start, removeSubExpr.end)
	return true
}

type match struct {
	val        string
	start, end int
}

func findStringSubmatch(re *regexp.Regexp, s string) []*match {
	indexes := re.FindStringSubmatchIndex(s)
	if len(indexes) <= 2 {
		return nil
	}

	n := len(indexes)
	matches := make([]*match, 0, n/2)
	for i := 0; i < n; i = i + 2 {
		start, end := indexes[i], indexes[i+1]
		val := s[start:end]
		matches = append(matches, &match{val: val, start: start, end: end})
	}
	return matches
}

// TagProducer is used to produce tags for metric name.
type TagProducer struct {
	defaultTags []*Tag

	normalExtractors []*TagExtractor
	prefixExtractors map[string][]*TagExtractor
}

// NewTagProducer creates a tag producer with default tags.
func NewTagProducer(defaultTags ...*Tag) *TagProducer {
	p := &TagProducer{
		defaultTags:      defaultTags,
		prefixExtractors: make(map[string][]*TagExtractor),
	}
	return p
}

// AddExtractor adds a tag extractor.
func (p *TagProducer) AddExtractor(e *TagExtractor) {
	prefixToken := e.prefixToken()
	if prefixToken == "" {
		p.normalExtractors = append(p.normalExtractors, e)
		return
	}

	extractors := p.prefixExtractors[prefixToken]
	extractors = append(extractors, e)
	p.prefixExtractors[prefixToken] = extractors
}

// Produce produces the tags for the given metric name.
func (p *TagProducer) Produce(metricName string) (string, []*Tag) {
	tags := make([]*Tag, len(p.defaultTags))
	copy(tags, p.defaultTags)

	removeChars := newIntervalSet()
	for _, extractor := range p.normalExtractors {
		extractor.extract(metricName, &tags, removeChars)
	}

	index := strings.Index(metricName, ".")
	if index != -1 {
		token := metricName[:index]
		extractors := p.prefixExtractors[token]
		for _, extractor := range extractors {
			extractor.extract(metricName, &tags, removeChars)
		}
	}

	return removeCharacters(metricName, removeChars), tags
}

func removeCharacters(s string, iset *IntervalSet) string {
	var (
		pos int
		buf bytes.Buffer
	)
	for _, interval := range iset.AllIntervals() {
		// TODO(kirk91): check index validatily
		if interval.Left != pos {
			buf.WriteString(s[pos:interval.Left]) // [left, right)
		}
		pos = interval.Right
	}
	if pos != len(s) {
		buf.WriteString(s[pos:])
	}
	return buf.String()
}
