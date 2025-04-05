package repository

import (
	"bytes"
	"fmt"

	"github.com/adrg/frontmatter"
	"github.com/go-playground/validator/v10"
	"github.com/mathieudr/readdeck-highlight-exporter/internal/model"
	"github.com/mathieudr/readdeck-highlight-exporter/internal/util"
	"gopkg.in/yaml.v2"
)

// NoteParser is responsible for parsing notes from raw content
type NoteParser interface {
	ParseNote(content []byte, path string) (model.ParsedNote, error)
}

type YAMLNoteParser struct {
	Validator *validator.Validate
	Hasher    *util.GobHasher
}

func NewYAMLNoteParser() *YAMLNoteParser {
	return &YAMLNoteParser{
		Validator: validator.New(),
		Hasher:    util.NewGobHasher(),
	}
}

func (p *YAMLNoteParser) ParseNote(content []byte, path string) (model.ParsedNote, error) {
	// Parse into a map once
	var rawMap map[string]interface{}
	textContent, err := frontmatter.Parse(bytes.NewReader(content), &rawMap)
	if err != nil {
		return model.ParsedNote{}, fmt.Errorf("could not parse frontmatter: %w", err)
	}

	// Convert to struct using existing yaml package
	yamlBytes, err := yaml.Marshal(rawMap)
	if err != nil {
		return model.ParsedNote{}, fmt.Errorf("could not remarshal frontmatter: %w", err)
	}

	var metadata model.NoteMetadata
	if err := yaml.Unmarshal(yamlBytes, &metadata); err != nil {
		return model.ParsedNote{}, fmt.Errorf("could not unmarshal to struct: %w", err)
	}

	// Rest of your validation logic...
	if err := p.Validator.Struct(&metadata); err != nil {
		return model.ParsedNote{}, fmt.Errorf("frontmatter is invalid: %w", err)
	}

	highlightIDs, err := p.decodeHighlightIDsHash(metadata.ReaddeckHash)
	if err != nil {
		return model.ParsedNote{}, err
	}

	return model.ParsedNote{
		Path:           path,
		Metadata:       metadata,
		Content:        string(textContent),
		HighlightIDs:   highlightIDs,
		RawFrontmatter: rawMap,
	}, nil
}

func (p *YAMLNoteParser) decodeHighlightIDsHash(hash string) ([]string, error) {
	if hash == "" {
		return []string{}, nil
	}

	ids, err := p.Hasher.Decode(hash)

	if err != nil {
		return nil, fmt.Errorf("could not decode IDs: %w", err)
	}

	return ids, nil
}
