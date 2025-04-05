package repository

import (
	"fmt"

	"github.com/mathieudr/readdeck-highlight-exporter/internal/model"
	"github.com/mathieudr/readdeck-highlight-exporter/internal/readdeck"
	"gopkg.in/yaml.v2"
)

type YAMLNoteUpdater struct {
	Generator *YAMLNoteGenerator
	Parser    *YAMLNoteParser
}

func NewYAMLNoteUpdater(generator *YAMLNoteGenerator, parser *YAMLNoteParser) *YAMLNoteUpdater {
	return &YAMLNoteUpdater{
		Generator: generator,
		Parser:    parser,
	}
}

func (u *YAMLNoteUpdater) UpdateNoteContent(existing model.ParsedNote, note model.Note) (NoteOperation, error) {
	metadata, err := u.updateMetadata(existing.Metadata, note.Bookmark, note.Highlights)
	if err != nil {
		return NoteOperation{}, err
	}

	var content []byte

	frontmatter, err := u.updateFrontmatter(existing.RawFrontmatter, metadata)
	if err != nil {
		return NoteOperation{}, err
	}
	content = append(content, frontmatter...)

	return NoteOperation{}, nil
}

func (u *YAMLNoteUpdater) updateMetadata(existing model.NoteMetadata, bookmark readdeck.Bookmark, highlights []readdeck.Highlight) (model.NoteMetadata, error) {
	metadata, err := u.Generator.generateMetadata(bookmark, highlights)
	if err != nil {
		return model.NoteMetadata{}, fmt.Errorf("Could not generate new metadata: %w", err)
	}

	return model.NoteMetadata{
		ID:           existing.ID,
		Aliases:      u.merge(existing.Aliases, metadata.Aliases),
		Tags:         u.merge(existing.Tags, metadata.Tags),
		Created:      existing.Created,
		ReaddeckID:   existing.ReaddeckID,
		Media:        metadata.Media,
		Type:         metadata.Type,
		Published:    metadata.Published,
		ArchiveUrl:   metadata.ArchiveUrl,
		Site:         metadata.Site,
		Authors:      u.merge(existing.Authors, metadata.Authors),
		ReaddeckHash: metadata.ReaddeckHash,
	}, nil
}

func (u *YAMLNoteUpdater) updateFrontmatter(existing map[string]interface{}, metadata model.NoteMetadata) ([]byte, error) {
	metadataBytes, err := yaml.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("could not marshal updated metadata: %w", err)
	}

	var updatedMap map[string]interface{}
	if err := yaml.Unmarshal(metadataBytes, &updatedMap); err != nil {
		return nil, fmt.Errorf("could not unmarshal to map: %w", err)
	}

	// Merge updated fields into original raw map
	for k, v := range updatedMap {
		existing[k] = v
	}

	// Create frontmatter with ALL fields preserved
	frontmatterBytes, err := yaml.Marshal(existing)
	if err != nil {
		return nil, fmt.Errorf("could not marshal frontmatter: %w", err)
	}

	return []byte(fmt.Sprintf("---\n%s---\n", frontmatterBytes)), nil
}

func (u *YAMLNoteUpdater) merge(list1 []string, list2 []string) []string {
	seen := make(map[string]struct{}, len(list1)+len(list2))
	unique := make([]string, 0, len(list1)+len(list2))

	for _, list := range [][]string{list1, list2} {
		for _, item := range list {
			if _, exists := seen[item]; !exists {
				seen[item] = struct{}{}
				unique = append(unique, item)
			}
		}
	}

	return unique
}
