package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

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

	sendJSON(w, result)
}

func (h *MediaHandler) SearchTitles(w http.ResponseWriter, r *http.Request) {
	searchTerm := r.PathValue("searchTerm")

	if len(searchTerm) == 0 {
		log.Printf("Empty search term")
		http.Error(w, "Empty search term", http.StatusBadRequest)
		return
	}

	result, err := h.svc.SearchEntries(searchTerm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, result)
}
