package model

import (
	"time"
)

// NoteMetadata represents the YAML frontmatter in a note file
type NoteMetadata struct {
	ID         string    `validate:"required" yaml:"id"`
	Aliases    []string  `yaml:"aliases,omitempty"`
	Tags       []string  `yaml:"tags,omitempty"`
	Created    time.Time `yaml:"created"`
	ReaddeckID string    `yaml:"readdeck-id"`
	Media      string    `yaml:"media"`
	Type       string    `yaml:"media-type"`
	Published  time.Time `yaml:"media-published"`
	ArchiveUrl string    `yaml:"readdeck-url"`
	Site       string    `yaml:"media-url"`
	Authors    []string  `yaml:"authors"`
}

// ParsedNote contains both metadata and content for a parsed note file
type ParsedNote struct {
	Path       string
	Metadata   NoteMetadata
	Content    string
	Highlights []ParsedHighlight
}

// ParsedHighlight represents a highlight extracted from note content
type ParsedHighlight struct {
	ID   string
	Text string
}

// NoteUpdate contains the changes needed for a note
type NoteUpdate struct {
	Path           string
	UpdateMetadata bool
	NewMetadata    NoteMetadata
	UpdateContent  bool
	NewContent     string
}
