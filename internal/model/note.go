package model

import "github.com/mathieudr/readdeck-highlight-exporter/internal/readdeck"

type Note struct {
	Path       string
	Bookmark   readdeck.Bookmark
	Highlights []readdeck.Highlight
}
