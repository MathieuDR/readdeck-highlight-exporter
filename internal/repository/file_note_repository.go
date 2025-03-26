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

func NewFileNoteRepository(basePath, fleetingDir string, parser NoteParser) *FileNoteRepository {
	return &FileNoteRepository{
		basePath:    basePath,
		fleetingDir: fleetingDir,
		parser:      parser,
	}
}

func (f *FileNoteRepository) UpsertAll(ctx context.Context, notes []model.Note) ([]model.Note, error) {
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

func (f *FileNoteRepository) readNoteFiles(filePaths []string) ([]model.ParsedNote, error) {
	results := make([]model.ParsedNote, 0, len(filePaths))

	// Can make parallel later
	for _, path := range filePaths {
		note, err := f.readNoteFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read note at %s: %w", path, err)
		}
		results = append(results, note)
	}

	return results, nil
}

func (f *FileNoteRepository) readNoteFile(filePath string) (model.ParsedNote, error) {
	panic("not implemented")
}

func (f *FileNoteRepository) createOrUpdateNote(note model.Note) error {
	return nil
}
