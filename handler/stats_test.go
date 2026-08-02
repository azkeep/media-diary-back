package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/azkeep/MediaDiary/backend-go/model"
)

func TestGetStats(t *testing.T) {
	tests := []struct {
		name               string
		urlPath            string
		mockStats          *model.StatsResponse
		mockExists         bool
		mockErr            error
		expectedStatusCode int
	}{
		{
			name:               "should return 200 OK and stats json when query title exists",
			urlPath:            "/api/stats?title=Matrix",
			mockStats:          &sampleStats,
			mockExists:         true,
			mockErr:            nil,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "should return 400 Bad Request when title query parameter is missing",
			urlPath:            "/api/stats",
			mockStats:          nil,
			mockExists:         false,
			mockErr:            nil,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should return 404 Not Found when statistics do not exist for title",
			urlPath:            "/api/stats?title=NonExistent",
			mockStats:          nil,
			mockExists:         false,
			mockErr:            nil,
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "should return 500 Internal Server Error when service fails",
			urlPath:            "/api/stats?title=Matrix",
			mockStats:          nil,
			mockExists:         false,
			mockErr:            fmt.Errorf("db stats error"),
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockService{
				stats:       tt.mockStats,
				statsExists: tt.mockExists,
				err:         tt.mockErr,
			}
			mux := setupTestServer(svc)
			req := httptest.NewRequest(http.MethodGet, tt.urlPath, nil)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			res := assertSingleJSONResponse[model.StatsResponse](t, rec, tt.expectedStatusCode)
			if res != nil && res.Title != sampleStats.Title {
				t.Errorf("got stats title %q, want %q", res.Title, sampleStats.Title)
			}
		})
	}
}

func TestGetTitlesRating(t *testing.T) {
	tests := []struct {
		name               string
		monthsParam        string
		mockRating         []model.MediaRating
		mockErr            error
		expectedStatusCode int
		expectedCount      int
	}{
		{
			name:               "should return 200 OK when months param is valid non-negative int",
			monthsParam:        "6",
			mockRating:         []model.MediaRating{sampleRating},
			mockErr:            nil,
			expectedStatusCode: http.StatusOK,
			expectedCount:      1,
		},
		{
			name:               "should return 400 Bad Request when months param is negative",
			monthsParam:        "-3",
			mockRating:         []model.MediaRating{},
			mockErr:            nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedCount:      0,
		},
		{
			name:               "should return 400 Bad Request when months param is non-numeric",
			monthsParam:        "invalid",
			mockRating:         nil,
			mockErr:            nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedCount:      0,
		},
		{
			name:               "should return 500 Internal Server Error when service fails",
			monthsParam:        "12",
			mockRating:         nil,
			mockErr:            fmt.Errorf("db ratings error"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedCount:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := setupTestServer(&mockService{
				ratings: tt.mockRating,
				err:     tt.mockErr,
			})
			urlPath := fmt.Sprintf("/api/entries/ratings/%s", tt.monthsParam)

			req := httptest.NewRequest(http.MethodGet, urlPath, nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			assertJSONSliceResponse[model.MediaRating](t, rec, tt.expectedStatusCode, tt.expectedCount)
		})
	}
}

func TestExportTitlesRating(t *testing.T) {
	tests := []struct {
		name               string
		monthsParam        string
		mockErr            error
		expectedStatusCode int
		expectedHeader     string
	}{
		{
			name:               "should return 200 OK with csv content headers when months param is valid",
			monthsParam:        "12",
			mockErr:            nil,
			expectedStatusCode: http.StatusOK,
			expectedHeader:     "text/csv; charset=utf-8",
		},
		{
			name:               "should return 400 Bad Request when months param is invalid",
			monthsParam:        "abc",
			mockErr:            nil,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should return 500 Internal Server Error when CSV export fails",
			monthsParam:        "6",
			mockErr:            fmt.Errorf("export error"),
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := setupTestServer(&mockService{err: tt.mockErr})
			urlPath := fmt.Sprintf("/api/entries/ratings/%s/export", tt.monthsParam)

			req := httptest.NewRequest(http.MethodGet, urlPath, nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatusCode {
				t.Fatalf("got status code %d, want %d", rec.Code, tt.expectedStatusCode)
			}

			if tt.expectedStatusCode == http.StatusOK {
				if contentType := rec.Header().Get("Content-Type"); contentType != tt.expectedHeader {
					t.Errorf("got Content-Type %q, want %q", contentType, tt.expectedHeader)
				}
			}
		})
	}
}
