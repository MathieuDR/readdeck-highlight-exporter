// internal/repository/note_generator.go
package repository

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/mathieudr/readdeck-highlight-exporter/internal/model"
	"github.com/mathieudr/readdeck-highlight-exporter/internal/readdeck"
	"github.com/mathieudr/readdeck-highlight-exporter/internal/util"
)

type NoteOperation struct {
	Metadata model.NoteMetadata
	Content  []byte
}

type NoteGenerator interface {
	GenerateNoteContent(note model.Note) (NoteOperation, error)
}

type NoteUpdater interface {
	UpdateNoteContent(existing model.ParsedNote, note model.Note) (NoteOperation, error)
}

type YAMLNoteGenerator struct {
	Hasher             *util.GobHasher
	HighlightFormatter *HighlightFormatter
}

func NewYAMLNoteGenerator(formatter *HighlightFormatter) *YAMLNoteGenerator {
	return &YAMLNoteGenerator{
		Hasher:             util.NewGobHasher(),
		HighlightFormatter: formatter,
	}
}

func (g *YAMLNoteGenerator) GenerateNoteContent(note model.Note) (NoteOperation, error) {
	metadata, err := g.generateMetadata(note.Bookmark, note.Highlights)
	if err != nil {
		return NoteOperation{}, err
	}

	var content []byte

	// Add frontmatter
	frontmatter, err := g.generateFrontmatter(metadata)
	if err != nil {
		return NoteOperation{}, err
	}
	content = append(content, frontmatter...)

	// Add title
	content = append(content, []byte(fmt.Sprintf("# %s\n\n", metadata.Aliases[0]))...)

	// Add highlights by color
	highlightSections := g.HighlightFormatter.FormatHighlightsByColor(note.Highlights)
	colorOrder := g.HighlightFormatter.GetSortedColorOrder(
		g.HighlightFormatter.groupHighlightsByColor(note.Highlights))

	for _, color := range colorOrder {
		if section, ok := highlightSections[color]; ok {
			content = append(content, section...)
		}
	}

	return NoteOperation{
		Metadata: metadata,
		Content:  content,
	}, nil
}

func (g *YAMLNoteGenerator) generateMetadata(bookmark readdeck.Bookmark, highlights []readdeck.Highlight) (model.NoteMetadata, error) {
	ids := make([]string, len(highlights))
	for i, h := range highlights {
		ids[i] = h.ID
	}

	hash, err := g.Hasher.Encode(ids)
	if err != nil {
		return model.NoteMetadata{}, fmt.Errorf("could not hash highlights: %w", err)
	}

	created := model.SimpleTime{Time: bookmark.Created}
	published := model.SimpleTime{Time: bookmark.Published}

	tags := make([]string, 0, 3+len(bookmark.Labels))
	tags = append(tags, []string{"highlights", "zettelkasten", "fleeting-note"}...)

	return model.NoteMetadata{
		ID:           util.GenerateId(bookmark.Title, time.Now()),
		Aliases:      []string{fmt.Sprintf("%s highlights", util.Capitalize(bookmark.Title))},
		Tags:         tags,
		Created:      created,
		ReaddeckID:   bookmark.ID,
		ReaddeckHash: hash,
		Media:        bookmark.Title,
		Type:         bookmark.Type,
		Published:    published,
		ArchiveUrl:   bookmark.Href,
		Site:         bookmark.SiteUrl,
		Authors:      bookmark.Authors,
	}, nil
}

func (g *YAMLNoteGenerator) generateFrontmatter(metadata model.NoteMetadata) ([]byte, error) {
	yamlData, err := yaml.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("could not format metadata: %w", err)
	}

	return []byte(fmt.Sprintf("---\n%s---\n", yamlData)), nil
}
