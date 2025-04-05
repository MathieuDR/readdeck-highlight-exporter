package repository

import (
	"bytes"
	"fmt"
	"github.com/adrg/frontmatter"
	"github.com/go-playground/validator/v10"
	"github.com/mathieudr/readdeck-highlight-exporter/internal/model"
	"github.com/mathieudr/readdeck-highlight-exporter/internal/util"
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
	var rawMap map[string]interface{}
	textContent, err := frontmatter.Parse(bytes.NewReader(content), &rawMap)
	if err != nil {
		return model.ParsedNote{}, fmt.Errorf("could not parse frontmatter: %w", err)
	}

	var matter model.NoteMetadata
	contentCopy := bytes.NewReader(content)
	_, err = frontmatter.Parse(contentCopy, &matter)
	if err != nil {
		return model.ParsedNote{}, fmt.Errorf("could not parse frontmatter into struct: %w", err)
	}

	if err := p.Validator.Struct(&matter); err != nil {
		return model.ParsedNote{}, fmt.Errorf("frontmatter is invalid: %w", err)
	}

	highlightIDs, err := p.decodeHighlightIDsHash(matter.ReaddeckHash)
	if err != nil {
		return model.ParsedNote{}, err
	}

	return model.ParsedNote{
		Path:           path,
		Metadata:       matter,
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
