// repository/note-parser.go
package repository

import (
	"bytes"
	"fmt"
	"os" // Added for Fprintf to Stderr
	"sort"
	"strings"
	"time"

	"github.com/adrg/frontmatter"
	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"

	"github.com/mathieudr/readdeck-highlight-exporter/internal/model"
	"github.com/mathieudr/readdeck-highlight-exporter/internal/readdeck"
	"github.com/mathieudr/readdeck-highlight-exporter/internal/util"
)

// Updated NoteParser interface
type NoteParser interface {
	ParseNote(content []byte, path string) (model.ParsedNote, error)
	GenerateFrontmatter(meta model.NoteMetadata) ([]byte, error)
	FormatHighlight(h readdeck.Highlight) string
	FormatGroupHeader(groupName string) string
	GetHighlightGroups(highlights []readdeck.Highlight) map[string][]readdeck.Highlight
	GetGroupOrder() []string
	Hasher() *util.GobHasher
}

// YAMLFrontmatterParser implementation
type YAMLFrontmatterParser struct {
	Validator             *validator.Validate
	hasher                *util.GobHasher
	highlightColorToGroup map[string]string
	groupOrder            []string
}

func NewYAMLFrontmatterParser() *YAMLFrontmatterParser {
	// Define color mapping and group order here
	colorMap := map[string]string{
		"yellow": "Highlights",
		"red":    "Contradictions",
		"blue":   "Questions",
		"green":  "Key Takeaways",
		// Add other colors
	}
	// Add "Other Highlights" to the explicit order, usually last
	order := []string{
		"Key Takeaways",
		"Highlights",
		"Contradictions",
		"Questions",
		"Other Highlights", // Handles fallback/unmapped colors
	}

	return &YAMLFrontmatterParser{
		Validator:             validator.New(),
		hasher:                util.NewGobHasher(), // Ensure util.NewGobHasher() exists
		highlightColorToGroup: colorMap,
		groupOrder:            order,
	}
}

func (p *YAMLFrontmatterParser) Hasher() *util.GobHasher {
	return p.hasher
}

func (p *YAMLFrontmatterParser) GetGroupOrder() []string {
	return p.groupOrder
}

// ParseNote parses the frontmatter and content, requiring readdeck-id.
func (p *YAMLFrontmatterParser) ParseNote(content []byte, path string) (model.ParsedNote, error) {
	var matter model.NoteMetadata
	textContentBytes, err := frontmatter.Parse(bytes.NewReader(content), &matter)
	if err != nil {
		// Fallback or specific error handling if needed, e.g., for files without '---'
		return model.ParsedNote{}, fmt.Errorf("could not parse frontmatter from %s: %w", path, err)
	}

	// Validate essential fields needed for idempotency/lookup
	if matter.ReaddeckID == "" {
		return model.ParsedNote{}, fmt.Errorf("frontmatter in %s is missing required field 'readdeck-id'", path)
	}

	highlightIDs, err := p.decodeHighlightIDsHash(matter.ReaddeckHash)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not decode highlight hash for %s: %v. Assuming no previously tracked highlights.\n", path, err)
		highlightIDs = []string{} // Proceed with empty list, updates might re-add highlights
	}

	return model.ParsedNote{
		Path:         path,
		Metadata:     matter,
		Content:      string(textContentBytes), // Raw content after frontmatter
		HighlightIDs: highlightIDs,
	}, nil
}

// GenerateFrontmatter marshals metadata to YAML bytes, wrapped in '---'
func (p *YAMLFrontmatterParser) GenerateFrontmatter(meta model.NoteMetadata) ([]byte, error) {
	// Ensure required fields are present before marshaling if necessary
	// (Validation happens elsewhere, but good practice)

	yamlData, err := yaml.Marshal(&meta)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata to YAML: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(yamlData)
	buf.WriteString("---\n")
	return buf.Bytes(), nil
}

// FormatHighlight formats as a plain paragraph text (trimmed).
func (p *YAMLFrontmatterParser) FormatHighlight(h readdeck.Highlight) string {
	trimmedText := strings.TrimSpace(h.Text)
	// Return just the text; calling function adds spacing/newlines
	return trimmedText
}

// FormatGroupHeader formats a group header line: "## Group Name\n"
func (p *YAMLFrontmatterParser) FormatGroupHeader(groupName string) string {
	return fmt.Sprintf("## %s\n", groupName)
}

// GetHighlightGroups categorizes highlights based on the color map.
func (p *YAMLFrontmatterParser) GetHighlightGroups(highlights []readdeck.Highlight) map[string][]readdeck.Highlight {
	grouped := make(map[string][]readdeck.Highlight)
	otherGroupName := "Other Highlights" // Consistent fallback name

	for _, h := range highlights {
		groupName, ok := p.highlightColorToGroup[strings.ToLower(h.Color)]
		if !ok || groupName == "" { // If color not found OR maps to empty string
			groupName = otherGroupName
		}
		grouped[groupName] = append(grouped[groupName], h)
	}
	// Sort highlights within each group by creation time
	for group := range grouped {
		sort.Slice(grouped[group], func(i, j int) bool {
			// Handle potential nil times if necessary
			return grouped[group][i].Created.Before(grouped[group][j].Created)
		})
	}
	return grouped
}

