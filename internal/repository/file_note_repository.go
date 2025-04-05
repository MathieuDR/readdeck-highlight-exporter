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
	noteService NoteService
}

func NewFileNoteRepository(basePath, fleetingDir string, noteService NoteService) *FileNoteRepository {
	return &FileNoteRepository{
		basePath:    basePath,
		fleetingDir: fleetingDir,
		noteService: noteService,
	}
}

func (f *FileNoteRepository) getFleetingNotesPath() string {
	return path.Join(f.basePath, f.fleetingDir)
}

func (f *FileNoteRepository) UpsertAll(ctx context.Context, notes []model.Note) ([]model.Note, error) {
	notePaths, err := f.findNotesInDirectory(f.getFleetingNotesPath())

	if err != nil {
		return nil, fmt.Errorf("could not find note paths: %w", err)
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
			return model.Note{}, fmt.Errorf("could not update note %s (%s): %w",
				bookmarkID, existingNote.Path, err)
		}
		return updatedNote, nil
	}

	newNote, err := f.createNote(note)
	if err != nil {
		return model.Note{}, fmt.Errorf("could not create note %s: %w",
			bookmarkID, err)
	}
	return newNote, nil
}

func (f *FileNoteRepository) updateNote(existingNote model.ParsedNote, note model.Note) (model.Note, error) {
	op, err := f.noteService.UpdateNoteContent(existingNote, note)

	if err != nil {
		return model.Note{}, fmt.Errorf("could not generate bytes for update: %w", err)
	}

	// TODO: Make immutable
	result := note
	result.Path = existingNote.Path

	err = f.writeBytes(op.Content, existingNote.Path)
	if err != nil {
		return model.Note{}, err
	}

	return result, nil
}

func (f *FileNoteRepository) createNote(note model.Note) (model.Note, error) {
	operation, err := f.noteService.GenerateNoteContent(note)
	if err != nil {
		return model.Note{}, fmt.Errorf("could not generate bytes for creation: %w", err)
	}

	// TODO: Make immutable
	result := note

	notePath := fmt.Sprintf("%s/%s.md", f.getFleetingNotesPath(), operation.Metadata.ID)
	result.Path = notePath

	err = f.writeBytes(operation.Content, notePath)
	if err != nil {
		return model.Note{}, err
	}

	return result, nil
}

func (f *FileNoteRepository) writeBytes(bytes []byte, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("cannot create directory for %s: %w", path, err)
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("cannot open file %s: %w", path, err)
	}
	defer file.Close()

	_, err = file.Write(bytes)
	if err != nil {
		return fmt.Errorf("cannot write file %s: %w", path, err)
	}

	return nil
}

func (f *FileNoteRepository) createLookup(parsedNotes []model.ParsedNote) map[string]model.ParsedNote {
	lookup := make(map[string]model.ParsedNote, len(parsedNotes))
	for _, p := range parsedNotes {
		if p.Metadata.ReaddeckID != "" {
			lookup[p.Metadata.ReaddeckID] = p
		}
	}
	return lookup
}

// LEARNING: Even though it's stateless it ONLY makes sense in the
// context of the FileNoteRepository and thus should be part of that domain
// Another VALID approach would to put this under a util package
func (f *FileNoteRepository) findNotesInDirectory(dirPath string) ([]string, error) {
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return nil, fmt.Errorf("cannot create directory %s: %w", dirPath, err)
	}

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

	parsedNote, err := f.noteService.ParseNote(content, filePath)
	if err != nil {
		return model.ParsedNote{}, fmt.Errorf("failed to parse note at %s: %w", filePath, err)
	}

	return parsedNote, nil
}

