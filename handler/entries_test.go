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
		mockEntries        []model.MediaEntry
		expectedStatusCode int
		expectedCount      int
	}{
		{
			name:               "should return 200 OK and json list when media entries exist",
			mockEntries:        []model.MediaEntry{sampleEntry},
			expectedStatusCode: http.StatusOK,
			expectedCount:      1,
		},
		{
			name:               "should return 204 No Content when no media entries exist",
			mockEntries:        []model.MediaEntry{},
			expectedStatusCode: http.StatusNoContent,
			expectedCount:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := setupTestServer(&mockService{entries: tt.mockEntries})
			req := httptest.NewRequest(http.MethodGet, "/api/entries/all", nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			result := assertJSONSliceResponse[model.MediaEntry](t, rec, tt.expectedStatusCode, tt.expectedCount)

			if len(result) > 0 && result[0].Title != sampleEntry.Title {
				t.Errorf("got title %q, want %q", result[0].Title, sampleEntry.Title)
			}
		})
	}
}

func TestGetEntriesByDate(t *testing.T) {
	tests := []struct {
		name               string
		urlPath            string
		mockEntries        []model.MediaEntry
		mockErr            error
		expectedStatusCode int
		expectedCount      int
	}{
		{
			name:               "should return 200 OK and json entries when date format is valid and entries exist",
			urlPath:            "/api/entries/date/2026-03-15",
			mockEntries:        []model.MediaEntry{sampleEntry},
			mockErr:            nil,
			expectedStatusCode: http.StatusOK,
			expectedCount:      1,
		},
		{
			name:               "should return 204 No Content when date format is valid but no entries exist",
			urlPath:            "/api/entries/date/2026-03-15",
			mockEntries:        []model.MediaEntry{},
			mockErr:            nil,
			expectedStatusCode: http.StatusNoContent,
			expectedCount:      0,
		},
		{
			name:               "should return 400 Bad Request when date format is invalid",
			urlPath:            "/api/entries/date/15-03-2026",
			mockEntries:        nil,
			mockErr:            nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedCount:      0,
		},
		{
			name:               "should return 500 Internal Server Error when service fails",
			urlPath:            "/api/entries/date/2026-03-15",
			mockEntries:        nil,
			mockErr:            fmt.Errorf("db querry error"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedCount:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := setupTestServer(&mockService{
				entries: tt.mockEntries,
				err:     tt.mockErr,
			})
			req := httptest.NewRequest(http.MethodGet, tt.urlPath, nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			assertJSONSliceResponse[model.MediaEntry](t, rec, tt.expectedStatusCode, tt.expectedCount)
		})
	}
}

func TestGetEntriesLaterThan(t *testing.T) {
	tests := []struct {
		name               string
		daysParam          string
		mockEntries        []model.MediaEntry
		mockErr            error
		expectedStatusCode int
		expectedCount      int
	}{
		{
			name:               "should return 200 OK when days parameter is valid integer",
			daysParam:          "30",
			mockEntries:        []model.MediaEntry{sampleEntry},
			mockErr:            nil,
			expectedStatusCode: http.StatusOK,
			expectedCount:      1,
		},
		{
			name:               "should return 400 Bad Request when days parameter is not a number",
			daysParam:          "abc",
			mockEntries:        nil,
			mockErr:            nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedCount:      0,
		},
		{
			name:               "should return 500 Internal Server Error when service fails fetching entries",
			daysParam:          "7",
			mockEntries:        nil,
			mockErr:            fmt.Errorf("db querry error"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedCount:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := setupTestServer(&mockService{
				entries: tt.mockEntries,
				err:     tt.mockErr,
			})
			urlPath := fmt.Sprintf("/api/entries/%s", tt.daysParam)
			req := httptest.NewRequest(http.MethodGet, urlPath, nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			assertJSONSliceResponse[model.MediaEntry](t, rec, tt.expectedStatusCode, tt.expectedCount)
		})
	}
}
