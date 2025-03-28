// repository/filenote_repository.go
package repository

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mathieudr/readdeck-highlight-exporter/internal/model"
	"github.com/mathieudr/readdeck-highlight-exporter/internal/readdeck"
	"github.com/mathieudr/readdeck-highlight-exporter/internal/util"
)

// FileNoteRepository struct and NewFileNoteRepository remain the same
type FileNoteRepository struct {
	basePath    string
	fleetingDir string
	parser      NoteParser
}

func NewFileNoteRepository(basePath, fleetingDir string, parser NoteParser) *FileNoteRepository {
	return &FileNoteRepository{
		basePath:    basePath,
		fleetingDir: fleetingDir,
		parser:      parser,
	}
}

func (f *FileNoteRepository) getFleetingNotesPath() string {
	dir := path.Join(f.basePath, f.fleetingDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		// Log or handle critical error - maybe return error from here?
		fmt.Fprintf(os.Stderr, "Critical: could not create directory %s: %v\n", dir, err)
	}
	return dir
}

// UpsertAll orchestrates finding, parsing, and processing notes.
func (f *FileNoteRepository) UpsertAll(ctx context.Context, notes []model.Note) ([]model.Note, error) {
	fleetingPath := f.getFleetingNotesPath()
	if _, err := os.Stat(fleetingPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("fleeting notes directory does not exist or could not be created: %s", fleetingPath)
	}

	notePaths, err := f.findNotesInDirectory(fleetingPath)
	if err != nil {
		return nil, fmt.Errorf("could not find existing notes: %w", err)
	}

	parsedNotes, err := f.readNoteFiles(notePaths) // Errors reading individual files are logged and skipped inside
	if err != nil {
		// This would be an error from readNoteFiles itself, not individual files
		return nil, fmt.Errorf("unexpected error reading note files: %w", err)
	}

	lookup := f.createLookup(parsedNotes) // Key: ReaddeckID, Value: ParsedNote
	result := make([]model.Note, len(notes))

	for i, noteToWrite := range notes {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		processedNote, err := f.processNote(noteToWrite, lookup)
		if err != nil {
			return nil, fmt.Errorf("failed to process note for Readdeck bookmark ID %s: %w", noteToWrite.Bookmark.ID, err)
		}
		result[i] = processedNote // Store note with its final path
	}

	return result, nil
}

// processNote decides whether to create or update a note.
func (f *FileNoteRepository) processNote(note model.Note, lookup map[string]model.ParsedNote) (model.Note, error) {
	bookmarkID := note.Bookmark.ID
	if bookmarkID == "" { // Safety check
		return model.Note{}, fmt.Errorf("attempted to process note with empty Readdeck bookmark ID (Title: %s)", note.Bookmark.Title)
	}

	existingNote, exists := lookup[bookmarkID]

	if exists {
		updatedNote, err := f.updateNote(existingNote, note)
		if err != nil {
			// Error already includes context from updateNote
			return model.Note{}, err
		}
		return updatedNote, nil
	}

	newNote, err := f.createNote(note)
	if err != nil {
		// Error already includes context from createNote
		return model.Note{}, err
	}
	return newNote, nil
}

// createNote generates content for a new note and writes it to a file.
func (f *FileNoteRepository) createNote(request model.Note) (model.Note, error) {
	// 1. Generate unique ID and path
	noteID := util.GenerateNoteID(request.Bookmark.Title) // Use util function
	filename := noteID + ".md"
	filePath := path.Join(f.getFleetingNotesPath(), filename)

	// 2. Prepare metadata
	allHighlightIDs := make([]string, len(request.Highlights))
	for i, h := range request.Highlights { allHighlightIDs[i] = h.ID }
	metadata, err := generateMetadata(noteID, request, allHighlightIDs, f.parser.Hasher())
	if err != nil {
		return model.Note{}, fmt.Errorf("create: failed to generate metadata for %s: %w", noteID, err)
	}

	// 3. Generate frontmatter YAML
	frontmatterBytes, err := f.parser.GenerateFrontmatter(metadata)
	if err != nil {
		return model.Note{}, fmt.Errorf("create: failed to generate frontmatter for %s: %w", noteID, err)
	}

	// 4. Generate initial body content
	groupedHighlights := f.parser.GetHighlightGroups(request.Highlights)
	bodyContent := generateInitialBody(request.Bookmark, groupedHighlights, f.parser) // Use helper

	// 5. Combine and write
	var finalContent bytes.Buffer
	finalContent.Write(frontmatterBytes)
	if bodyContent != "" {
		finalContent.WriteString("\n") // Separator line between frontmatter and body
		finalContent.WriteString(bodyContent) // Already ends with \n if not empty
	}

	err = os.WriteFile(filePath, finalContent.Bytes(), 0644)
	if err != nil {
		return model.Note{}, fmt.Errorf("create: failed to write new note file %s: %w", filePath, err)
	}

	// 6. Update Note model with path
	request.Path = filePath
	fmt.Printf("Created note: %s\n", filePath)
	return request, nil
}

