package repository

import (
	"bytes"
	"fmt"
	"strings"

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

	highlights := u.getHighlights(existing.HighlightIDs, note.Highlights)
	bodyBytes := u.appendHighlightsToSections(existing.Content, highlights)
	content = append(content, bodyBytes...)

	return NoteOperation{
		Metadata: metadata,
		Content:  content,
	}, nil
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

func (u *YAMLNoteUpdater) findReferenceSection(sections []model.Section) *model.Section {
	for i, section := range sections {
		if section.Type == model.H2 && section.Title == "References" {
			return &sections[i]
		}
	}

	return nil
}

func (u *YAMLNoteUpdater) appendHighlightsToSections(sections []model.Section, highlights []readdeck.Highlight) []byte {
	var buffer bytes.Buffer
	referenceSection := u.findReferenceSection(sections)

	if len(highlights) == 0 {
		for _, section := range sections {
			writeSection(&buffer, section)
		}
		return buffer.Bytes()
	}

	// Reuse the formatter's grouping and ordering logic to ensure consistent
	// presentation between new notes and updated notes
	highlightGroups := u.Generator.HighlightFormatter.groupHighlightsByColor(highlights)
	colorOrder := u.Generator.HighlightFormatter.GetSortedColorOrder(highlightGroups)
	highlightBodies := make(map[string][]byte)
	for color, hs := range highlightGroups {
		highlightBodies[color] = u.Generator.HighlightFormatter.highlightBodyBytes(hs)
	}

	// Track which highlight groups have been handled so we know which ones
	// need new sections at the end of the document
	processedColors := make(map[string]bool)

	for _, section := range sections {
		if &section != referenceSection {
			writeSection(&buffer, section)
		}

		if section.Type == model.H2 {
			for color := range highlightGroups {
				friendlyName := u.Generator.HighlightFormatter.colorToFriendlyName(color)

				if friendlyName == section.Title && len(highlightGroups[color]) > 0 {
					// Use the highlight bodies directly instead of trying to extract them
					buffer.Write(highlightBodies[color])
					processedColors[color] = true
					break
				}
			}

		}
	}

	// For highlight colors without matching sections, add them as new sections
	// at the end of the document in the proper order
	for _, color := range colorOrder {
		if !processedColors[color] && len(highlightGroups[color]) > 0 {
			if buffer.Len() > 0 {
				buffer.WriteString("\n\n")
			}
			title := u.Generator.HighlightFormatter.highlightTitleBytes(color)
			buffer.Write(title)
			buffer.Write(highlightBodies[color])
		}
	}

	if referenceSection != nil {
		if buffer.Len() > 0 {
			buffer.WriteString("\n\n")
		}
		writeSection(&buffer, *referenceSection)
	}

	return buffer.Bytes()
}

func writeSection(buffer *bytes.Buffer, section model.Section) {
	if section.Type != model.None {
		level := 0
		switch section.Type {
		case model.H1:
			level = 1
		case model.H2:
			level = 2
		case model.H3:
			level = 3
		case model.H4:
			level = 4
		case model.H5:
			level = 5
		case model.H6:
			level = 6
		}

		buffer.WriteString(strings.Repeat("#", level))
		buffer.WriteString(" ")
		buffer.WriteString(section.Title)
		buffer.WriteString("\n")
	}

	if section.Content != "" {
		buffer.WriteString(section.Content)
	}
}

func (u *YAMLNoteUpdater) getHighlights(existingIds []string, highlights []readdeck.Highlight) []readdeck.Highlight {
	ids := make([]string, len(highlights))
	for i, h := range highlights {
		ids[i] = h.ID
	}

	newIds := u.diffHighlights(existingIds, ids)
	lookup := make(map[string]readdeck.Highlight, len(highlights))
	for _, h := range highlights {
		lookup[h.ID] = h
	}

	result := make([]readdeck.Highlight, len(newIds))
	for i, id := range newIds {
		result[i] = lookup[id]
	}

	return result
}

func (u *YAMLNoteUpdater) diffHighlights(existingIds []string, newIds []string) []string {
	// Returns the IDs from newIds that aren't in existingIds
	result := make([]string, 0, len(newIds))
	existing := make(map[string]struct{}, len(existingIds))

	for _, id := range existingIds {
		existing[id] = struct{}{}
	}

	for _, id := range newIds {

		if _, exists := existing[id]; !exists {
			result = append(result, id)
		}
	}

	return result
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
