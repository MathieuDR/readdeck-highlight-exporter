package repository

import (
	"context"

	"github.com/mathieudr/readdeck-highlight-exporter/internal/model"
)

type NoteRepository interface {
	Upsert(ctx context.Context, note model.Note) (model.Note, error)
}
