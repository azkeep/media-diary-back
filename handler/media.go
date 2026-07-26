package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/azkeep/MediaDiary/backend-go/model"
	"github.com/azkeep/MediaDiary/backend-go/service"
)

type MediaHandler struct {
	svc service.MediaService
}

func NewMediaHandler(svc service.MediaService) *MediaHandler {
	return &MediaHandler{svc: svc}
}

// CORSMiddleware adds CORS headers to request
func CORSMiddleware(next http.Handler, allowedOrigin string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RegisterRoutes registers all API routes to the ServeMux
func (h *MediaHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/entries/all", h.GetAllEntries)
	mux.HandleFunc("GET /api/entries/date/{date}", h.GetEntriesByDate)
	mux.HandleFunc("GET /api/entries/{days}", h.GetEntriesLaterThan)
	mux.HandleFunc("POST /api/entries", h.AddEntry)
	mux.HandleFunc("PUT /api/entries/{entryId}", h.EditEntry)
	mux.HandleFunc("DELETE /api/entries/{entryId}", h.DeleteEntry)
	mux.HandleFunc("GET /api/stats", h.GetStats)
	mux.HandleFunc("GET /api/entries/ratings/{months}", h.GetTitlesRating)
}

func (h *MediaHandler) GetEntriesLaterThan(w http.ResponseWriter, r *http.Request) {
	daysStr := r.PathValue("days")
	days, err := strconv.ParseInt(daysStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid days parameter", http.StatusBadRequest)
		return
	}

	targetDate := time.Now().AddDate(0, 0, -int(days))
	ld := model.LocalDate(targetDate)

	result, err := h.svc.GetMediaLaterThan(ld)
	if err != nil {
		log.Printf("Error fetching entries: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, result)
}

func (h *MediaHandler) GetEntriesByDate(w http.ResponseWriter, r *http.Request) {
	dateStr := r.PathValue("date")
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, "invalid date format, must be YYYY-MM-DD", http.StatusBadRequest)
		return
	}
	ld := model.LocalDate(t)

	log.Printf("Fetching entries for date: %s", dateStr)
	result, err := h.svc.GetMediaByDate(ld)
	if err != nil {
		log.Printf("Error fetching entries: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, result)
}

func (h *MediaHandler) GetAllEntries(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.GetAllMedia()
	if err != nil {
		log.Printf("Error fetching entries: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, result)
}

func (h *MediaHandler) AddEntry(w http.ResponseWriter, r *http.Request) {
	var m model.MediaSelected
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Adding new entry: %s", m.Title)
	if err := h.svc.Save(&m); err != nil {
		log.Printf("Error saving entry: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err := json.NewEncoder(w).Encode(m)
	if err != nil {
		return
	}
}

func (h *MediaHandler) EditEntry(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("entryId")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid entryId parameter", http.StatusBadRequest)
		return
	}

	var m model.MediaSelected
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	m.ID = id

	log.Printf("Updating entry ID: %d", id)
	if err := h.svc.Update(&m); err != nil {
		log.Printf("Error updating entry: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(m)
	if err != nil {
		return
	}
}

func (h *MediaHandler) DeleteEntry(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("entryId")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid entryId parameter", http.StatusBadRequest)
		return
	}

	log.Printf("Deleting entry ID: %d", id)
	if err := h.svc.Delete(id); err != nil {
		log.Printf("Error deleting entry: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func sendJSON(w http.ResponseWriter, result []model.MediaSelected) {
	if len(result) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(result)
	if err != nil {
		return
	}
}

func sendJSONRating(w http.ResponseWriter, result []model.MediaRating) {
	if len(result) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(result)
	if err != nil {
		return
	}
}

func (h *MediaHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")
	if title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	stats, exists, err := h.svc.GetTitleStats(title)
	if err != nil {
		log.Printf("Error pulling metadata statistics for %s: %v", title, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "No data found for the given title", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		log.Printf("Failed encoding response: %v", err)
	}
}

func (h *MediaHandler) GetTitlesRating(w http.ResponseWriter, r *http.Request) {
	monStr := r.PathValue("months")

	months, err := strconv.Atoi(monStr)
	if err != nil || months < 0 {
		http.Error(w, "invalid months parameter", http.StatusBadRequest)
		return
	}

	result, err := h.svc.GetAllTitlesByRating(months)

	if err != nil {
		log.Printf("Error fetching titles: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSONRating(w, result)
}
