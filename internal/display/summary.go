package display

import (
	"fmt"
	"strings"
	"time"

	"github.com/mathieudr/readdeck-highlight-exporter/internal/readdeck"
	"github.com/mathieudr/readdeck-highlight-exporter/internal/repository"
)

func PrintSummary(results []repository.OperationResult, printTiming bool, duration time.Duration) {
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

	fmt.Println("\n" + HeaderColor("Export Summary"))
	fmt.Println(HeaderColor("==================================="))

	fmt.Printf("Processed %d notes (%s created, %s updated, %s unchanged)\n",
		len(results),
		CreatedColor(fmt.Sprintf("%d", created)),
		UpdatedColor(fmt.Sprintf("%d", updated)),
		UnchangedColor(fmt.Sprintf("%d", unchanged)))

	if newHighlights > 0 {
		fmt.Printf("Total highlights: %d (%s new added)\n",
			totalHighlights,
			CreatedColor(fmt.Sprintf("+%d", newHighlights)))
	} else {
		fmt.Printf("Total highlights: %d\n", totalHighlights)
	}

	if printTiming {
		var timeStr string
		if duration.Seconds() < 10 {
			timeStr = fmt.Sprintf("%dms", duration.Milliseconds())
		} else {
			timeStr = fmt.Sprintf("%.2fs", duration.Seconds())
		}
		fmt.Printf("Time: %s\n", TimeColor(timeStr))
	}
}

func PrintDetails(results []repository.OperationResult) {
	fmt.Println("\n" + HeaderColor("Notes Detail"))
	fmt.Println(HeaderColor("==================================="))

	// Group by type for better organization
	createdNotes := filterByType(results, "created")
	updatedNotes := filterByType(results, "updated")
	unchangedNotes := filterByType(results, "unchanged")

	// Print created notes first
	if len(createdNotes) > 0 {
		fmt.Println(BoldCreated("✨ Created:"))
		for _, r := range createdNotes {
			printNoteDetail(r, true)
			fmt.Println("")
		}
	}

	// Print updated notes next
	if len(updatedNotes) > 0 {
		fmt.Println(BoldUpdated("🔄 Updated:"))
		for _, r := range updatedNotes {
			printNoteDetail(r, true)
			fmt.Println("")
		}
	}

	// Print unchanged notes with less detail
	if len(unchangedNotes) > 0 {
		fmt.Println(BoldUnchanged("⏭️ Unchanged:"))
		for _, r := range unchangedNotes {
			printNoteDetail(r, false)
			fmt.Println("")
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

	fmt.Printf("%s\n", BoldTitle(note.Bookmark.Title))

	if detailed {
		fmt.Printf("    Highlights: %d", len(note.Highlights))
		if r.Type == "updated" && r.HighlightsAdded > 0 {
			fmt.Printf(" (%s)", CreatedColor(fmt.Sprintf("+%d", r.HighlightsAdded)))
		}
		fmt.Println()

		fmt.Printf("    Path: %s\n", note.Path)

		colorCounts := getColorBreakdown(note.Highlights)
		if len(colorCounts) > 0 {
			fmt.Printf("    Types: %s\n", formatColorBreakdown(colorCounts))
		}
	} else {
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
	config := repository.DefaultColorConfig()

	colorInfo := []string{}
	for color, count := range colorCounts {
		friendlyName := color
		if name, ok := config.ColorNames[color]; ok {
			friendlyName = name
		}

		segment := fmt.Sprintf("%d %s", count, friendlyName)

		switch color {
		case "yellow":
			colorInfo = append(colorInfo, Yellow(segment))
		case "red":
			colorInfo = append(colorInfo, Red(segment))
		case "blue":
			colorInfo = append(colorInfo, Blue(segment))
		case "green":
			colorInfo = append(colorInfo, Green(segment))
		default:
			colorInfo = append(colorInfo, White(segment))
		}
	}
	return strings.Join(colorInfo, ", ")
}

