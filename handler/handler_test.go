package handler

import (
	"bytes"
	"encoding/json"
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
	entries        []model.MediaEntry
	stats          *model.StatsResponse
	statsExists    bool
	ratings        []model.MediaRating
	cursorResponse *model.CursorResponse
	err            error
}

func (m *mockService) GetAllMedia() ([]model.MediaEntry, error) {
	return m.entries, m.err
}
func (m *mockService) GetMediaPaginated(cursor string, limit int) (*model.CursorResponse, error) {
	return m.cursorResponse, m.err
}
func (m *mockService) GetMediaByDate(date model.LocalDate) ([]model.MediaEntry, error) {
	return m.entries, m.err
}
func (m *mockService) GetMediaForNDays(date model.LocalDate) ([]model.MediaEntry, error) {
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

func assertCursorResponse(t *testing.T,
	rec *httptest.ResponseRecorder,
	wantStatus int,
	wantCount int,
	wantHasMore bool) *model.CursorResponse {
	t.Helper()

	if rec.Code != wantStatus {
		t.Fatalf("got status code %d, want %d", rec.Code, wantStatus)
	}

	if wantStatus >= 400 {
		return nil
	}

	if contentType := rec.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("got content-type %q, want %q", contentType, "application/json")
	}

	var result model.CursorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}

	if len(result.Data) != wantCount {
		t.Errorf("got %d entries data slice, want %d", len(result.Data), wantCount)
	}

	if result.HasMore != wantHasMore {
		t.Errorf("got has_more %v, want %v", result.HasMore, wantHasMore)
	}

	return &result
}
