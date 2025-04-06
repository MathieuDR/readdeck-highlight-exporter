package readdeck

import (
	"context"
	"time"
)

type Highlight struct {
	ID               string    `json:"id"`
	Href             string    `json:"href"`
	Text             string    `json:"text"`
	Created          time.Time `json:"created"`
	Color            string    `json:"color"`
	BookmarkID       string    `json:"bookmark_id"`
	BookmarkHref     string    `json:"bookmark_href"`
	BookmarkURL      string    `json:"bookmark_url"`
	BookmarkTitle    string    `json:"bookmark_title"`
	BookmarkSiteName string    `json:"bookmark_site_name"`
}

type Bookmark struct {
	Authors     []string  `json:"authors"`
	Created     time.Time `json:"created"`
	Description string    `json:"description"`
	Type        string    `json:"document_type"`
	Href        string    `json:"href"`
	ID          string    `json:"id"`
	Labels      []string  `json:"labels"`
	Published   time.Time `json:"published"`
	SiteUrl     string    `json:"url"`
	Title       string    `json:"title"`
}

type Client interface {
	GetHighlights(ctx context.Context, since *time.Time) ([]Highlight, error)
	GetBookmark(ctx context.Context, bookmarkId string) (Bookmark, error)
}
