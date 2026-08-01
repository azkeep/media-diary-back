package handler

import (
	"encoding/json"
	"github.com/azkeep/MediaDiary/backend-go/model"
	"log"
	"net/http"
	"strconv"
	"time"
)

const defaultPageLimit = 50

func (h *MediaHandler) GetEntriesForNDays(w http.ResponseWriter, r *http.Request) {
	daysStr := r.PathValue("days")
	days, err := strconv.ParseInt(daysStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid days parameter", http.StatusBadRequest)
		return
	}

	targetDate := time.Now().AddDate(0, 0, -int(days))
	ld := model.LocalDate(targetDate)

	result, err := h.svc.GetMediaForNDays(ld)
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
	cursor := r.URL.Query().Get("cursor")
	limitStr := r.URL.Query().Get("limit")

	limit := defaultPageLimit
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	result, err := h.svc.GetMediaPaginated(cursor, limit)
	if err != nil {
		log.Printf("Error fetching entries: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, result)
}

func (h *MediaHandler) SearchTitles(w http.ResponseWriter, r *http.Request) {
	searchTerm := r.PathValue("searchTerm")
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

	result, err := h.svc.SearchEntriesPaginated(searchTerm, cursor, limit)
	if err != nil {
		log.Printf("Error searching entries: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, result)
}

func (h *MediaHandler) AddEntries(w http.ResponseWriter, r *http.Request) {
	var entries []model.MediaEntry
	if err := json.NewDecoder(r.Body).Decode(&entries); err != nil {
		http.Error(w, "invalid request body: expected array of entries", http.StatusBadRequest)
		return
	}

	log.Printf("Adding %d entries...", len(entries))

	if err := h.svc.SaveBatch(entries); err != nil {
		log.Printf("Error saving batch: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(entries); err != nil {
		log.Printf("Error encoding response: %v", err)
		return
	}

	log.Printf("%d entries added successfully", len(entries))
}

func (h *MediaHandler) EditEntries(w http.ResponseWriter, r *http.Request) {
	var entries []model.MediaEntry
	if err := json.NewDecoder(r.Body).Decode(&entries); err != nil {
		http.Error(w, "invalid request body: expected array of entries", http.StatusBadRequest)
		return
	}

	log.Printf("Updating %d entries...", len(entries))

	updatedEntries, err := h.svc.UpdateBatch(entries)
	if err != nil {
		log.Printf("Error updating entries batch: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(updatedEntries); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

func (h *MediaHandler) DeleteEntries(w http.ResponseWriter, r *http.Request) {
	var ids []int64
	if err := json.NewDecoder(r.Body).Decode(&ids); err != nil {
		http.Error(w, "invalid request body: expected array of int64 IDs", http.StatusBadRequest)
		return
	}

	log.Printf("Deleting %d entries...", len(ids))

	if err := h.svc.DeleteBatch(ids); err != nil {
		log.Printf("Error deleting batch: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
