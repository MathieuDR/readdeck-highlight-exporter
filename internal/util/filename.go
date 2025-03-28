package util

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Regular expression for finding invalid filename characters (allow alphanumeric and hyphen)
var slugInvalidChars = regexp.MustCompile(`[^a-z0-9-]`)

// Regular expression for multiple hyphens
var slugMultiHyphens = regexp.MustCompile(`-+`)

// Slugify creates a URL-friendly slug from a string.
func Slugify(title string) string {
	if title == "" {
		return "untitled" // Default slug for empty titles
	}
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")           // Replace spaces first
	slug = slugInvalidChars.ReplaceAllString(slug, "")  // Remove invalid chars
	slug = slugMultiHyphens.ReplaceAllString(slug, "-") // Collapse multiple hyphens
	slug = strings.Trim(slug, "-")                      // Trim leading/trailing hyphens

	// Limit slug length for sanity? E.g., max 50 chars
	maxSlugLength := 50
	if len(slug) > maxSlugLength {
		slug = slug[:maxSlugLength]
	}

	if slug == "" {
		return "untitled" // Default if slug becomes empty after cleaning
	}
	return slug
}

// GenerateNoteID creates a unique ID based on timestamp and title slug.
// This ID is suitable for use as a filename base and in note metadata.
func GenerateNoteID(title string) string {
	timestamp := time.Now().Unix()
	slug := Slugify(title)
	return fmt.Sprintf("%d-%s", timestamp, slug)
}
