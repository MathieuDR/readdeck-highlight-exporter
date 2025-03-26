package repository

import (
	"bytes"
	"fmt"

	"github.com/adrg/frontmatter"
	"github.com/go-playground/validator/v10"
	"github.com/mathieudr/readdeck-highlight-exporter/internal/model"
)

// NoteParser defines the interface for parsing note files
type NoteParser interface {
	ParseNote(content []byte, path string) (model.ParsedNote, error)
	GenerateNoteContent(note model.Note) ([]byte, error)
}

// YAMLFrontmatterParser implements NoteParser for YAML frontmatter notes
type YAMLFrontmatterParser struct{}

func NewYAMLFrontmatterParser() *YAMLFrontmatterParser {
	return &YAMLFrontmatterParser{}
}

// ParseNote parses a note file content into a ParsedNote
func (p *YAMLFrontmatterParser) ParseNote(content []byte, path string) (model.ParsedNote, error) {
	var matter model.NoteMetadata
	text_content, err := frontmatter.MustParse(bytes.NewReader(content), &matter)

	if err != nil {
		return model.ParsedNote{}, fmt.Errorf("Could not parse frontmatter: %w", err)
	}

	validator := validator.New()
	err = validator.Struct(&matter)

	if err != nil {
		return model.ParsedNote{}, fmt.Errorf("Frontmatter is invalid: %w", err)
	}

	return model.ParsedNote{
		Path:       path,
		Metadata:   matter,
		Content:    string(text_content),
		Highlights: nil,
	}, nil
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
