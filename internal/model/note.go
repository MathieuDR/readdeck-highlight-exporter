package model

import (
	"time"
)

type Note struct {
	ID           string
	Title        string
	Content      string
	Source       string
	BookmarkID   string
	CreatedAt    time.Time
	HighlightIDs []string
	Tags         []string
}
