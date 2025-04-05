package repository

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

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
	Validator    *validator.Validate
	Hasher       *util.GobHasher
	headingRegex *regexp.Regexp
}

func NewYAMLNoteParser() *YAMLNoteParser {
	return &YAMLNoteParser{
		Validator:    validator.New(),
		Hasher:       util.NewGobHasher(),
		headingRegex: regexp.MustCompile(`^(#{1,6})\s+(.*)$`),
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

	// Validate metadata
	if err := p.Validator.Struct(&metadata); err != nil {
		return model.ParsedNote{}, fmt.Errorf("frontmatter is invalid: %w", err)
	}

	highlightIDs, err := p.decodeHighlightIDsHash(metadata.ReaddeckHash)
	if err != nil {
		return model.ParsedNote{}, err
	}

	// Parse the content into sections
	sections, err := p.ParseContent(string(textContent))
	if err != nil {
		return model.ParsedNote{}, fmt.Errorf("could not parse content sections: %w", err)
	}

	return model.ParsedNote{
		Path:           path,
		Metadata:       metadata,
		Content:        sections,
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

func (p *YAMLNoteParser) headingTypeFromLevel(level int) model.SectionType {
	switch level {
	case 1:
		return model.H1
	case 2:
		return model.H2
	case 3:
		return model.H3
	case 4:
		return model.H4
	case 5:
		return model.H5
	case 6:
		return model.H6
	default:
		return model.None
	}
}

func (p *YAMLNoteParser) ParseContent(input string) ([]model.Section, error) {
	var sections []model.Section

	lines := strings.Split(input, "\n")

	var contentBuffer bytes.Buffer
	currentSection := model.Section{
		Type:  model.None,
		Title: "",
	}

	for _, line := range lines {
		matches := p.headingRegex.FindStringSubmatch(line)

		if len(matches) > 0 {
			// Found a heading - store any accumulated content in the current section
			// Only store content if there's actual content or if it's a heading section
			content := strings.TrimSpace(contentBuffer.String())
			if content != "" || currentSection.Type != model.None {
				currentSection.Content = content
				sections = append(sections, currentSection)
				contentBuffer.Reset()
			}

			// New heading
			headingLevel := len(matches[1])
			title := strings.TrimSpace(matches[2])

			currentSection = model.Section{
				Type:  p.headingTypeFromLevel(headingLevel),
				Title: title,
			}
		} else {
			contentBuffer.WriteString(line)
			contentBuffer.WriteString("\n")
		}
	}

	// Last section
	content := strings.TrimSpace(contentBuffer.String())
	if content != "" || currentSection.Type != model.None {
		currentSection.Content = content
		sections = append(sections, currentSection)
	}

	return sections, nil
}

