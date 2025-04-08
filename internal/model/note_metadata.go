package model

type NoteMetadata struct {
	ID           string     `validate:"required" yaml:"id"`
	Aliases      []string   `yaml:"aliases,omitempty"`
	Tags         []string   `yaml:"tags,omitempty"`
	Created      SimpleTime `yaml:"created"`
	ReaddeckID   string     `yaml:"readdeck-id"`
	ReaddeckHash string     `yaml:"readdeck-hash"`
	Media        string     `yaml:"media"`
	Type         string     `yaml:"media-type"`
	Published    SimpleTime `yaml:"media-published"`
	ArchiveUrl   string     `yaml:"readdeck-url"`
	Site         string     `yaml:"media-url"`
	Authors      []string   `yaml:"authors"`
}

type ParsedNote struct {
	Path           string
	Metadata       NoteMetadata
	Content        []Section
	HighlightIDs   []string
	RawFrontmatter map[string]interface{}
}

type SectionType string

const (
	H1   SectionType = "h1"
	H2   SectionType = "h2"
	H3   SectionType = "h3"
	H4   SectionType = "h4"
	H5   SectionType = "h5"
	H6   SectionType = "h6"
	None SectionType = "none"
)

type Section struct {
	Type    SectionType
	Title   string
	Content string
}
