package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

func (h *MediaHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")
	if title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	stats, exists, err := h.svc.GetStats(title)
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

func (h *MediaHandler) GetRatings(w http.ResponseWriter, r *http.Request) {
	monStr := r.PathValue("months")

	months, err := strconv.Atoi(monStr)
	if err != nil || months < 0 {
		http.Error(w, "invalid months parameter", http.StatusBadRequest)
		return
	}

	result, err := h.svc.GetRatings(months)

	if err != nil {
		log.Printf("Error fetching titles: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, result)
}

func (h *MediaHandler) ExportRatingsCSV(w http.ResponseWriter, r *http.Request) {
	monStr := r.PathValue("months")

	months, err := strconv.Atoi(monStr)
	if err != nil || months < 0 {
		http.Error(w, "invalid months parameter", http.StatusBadRequest)
		return
	}

	filename := fmt.Sprintf("%s-ratings-last-%d-months.csv", time.Now().Format("20060102150405"), months)
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	if err := h.svc.ExportRatingsCSV(w, months); err != nil {
		log.Printf("Error exporting ratings to CSV: %v", err)
		http.Error(w, "Failed to export CSV", http.StatusInternalServerError)
		return
	}
}
