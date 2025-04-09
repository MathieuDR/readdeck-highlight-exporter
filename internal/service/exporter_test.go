package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mathieudr/readdeck-highlight-exporter/internal/model"
	"github.com/mathieudr/readdeck-highlight-exporter/internal/readdeck"
	"github.com/mathieudr/readdeck-highlight-exporter/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockReaddeckClient struct {
	mock.Mock
}

type MockNoteRepository struct {
	mock.Mock
}

func (m *MockNoteRepository) UpsertAll(ctx context.Context, note []model.Note) ([]repository.OperationResult, error) {
	args := m.Called(ctx, note)
	return args.Get(0).([]repository.OperationResult), args.Error(1)
}

func (m *MockReaddeckClient) GetHighlights(ctx context.Context, since *time.Time) ([]readdeck.Highlight, error) {
	args := m.Called(ctx)
	return args.Get(0).([]readdeck.Highlight), args.Error(1)
}

func (m *MockReaddeckClient) GetBookmark(ctx context.Context, bookmarkId string) (readdeck.Bookmark, error) {
	args := m.Called(ctx, bookmarkId)
	return args.Get(0).(readdeck.Bookmark), args.Error(1)
}

func TestExport(t *testing.T) {
	mockClient := new(MockReaddeckClient)
	mockRepo := new(MockNoteRepository)
	exporter := NewExporter(mockClient, mockRepo)

	ctx := context.Background()

	highlight1 := readdeck.Highlight{ID: "h1", BookmarkID: "book1", Text: "First highlight"}
	highlight2 := readdeck.Highlight{ID: "h2", BookmarkID: "book2", Text: "Second highlight"}
	highlights := []readdeck.Highlight{highlight1, highlight2}

	bookmark1 := readdeck.Bookmark{ID: "book1", Title: "Test Book 1"}
	bookmark2 := readdeck.Bookmark{ID: "book2", Title: "Test Book 2"}

	expectedNote1 := model.Note{
		Path:       "/path/to/book1.md",
		Bookmark:   bookmark1,
		Highlights: []readdeck.Highlight{highlight1},
	}

	expectedNote2 := model.Note{
		Path:       "/path/to/book2.md",
		Bookmark:   bookmark2,
		Highlights: []readdeck.Highlight{highlight2},
	}

	expectedResults := []repository.OperationResult{
		{
			Type:            "created",
			Note:            expectedNote1,
			HighlightsAdded: 1,
		},
		{
			Type:            "created",
			Note:            expectedNote2,
			HighlightsAdded: 1,
		},
	}

	mockClient.On("GetHighlights", ctx).Return(highlights, nil)
	mockClient.On("GetBookmark", ctx, "book1").Return(bookmark1, nil)
	mockClient.On("GetBookmark", ctx, "book2").Return(bookmark2, nil)
	mockRepo.On("UpsertAll", ctx, mock.Anything).Return(expectedResults, nil)

	operationResults, err := exporter.Export(ctx)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(operationResults))

	noteMap := make(map[string]model.Note)
	for _, result := range operationResults {
		noteMap[result.Note.Bookmark.ID] = result.Note
	}

	assert.Equal(t, expectedNote1, noteMap["book1"])
	assert.Equal(t, expectedNote2, noteMap["book2"])

	// VerifyOnExit
	mockClient.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
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

	resultMap := make(map[string]model.Note)
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
