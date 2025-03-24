package repository

import (
	"context"
	"github.com/mathieudr/readdeck-highlight-exporter/internal/model"
)

// NoteRepository defines the interface for saving notes
type NoteRepository interface {
	// SaveNote saves a note to the Zettelkasten
	SaveNote(ctx context.Context, note model.Note) error

	// NoteExists checks if a note with similar content already exists
	NoteExists(ctx context.Context, note model.Note) (bool, error)

	// GetExistingNotes retrieves all existing notes for a given bookmark
	GetExistingNotes(ctx context.Context, bookmarkID string) ([]model.Note, error)
}
