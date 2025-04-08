package repository_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/mathieudr/readdeck-highlight-exporter/internal/model"
	"github.com/mathieudr/readdeck-highlight-exporter/internal/repository"
	"github.com/mathieudr/readdeck-highlight-exporter/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestYAMLFrontmatterParser_ParseNote(t *testing.T) {
	testTime, _ := time.Parse(time.RFC3339, "2025-03-26T14:00:00Z")
	publishTime, _ := time.Parse(time.RFC3339, "2020-03-26T14:00:00Z")

	simpleTestTime := model.SimpleTime{Time: testTime}
	simplePublishTime := model.SimpleTime{Time: publishTime}
	hash, err := util.NewGobHasher().Encode([]string{"h1", "h2"})

	require.NoError(t, err, "Could not hash for tests")

	tests := []struct {
		name    string
		content []byte
		path    string
		want    model.ParsedNote
		wantErr bool
	}{
		{
			name: "basic valid note",
			content: []byte(`---
id: note123
created: 2025-03-26 14:00
readdeck-id: rd456
tags:
  - go
  - programming
---
This is the content of the note.

Multiple paragraphs are supported.`),
			path: "/path/to/note.md",
			want: model.ParsedNote{
				Metadata: model.NoteMetadata{
					ID:         "note123",
					Created:    simpleTestTime,
					ReaddeckID: "rd456",
					Tags:       []string{"go", "programming"},
				},
				HighlightIDs: []string{},
				Content: []model.Section{
					{
						Type:    model.None,
						Title:   "",
						Content: "This is the content of the note.\n\nMultiple paragraphs are supported.",
					},
				},
				Path: "/path/to/note.md",
				RawFrontmatter: map[string]interface{}{
					"id":          "note123",
					"created":     "2025-03-26 14:00",
					"readdeck-id": "rd456",
					"tags":        []interface{}{"go", "programming"},
				},
			},
			wantErr: false,
		},
		{
			name: "note with all metadata fields",
			content: []byte(fmt.Sprintf(`---
id: full-note
aliases:
  - "alias1"
  - "alias2"
tags:
  - research
  - papers
created: 2025-03-26 14:00
readdeck-id: rd789
media: "Rework"
media-type: article
readdeck-hash: %s
media-published: 2020-03-26 14:00
readdeck-url: https://read.deck.com
authors:
  - Jason
  - Bourne
media-url: https://bourne.identity
---
Full content here.`, hash)),
			path: "/path/to/full.md",
			want: model.ParsedNote{
				Metadata: model.NoteMetadata{
					ID:           "full-note",
					Aliases:      []string{"alias1", "alias2"},
					Tags:         []string{"research", "papers"},
					Created:      simpleTestTime,
					Published:    simplePublishTime,
					Media:        "Rework",
					Type:         "article",
					ArchiveUrl:   "https://read.deck.com",
					Authors:      []string{"Jason", "Bourne"},
					Site:         "https://bourne.identity",
					ReaddeckID:   "rd789",
					ReaddeckHash: hash,
				},
				HighlightIDs: []string{"h1", "h2"},
				Content: []model.Section{
					{
						Type:    model.None,
						Title:   "",
						Content: "Full content here.",
					},
				},
				Path: "/path/to/full.md",
				RawFrontmatter: map[string]interface{}{
					"id":              "full-note",
					"aliases":         []interface{}{"alias1", "alias2"},
					"tags":            []interface{}{"research", "papers"},
					"created":         "2025-03-26 14:00",
					"readdeck-id":     "rd789",
					"media":           "Rework",
					"media-type":      "article",
					"readdeck-hash":   hash,
					"media-published": "2020-03-26 14:00",
					"readdeck-url":    "https://read.deck.com",
					"authors":         []interface{}{"Jason", "Bourne"},
					"media-url":       "https://bourne.identity",
				},
			},
			wantErr: false,
		},
		{
			name: "note with invalid types",
			content: []byte(`---
id: note123
created: BadDate
readdeck-id: 
  - rd456
  - rd457
publish: 19
tags:
  - go
  - programming
---
This is the content of the note.

Multiple paragraphs are supported.`),
			path:    "/path/to/invalid/note.md",
			wantErr: true,
		},
		{
			name:    "empty content",
			content: []byte{},
			path:    "/path/to/empty.md",
			wantErr: true,
		},
		{
			name: "missing frontmatter delimiters",
			content: []byte(`id: note123
created: 2025-03-26 14:00
This is not proper frontmatter.`),
			path:    "/path/to/bad.md",
			wantErr: true,
		},
		{
			name: "invalid yaml in frontmatter",
			content: []byte(`---
id: [broken
created: not-a-date
---
Content after invalid YAML.`),
			path:    "/path/to/invalid.md",
			wantErr: true,
		},
		{
			name: "missing required fields",
			content: []byte(`---
tags:
  - test
---
Content with missing required fields.`),
			path:    "/path/to/missing.md",
			wantErr: true,
		},
		{
			name: "section without header",
			content: []byte(`---
id: note123
created: 2025-03-26 14:00
---

abc
`),
			path: "/path/to/no-header-content.md",
			want: model.ParsedNote{
				Metadata: model.NoteMetadata{
					ID:      "note123",
					Created: simpleTestTime,
				},
				HighlightIDs: []string{},
				Content: []model.Section{{
					Type:    model.None,
					Title:   "",
					Content: "abc",
				}},
				Path: "/path/to/no-header-content.md",
				RawFrontmatter: map[string]interface{}{
					"id":      "note123",
					"created": "2025-03-26 14:00",
				},
			},
			wantErr: false,
		},
		{
			name: "empty document after frontmatter",
			content: []byte(`---
id: note123
created: 2025-03-26 14:00
---
`),
			path: "/path/to/empty-content.md",
			want: model.ParsedNote{
				Metadata: model.NoteMetadata{
					ID:      "note123",
					Created: simpleTestTime,
				},
				HighlightIDs: []string{},
				Content:      []model.Section{},
				Path:         "/path/to/empty-content.md",
				RawFrontmatter: map[string]interface{}{
					"id":      "note123",
					"created": "2025-03-26 14:00",
				},
			},
			wantErr: false,
		},
		{
			name: "codeblock are not headers",
			content: []byte(`---
id: headers-note
created: 2025-03-26 14:00
---
# Main title
Some intro text

## Section 1
Content for section 1

` + "```" + `
# My comment
defp some_func(a, b), do: a + b
` + "```" + `

## Section 2
Multi-paragraph content

More text here.

# Footer
Final notes

`),
			path: "/path/to/headers.md",
			want: model.ParsedNote{
				Metadata: model.NoteMetadata{
					ID:      "headers-note",
					Created: simpleTestTime,
				},
				HighlightIDs: []string{},
				Content: []model.Section{
					{
						Type:    model.H1,
						Title:   "Main title",
						Content: "Some intro text\n\n",
					},
					{
						Type:    model.H2,
						Title:   "Section 1",
						Content: "Content for section 1\n\n```\n# My comment\ndefp some_func(a, b), do: a + b\n```\n\n",
					},
					{
						Type:    model.H2,
						Title:   "Section 2",
						Content: "Multi-paragraph content\n\nMore text here.\n\n",
					},
					{
						Type:    model.H1,
						Title:   "Footer",
						Content: "Final notes",
					},
				},
				Path: "/path/to/headers.md",
				RawFrontmatter: map[string]interface{}{
					"id":      "headers-note",
					"created": "2025-03-26 14:00",
				},
			},
			wantErr: false,
		},
		{
			name: "content with headers",
			content: []byte(`---
id: headers-note
created: 2025-03-26 14:00
---
# Main title
Some intro text

## Section 1
Content for section 1

## Section 2
Multi-paragraph content

More text here.

# Footer
Final notes`),
			path: "/path/to/headers.md",
			want: model.ParsedNote{
				Metadata: model.NoteMetadata{
					ID:      "headers-note",
					Created: simpleTestTime,
				},
				HighlightIDs: []string{},
				Content: []model.Section{
					{
						Type:    model.H1,
						Title:   "Main title",
						Content: "Some intro text\n\n",
					},
					{
						Type:    model.H2,
						Title:   "Section 1",
						Content: "Content for section 1\n\n",
					},
					{
						Type:    model.H2,
						Title:   "Section 2",
						Content: "Multi-paragraph content\n\nMore text here.\n\n",
					},
					{
						Type:    model.H1,
						Title:   "Footer",
						Content: "Final notes",
					},
				},
				Path: "/path/to/headers.md",
				RawFrontmatter: map[string]interface{}{
					"id":      "headers-note",
					"created": "2025-03-26 14:00",
				},
			},
			wantErr: false,
		},
		{
			name: "headers without content",
			content: []byte(`---
id: empty-sections
created: 2025-03-26 14:00
---
# Title

## Section 1

## Section 2`),
			path: "/path/to/empty-sections.md",
			want: model.ParsedNote{
				Metadata: model.NoteMetadata{
					ID:      "empty-sections",
					Created: simpleTestTime,
				},
				HighlightIDs: []string{},
				Content: []model.Section{
					{
						Type:    model.H1,
						Title:   "Title",
						Content: "\n",
					},
					{
						Type:    model.H2,
						Title:   "Section 1",
						Content: "\n",
					},
					{
						Type:    model.H2,
						Title:   "Section 2",
						Content: "",
					},
				},
				Path: "/path/to/empty-sections.md",
				RawFrontmatter: map[string]interface{}{
					"id":      "empty-sections",
					"created": "2025-03-26 14:00",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := repository.NewYAMLNoteParser()
			got, err := p.ParseNote(tt.content, tt.path)

			if tt.wantErr {
				assert.Error(t, err, "Expected an error for test case: %s", tt.name)
				return
			}

			require.NoError(t, err, "Unexpected error in test case: %s", tt.name)
			assert.Equal(t, tt.want.Metadata, got.Metadata)
			assert.Equal(t, tt.want.Path, got.Path)
			assert.Equal(t, tt.want.HighlightIDs, got.HighlightIDs)
			assert.Equal(t, tt.want.RawFrontmatter, got.RawFrontmatter)

			// Compare sections
			assert.Equal(t, len(tt.want.Content), len(got.Content),
				fmt.Sprintf("Number of sections should match, \n %+v", got.Content))

			for i := range tt.want.Content {
				if i < len(got.Content) {
					assert.Equal(t, tt.want.Content[i].Type, got.Content[i].Type,
						"Section %d type should match", i)
					assert.Equal(t, tt.want.Content[i].Title, got.Content[i].Title,
						"Section %d title should match", i)
					assert.Equal(t, tt.want.Content[i].Content, got.Content[i].Content,
						"Section %d content should match", i)
				}
			}
		})
	}
}
