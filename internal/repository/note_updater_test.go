package repository

import (
	"reflect"
	"sort"
	"testing"

	"github.com/mathieudr/readdeck-highlight-exporter/internal/readdeck"
)

func TestYAMLNoteUpdater_diffHighlights(t *testing.T) {
	tests := []struct {
		name        string
		existingIds []string
		newIds      []string
		want        []string
	}{
		{
			name:        "empty slices",
			existingIds: []string{},
			newIds:      []string{},
			want:        []string{},
		},
		{
			name:        "empty existing IDs",
			existingIds: []string{},
			newIds:      []string{"a", "b", "c"},
			want:        []string{"a", "b", "c"},
		},
		{
			name:        "empty new IDs",
			existingIds: []string{"a", "b", "c"},
			newIds:      []string{},
			want:        []string{},
		},
		{
			name:        "no overlap",
			existingIds: []string{"a", "b", "c"},
			newIds:      []string{"d", "e", "f"},
			want:        []string{"d", "e", "f"},
		},
		{
			name:        "complete overlap",
			existingIds: []string{"a", "b", "c"},
			newIds:      []string{"a", "b", "c"},
			want:        []string{},
		},
		{
			name:        "partial overlap",
			existingIds: []string{"a", "b", "c"},
			newIds:      []string{"c", "d", "e"},
			want:        []string{"d", "e"},
		},
		{
			name:        "duplicate IDs in new",
			existingIds: []string{"a", "b"},
			newIds:      []string{"c", "c", "d"},
			want:        []string{"c", "c", "d"},
		},
		{
			name:        "duplicate IDs in existing",
			existingIds: []string{"a", "a", "b"},
			newIds:      []string{"a", "c", "d"},
			want:        []string{"c", "d"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We don't actually need real implementations of these for the test
			// since diffHighlights doesn't use them
			u := &YAMLNoteUpdater{}

			got := u.diffHighlights(tt.existingIds, tt.newIds)

			// Sort the slices for stable comparison
			sort.Strings(got)
			sort.Strings(tt.want)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("diffHighlights() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestYAMLNoteUpdater_getHighlights(t *testing.T) {
	// Create some test highlights
	h1 := readdeck.Highlight{ID: "h1", Text: "Highlight 1"}
	h2 := readdeck.Highlight{ID: "h2", Text: "Highlight 2"}
	h3 := readdeck.Highlight{ID: "h3", Text: "Highlight 3"}
	h4 := readdeck.Highlight{ID: "h4", Text: "Highlight 4"}

	tests := []struct {
		name        string
		existingIds []string
		highlights  []readdeck.Highlight
		want        []readdeck.Highlight
	}{
		{
			name:        "empty slices",
			existingIds: []string{},
			highlights:  []readdeck.Highlight{},
			want:        []readdeck.Highlight{},
		},
		{
			name:        "no existing IDs",
			existingIds: []string{},
			highlights:  []readdeck.Highlight{h1, h2, h3},
			want:        []readdeck.Highlight{h1, h2, h3},
		},
		{
			name:        "no new highlights",
			existingIds: []string{"h1", "h2", "h3"},
			highlights:  []readdeck.Highlight{},
			want:        []readdeck.Highlight{},
		},
		{
			name:        "no overlap",
			existingIds: []string{"h1", "h2"},
			highlights:  []readdeck.Highlight{h3, h4},
			want:        []readdeck.Highlight{h3, h4},
		},
		{
			name:        "complete overlap",
			existingIds: []string{"h1", "h2", "h3"},
			highlights:  []readdeck.Highlight{h1, h2, h3},
			want:        []readdeck.Highlight{},
		},
		{
			name:        "partial overlap",
			existingIds: []string{"h1", "h2"},
			highlights:  []readdeck.Highlight{h1, h3, h4},
			want:        []readdeck.Highlight{h3, h4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can create a simple updater since we're only testing getHighlights
			u := &YAMLNoteUpdater{
				// Create a mock implementation of diffHighlights to simulate real behavior
				// This is optional since we're testing a method that calls another method
			}

			got := u.getHighlights(tt.existingIds, tt.highlights)

			// Compare the results
			if !equalHighlightSets(got, tt.want) {
				t.Errorf("getHighlights() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function to compare two slices of highlights
// We need this because the order of highlights might not be guaranteed
func equalHighlightSets(a, b []readdeck.Highlight) bool {
	if len(a) != len(b) {
		return false
	}

	// Create maps of highlight IDs for comparison
	mapA := make(map[string]readdeck.Highlight)
	mapB := make(map[string]readdeck.Highlight)

	for _, h := range a {
		mapA[h.ID] = h
	}

	for _, h := range b {
		mapB[h.ID] = h
	}

	// Compare the maps
	return reflect.DeepEqual(mapA, mapB)
}
