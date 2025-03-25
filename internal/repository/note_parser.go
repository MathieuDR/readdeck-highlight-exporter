// internal/repository/note_parser.go
package repository

import (
	"github.com/mathieudr/readdeck-highlight-exporter/internal/model"
)

// NoteParser defines the interface for parsing note files
type NoteParser interface {
	ParseNote(content []byte, path string) (model.ParsedNote, error)
	GenerateNoteContent(note model.Note) ([]byte, error)
}

// YAMLFrontmatterParser implements NoteParser for YAML frontmatter notes
type YAMLFrontmatterParser struct{}

// ParseNote parses a note file content into a ParsedNote
func (p *YAMLFrontmatterParser) ParseNote(content []byte, path string) (model.ParsedNote, error) {
	// Will be implemented later
	return model.ParsedNote{}, nil
}

// GenerateNoteContent generates note content from a model.Note
func (p *YAMLFrontmatterParser) GenerateNoteContent(note model.Note) ([]byte, error) {
	// Will be implemented later
	return nil, nil
}

// parseHighlightsFromContent extracts highlights from note content
func (p *YAMLFrontmatterParser) parseHighlightsFromContent(content string) []model.ParsedHighlight {
	// Will be implemented later
	return nil
}

