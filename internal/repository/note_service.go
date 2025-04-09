// internal/repository/note_service.go
package repository

import (
	"github.com/mathieudr/readdeck-highlight-exporter/internal/model"
)

type NoteService interface {
	ParseNote(content []byte, path string) (model.ParsedNote, error)
	GenerateNoteContent(note model.Note) (NoteOperation, error)
	UpdateNoteContent(existing model.ParsedNote, note model.Note) (NoteOperation, error)
}

type ComprehensiveNoteService struct {
	Parser    NoteParser
	Generator NoteGenerator
	Updater   NoteUpdater
}

func NewNoteService(baseUrl string) NoteService {
	formatter := NewHighlightFormatter(DefaultColorConfig())
	parser := NewYAMLNoteParser()
	generator := NewYAMLNoteGenerator(formatter, baseUrl)
	updater := NewYAMLNoteUpdater(generator, parser)

	return &ComprehensiveNoteService{
		Parser:    parser,
		Generator: generator,
		Updater:   updater,
	}
}

func NewCustomNoteService(parser NoteParser, generator NoteGenerator, updater NoteUpdater) NoteService {
	return &ComprehensiveNoteService{
		Parser:    parser,
		Generator: generator,
		Updater:   updater,
	}
}

var _ NoteService = (*ComprehensiveNoteService)(nil)
var _ NoteParser = (*ComprehensiveNoteService)(nil)
var _ NoteGenerator = (*ComprehensiveNoteService)(nil)
var _ NoteUpdater = (*ComprehensiveNoteService)(nil)

func (s *ComprehensiveNoteService) ParseNote(content []byte, path string) (model.ParsedNote, error) {
	return s.Parser.ParseNote(content, path)
}

func (s *ComprehensiveNoteService) GenerateNoteContent(note model.Note) (NoteOperation, error) {
	return s.Generator.GenerateNoteContent(note)
}

func (s *ComprehensiveNoteService) UpdateNoteContent(existing model.ParsedNote, note model.Note) (NoteOperation, error) {
	return s.Updater.UpdateNoteContent(existing, note)
}
