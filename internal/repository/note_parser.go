package repository

import (
	"bytes"
	"fmt"

	"github.com/adrg/frontmatter"
	"github.com/go-playground/validator/v10"
	"github.com/mathieudr/readdeck-highlight-exporter/internal/model"
	"github.com/mathieudr/readdeck-highlight-exporter/internal/util"
)

type NoteParser interface {
	ParseNote(content []byte, path string) (model.ParsedNote, error)
	GenerateNoteContent(note model.Note) ([]byte, string, error)
}

type YAMLFrontmatterParser struct {
	Validator *validator.Validate
	Hasher    *util.GobHasher
}

func NewYAMLFrontmatterParser() *YAMLFrontmatterParser {
	return &YAMLFrontmatterParser{
		Validator: validator.New(),
		Hasher:    util.NewGobHasher(),
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

	highlightIDs, err := p.decodeHighlightIDsHash(matter.ReaddeckHash)
	if err != nil {
		return model.ParsedNote{}, err
	}

	return model.ParsedNote{
		Path:         path,
		Metadata:     matter,
		Content:      string(textContent),
		HighlightIDs: highlightIDs,
	}, nil
}

func (p *YAMLFrontmatterParser) GenerateNoteContent(note model.Note) ([]byte, string, error) {
	return nil, nil
}

func (p *YAMLFrontmatterParser) decodeHighlightIDsHash(hash string) ([]string, error) {
	if hash == "" {
		return []string{}, nil
	}

	ids, err := p.Hasher.Decode(hash)

	if err != nil {
		return nil, fmt.Errorf("Could not decode id's: %w", err)
	}

	return ids, err
}
