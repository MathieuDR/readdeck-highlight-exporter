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
	// Create a temporary test directory structure
	tempDir := t.TempDir()

	// Set up test directories and files
	testDirs := []string{
		filepath.Join(tempDir, "notes"),
		filepath.Join(tempDir, "notes/subfolder"),
		filepath.Join(tempDir, "empty"),
	}

	for _, dir := range testDirs {
		err := os.MkdirAll(dir, 0755)
		require.NoError(t, err)
	}

	// Create test files
	testFiles := map[string]bool{
		filepath.Join(tempDir, "notes/note1.md"):           true,
		filepath.Join(tempDir, "notes/note2.md"):           true,
		filepath.Join(tempDir, "notes/subfolder/note3.md"): true,
		filepath.Join(tempDir, "notes/not-a-note.txt"):     false, // Should be ignored
		filepath.Join(tempDir, "notes/.hidden.md"):         false, // Should be ignored
	}

	for path, _ := range testFiles {
		err := os.WriteFile(path, []byte("test content"), 0644)
		require.NoError(t, err)
	}

	// Create mock parser
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

// Mock parser for testing
type mockNoteParser struct{}

func (m *mockNoteParser) ParseNote(content []byte, path string) (model.ParsedNote, error) {
	return model.ParsedNote{}, nil
}

func (m *mockNoteParser) GenerateNoteContent(note model.Note) ([]byte, error) {
	return nil, nil
}

func TestFileNoteRepository_readNoteFile(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		basePath    string
		fleetingDir string
		parser      NoteParser
		// Named input parameters for target function.
		filePath string
		want     model.ParsedNote
		wantErr  bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFileNoteRepository(tt.basePath, tt.fleetingDir, tt.parser)
			got, gotErr := f.readNoteFile(tt.filePath)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("readNoteFile() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("readNoteFile() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("readNoteFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

