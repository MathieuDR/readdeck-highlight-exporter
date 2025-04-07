package integration_test

import (
	"context"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/mathieudr/readdeck-highlight-exporter/internal/model"
	"github.com/mathieudr/readdeck-highlight-exporter/internal/readdeck"
	"github.com/mathieudr/readdeck-highlight-exporter/internal/repository"
)

// findProjectRoot returns the absolute path to the project root directory
func findProjectRoot() (string, error) {
	// Start from the current file
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)

	// Walk up until we find go.mod
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root without finding go.mod
			return "", os.ErrNotExist
		}
		dir = parent
	}
}

// TestGenerateFirstThreeNotes is a one-time integration test to validate the export flow
func TestGenerateFirstThreeNotes(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test. Set `export RUN_INTEGRATION_TEST=true` to run")
	}

	// Get credentials from environment
	baseURL := os.Getenv("BASE_URL")
	token := os.Getenv("AUTH_TOKEN")

	if baseURL == "" || token == "" {
		t.Fatal("READDECK_BASE_URL and READDECK_AUTH_TOKEN environment variables must be set")
	}

	// Find project root directory using go.mod as a marker
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}

	// Setup test artifacts directory with absolute path
	artifactsDir := filepath.Join(projectRoot, "test_artifacts")
	if err := os.MkdirAll(artifactsDir, 0755); err != nil {
		t.Fatalf("Failed to create test artifacts directory: %v", err)
	}

	t.Logf("Artifacts will be saved to: %s", artifactsDir)

	// Create client with a reasonable timeout
	httpClient := http.Client{
		Timeout: 30 * time.Second,
	}
	readdeckClient := readdeck.NewHttpClient(httpClient, baseURL, token, 100)

	// Get highlights
	ctx := context.Background()
	highlights, err := readdeckClient.GetHighlights(ctx, nil)
	if err != nil {
		t.Fatalf("Failed to get highlights: %v", err)
	}

	t.Logf("Found %d highlights", len(highlights))
	if len(highlights) == 0 {
		t.Fatal("No highlights found, cannot proceed with test")
	}

	// Limit to first 3 bookmarks
	bookmarkMap := make(map[string][]readdeck.Highlight)
	for _, h := range highlights {
		if h.BookmarkID != "" {
			bookmarkMap[h.BookmarkID] = append(bookmarkMap[h.BookmarkID], h)
		}
	}

	t.Logf("Found %d unique bookmarks", len(bookmarkMap))

	// Process only first 3 bookmarks
	count := 0
	processedBookmarks := make(map[string][]readdeck.Highlight)
	for id, hs := range bookmarkMap {
		processedBookmarks[id] = hs
		count++
		// if count >= 3 {
		// 	break
		// }
	}

	// Setup repository with absolute path
	formatter := repository.NewHighlightFormatter(repository.DefaultColorConfig())
	parser := repository.NewYAMLNoteParser()
	generator := repository.NewYAMLNoteGenerator(formatter)
	updater := repository.NewYAMLNoteUpdater(generator, parser)
	noteService := repository.NewCustomNoteService(parser, generator, updater)
	fleetingNotes := path.Join(projectRoot, "test_artifacts")
	noteRepo := repository.NewFileNoteRepository(fleetingNotes, noteService)

	// Manually get bookmark details and construct notes
	notes := make([]model.Note, 0, len(processedBookmarks))
	for id, hs := range processedBookmarks {
		bookmark, err := readdeckClient.GetBookmark(ctx, id)
		if err != nil {
			t.Fatalf("Failed to get bookmark %s: %v", id, err)
		}

		notes = append(notes, model.Note{
			Bookmark:   bookmark,
			Highlights: hs,
		})
	}

	// Save the notes
	savedNotes, err := noteRepo.UpsertAll(ctx, notes)
	if err != nil {
		t.Fatalf("Failed to save notes: %v", err)
	}

	// Log the results
	t.Logf("Successfully saved %d notes:", len(savedNotes))
	for i, note := range savedNotes {
		t.Logf("%d. %s - %s (%d highlights)",
			i+1,
			note.Bookmark.Title,
			filepath.Base(note.Path),
			len(note.Highlights),
		)
	}

	// Validate the files exist
	for _, note := range savedNotes {
		if _, err := os.Stat(note.Path); os.IsNotExist(err) {
			t.Errorf("Expected file %s does not exist", note.Path)
		}
	}

	t.Log("Test completed successfully!")
}
