package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/azkeep/MediaDiary/backend-go/model"
)

var (
	movieType   = "Movie"
	sampleEntry = model.MediaEntry{
		ID:         1,
		Title:      "Matrix",
		Date:       model.LocalDate(time.Now()),
		IsFinished: true,
		Type:       &movieType,
	}

	sampleStats = model.StatsResponse{
		Title: "Matrix",
		Total: 5,
	}

	sampleRating = model.MediaRating{
		Title:  "Matrix",
		Rating: 10,
	}
)

type mockService struct {
	entries     []model.MediaEntry
	stats       *model.StatsResponse
	statsExists bool
	ratings     []model.MediaRating
	err         error
}

func (m *mockService) GetAllMedia() ([]model.MediaEntry, error) {
	return m.entries, m.err
}
func (m *mockService) GetMediaByDate(date model.LocalDate) ([]model.MediaEntry, error) {
	return m.entries, m.err
}
func (m *mockService) GetMediaLaterThan(date model.LocalDate) ([]model.MediaEntry, error) {
	return m.entries, m.err
}
func (m *mockService) SaveBatch(entries []model.MediaEntry) error {
	return m.err
}
func (m *mockService) UpdateBatch(entries []model.MediaEntry) ([]model.MediaEntry, error) {
	return nil, m.err
}
func (m *mockService) DeleteBatch(ids []int64) error {
	return m.err
}
func (m *mockService) GetTitleStats(title string) (*model.StatsResponse, bool, error) {
	return m.stats, m.statsExists, m.err
}
func (m *mockService) GetAllTitlesByRating(months int) ([]model.MediaRating, error) {
	return m.ratings, m.err
}
func (m *mockService) ImportFromCSV(r io.Reader) error {
	return m.err
}
func (m *mockService) ExportRatingsToCSV(w io.Writer, months int) error {
	if m.err != nil {
		return m.err
	}
	_, err := w.Write([]byte("Title;Rating\nMatrix;10\n"))
	return err
}
func (m *mockService) SearchEntries(searchTerm string) ([]model.MediaEntry, error) {
	return m.entries, m.err
}

func setupTestServer(svc *mockService) *http.ServeMux {
	h := NewMediaHandler(svc)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	return mux
}

func createMultipartCSVRequest(filename string, fileContent string) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if filename != "" {
		part, err := writer.CreateFormFile("file", filename)
		if err != nil {
			return nil, err
		}
		_, err = part.Write([]byte(fileContent))
		if err != nil {
			return nil, err
		}
	}

	err := writer.Close()
	if err != nil {
		return nil, err
	}

	req := httptest.NewRequest(http.MethodPost, "/api/entries/import", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, nil
}

func assertJSONSliceResponse[T any](t *testing.T, rec *httptest.ResponseRecorder, wantStatus int, wantCount int) []T {
	t.Helper()

	if rec.Code != wantStatus {
		t.Errorf("got status code %d, want %d", rec.Code, wantStatus)
	}

	if wantStatus == http.StatusNoContent || wantStatus >= 400 {
		return nil
	}

	if contentType := rec.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("got content-type %q, want %q", contentType, "application/json")
	}

	var result []T
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}

	if len(result) != wantCount {
		t.Errorf("got %d entries in response, want %d", len(result), wantCount)
	}

	return result
}

func assertSingleJSONResponse[T any](t *testing.T, rec *httptest.ResponseRecorder, wantStatus int) *T {
	t.Helper()

	if rec.Code != wantStatus {
		t.Errorf("got status code %d, want %d", rec.Code, wantStatus)
	}

	if wantStatus >= 400 {
		return nil
	}

	if contentType := rec.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("got content-type %q, want %q", contentType, "application/json")
	}

	var result T
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}

	return &result
}

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

func TestImportCSV(t *testing.T) {
	tests := []struct {
		name               string
		filename           string
		content            string
		mockErr            error
		expectedStatusCode int
	}{
		{
			name:               "should return 200 OK when uploading valid CSV file",
			filename:           "data.csv",
			content:            "Date;Title,IsFinished\n01.01.2026;Matrix;true",
			mockErr:            nil,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "should return 400 Bad Request when file extension is not csv",
			filename:           "data.txt",
			content:            "some text content",
			mockErr:            nil,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should return 400 Bad Request when file form field is missing",
			filename:           "",
			content:            "",
			mockErr:            nil,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should return 400 Bad Request when service import fails",
			filename:           "data.csv",
			content:            "invalid csv content",
			mockErr:            fmt.Errorf("csv parse error"),
			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := setupTestServer(&mockService{err: tt.mockErr})

			req, err := createMultipartCSVRequest(tt.filename, tt.content)
			if err != nil {
				t.Fatalf("failed creating multipart request: %v", err)
			}
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			assertSingleJSONResponse[map[string]string](t, rec, tt.expectedStatusCode)
		})
	}
}

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

func TestSearchTitles(t *testing.T) {
	tests := []struct {
		name               string
		searchTerm         string
		mockEntries        []model.MediaEntry
		mockErr            error
		expectedStatusCode int
		expectedCount      int
	}{
		{
			name:               "should return 200 OK when search term yields results",
			searchTerm:         "matrix",
			mockEntries:        []model.MediaEntry{sampleEntry},
			mockErr:            nil,
			expectedStatusCode: http.StatusOK,
			expectedCount:      1,
		},
		{
			name:               "should return 500 Internal Server Error when service search fails",
			searchTerm:         "Matrix",
			mockEntries:        nil,
			mockErr:            fmt.Errorf("search index error"),
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
			urlPath := fmt.Sprintf("/api/entries/search/%s", tt.searchTerm)

			req := httptest.NewRequest(http.MethodGet, urlPath, nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			assertJSONSliceResponse[model.MediaEntry](t, rec, tt.expectedStatusCode, tt.expectedCount)
		})
	}
}
