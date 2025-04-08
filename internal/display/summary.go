package display

import (
	"fmt"
	"strings"
	"time"

	"github.com/mathieudr/readdeck-highlight-exporter/internal/readdeck"
	"github.com/mathieudr/readdeck-highlight-exporter/internal/repository"
)

func PrintSummary(results []repository.OperationResult, printTiming bool, duration time.Duration) {
	// Count operations by type
	created, updated, unchanged := 0, 0, 0
	totalHighlights, newHighlights := 0, 0

	for _, r := range results {
		switch r.Type {
		case "created":
			created++
		case "updated":
			updated++
		case "unchanged":
			unchanged++
		}

		totalHighlights += len(r.Note.Highlights)
		newHighlights += r.HighlightsAdded
	}

	fmt.Println("\nExport Summary")
	fmt.Println("===================================")
	fmt.Printf("Processed %d notes (%d created, %d updated, %d unchanged)\n",
		len(results), created, updated, unchanged)

	if newHighlights > 0 {
		fmt.Printf("Total highlights: %d (%d new added)\n", totalHighlights, newHighlights)
	} else {
		fmt.Printf("Total highlights: %d\n", totalHighlights)
	}

	if printTiming {
		fmt.Printf("Time: %.2fs\n", duration.Seconds())
	}
}

func PrintDetails(results []repository.OperationResult) {
	fmt.Println("\nNotes Detail")
	fmt.Println("===================================")

	// Group by type for better organization
	createdNotes := filterByType(results, "created")
	updatedNotes := filterByType(results, "updated")
	unchangedNotes := filterByType(results, "unchanged")

	// Print created notes first
	if len(createdNotes) > 0 {
		fmt.Println("\n✨ Created:")
		for _, r := range createdNotes {
			printNoteDetail(r, true)
		}
	}

	// Print updated notes next
	if len(updatedNotes) > 0 {
		fmt.Println("\n🔄 Updated:")
		for _, r := range updatedNotes {
			printNoteDetail(r, true)
		}
	}

	// Print unchanged notes with less detail
	if len(unchangedNotes) > 0 {
		fmt.Println("\n⏭️ Unchanged:")
		for _, r := range unchangedNotes {
			printNoteDetail(r, false)
		}
	}
}

func filterByType(results []repository.OperationResult, opType string) []repository.OperationResult {
	filtered := []repository.OperationResult{}
	for _, r := range results {
		if r.Type == opType {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

func printNoteDetail(r repository.OperationResult, detailed bool) {
	note := r.Note

	// Basic information
	fmt.Printf("  • %s\n", note.Bookmark.Title)

	// For created/updated or verbose mode, show more details
	if detailed {
		fmt.Printf("    Highlights: %d", len(note.Highlights))
		if r.Type == "updated" && r.HighlightsAdded > 0 {
			fmt.Printf(" (+%d new)", r.HighlightsAdded)
		}
		fmt.Println()

		fmt.Printf("    Path: %s\n", note.Path)

		// Get color breakdown with friendly names
		colorCounts := getColorBreakdown(note.Highlights)
		if len(colorCounts) > 0 {
			fmt.Printf("    Types: %s\n", formatColorBreakdown(colorCounts))
		}
	} else {
		// Minimal output for unchanged notes
		fmt.Printf("    Highlights: %d\n", len(note.Highlights))
	}
}

func getColorBreakdown(highlights []readdeck.Highlight) map[string]int {
	colorCounts := make(map[string]int)
	for _, h := range highlights {
		colorCounts[h.Color]++
	}
	return colorCounts
}

func formatColorBreakdown(colorCounts map[string]int) string {
	// Use the same color config from the formatter for consistent naming
	config := repository.DefaultColorConfig()

	colorInfo := []string{}
	for color, count := range colorCounts {
		friendlyName := color
		if name, ok := config.ColorNames[color]; ok {
			friendlyName = name
		}
		colorInfo = append(colorInfo, fmt.Sprintf("%d %s", count, friendlyName))
	}
	return strings.Join(colorInfo, ", ")
}
