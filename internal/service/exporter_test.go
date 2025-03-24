package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mathieudr/readdeck-highlight-exporter/internal/readdeck"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockReaddeckClient struct {
	mock.Mock
}

func (m *MockReaddeckClient) GetHighlights(ctx context.Context) ([]readdeck.Highlight, error) {
	args := m.Called(ctx)
	return args.Get(0).([]readdeck.Highlight), args.Error(1)
}

func (m *MockReaddeckClient) GetBookmark(ctx context.Context, bookmarkId string) (readdeck.Bookmark, error) {
	args := m.Called(ctx, bookmarkId)
	return args.Get(0).(readdeck.Bookmark), args.Error(1)
}

func TestResolveBookmarks(t *testing.T) {
	mockClient := new(MockReaddeckClient)
	exporter := NewExporter(mockClient, nil)

	ctx := context.Background()

	bookmark1 := readdeck.Bookmark{
		ID:    "book1",
		Title: "Test Book 1",
	}

	bookmark2 := readdeck.Bookmark{
		ID:    "book2",
		Title: "Test Book 2",
	}

	highlight1 := readdeck.Highlight{ID: "h1", BookmarkID: "book1"}
	highlight2 := readdeck.Highlight{ID: "h2", BookmarkID: "book1"}
	highlight3 := readdeck.Highlight{ID: "h3", BookmarkID: "book2"}

	input := map[string][]readdeck.Highlight{
		"book1": {highlight1, highlight2},
		"book2": {highlight3},
	}

	mockClient.On("GetBookmark", ctx, "book1").Return(bookmark1, nil)
	mockClient.On("GetBookmark", ctx, "book2").Return(bookmark2, nil)

	result, err := exporter.resolveBookmarks(ctx, input)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))

	resultMap := make(map[string]BookmarkHighlights)
	for _, item := range result {
		resultMap[item.Bookmark.ID] = item
	}

	assert.Contains(t, resultMap, "book1")
	assert.Equal(t, "Test Book 1", resultMap["book1"].Bookmark.Title)
	assert.Equal(t, 2, len(resultMap["book1"].Highlights))
	assert.Equal(t, "h1", resultMap["book1"].Highlights[0].ID)
	assert.Equal(t, "h2", resultMap["book1"].Highlights[1].ID)

	assert.Contains(t, resultMap, "book2")
	assert.Equal(t, "Test Book 2", resultMap["book2"].Bookmark.Title)
	assert.Equal(t, 1, len(resultMap["book2"].Highlights))
	assert.Equal(t, "h3", resultMap["book2"].Highlights[0].ID)

	// Verify all expectations were met
	mockClient.AssertExpectations(t)
}

func TestResolveBookmarksError(t *testing.T) {
	mockClient := new(MockReaddeckClient)
	exporter := NewExporter(mockClient, nil)

	ctx := context.Background()

	highlight := readdeck.Highlight{ID: "h1", BookmarkID: "book1"}
	input := map[string][]readdeck.Highlight{
		"book1": {highlight},
	}

	expectedErr := errors.New("API error")
	mockClient.On("GetBookmark", ctx, "book1").Return(readdeck.Bookmark{}, expectedErr)

	result, err := exporter.resolveBookmarks(ctx, input)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "book1")
	assert.Contains(t, err.Error(), expectedErr.Error())

	mockClient.AssertExpectations(t)
}

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

	grouped := exporter.groupHighlightsByBookmark(highlights)

	assert.Equal(t, 3, len(grouped), "Should have 3 bookmark groups")
	assert.Equal(t, 2, len(grouped["book1"]), "book1 should have 2 highlights")
	assert.Equal(t, 2, len(grouped["book2"]), "book2 should have 2 highlights")
	assert.Equal(t, 1, len(grouped["book3"]), "book3 should have 1 highlight")

	assert.Equal(t, "h1", grouped["book1"][0].ID)
	assert.Equal(t, "h2", grouped["book1"][1].ID)
}
