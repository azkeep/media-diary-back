package handler

import (
	"encoding/json"
	"github.com/azkeep/MediaDiary/backend-go/service"
	"log"
	"net/http"
	"strconv"
)

const defaultPageLimit = 50

type MediaHandler struct {
	svc service.MediaService
}

func NewMediaHandler(svc service.MediaService) *MediaHandler {
	return &MediaHandler{svc: svc}
}

func (h *MediaHandler) RegisterRoutes(mux *http.ServeMux) {
	// Entries
	mux.HandleFunc("GET /api/entries/all", h.GetAllEntries)
	mux.HandleFunc("GET /api/entries/date/{date}", h.GetEntriesByDate)
	mux.HandleFunc("GET /api/entries/{days}", h.GetEntriesForNDays)
	mux.HandleFunc("GET /api/entries/search/{searchTerm}", h.SearchEntries)
	mux.HandleFunc("POST /api/entries", h.AddEntries)
	mux.HandleFunc("PUT /api/entries", h.EditEntries)
	mux.HandleFunc("DELETE /api/entries", h.DeleteEntries)

	// Analytics & Search
	mux.HandleFunc("GET /api/stats", h.GetStats)
	mux.HandleFunc("GET /api/entries/ratings/{months}", h.GetTitlesRating)
	mux.HandleFunc("GET /api/entries/ratings/{months}/export", h.GetTitlesRatingCSV)

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
