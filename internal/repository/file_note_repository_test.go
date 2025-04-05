package repository

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/mathieudr/readdeck-highlight-exporter/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileNoteRepository_findNotesInDirectory(t *testing.T) {
	tempDir := t.TempDir()

	testDirs := []string{
		filepath.Join(tempDir, "notes"),
		filepath.Join(tempDir, "notes/subfolder"),
		filepath.Join(tempDir, "empty"),
	}

	for _, dir := range testDirs {
		err := os.MkdirAll(dir, 0755)
		require.NoError(t, err)
	}

	testFiles := map[string]bool{
		filepath.Join(tempDir, "notes/note1.md"):           true,
		filepath.Join(tempDir, "notes/note2.md"):           true,
		filepath.Join(tempDir, "notes/subfolder/note3.md"): true,
		filepath.Join(tempDir, "notes/not-a-note.txt"):     false, // Should be ignored
		filepath.Join(tempDir, "notes/.hidden.md"):         false, // Should be ignored
	}

	for path := range testFiles {
		err := os.WriteFile(path, []byte("test content"), 0644)
		require.NoError(t, err)
	}

	mockParser := &mockNoteParser{}

	// Define test cases
	tests := []struct {
		name        string
		basePath    string
		fleetingDir string
		dirPath     string
		want        []string
		wantErr     bool
	}{
		{
			name:        "Find markdown files in directory with subdirectories",
			basePath:    tempDir,
			fleetingDir: "notes",
			dirPath:     filepath.Join(tempDir, "notes"),
			want: []string{
				filepath.Join(tempDir, "notes/note1.md"),
				filepath.Join(tempDir, "notes/note2.md"),
				filepath.Join(tempDir, "notes/subfolder/note3.md"),
			},
			wantErr: false,
		},
		{
			name:        "Empty directory returns empty slice",
			basePath:    tempDir,
			fleetingDir: "empty",
			dirPath:     filepath.Join(tempDir, "empty"),
			want:        []string{},
			wantErr:     false,
		},
		{
			name:        "Non-existent directory returns error",
			basePath:    tempDir,
			fleetingDir: "nonexistent",
			dirPath:     filepath.Join(tempDir, "nonexistent"),
			want:        nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFileNoteRepository(tt.basePath, tt.fleetingDir, mockParser)
			got, err := f.findNotesInDirectory(tt.dirPath)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Sort the results for consistent comparison
			sort.Strings(got)
			sort.Strings(tt.want)

			assert.Equal(t, tt.want, got)
		})
	}
}

type mockNoteParser struct{}

func (m *mockNoteParser) ParseNote(content []byte, path string) (model.ParsedNote, error) {
	return model.ParsedNote{}, nil
}

func (m *mockNoteParser) GenerateNoteContent(note model.Note) (NoteOperation, error) {
	return NoteOperation{
		Metadata: model.NoteMetadata{ID: "my-id"},
		Content:  []byte{},
	}, nil
}

func (m *mockNoteParser) UpdateNoteContent(existing model.ParsedNote, note model.Note) (NoteOperation, error) {
	return NoteOperation{
		Metadata: model.NoteMetadata{ID: "my-id"},
		Content:  []byte{},
	}, nil
}

func TestFileNoteRepository_createLookup(t *testing.T) {
	tests := []struct {
		name        string
		parsedNotes []model.ParsedNote
		want        map[string]model.ParsedNote
	}{
		{
			name: "Create lookup from parsed notes with ReaddeckID",
			parsedNotes: []model.ParsedNote{
				{
					Path:     "path1",
					Metadata: model.NoteMetadata{ID: "id1", ReaddeckID: "readdeck1"},
					Content:  []model.Section{},
				},
				{
					Path:     "path2",
					Metadata: model.NoteMetadata{ID: "id2", ReaddeckID: "readdeck2"},
					Content:  []model.Section{},
				},
			},
			want: map[string]model.ParsedNote{
				"readdeck1": {
					Path:     "path1",
					Metadata: model.NoteMetadata{ID: "id1", ReaddeckID: "readdeck1"},
					Content:  []model.Section{},
				},
				"readdeck2": {
					Path:     "path2",
					Metadata: model.NoteMetadata{ID: "id2", ReaddeckID: "readdeck2"},
					Content:  []model.Section{},
				},
			},
		},
		{
			name: "Notes without ReaddeckID are not included in lookup",
			parsedNotes: []model.ParsedNote{
				{
					Path:     "path1",
					Metadata: model.NoteMetadata{ID: "id1", ReaddeckID: "readdeck1"},
					Content:  []model.Section{},
				},
				{
					Path:     "path2",
					Metadata: model.NoteMetadata{ID: "id2", ReaddeckID: ""},
					Content:  []model.Section{},
				},
			},
			want: map[string]model.ParsedNote{
				"readdeck1": {
					Path:     "path1",
					Metadata: model.NoteMetadata{ID: "id1", ReaddeckID: "readdeck1"},
					Content:  []model.Section{},
				},
			},
		},
		{
			name:        "Empty input returns empty map",
			parsedNotes: []model.ParsedNote{},
			want:        map[string]model.ParsedNote{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFileNoteRepository("", "", nil)
			got := f.createLookup(tt.parsedNotes)
			assert.Equal(t, tt.want, got)
		})
	}
}
