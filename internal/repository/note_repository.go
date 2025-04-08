package repository

import (
	"context"

	"github.com/mathieudr/readdeck-highlight-exporter/internal/model"
)

type NoteRepository interface {
	UpsertAll(ctx context.Context, notes []model.Note) ([]OperationResult, error)
}

var _ NoteRepository = (*FileNoteRepository)(nil)