// updateNote handles merging new highlights into an existing note non-destructively.
func (f *FileNoteRepository) updateNote(existingNote model.ParsedNote, request model.Note) (model.Note, error) {
	// 1. Calculate expected hash of *all* incoming highlights
	allIncomingHighlightIDs := make([]string, len(request.Highlights))
	for i, h := range request.Highlights { allIncomingHighlightIDs[i] = h.ID }
	sort.Strings(allIncomingHighlightIDs)
	expectedNewHash, err := f.parser.Hasher().Encode(allIncomingHighlightIDs)
	if err != nil {
		return model.Note{}, fmt.Errorf("update: failed to encode incoming highlight IDs for %s: %w", existingNote.Path, err)
	}

	// 2. Check if highlights have changed (using the hash)
	if existingNote.Metadata.ReaddeckHash == expectedNewHash {
		//fmt.Printf("Skipping update for %s: highlights unchanged (hash: %s)\n", existingNote.Path, expectedNewHash)
		request.Path = existingNote.Path // Ensure path is set on the request model
		return request, nil
	}

	fmt.Printf("Updating note %s: highlights changed (old hash: %s, new hash: %s)\n",
		existingNote.Path, existingNote.Metadata.ReaddeckHash, expectedNewHash)

	// 3. Identify *new* highlights to append
	existingIDsSet := make(map[string]struct{})
	for _, id := range existingNote.HighlightIDs { existingIDsSet[id] = struct{}{} }
	newHighlights := []readdeck.Highlight{}
	for _, h := range request.Highlights {
		if _, exists := existingIDsSet[h.ID]; !exists {
			newHighlights = append(newHighlights, h)
		}
	}

	// 4. Group the *new* highlights
	newHighlightsByGroup := f.parser.GetHighlightGroups(newHighlights)

	// 5. Prepare updated body content by appending
	updatedBodyContent := f.appendHighlightsToBody(existingNote.Content, newHighlightsByGroup)

	// 6. Prepare updated metadata using existing Note ID and *all* current highlights
	// Note: Use existingNote.Metadata.ID as the canonical ID for this file.
	if existingNote.Metadata.ID == "" {
		// This case shouldn't happen if createNote always sets an ID, but handle defensively
		fmt.Fprintf(os.Stderr, "Warning: Existing note %s is missing 'id' in frontmatter. Generating a new one might be needed if lookup depends on it.\n", existingNote.Path)
		// Optionally generate one now if required downstream? For now, use ReaddeckID as fallback? Risky.
		// Let's assume createNote sets it. If not, this could be an issue.
		return model.Note{}, fmt.Errorf("update: existing note %s is missing 'id' in frontmatter", existingNote.Path)
	}
	updatedMetadata, err := generateMetadata(existingNote.Metadata.ID, request, allIncomingHighlightIDs, f.parser.Hasher())
	if err != nil {
		return model.Note{}, fmt.Errorf("update: failed to generate updated metadata for %s: %w", existingNote.Path, err)
	}

	// 7. Generate new frontmatter YAML
	newFrontmatterBytes, err := f.parser.GenerateFrontmatter(updatedMetadata)
	if err != nil {
		return model.Note{}, fmt.Errorf("update: failed to generate updated frontmatter for %s: %w", existingNote.Path, err)
	}

	// 8. Combine new frontmatter and updated body
	var finalContent bytes.Buffer
	finalContent.Write(newFrontmatterBytes)
	if updatedBodyContent != "" {
		finalContent.WriteString("\n") // Separator line
		finalContent.WriteString(updatedBodyContent) // Already ends with \n if not empty
	}

	// 9. Write the updated content back to the existing file path
	err = os.WriteFile(existingNote.Path, finalContent.Bytes(), 0644)
	if err != nil {
		return model.Note{}, fmt.Errorf("update: failed to write updated note file %s: %w", existingNote.Path, err)
	}

	// 10. Return the request model updated with the correct path
	request.Path = existingNote.Path
	fmt.Printf("Updated note: %s\n", existingNote.Path)
	return request, nil
}

