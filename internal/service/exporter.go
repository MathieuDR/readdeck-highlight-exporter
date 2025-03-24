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

// func (e *Exporter) Export(ctx context.Context) ([]model.Note, error) {
func (e *Exporter) Export(ctx context.Context) (map[string][]readdeck.Highlight, error) {
	highlights, err := e.readdeckClient.GetHighlights(ctx)

	if err != nil {
		return nil, err
	}

	groupedHiglights := e.groupHighlightsByBookmark(highlights)

	// 1. Fetch highlights from readdeck
	// 2. Group highlights by bookmark
	// 3. Fetch bookmark details
	// 4. For each bookmark, create notes
	// 5. Check for existing notes (idempotency)
	// 6. Save new/modified notes
	// 7. Return the exported notes

	return groupedHiglights, nil
}

func (e *Exporter) resolveBookmarks(ctx context.Context, dict map[string][]readdeck.Highlight) (map[*readdeck.Bookmark][]readdeck.Highlight, error) {
	res := make(map[*readdeck.Bookmark][]readdeck.Highlight)
	for id, highlights := range dict {
		b, err := e.readdeckClient.GetBookmark(ctx, id)

		if err != nil {
			return nil, fmt.Errorf("Could not retrieve bookmark with id %s: %w", id, err)
		}

		res[&b] = highlights
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

func (e *Exporter) convertToNotes(ctx context.Context, groupedHighlights map[string][]readdeck.Highlight) ([]model.Note, error) {
	return nil, nil
}