// decodeHighlightIDsHash decodes the base64 gob-encoded hash.
func (p *YAMLFrontmatterParser) decodeHighlightIDsHash(hash string) ([]string, error) {
	if hash == "" {
		return []string{}, nil
	}
	ids, err := p.hasher.Decode(hash) // Assumes Hasher().Decode exists
	if err != nil {
		// Propagate error for handling upstream
		return nil, fmt.Errorf("could not decode highlight IDs hash: %w", err)
	}
	if ids == nil { // Ensure non-nil slice is returned
		return []string{}, nil
	}
	return ids, nil
}

// Helper to generate the *initial* body content for a new note
func generateInitialBody(bookmark readdeck.Bookmark, groupedHighlights map[string][]readdeck.Highlight, parser NoteParser) string {
	var body strings.Builder

	// 1. Title (H1)
	if bookmark.Title != "" {
		body.WriteString(fmt.Sprintf("# %s\n\n", strings.TrimSpace(bookmark.Title)))
	}

	// 2. Description (optional)
	if bookmark.Description != "" {
		body.WriteString(fmt.Sprintf("%s\n\n", strings.TrimSpace(bookmark.Description)))
	}

	// 3. Highlights grouped and ordered
	groupOrder := parser.GetGroupOrder() // Includes "Other Highlights"

	writeGroup := func(groupName string, highlights []readdeck.Highlight) {
		body.WriteString(parser.FormatGroupHeader(groupName)) // Includes \n
		body.WriteString("\n")                                // Blank line after header

		addedHighlightCount := 0
		for _, h := range highlights {
			formattedHighlight := parser.FormatHighlight(h) // Just the text
			if formattedHighlight != "" {
				if addedHighlightCount > 0 {
					body.WriteString("\n\n") // Blank line *between* highlights
				}
				body.WriteString(formattedHighlight)
				addedHighlightCount++
			}
		}
		// Add trailing newlines *after* the last highlight of the group
		if addedHighlightCount > 0 {
			body.WriteString("\n\n")
		} else {
			// If group exists but had no valid highlights, still add a newline after header
			// The initial newline after header is already added.
			// Ensure at least one newline follows the header.
		}
	}

	// Write groups in the specified order
	for _, groupName := range groupOrder {
		if highlights, ok := groupedHighlights[groupName]; ok && len(highlights) > 0 {
			// Check if there are actually *non-empty* highlights to write
			hasContent := false
			for _, h := range highlights {
				if parser.FormatHighlight(h) != "" {
					hasContent = true
					break
				}
			}
			if hasContent {
				writeGroup(groupName, highlights)
			}
		}
	}

	// Remove potentially excessive trailing newlines, ensure at least one if content exists
	finalBody := strings.TrimSpace(body.String())
	if finalBody != "" {
		return finalBody + "\n"
	}
	return "" // Return empty string if nothing was generated
}

// Helper to generate NoteMetadata (factored out for reuse)
func generateMetadata(noteID string, note model.Note, allHighlightIDs []string, hasher *util.GobHasher) (model.NoteMetadata, error) {
	// Sort IDs for consistent hashing
	sort.Strings(allHighlightIDs)
	hash, err := hasher.Encode(allHighlightIDs)
	if err != nil {
		return model.NoteMetadata{}, fmt.Errorf("failed to encode highlight IDs: %w", err)
	}

	createdTime := note.Bookmark.Created
	if createdTime.IsZero() {
		createdTime = time.Now()
	}
	publishedTime := note.Bookmark.Published

	aliases := []string{}
	if note.Bookmark.Title != "" {
		aliases = append(aliases, note.Bookmark.Title)
	}

	tags := []string{"readdeck/highlight"} // Base tag
	if len(note.Bookmark.Labels) > 0 {
		tags = append(tags, note.Bookmark.Labels...)
	}

	authors := note.Bookmark.Authors
	if len(authors) == 0 {
		authors = nil // Use nil for cleaner YAML
	}

	metadata := model.NoteMetadata{
		ID:           noteID, // The generated <TIMESTAMP>-<SLUG>
		Aliases:      aliases,
		Tags:         tags,
		Created:      createdTime,
		ReaddeckID:   note.Bookmark.ID, // Source ID
		ReaddeckHash: hash,             // Hash of all current highlight IDs
		Media:        note.Bookmark.Title,
		Type:         note.Bookmark.Type,
		Published:    publishedTime,
		ArchiveUrl:   note.Bookmark.Href,
		Site:         note.Bookmark.SiteUrl,
		Authors:      authors,
	}

	// Clean potentially empty slices AFTER creation
	if len(metadata.Aliases) == 0 {
		metadata.Aliases = nil
	}
	// Keep "readdeck/highlight" tag even if no labels? Yes.

	return metadata, nil
}

