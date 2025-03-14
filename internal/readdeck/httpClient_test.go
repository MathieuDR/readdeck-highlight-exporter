package readdeck

import (
	"context"
	"net/http"
	"os"
	"testing"
)

func TestGetHighlightsIntegration(t *testing.T) {
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
