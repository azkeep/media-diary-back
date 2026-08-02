package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/azkeep/MediaDiary/backend-go/model"
)

func TestGetAllEntries(t *testing.T) {
	tests := []struct {
		name               string
		mockCursorResp     *model.CursorResponse
		mockErr            error
		expectedStatusCode int
		expectedCount      int
		expectedHasMore    bool
	}{
		{
			name: "should return 200 OK and paginated response when entries exist",
			mockCursorResp: &model.CursorResponse{
				Data:       []model.MediaEntry{sampleEntry},
				NextCursor: "encoded_cursor_string",
				HasMore:    true,
			},
			mockErr:            nil,
			expectedStatusCode: http.StatusOK,
			expectedCount:      1,
			expectedHasMore:    true,
		},
		{
			name: "should return 200 OK with empty data array when no media entries exist",
			mockCursorResp: &model.CursorResponse{
				Data:       []model.MediaEntry{},
				NextCursor: "",
				HasMore:    false,
			},
			mockErr:            nil,
			expectedStatusCode: http.StatusOK,
			expectedCount:      0,
			expectedHasMore:    false,
		},
		{
			name:               "should return 500 Internal Server Error when service fails",
			mockCursorResp:     nil,
			mockErr:            fmt.Errorf("database query error"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedCount:      0,
			expectedHasMore:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := setupTestServer(&mockService{
				cursorResponse: tt.mockCursorResp,
				err:            tt.mockErr,
			})
			req := httptest.NewRequest(http.MethodGet, "/api/entries/all", nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			resp := assertCursorResponse(t,
				rec,
				tt.expectedStatusCode,
				tt.expectedCount,
				tt.expectedHasMore)

			if resp != nil && len(resp.Data) > 0 && resp.Data[0].Title != sampleEntry.Title {
				t.Errorf("got title %q, want %q", resp.Data[0].Title, sampleEntry.Title)
			}
		})
	}
}

func TestGetEntriesByDate(t *testing.T) {
	tests := []struct {
		name               string
		urlPath            string
		mockCursorResp     *model.CursorResponse
		mockErr            error
		expectedStatusCode int
		expectedCount      int
		expectedHasMore    bool
	}{
		{
			name:    "should return 200 OK and json entries when date format is valid and entries exist",
			urlPath: "/api/entries/date/2026-03-15",
			mockCursorResp: &model.CursorResponse{
				Data:       []model.MediaEntry{sampleEntry},
				NextCursor: "",
				HasMore:    false,
			},
			mockErr:            nil,
			expectedStatusCode: http.StatusOK,
			expectedCount:      1,
			expectedHasMore:    false,
		},
		{
			name:    "should return 200 OK with empty slice when date format is valid but no entries exist",
			urlPath: "/api/entries/date/2026-03-15",
			mockCursorResp: &model.CursorResponse{
				Data:       []model.MediaEntry{},
				NextCursor: "",
				HasMore:    false,
			},
			mockErr:            nil,
			expectedStatusCode: http.StatusOK,
			expectedCount:      0,
			expectedHasMore:    false,
		},
		{
			name:               "should return 400 Bad Request when date format is invalid",
			urlPath:            "/api/entries/date/15-03-2026",
			mockCursorResp:     nil,
			mockErr:            nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedCount:      0,
			expectedHasMore:    false,
		},
		{
			name:               "should return 500 Internal Server Error when service fails",
			urlPath:            "/api/entries/date/2026-03-15",
			mockCursorResp:     nil,
			mockErr:            fmt.Errorf("db querry error"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedCount:      0,
			expectedHasMore:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := setupTestServer(&mockService{
				cursorResponse: tt.mockCursorResp,
				err:            tt.mockErr,
			})
			req := httptest.NewRequest(http.MethodGet, tt.urlPath, nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			resp := assertCursorResponse(t,
				rec,
				tt.expectedStatusCode,
				tt.expectedCount,
				tt.expectedHasMore)

			if resp != nil && len(resp.Data) > 0 && resp.Data[0].Title != sampleEntry.Title {
				t.Errorf("got title %q, want %q", resp.Data[0].Title, sampleEntry.Title)
			}
		})
	}
}

func TestGetEntriesLaterThan(t *testing.T) {
	tests := []struct {
		name               string
		daysParam          string
		mockCursorResp     *model.CursorResponse
		mockErr            error
		expectedStatusCode int
		expectedCount      int
		expectedHasMore    bool
	}{
		{
			name:      "should return 200 OK when days parameter is valid integer",
			daysParam: "30",
			mockCursorResp: &model.CursorResponse{
				Data:       []model.MediaEntry{sampleEntry},
				NextCursor: "encoded_cursor_string",
				HasMore:    true,
			},
			mockErr:            nil,
			expectedStatusCode: http.StatusOK,
			expectedCount:      1,
			expectedHasMore:    true,
		},
		{
			name:               "should return 400 Bad Request when days parameter is not a number",
			daysParam:          "abc",
			mockCursorResp:     nil,
			mockErr:            nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedCount:      0,
			expectedHasMore:    false,
		},
		{
			name:               "should return 500 Internal Server Error when service fails fetching entries",
			daysParam:          "7",
			mockCursorResp:     nil,
			mockErr:            fmt.Errorf("db query error"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedCount:      0,
			expectedHasMore:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := setupTestServer(&mockService{
				cursorResponse: tt.mockCursorResp,
				err:            tt.mockErr,
			})
			urlPath := fmt.Sprintf("/api/entries/%s", tt.daysParam)
			req := httptest.NewRequest(http.MethodGet, urlPath, nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			resp := assertCursorResponse(t,
				rec,
				tt.expectedStatusCode,
				tt.expectedCount,
				tt.expectedHasMore)

			if resp != nil && len(resp.Data) > 0 && resp.Data[0].Title != sampleEntry.Title {
				t.Errorf("got title %q, want %q", resp.Data[0].Title, sampleEntry.Title)
			}
		})
	}
}

func TestSearchTitles(t *testing.T) {
	tests := []struct {
		name               string
		searchTerm         string
		mockCursorResp     *model.CursorResponse
		mockErr            error
		expectedStatusCode int
		expectedCount      int
		expectedHasMore    bool
	}{
		{
			name:       "should return 200 OK when search term yields results",
			searchTerm: "matrix",
			mockCursorResp: &model.CursorResponse{
				Data:       []model.MediaEntry{sampleEntry},
				NextCursor: "encoded_cursor_string",
				HasMore:    true,
			},
			mockErr:            nil,
			expectedStatusCode: http.StatusOK,
			expectedCount:      1,
			expectedHasMore:    true,
		},
		{
			name:               "should return 500 Internal Server Error when service search fails",
			searchTerm:         "Matrix",
			mockCursorResp:     nil,
			mockErr:            fmt.Errorf("search index error"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedCount:      0,
			expectedHasMore:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := setupTestServer(&mockService{
				cursorResponse: tt.mockCursorResp,
				err:            tt.mockErr,
			})
			urlPath := fmt.Sprintf("/api/entries/search/%s", tt.searchTerm)

			req := httptest.NewRequest(http.MethodGet, urlPath, nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			resp := assertCursorResponse(t,
				rec,
				tt.expectedStatusCode,
				tt.expectedCount,
				tt.expectedHasMore)

			if resp != nil && len(resp.Data) > 0 && resp.Data[0].Title != sampleEntry.Title {
				t.Errorf("got title %q, want %q", resp.Data[0].Title, sampleEntry.Title)
			}
		})
	}
}
