// internal/repository/file_note_repository.go
package repository

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/mathieudr/readdeck-highlight-exporter/internal/model"
)

type FileNoteRepository struct {
	basePath    string
	fleetingDir string
	parser      NoteParser
}

// NewFileNoteRepository creates a new FileNoteRepository
func NewFileNoteRepository(basePath, fleetingDir string, parser NoteParser) *FileNoteRepository {
	return &FileNoteRepository{
		basePath:    basePath,
		fleetingDir: fleetingDir,
		parser:      parser,
	}
}

// UpsertAll implements NoteRepository.UpsertAll
func (f *FileNoteRepository) UpsertAll(ctx context.Context, notes []model.Note) ([]model.Note, error) {
	// Will be implemented later - focusing on tests first
	panic("not implemented")
}

// LEARNING: Even though it's stateless it ONLY makes sense in the
// context of the FileNoteRepository and thus should be part of that domain
// Another VALID approach would to put this under a util package
func (f *FileNoteRepository) findNotesInDirectory(dirPath string) ([]string, error) {
	notePaths := make([]string, 0)

	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if strings.HasPrefix(filepath.Base(path), ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !d.IsDir() && strings.HasSuffix(path, ".md") {
			notePaths = append(notePaths, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory %s: %w", dirPath, err)
	}

	return notePaths, nil
}

// readNoteFile reads a file and returns its content
func (f *FileNoteRepository) readNoteFile(filePath string) ([]byte, error) {
	// Will be implemented later
	return nil, nil
}

// createOrUpdateNote creates or updates a note file
func (f *FileNoteRepository) createOrUpdateNote(note model.Note) error {
	// Will be implemented later
	return nil
}
