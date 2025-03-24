package service

import (
	"testing"
	"time"

	"github.com/mathieudr/readdeck-highlight-exporter/internal/readdeck"
	"github.com/stretchr/testify/assert"
)

func TestGroupHighlightsByBookmark(t *testing.T) {
	exporter := NewExporter(nil, nil)

	highlights := []readdeck.Highlight{
		{
			ID:         "h1",
			BookmarkID: "book1",
			Text:       "First highlight from book 1",
			Created:    time.Now(),
		},
		{
			ID:         "h2",
			BookmarkID: "book1",
			Text:       "Second highlight from book 1",
			Created:    time.Now(),
		},
		{
			ID:         "h3",
			BookmarkID: "book2",
			Text:       "First highlight from book 2",
			Created:    time.Now(),
		},
		{
			ID:         "h4",
			BookmarkID: "book3",
			Text:       "First highlight from book 3",
			Created:    time.Now(),
		},
		{
			ID:         "h5",
			BookmarkID: "book2",
			Text:       "Second highlight from book 2",
			Created:    time.Now(),
		},
	}

	grouped := exporter.GroupHighlightsByBookmark(highlights)

	assert.Equal(t, 3, len(grouped), "Should have 3 bookmark groups")
	assert.Equal(t, 2, len(grouped["book1"]), "book1 should have 2 highlights")
	assert.Equal(t, 2, len(grouped["book2"]), "book2 should have 2 highlights")
	assert.Equal(t, 1, len(grouped["book3"]), "book3 should have 1 highlight")

	assert.Equal(t, "h1", grouped["book1"][0].ID)
	assert.Equal(t, "h2", grouped["book1"][1].ID)
}

