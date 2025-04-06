package readdeck

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetHighlightsIntegration(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test. Set `export RUN_INTEGRATION_TEST=true` to run")
	}

	ctx := context.TODO()
	url := os.Getenv("BASE_URL")
	token := os.Getenv("AUTH_TOKEN")

	client := NewHttpClient(http.Client{}, url, token)
	highlights, err := client.GetHighlights(ctx)

	if err != nil {
		t.Fatalf("error while getting highlights: %s", err)
	}

	t.Logf("Number of highlights: %d", len(highlights))

	if len(highlights) > 0 {
		t.Logf("First highlight: %+v", highlights[0])
		t.Logf("Sample highlight - ID: %s, Text: %.50s...",
			highlights[0].ID, highlights[0].Text)
	} else {
		t.Log("No highlights found")
	}
}

func TestBookmarkIntegration(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test. Set `export RUN_INTEGRATION_TEST=true` to run")
	}

	ctx := context.TODO()
	url := os.Getenv("BASE_URL")
	token := os.Getenv("AUTH_TOKEN")

	client := NewHttpClient(http.Client{}, url, token)
	id := "DUvg9NZ93QP9pRbuzHVuyd"
	bookmark, err := client.GetBookmark(ctx, id)

	assert.NoError(t, err, "Should not error")
	assert.NotNil(t, bookmark, "Should have a bookmark")
	created, _ := time.Parse(time.RFC3339Nano, "2025-03-23T20:16:09.387720195Z")

	assert.Equal(t, bookmark.ID, id)
	assert.Empty(t, bookmark.Authors)
	assert.Equal(t, bookmark.Created, created)
	assert.Empty(t, bookmark.Description)
	assert.Equal(t, bookmark.Type, "article")
	assert.Equal(t, bookmark.Href, "https://readlater.deraedt.dev/api/bookmarks/DUvg9NZ93QP9pRbuzHVuyd")
	assert.Equal(t, bookmark.SiteUrl, "https://www.paulgraham.com/schlep.html")
	assert.Equal(t, bookmark.Labels, []string{"business", "entrepreneurial"})
	assert.Equal(t, bookmark.Published, time.Time{})
	assert.Equal(t, bookmark.Title, "Schlep Blindness")
}
