package handler

import (
	"encoding/json"
	"github.com/azkeep/MediaDiary/backend-go/service"
	"log"
	"net/http"
)

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
	mux.HandleFunc("GET /api/entries/{days}", h.GetEntriesLaterThan)
	mux.HandleFunc("POST /api/entries", h.AddEntries)
	mux.HandleFunc("PUT /api/entries", h.EditEntries)
	mux.HandleFunc("DELETE /api/entries", h.DeleteEntries)

	// Analytics & Search
	mux.HandleFunc("GET /api/stats", h.GetStats)
	mux.HandleFunc("GET /api/entries/ratings/{months}", h.GetTitlesRating)
	mux.HandleFunc("GET /api/entries/ratings/{months}/export", h.ExportTitlesRating)
	mux.HandleFunc("GET /api/entries/search/{searchTerm}", h.SearchTitles)

	// Bulk Operations
	mux.HandleFunc("POST /api/entries/import", h.ImportCSV)
}

func sendJSON[T any](w http.ResponseWriter, result []T) {
	if len(result) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(result)
	if err != nil {
		log.Printf("Failed encoding response: %v", err)
		return
	}
}
