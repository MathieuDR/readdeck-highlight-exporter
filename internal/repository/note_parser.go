package repository

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/adrg/frontmatter"
	"github.com/go-playground/validator/v10"
	"github.com/mathieudr/readdeck-highlight-exporter/internal/model"
	"github.com/mathieudr/readdeck-highlight-exporter/internal/readdeck"
	"github.com/mathieudr/readdeck-highlight-exporter/internal/util"
	"gopkg.in/yaml.v2"
)

type NoteParser interface {
	ParseNote(content []byte, path string) (model.ParsedNote, error)
	GenerateNoteContent(note model.Note) (NoteOperation, error)
}

type ColorConfig struct {
	ColorNames map[string]string
	ColorOrder []string
}

type NoteOperation struct {
	Metadata model.NoteMetadata
	Content  []byte
}

type YAMLFrontmatterParser struct {
	Validator   *validator.Validate
	Hasher      *util.GobHasher
	ColorConfig ColorConfig
}

func NewYAMLFrontmatterParser() *YAMLFrontmatterParser {
	return &YAMLFrontmatterParser{
		Validator: validator.New(),
		Hasher:    util.NewGobHasher(),
		ColorConfig: ColorConfig{
			ColorNames: map[string]string{
				"yellow": "General highlights",
				"red":    "Thought-provoking insights",
				"blue":   "Important references",
				"green":  "Key takeaways",
			},
			ColorOrder: []string{"green", "red", "yellow", "blue"},
		},
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

func (p *YAMLFrontmatterParser) GenerateNoteContent(note model.Note) (NoteOperation, error) {
	metadata := p.generateMetadata(note.Bookmark)
	bytes := make([]byte, 0)
	frontmatter, err := p.generateFrontmatter(metadata)

	if err != nil {
		return NoteOperation{}, err
	}

	bytes = append(bytes, frontmatter...)
	bytes = append(bytes, []byte(fmt.Sprintf("\n%s\n\n", metadata.Aliases[0]))...)

	groups := p.groupHighlightsByColor(note.Highlights)
	for color, highlights := range groups {
		bytes = append(bytes, p.highlightTitleBytes(color)...)
		bytes = append(bytes, p.highlightBodyBytes(highlights)...)
	}

	return NoteOperation{
		Metadata: metadata,
		Content:  bytes,
	}, nil
}

func (p *YAMLFrontmatterParser) highlightTitleBytes(color string) []byte {
	colour := p.colourToFriendlyName(color)
	return []byte(fmt.Sprintf("## %s\n", colour))
}

func (p *YAMLFrontmatterParser) highlightBodyBytes(highlights []readdeck.Highlight) []byte {
	result := make([]byte, 0)
	for _, h := range highlights {
		highlightBytes := []byte(fmt.Sprintf("%s\n\n", h.Text))
		result = append(result, highlightBytes...)
	}

	return result
}

func (p *YAMLFrontmatterParser) groupHighlightsByColor(highlights []readdeck.Highlight) map[string][]readdeck.Highlight {
	res := make(map[string][]readdeck.Highlight)

	for _, h := range highlights {
		res[h.Color] = append(res[h.Color], h)
	}

	return p.sortHighlightGroups(res)
}

func (p *YAMLFrontmatterParser) generateMetadata(bookmark readdeck.Bookmark) model.NoteMetadata {
	return model.NoteMetadata{
		ID:         util.GenerateId(bookmark.Title, time.Now()),
		Aliases:    []string{fmt.Sprintf("%s highlights", util.Capitalize(bookmark.Title))},
		Tags:       bookmark.Labels,
		Created:    bookmark.Created,
		ReaddeckID: bookmark.ID,
		Media:      bookmark.Title,
		Type:       bookmark.Type,
		Published:  bookmark.Published,
		ArchiveUrl: bookmark.Href,
		Site:       bookmark.SiteUrl,
		Authors:    bookmark.Authors,
	}
}

func (p *YAMLFrontmatterParser) generateFrontmatter(metadata model.NoteMetadata) ([]byte, error) {
	yamlData, err := yaml.Marshal(metadata)

	if err != nil {
		return nil, fmt.Errorf("Could not format metada: %w", err)
	}

	return yamlData, nil
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

func (p *YAMLFrontmatterParser) colourToFriendlyName(color string) string {
	if name, ok := p.ColorConfig.ColorNames[color]; ok {
		return name
	}

	capitalizedColor := util.Capitalize(strings.ToLower(color))
	return fmt.Sprintf("%s highlights", capitalizedColor)
}

func (p *YAMLFrontmatterParser) sortHighlightGroups(highlights map[string][]readdeck.Highlight) map[string][]readdeck.Highlight {
	sortedHighlights := make(map[string][]readdeck.Highlight)

	for _, color := range p.ColorConfig.ColorOrder {
		if highlightList, ok := highlights[color]; ok {
			sortedHighlights[color] = highlightList
		}
	}

	var remainingColors []string
	for color := range highlights {
		if _, exists := sortedHighlights[color]; !exists {
			remainingColors = append(remainingColors, color)
		}
	}

	sort.Strings(remainingColors)

	for _, color := range remainingColors {
		sortedHighlights[color] = highlights[color]
	}

	return sortedHighlights
}
