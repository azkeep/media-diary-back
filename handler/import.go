package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func (h *MediaHandler) ImportCSV(w http.ResponseWriter, r *http.Request) {
	// Restrict payload size to 10MB
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Missing or invalid 'file' field", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if !strings.HasSuffix(
		strings.ToLower(header.Filename), ".csv") {
		http.Error(w, "Uploaded file must have a .csv extension", http.StatusBadRequest)
		return
	}

	log.Printf("Starting CSV import process for file: %s", header.Filename)

	if err := h.svc.ImportCSV(file); err != nil {
		log.Printf("CSV import failed: %v", err)
		http.Error(w, fmt.Sprintf("Import failed: %v", err), http.StatusBadRequest)
		return
	}

	log.Printf("CSV data successfully validated and imported.")

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "CSV data successfully validated and imported.",
	})
}
