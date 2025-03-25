package model

import (
	"time"
)

// NoteMetadata represents the YAML frontmatter in a note file
type NoteMetadata struct {
	ID         string    `yaml:"id"`
	Aliases    []string  `yaml:"aliases,omitempty"`
	Tags       []string  `yaml:"tags,omitempty"`
	Created    time.Time `yaml:"created"`
	ReadDeckID string    `yaml:"readdeck-id,omitempty"`
	Processed  bool      `yaml:"processed,omitempty"`
	Publish    bool      `yaml:"publish,omitempty"`
	Reference  bool      `yaml:"reference,omitempty"`
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
