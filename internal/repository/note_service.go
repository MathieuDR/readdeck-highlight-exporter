package repository

import (
	"github.com/mathieudr/readdeck-highlight-exporter/internal/model"
)

type NoteService struct {
	Parser    NoteParser
	Generator NoteGenerator
	Updater   NoteUpdater
}

func NewNoteService() *NoteService {
	formatter := NewHighlightFormatter(DefaultColorConfig())
	parser := NewYAMLNoteParser()
	generator := NewYAMLNoteGenerator(formatter)
	updater := NewYAMLNoteUpdater(generator, parser)

	return &NoteService{
		Parser:    parser,
		Generator: generator,
		Updater:   updater,
	}
}

// Implement all the required interfaces
var _ NoteParser = (*NoteService)(nil)
var _ NoteGenerator = (*NoteService)(nil)
var _ NoteUpdater = (*NoteService)(nil)

func (s *NoteService) ParseNote(content []byte, path string) (model.ParsedNote, error) {
	return s.Parser.ParseNote(content, path)
}

func (s *NoteService) GenerateNoteContent(note model.Note) (NoteOperation, error) {
	return s.Generator.GenerateNoteContent(note)
}

func (s *NoteService) UpdateNoteContent(existing model.ParsedNote, note model.Note) (NoteOperation, error) {
	return s.Updater.UpdateNoteContent(existing, note)
}
