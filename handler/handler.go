package handler

import (
	"encoding/json"
	"fmt"
	"github.com/azkeep/MediaDiary/backend-go/model"
	"github.com/azkeep/MediaDiary/backend-go/service"
	"log"
	"net/http"
	"strconv"
	"time"
)

const defaultPageLimit = 50

type MediaHandler struct {
	svc service.MediaService
}

func NewMediaHandler(svc service.MediaService) *MediaHandler {
	return &MediaHandler{svc: svc}
}

func (h *MediaHandler) RegisterRoutes(mux *http.ServeMux) {
	// Entries CRUD & Query operations
	mux.HandleFunc("GET /api/entries/all", h.ListEntries)
	mux.HandleFunc("GET /api/entries/date/{date}", h.ListEntriesByDate)
	mux.HandleFunc("GET /api/entries/{days}", h.ListEntriesSince)
	mux.HandleFunc("GET /api/entries/{startDate}/{finishDate}", h.ListEntriesBetween)
	mux.HandleFunc("GET /api/entries/search/{searchTerm}", h.SearchEntries)
	mux.HandleFunc("POST /api/entries", h.SaveEntries)
	mux.HandleFunc("PUT /api/entries", h.UpdateEntries)
	mux.HandleFunc("DELETE /api/entries", h.DeleteEntries)

	// Analytics & Report
	mux.HandleFunc("GET /api/stats", h.GetStats)
	mux.HandleFunc("GET /api/entries/ratings/{months}", h.GetRatings)
	mux.HandleFunc("GET /api/entries/ratings/{months}/export", h.ExportRatingsCSV)
	mux.HandleFunc("GET /api/entries/ratings/{startDate}/{finishDate}", h.GetRatingsBetween)
	mux.HandleFunc("GET /api/entries/ratings/{startDate}/{finishDate}/export", h.ExportRatingsCSVBetween)

	// Bulk Operations
	mux.HandleFunc("POST /api/entries/import", h.ImportCSV)
}

func sendJSON[T any](w http.ResponseWriter, result T) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("Failed encoding response: %v", err)
		return
	}
}

// parsePaginationParams extracts and validates cursor and limit query parameters from the HTTP request.
// If defaultLimit > 0, invalid or missing limit parameters will fall back to defaultLimit.
func parsePaginationParams(r *http.Request) (string, int) {
	cursor := r.URL.Query().Get("cursor")
	limitStr := r.URL.Query().Get("limit")

	limit := 0
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil {
			limit = parsed
		}
	} else if cursor != "" {
		limit = defaultPageLimit
	}

	if limit == 0 && limitStr == "" {
		limit = defaultPageLimit
	}

	return cursor, limit
}

func parseDateParam(r *http.Request, paramName string) (model.LocalDate, error) {
	dateStr := r.PathValue(paramName)
	if dateStr == "" {
		return model.LocalDate{}, fmt.Errorf("missing path parameter %q", paramName)
	}

	t, err := time.Parse(model.DateFormat, dateStr)
	if err != nil {
		return model.LocalDate{}, fmt.Errorf("invalid date format for %q, expected %s", paramName, model.DateExpected)
	}

	return model.LocalDate(t), nil
}
