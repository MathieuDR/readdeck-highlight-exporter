package model

import (
	"time"
)

// NoteMetadata represents the YAML frontmatter in a note file
type NoteMetadata struct {
	ID           string    `validate:"required" yaml:"id"`
	Aliases      []string  `yaml:"aliases,omitempty"`
	Tags         []string  `yaml:"tags,omitempty"`
	Created      time.Time `yaml:"created"`
	ReaddeckID   string    `yaml:"readdeck-id"`
	ReaddeckHash string    `yaml:"readdeck-hash"`
	Media        string    `yaml:"media"`
	Type         string    `yaml:"media-type"`
	Published    time.Time `yaml:"media-published"`
	ArchiveUrl   string    `yaml:"readdeck-url"`
	Site         string    `yaml:"media-url"`
	Authors      []string  `yaml:"authors"`
}

type ParsedNote struct {
	Path         string
	Metadata     NoteMetadata
	Content      string
	HighlightIDs []string
}
