package service

import (
	"context"
	"fmt"

	"github.com/mathieudr/readdeck-highlight-exporter/internal/model"
	"github.com/mathieudr/readdeck-highlight-exporter/internal/readdeck"
	"github.com/mathieudr/readdeck-highlight-exporter/internal/repository"
)

type Exporter struct {
	readdeckClient readdeck.Client
	noteRepository repository.NoteRepository
}

func NewExporter(client readdeck.Client, repo repository.NoteRepository) *Exporter {
	return &Exporter{
		readdeckClient: client,
		noteRepository: repo,
	}
}

// Entrypoint
// Needs to get highlights, details, parse them and save them
func (e *Exporter) Export(ctx context.Context) ([]model.Note, error) {
	highlights, err := e.readdeckClient.GetHighlights(ctx)

	if err != nil {
		return nil, err
	}

	groupedHiglights := e.groupHighlightsByBookmark(highlights)
	bookmarkHighlights, err := e.resolveBookmarks(ctx, groupedHiglights)

	if err != nil {
		return nil, err
	}

	return e.noteRepository.UpsertAll(ctx, bookmarkHighlights)
}

func (e *Exporter) resolveBookmarks(ctx context.Context, dict map[string][]readdeck.Highlight) ([]model.Note, error) {
	res := make([]model.Note, 0, len(dict))

	for id, highlights := range dict {
		b, err := e.readdeckClient.GetBookmark(ctx, id)

		if err != nil {
			return nil, fmt.Errorf("Could not retrieve bookmark with id %s: %w", id, err)
		}

		res = append(res, model.Note{
			Bookmark:   b,
			Highlights: highlights,
		})
	}

	return res, nil
}

func (e *Exporter) groupHighlightsByBookmark(highlights []readdeck.Highlight) map[string][]readdeck.Highlight {
	res := make(map[string][]readdeck.Highlight)

	for _, h := range highlights {
		res[h.BookmarkID] = append(res[h.BookmarkID], h)
	}

	return res
}
