package readdeck

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
)

type HttpClient struct {
	client   http.Client
	baseUrl  string
	token    string
	pageSize int
}

// LEARNING
// This is a compile-time check to make sure it implements the interface
var _ Client = (*HttpClient)(nil)

type highlightsCall struct {
	Highlights  []Highlight
	CurrentPage int
	PageSize    int
	TotalPages  int
}

func NewHttpClient(client http.Client, baseUrl, authToken string, pagesize int) *HttpClient {
	if pagesize < 10 {
		pagesize = 10
	}

	if pagesize > 100 {
		pagesize = 100
	}

	return &HttpClient{
		client:   client,
		baseUrl:  baseUrl,
		token:    authToken,
		pageSize: pagesize,
	}
}

func (c HttpClient) GetHighlights(ctx context.Context) ([]Highlight, error) {
	var highlights []Highlight
	totalPages := 1
	for i := 0; i < totalPages; i++ {
		log.Printf("Requesting page %d, offset: %d", (i + 1), i*c.pageSize)

		call, err := c.doHighlightCall(ctx, c.pageSize, i*c.pageSize)

		if err != nil {
			return nil, fmt.Errorf("Error while fetching highlights: %w", err)
		}

		totalPages = call.TotalPages
		highlights = append(highlights, call.Highlights...)
	}

	return c.reverseList(highlights), nil
}

func (c HttpClient) reverseList(list []Highlight) []Highlight {
	result := make([]Highlight, len(list))
	for i, j := 0, len(list)-1; j >= 0; i, j = i+1, j-1 {
		result[i] = list[j]
	}
	return result
}

func (c HttpClient) doHighlightCall(ctx context.Context, limit int, offset int) (highlightsCall, error) {
	request, err := c.createHighlightsRequest(ctx, limit, offset)
	if err != nil {
		return highlightsCall{}, err
	}

	resp, err := c.client.Do(request)
	if err != nil {
		return highlightsCall{}, fmt.Errorf("HTTP Request failed: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return highlightsCall{}, fmt.Errorf("Non success status code: %d", resp.StatusCode)
	}

	return parseHighlightResponse(resp, limit, offset)
}

func (c HttpClient) createHighlightsRequest(ctx context.Context, limit, offset int) (*http.Request, error) {
	endpoint := fmt.Sprintf("%s/api/bookmarks/annotations", c.baseUrl)
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("Could not create request: %w", err)
	}

	queries := req.URL.Query()
	queries.Add("limit", strconv.Itoa(limit))
	queries.Add("offset", strconv.Itoa(offset))
	req.URL.RawQuery = queries.Encode()

	c.addCommonHeaders(req)

	return req, nil
}

func parseHighlightResponse(response *http.Response, limit, offset int) (highlightsCall, error) {
	defer response.Body.Close()
	jsonBytes, err := io.ReadAll(response.Body)

	var result []Highlight
	// Marchalling is converting data to be 'transmitted'
	// Unmarshalling is converting data that has been encoded for data transmission back into useable data structures
	// in our case JSON -> Go Struct
	// It's very similar to decoding, if not the same but the nuance is that to unmarshal you need the whole binary data
	// vs decoding (in go) can be done in a stream
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return highlightsCall{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	currentPage, totalPages, err := extractPaginationInfo(response, limit, offset)
	if err != nil {
		return highlightsCall{}, err
	}

	return highlightsCall{
		PageSize:    limit,
		CurrentPage: currentPage,
		TotalPages:  totalPages,
		Highlights:  result,
	}, nil
}

func extractPaginationInfo(resp *http.Response, limit, offset int) (int, int, error) {
	tmp := resp.Header.Get("total-pages")
	totalPages, err := strconv.Atoi(tmp)
	if err != nil {
		return 0, 0, fmt.Errorf("could not determine total pages: %s", tmp)
	}

	tmp = resp.Header.Get("current-page")
	currentPage, err := strconv.Atoi(tmp)
	if err != nil {
		currentPage = (offset / limit) + 1
	}

	return currentPage, totalPages, nil
}

func (c HttpClient) GetBookmark(ctx context.Context, bookmarkId string) (Bookmark, error) {
	request, err := c.createBookmarkRequest(ctx, bookmarkId)

	if err != nil {
		return Bookmark{}, fmt.Errorf("Could not create request: %w", err)
	}

	resp, err := c.client.Do(request)

	if err != nil {
		return Bookmark{}, fmt.Errorf("HTTP Request failed: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Bookmark{}, fmt.Errorf("Non success status code: %d", resp.StatusCode)
	}

	return parseBookmark(resp)
}

func (c HttpClient) createBookmarkRequest(ctx context.Context, id string) (*http.Request, error) {
	endpoint := fmt.Sprintf("%s/api/bookmarks/%s", c.baseUrl, id)
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)

	if err != nil {
		return nil, fmt.Errorf("Could not create request: %w", err)
	}

	c.addCommonHeaders(req)

	return req, nil
}

func (c HttpClient) addCommonHeaders(req *http.Request) *http.Request {
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.token))

	return req
}

func parseBookmark(resp *http.Response) (Bookmark, error) {
	defer resp.Body.Close()

	var result Bookmark

	jsonBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return Bookmark{}, fmt.Errorf("Failed to read bytes: %w", err)
	}

	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return Bookmark{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result, nil
}
