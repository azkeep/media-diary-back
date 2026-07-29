package handler

import (
	"encoding/json"
	"github.com/azkeep/MediaDiary/backend-go/model"
	"log"
	"net/http"
	"strconv"
	"time"
)

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
	var m model.MediaEntry
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

	var m model.MediaEntry
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
