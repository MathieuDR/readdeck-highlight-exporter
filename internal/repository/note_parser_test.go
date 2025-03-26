package repository_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/mathieudr/readdeck-highlight-exporter/internal/model"
	"github.com/mathieudr/readdeck-highlight-exporter/internal/repository"
)

func TestYAMLFrontmatterParser_ParseNote(t *testing.T) {
	// Setup test time
	testTime, _ := time.Parse(time.RFC3339, "2025-03-26T14:00:00Z")

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
created: 2025-03-26T14:00:00Z
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
					Created:    testTime,
					ReaddeckID: "rd456",
					Tags:       []string{"go", "programming"},
				},
				Content: "This is the content of the note.\n\nMultiple paragraphs are supported.",
				Path:    "/path/to/note.md",
			},
			wantErr: false,
		},
		{
			name: "note with all metadata fields",
			content: []byte(`---
id: full-note
aliases:
  - "alias1"
  - "alias2"
tags:
  - research
  - papers
created: 2025-03-26T14:00:00Z
readdeck-id: rd789
publish: true
other: true
---
Full content here.`),
			path: "/path/to/full.md",
			want: model.ParsedNote{
				Metadata: model.NoteMetadata{
					ID:         "full-note",
					Aliases:    []string{"alias1", "alias2"},
					Tags:       []string{"research", "papers"},
					Created:    testTime,
					ReaddeckID: "rd789",
					Publish:    true,
				},
				Content: "Full content here.",
				Path:    "/path/to/full.md",
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
created: 2025-03-26T14:00:00Z
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
			name: "empty document after frontmatter",
			content: []byte(`---
id: note123
created: 2025-03-26T14:00:00Z
---
`),
			path: "/path/to/empty-content.md",
			want: model.ParsedNote{
				Metadata: model.NoteMetadata{
					ID:      "note123",
					Created: testTime,
				},
				Content: "",
				Path:    "/path/to/empty-content.md",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := repository.NewYAMLFrontmatterParser()
			got, err := p.ParseNote(tt.content, tt.path)

			if err != nil {
				if !tt.wantErr {
					t.Errorf("ParseNote() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if tt.wantErr {
				t.Logf("\n%+v\n\n", got)
				t.Fatal("ParseNote() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseNote() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
