package handler

import (
	"encoding/json"
	"fmt"
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
		log.Printf("Error fetching ratings: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, result)
}

func (h *MediaHandler) GetRatingsBetween(w http.ResponseWriter, r *http.Request) {
	sd, err := parseDateParam(r, "startDate")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fd, err := parseDateParam(r, "finishDate")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if fd.Time().Before(sd.Time()) {
		http.Error(w, fmt.Sprintf("finish date cannot be before start date"), http.StatusBadRequest)
		return
	}

	result, err := h.svc.GetRatingsBetween(sd, fd)

	if err != nil {
		log.Printf("Error fetching ratings between %v and %v: %v", sd, fd, err)
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

	filename, err := h.svc.ExportRatingsCSV(w, months)
	if err != nil {
		log.Printf("Error exporting ratings to CSV: %v", err)
		http.Error(w, "Failed to export CSV", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
}

func (h *MediaHandler) ExportRatingsCSVBetween(w http.ResponseWriter, r *http.Request) {
	sd, err := parseDateParam(r, "startDate")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fd, err := parseDateParam(r, "finishDate")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if fd.Time().Before(sd.Time()) {
		http.Error(w, fmt.Sprintf("finish date cannot be before start date"), http.StatusBadRequest)
		return
	}

	filename, err := h.svc.ExportRatingsCSVBetween(w, sd, fd)
	if err != nil {
		log.Printf("Error exporting ratings to CSV: %v", err)
		http.Error(w, "Failed to export CSV", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

}

func (h *MediaHandler) GetTimeline(w http.ResponseWriter, r *http.Request) {
	cursor, limit := parsePaginationParams(r)

	result, err := h.svc.GetTimeline(cursor, limit)
	if err != nil {
		log.Printf("Error fetching timeline: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, result)
}

func (h *MediaHandler) ExportTimelineCSV(w http.ResponseWriter, r *http.Request) {
	filename, err := h.svc.ExportTimelineCSV(w)
	if err != nil {
		log.Printf("Error exporting timeline CSV: %v", err)
		http.Error(w, "Failed to export CSV", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
}
