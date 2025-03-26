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
type YAMLFrontmatterParser struct {
	Validator *validator.Validate
}

func NewYAMLFrontmatterParser() *YAMLFrontmatterParser {
	return &YAMLFrontmatterParser{
		Validator: validator.New(),
	}
}

func (p *YAMLFrontmatterParser) ParseNote(content []byte, path string) (model.ParsedNote, error) {
	var matter model.NoteMetadata
	textContent, err := frontmatter.MustParse(bytes.NewReader(content), &matter)

	if err != nil {
		return model.ParsedNote{}, fmt.Errorf("Could not parse frontmatter: %w", err)
	}

	err = p.Validator.Struct(&matter)

	if err != nil {
		return model.ParsedNote{}, fmt.Errorf("Frontmatter is invalid: %w", err)
	}

	return model.ParsedNote{
		Path:       path,
		Metadata:   matter,
		Content:    string(textContent),
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
