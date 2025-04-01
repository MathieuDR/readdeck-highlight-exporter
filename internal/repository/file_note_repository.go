package repository

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"
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

func (f *FileNoteRepository) getFleetingNotesPath() string {
	return path.Join(f.basePath, f.fleetingDir)
}

func (f *FileNoteRepository) UpsertAll(ctx context.Context, notes []model.Note) ([]model.Note, error) {
	notePaths, err := f.findNotesInDirectory(f.getFleetingNotesPath())

	if err != nil {
		return nil, fmt.Errorf("Could not find note paths: %w", err)
	}

	parsedNotes, err := f.readNoteFiles(notePaths)
	if err != nil {
		return nil, err
	}

	lookup := f.createLookup(parsedNotes)
	result := make([]model.Note, len(notes))

	for i, toWriteNote := range notes {
		var err error

		result[i], err = f.processNote(toWriteNote, lookup)

		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (f *FileNoteRepository) processNote(note model.Note, lookup map[string]model.ParsedNote) (model.Note, error) {
	bookmarkID := note.Bookmark.ID
	existingNote, exists := lookup[bookmarkID]

	if exists {
		updatedNote, err := f.updateNote(existingNote, note)
		if err != nil {
			return model.Note{}, fmt.Errorf("Could not update note %s (%s): %w",
				bookmarkID, existingNote.Path, err)
		}
		return updatedNote, nil
	}

	newNote, err := f.createNote(note)
	if err != nil {
		return model.Note{}, fmt.Errorf("Could not create note %s: %w",
			bookmarkID, err)
	}
	return newNote, nil
}

func (f *FileNoteRepository) updateNote(existingNote model.ParsedNote, request model.Note) (model.Note, error) {
	return model.Note{}, nil

}

// TODO: Make request a copy, so it's immutable
func (f *FileNoteRepository) createNote(request model.Note) (model.Note, error) {
	operation, err := f.parser.GenerateNoteContent(request)
	if err != nil {
		return model.Note{}, fmt.Errorf("Could not generate bytes: %w", err)
	}

	path := fmt.Sprintf("%s/%s.md", f.getFleetingNotesPath(), operation.Metadata.ID)
	request.Path = path
	err = f.writeBytes(operation.Content, path)

	if err != nil {
		return model.Note{}, err
	}

	return request, nil
}

func (f *FileNoteRepository) writeBytes(bytes []byte, path string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("Can not open file %s: %w", path, err)
	}

	_, err = file.Write(bytes)

	if err != nil {
		return fmt.Errorf("Can not write file %s: %w", path, err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("Can not close file %s: %w", path, err)
	}

	return nil
}

func (f *FileNoteRepository) createLookup(parsedNotes []model.ParsedNote) map[string]model.ParsedNote {
	lookup := make(map[string]model.ParsedNote, len(parsedNotes))
	for _, p := range parsedNotes {
		if p.Metadata.ID != "" {
			lookup[p.Metadata.ID] = p
		}
	}
	return lookup
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
	content, err := os.ReadFile(filePath)
	if err != nil {
		return model.ParsedNote{}, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	parsedNote, err := f.parser.ParseNote(content, filePath)
	if err != nil {
		return model.ParsedNote{}, fmt.Errorf("failed to parse note at %s: %w", filePath, err)
	}

	return parsedNote, nil
}
