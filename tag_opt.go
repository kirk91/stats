package stats

// TagExtractStrategy represents the extract strategy from metric for dynamic tags.
type TagExtractStrategy struct {
	Name   string
	Regex  string
	SubStr string
}

// TagOption controls how tags are set.
type TagOption struct {
	DefaultTags          map[string]string
	TagExtractStrategies []TagExtractStrategy
}

// NewTagOption creates an empty TagOption.
func NewTagOption() *TagOption {
	return &TagOption{}
}

func (t TagOption) defaultTags() []*Tag {
	var tags = make([]*Tag, 0, len(t.DefaultTags))
	for k, v := range t.DefaultTags {
		tags = append(tags, &Tag{k, v})
	}
	return tags
}

// WithDefaultTags sets default tags.
func (t *TagOption) WithDefaultTags(tags map[string]string) *TagOption {
	t.DefaultTags = tags
	return t
}

// WithTagExtractStrategies sets tag extract strategies.
func (t *TagOption) WithTagExtractStrategies(strategies ...TagExtractStrategy) *TagOption {
	t.TagExtractStrategies = strategies
	return t
}
