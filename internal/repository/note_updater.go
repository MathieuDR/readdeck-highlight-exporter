package repository

import (
	"github.com/mathieudr/readdeck-highlight-exporter/internal/model"
)

// YAMLNoteUpdater handles updating existing notes with YAML frontmatter
type YAMLNoteUpdater struct {
	Generator *YAMLNoteGenerator
	Parser    *YAMLNoteParser
}

func NewYAMLNoteUpdater(generator *YAMLNoteGenerator, parser *YAMLNoteParser) *YAMLNoteUpdater {
	return &YAMLNoteUpdater{
		Generator: generator,
		Parser:    parser,
	}
}

func (u *YAMLNoteUpdater) UpdateNoteContent(existing model.ParsedNote, note model.Note) (NoteOperation, error) {
	return NoteOperation{}, nil
}
