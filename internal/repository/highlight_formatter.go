package repository

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mathieudr/readdeck-highlight-exporter/internal/readdeck"
	"github.com/mathieudr/readdeck-highlight-exporter/internal/util"
)

type ColorConfig struct {
	ColorNames map[string]string
	ColorOrder []string
}

func DefaultColorConfig() ColorConfig {
	return ColorConfig{
		ColorNames: map[string]string{
			"yellow": "General highlights",
			"red":    "Thought-provoking insights",
			"blue":   "Important references",
			"green":  "Key takeaways",
		},
		ColorOrder: []string{"green", "red", "yellow", "blue"},
	}
}

type HighlightFormatter struct {
	ColorConfig ColorConfig
}

func NewHighlightFormatter(config ColorConfig) *HighlightFormatter {
	return &HighlightFormatter{
		ColorConfig: config,
	}
}

func (f *HighlightFormatter) FormatHighlightsByColor(highlights []readdeck.Highlight) map[string][]byte {
	groups := f.groupHighlightsByColor(highlights)
	result := make(map[string][]byte)

	for color, hs := range groups {
		var sectionBytes []byte
		sectionBytes = append(sectionBytes, f.highlightTitleBytes(color)...)
		sectionBytes = append(sectionBytes, f.highlightBodyBytes(hs)...)
		result[color] = sectionBytes
	}

	return result
}

func (f *HighlightFormatter) GetSortedColorOrder(highlights map[string][]readdeck.Highlight) []string {
	var result []string

	for _, color := range f.ColorConfig.ColorOrder {
		if _, ok := highlights[color]; ok {
			result = append(result, color)
		}
	}

	var remainingColors []string
	for color := range highlights {
		if !containsString(result, color) {
			remainingColors = append(remainingColors, color)
		}
	}

	sort.Strings(remainingColors)
	result = append(result, remainingColors...)

	return result
}

func (f *HighlightFormatter) highlightTitleBytes(color string) []byte {
	friendlyColor := f.colorToFriendlyName(color)
	return []byte(fmt.Sprintf("## %s\n", friendlyColor))
}

func (f *HighlightFormatter) highlightBodyBytes(highlights []readdeck.Highlight) []byte {
	var result []byte
	for _, h := range highlights {
		highlightBytes := []byte(fmt.Sprintf("%s\n\n", h.Text))
		result = append(result, highlightBytes...)
	}
	return result
}

func (f *HighlightFormatter) groupHighlightsByColor(highlights []readdeck.Highlight) map[string][]readdeck.Highlight {
	result := make(map[string][]readdeck.Highlight)
	for _, h := range highlights {
		result[h.Color] = append(result[h.Color], h)
	}
	return result
}

func (f *HighlightFormatter) colorToFriendlyName(color string) string {
	if name, ok := f.ColorConfig.ColorNames[color]; ok {
		return name
	}

	capitalizedColor := util.Capitalize(strings.ToLower(color))
	return fmt.Sprintf("%s highlights", capitalizedColor)
}

// Helper function to check if a string is in a slice
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
