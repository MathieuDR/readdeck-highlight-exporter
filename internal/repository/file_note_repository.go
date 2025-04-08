// internal/repository/file_note_repository.go
package repository

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/mathieudr/readdeck-highlight-exporter/internal/model"
)

type OperationResult struct {
	Type            string // "created", "updated", "unchanged"
	Note            model.Note
	HighlightsAdded int
}

type FileNoteRepository struct {
	fleetingPath string
	noteService  NoteService
	verbose      bool
}

func NewFileNoteRepository(fleetingPath string, noteService NoteService, verbose bool) *FileNoteRepository {
	return &FileNoteRepository{
		fleetingPath: fleetingPath,
		noteService:  noteService,
		verbose:      verbose,
	}
}

func (f *FileNoteRepository) UpsertAll(ctx context.Context, notes []model.Note) ([]OperationResult, error) {
	notePaths, err := f.findNotesInDirectory(f.fleetingPath)
	if err != nil {
		return nil, fmt.Errorf("could not find note paths: %w", err)
	}

	parsedNotes, skippedPaths := f.readNoteFiles(notePaths)

	if f.verbose && len(skippedPaths) > 0 {
		fmt.Printf("\nSkipped %d note(s) due to parsing errors:\n", len(skippedPaths))
		for path, err := range skippedPaths {
			fmt.Printf("  - %s: %v\n", path, err)
		}
	}

	lookup := f.createLookup(parsedNotes)
	results := make([]OperationResult, 0, len(notes))

	for _, toWriteNote := range notes {
		result, err := f.processNote(toWriteNote, lookup)
		if err != nil {
			if f.verbose {
				fmt.Printf("Warning: Failed to process note for bookmark %s: %v\n",
					toWriteNote.Bookmark.ID, err)
			}
			continue
		}
		results = append(results, result)
	}

	return results, nil
}

func (f *FileNoteRepository) processNote(note model.Note, lookup map[string]model.ParsedNote) (OperationResult, error) {
	bookmarkID := note.Bookmark.ID
	existingNote, exists := lookup[bookmarkID]

	if exists {
		updatedNote, highlightsAdded, err := f.updateNote(existingNote, note)
		if err != nil {
			return OperationResult{}, fmt.Errorf("could not update note %s (%s): %w",
				bookmarkID, existingNote.Path, err)
		}

		opType := "updated"
		if highlightsAdded == 0 {
			opType = "unchanged"
		}

		return OperationResult{
			Type:            opType,
			Note:            updatedNote,
			HighlightsAdded: highlightsAdded,
		}, nil
	}

	newNote, err := f.createNote(note)
	if err != nil {
		return OperationResult{}, fmt.Errorf("could not create note %s: %w",
			bookmarkID, err)
	}

	return OperationResult{
		Type:            "created",
		Note:            newNote,
		HighlightsAdded: len(note.Highlights),
	}, nil
}

func (f *FileNoteRepository) updateNote(existingNote model.ParsedNote, note model.Note) (model.Note, int, error) {
	newHighlightsCount := 0
	existingIDs := make(map[string]bool)
	for _, id := range existingNote.HighlightIDs {
		existingIDs[id] = true
	}

	for _, h := range note.Highlights {
		if !existingIDs[h.ID] {
			newHighlightsCount++
		}
	}

	op, err := f.noteService.UpdateNoteContent(existingNote, note)
	if err != nil {
		return model.Note{}, 0, fmt.Errorf("could not generate bytes for update: %w", err)
	}

	result := note
	result.Path = existingNote.Path

	err = f.writeBytes(op.Content, existingNote.Path)
	if err != nil {
		return model.Note{}, 0, err
	}

	return result, newHighlightsCount, nil
}

func (f *FileNoteRepository) createNote(note model.Note) (model.Note, error) {
	operation, err := f.noteService.GenerateNoteContent(note)
	if err != nil {
		return model.Note{}, fmt.Errorf("could not generate bytes for creation: %w", err)
	}

	// TODO: Make immutable
	result := note

	notePath := fmt.Sprintf("%s/%s.md", f.fleetingPath, operation.Metadata.ID)
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

func (f *FileNoteRepository) readNoteFiles(filePaths []string) ([]model.ParsedNote, map[string]error) {
	results := make([]model.ParsedNote, 0, len(filePaths))
	skippedPaths := make(map[string]error)

	for _, path := range filePaths {
		note, err := f.readNoteFile(path)
		if err != nil {
			skippedPaths[path] = err
			continue // Skip this file but continue with others
		}
		results = append(results, note)
	}

	return results, skippedPaths
}

func (f *FileNoteRepository) readNoteFile(filePath string) (model.ParsedNote, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return model.ParsedNote{}, fmt.Errorf("failed to read file: %w", err)
	}

	parsedNote, err := f.noteService.ParseNote(content, filePath)
	if err != nil {
		return model.ParsedNote{}, fmt.Errorf("failed to parse note: %w", err)
	}

	return parsedNote, nil
}
