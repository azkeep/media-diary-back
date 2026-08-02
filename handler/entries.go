package handler

import (
	"encoding/json"
	"fmt"
	"github.com/azkeep/MediaDiary/backend-go/model"
	"log"
	"net/http"
	"strconv"
	"time"
)

func (h *MediaHandler) ListEntriesSince(w http.ResponseWriter, r *http.Request) {
	daysStr := r.PathValue("days")
	days, err := strconv.ParseInt(daysStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid days parameter", http.StatusBadRequest)
		return
	}

	cursor, limit := parsePaginationParams(r)

	targetDate := time.Now().AddDate(0, 0, -int(days))
	ld := model.LocalDate(targetDate)

	result, err := h.svc.ListSincePaginated(ld, cursor, limit)
	if err != nil {
		log.Printf("Error fetching entries for N days: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, result)
}

func (h *MediaHandler) ListEntriesByDate(w http.ResponseWriter, r *http.Request) {
	dateStr := r.PathValue("date")
	t, err := time.Parse(model.DateFormat, dateStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid date format, must be %s", model.DateExpected), http.StatusBadRequest)
		return
	}
	ld := model.LocalDate(t)

	log.Printf("Fetching entries for date: %s", dateStr)
	result, err := h.svc.ListByDate(ld)
	if err != nil {
		log.Printf("Error fetching entries: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, result)
}

func (h *MediaHandler) ListEntries(w http.ResponseWriter, r *http.Request) {
	cursor, limit := parsePaginationParams(r)

	result, err := h.svc.List(cursor, limit)
	if err != nil {
		log.Printf("Error fetching entries: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, result)
}

func (h *MediaHandler) SearchEntries(w http.ResponseWriter, r *http.Request) {
	searchTerm := r.PathValue("searchTerm")
	cursor, limit := parsePaginationParams(r)

	result, err := h.svc.Search(searchTerm, cursor, limit)
	if err != nil {
		log.Printf("Error searching entries: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, result)
}

func (h *MediaHandler) SaveEntries(w http.ResponseWriter, r *http.Request) {
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

func (h *MediaHandler) UpdateEntries(w http.ResponseWriter, r *http.Request) {
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
