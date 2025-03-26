package repository

import (
	"bytes"
	"fmt"

	"github.com/adrg/frontmatter"
	"github.com/go-playground/validator/v10"
	"github.com/mathieudr/readdeck-highlight-exporter/internal/model"
)

type NoteParser interface {
	ParseNote(content []byte, path string) (model.ParsedNote, error)
	GenerateNoteContent(note model.Note) ([]byte, error)
}

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
		Path:         path,
		Metadata:     matter,
		Content:      string(textContent),
		HighlightIDs: p.decodeHighlightIDsHash(matter.ReaddeckHash),
	}, nil
}

func (p *YAMLFrontmatterParser) GenerateNoteContent(note model.Note) ([]byte, error) {
	return nil, nil
}

func (p *YAMLFrontmatterParser) decodeHighlightIDsHash(hash string) []string {
	return nil
}