// appendHighlightsToBody intelligently adds new highlights (as plain paragraphs)
// to the end of their respective group sections in the existing content.
func (f *FileNoteRepository) appendHighlightsToBody(existingBody string, newHighlightsByGroup map[string][]readdeck.Highlight) string {
	if len(newHighlightsByGroup) == 0 {
		return existingBody // No new highlights to add
	}

	type section struct {
		groupName string   // Just the "Group Name" part
		lines     []string // Lines belonging to this section (excluding header)
	}
	var sections []*section // Use pointers to modify in place
	var currentSection *section
	var nonGroupLines []string

	scanner := bufio.NewScanner(strings.NewReader(existingBody))
	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		if strings.HasPrefix(trimmedLine, "## ") {
			groupName := strings.TrimSpace(strings.TrimPrefix(trimmedLine, "## "))
			s := §ion{groupName: groupName, lines: []string{}} // Store header line
			sections = append(sections, s)
			currentSection = s // Point to the newly added section
		} else if currentSection != nil {
			currentSection.lines = append(currentSection.lines, line)
		} else {
			nonGroupLines = append(nonGroupLines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Error scanning existing body content: %v. Returning original content.\n", err)
		return existingBody
	}

	sectionMap := make(map[string]*section)
	for _, s := range sections {
		sectionMap[s.groupName] = s
	}

	groupsToAppend := make(map[string][]readdeck.Highlight)
	for name, highlights := range newHighlightsByGroup {
		if len(highlights) > 0 { groupsToAppend[name] = highlights }
	}

	// Append new highlights to existing sections
	for groupName, highlights := range groupsToAppend {
		if s, exists := sectionMap[groupName]; exists {
			// Ensure separation before the first new highlight
			if len(s.lines) > 0 && strings.TrimSpace(s.lines[len(s.lines)-1]) != "" {
				s.lines = append(s.lines, "") // Add blank line separator
			} else if len(s.lines) == 0 {
				// If section existed but was empty, add a newline after header implicit in structure
			}

			addedHighlightCount := 0
			for _, h := range highlights {
				formattedHighlight := f.parser.FormatHighlight(h)
				if formattedHighlight != "" {
					if addedHighlightCount > 0 {
						s.lines = append(s.lines, "") // Blank line *between* new highlights
					}
					s.lines = append(s.lines, formattedHighlight)
					addedHighlightCount++
				}
			}
			delete(groupsToAppend, groupName) // Mark as processed
		}
	}

	// Reconstruct the body
	var result strings.Builder
	if len(nonGroupLines) > 0 {
		result.WriteString(strings.Join(nonGroupLines, "\n") + "\n")
	}

	for _, s := range sections {
		result.WriteString(f.parser.FormatGroupHeader(s.groupName)) // "## GroupName\n"
		if len(s.lines) > 0 {
			result.WriteString(strings.Join(s.lines, "\n") + "\n")
		} else {
			result.WriteString("\n") // Ensure at least one blank line follows header if section is empty
		}
	}

	// Append entirely new groups at the end, respecting order
	groupOrder := f.parser.GetGroupOrder() // Includes "Other Highlights"
	for _, groupName := range groupOrder {
		if highlights, ok := groupsToAppend[groupName]; ok {
			// Ensure separation from previous content
			if result.Len() > 0 {
				// Need two newlines if last char wasn't newline, or one otherwise
				trimmedResult := strings.TrimRight(result.String(), "\n")
				result.Reset()
				result.WriteString(trimmedResult)
				result.WriteString("\n\n")
			}

			result.WriteString(f.parser.FormatGroupHeader(groupName))
			result.WriteString("\n") // Blank line after header
			addedHighlightCount := 0
			for _, h := range highlights {
				formattedHighlight := f.parser.FormatHighlight(h)
				if formattedHighlight != "" {
					if addedHighlightCount > 0 { result.WriteString("\n\n") } // Blank line between
					result.WriteString(formattedHighlight)
					addedHighlightCount++
				}
			}
			result.WriteString("\n") // Ensure newline after last highlight in group
		}
	}

	// Final cleanup: ensure single trailing newline if content exists
	finalBody := result.String()
	if finalBody != "" {
		finalBody = strings.TrimRight(finalBody, " \t\n\r") + "\n"
	}

	return finalBody
}


// createLookup builds map from Readdeck Bookmark ID to ParsedNote.
func (f *FileNoteRepository) createLookup(parsedNotes []model.ParsedNote) map[string]model.ParsedNote {
	lookup := make(map[string]model.ParsedNote, len(parsedNotes))
	for _, p := range parsedNotes {
		// Note: p.Metadata comes directly from parsing the file
		if p.Metadata.ReaddeckID != "" {
			if _, exists := lookup[p.Metadata.ReaddeckID]; exists {
				// Log duplicate Readdeck IDs found across different files
				fmt.Fprintf(os.Stderr, "Warning: Duplicate Readdeck ID '%s' found. Note '%s' overwrites previous entry in lookup.\n", p.Metadata.ReaddeckID, p.Path)
			}
			lookup[p.Metadata.ReaddeckID] = p
		}
		// Notes missing ReaddeckID are skipped (error during ParseNote prevents them)
	}
	return lookup
}

// findNotesInDirectory finds all .md files recursively, skipping hidden ones.
func (f *FileNoteRepository) findNotesInDirectory(dirPath string) ([]string, error) {
	notePaths := make([]string, 0)
	err := filepath. WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Error accessing path (e.g., permissions)
			fmt.Fprintf(os.Stderr, "Warning: Error accessing path %s during scan: %v. Skipping.\n", path, err)
			// Allow WalkDir to continue if possible, return nil to ignore error for this entry
			// Return err directly if you want to halt the walk on permission errors etc.
			return nil // Or return err to stop walk
		}

		// Skip hidden files and directories
		if strings.HasPrefix(filepath.Base(path), ".") {
			if d.IsDir() {
				return filepath.SkipDir // Don't descend into hidden directories
			}
			return nil // Skip hidden files
		}

		// Process only non-directories with .md extension
		if !d.IsDir() && strings.HasSuffix(strings.ToLower(path), ".md") {
			absPath, err := filepath.Abs(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not get absolute path for %s: %v. Skipping.\n", path, err)
				return nil // Skip this file
			}
			notePaths = append(notePaths, absPath)
		}
		return nil // Continue walking
	})

	if err != nil {
		// This error is from WalkDir itself, not the callback function returning an error
		return nil, fmt.Errorf("error walking directory %s: %w", dirPath, err)
	}
	return notePaths, nil
}

// readNoteFiles reads and parses multiple note files, skipping unparseable ones.
func (f *FileNoteRepository) readNoteFiles(filePaths []string) ([]model.ParsedNote, error) {
	results := make([]model.ParsedNote, 0, len(filePaths))
	// Consider using errgroup for concurrency later if needed
	for _, path := range filePaths {
		note, err := f.readNoteFile(path)
		if err != nil {
			// Log error and skip file (error already includes path)
			fmt.Fprintf(os.Stderr, "Warning: skipping note file due to error: %v\n", err)
			continue
		}
		results = append(results, note)
	}
	// Return successfully parsed notes, overall function doesn't error out for single file issues
	return results, nil
}

// readNoteFile reads and parses a single note file.
func (f *FileNoteRepository) readNoteFile(filePath string) (model.ParsedNote, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return model.ParsedNote{}, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	if len(bytes.TrimSpace(content)) == 0 {
		return model.ParsedNote{}, fmt.Errorf("file %s is empty", filePath)
	}

	// ParseNote now returns errors for missing frontmatter or readdeck-id
	parsedNote, err := f.parser.ParseNote(content, filePath)
	if err != nil {
		// Error context should be sufficient from ParseNote
		return model.ParsedNote{}, err
	}
	return parsedNote, nil
}
