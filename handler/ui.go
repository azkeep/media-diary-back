package handler

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/azkeep/MediaDiary/backend-go/model"
	"github.com/azkeep/MediaDiary/backend-go/service"
)

const (
	days = 314
)

type UIHandler struct {
	svc  service.MediaService
	tmpl *template.Template
}

func NewUIHandler(svc service.MediaService, tmpl *template.Template) *UIHandler {
	return &UIHandler{
		svc:  svc,
		tmpl: tmpl,
	}
}

func (h *UIHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /ui/recent", h.ShowRecent)
}

func (h *UIHandler) ShowRecent(w http.ResponseWriter, r *http.Request) {
	daysAgo := time.Now().AddDate(0, 0, -days)
	ld := model.LocalDate(daysAgo)

	list, err := h.svc.GetMediaLaterThan(ld)
	if err != nil {
		log.Printf("UI Error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Recent": list,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.Execute(w, data); err != nil {
		log.Printf("Template rendering error: %v", err)
	}
}
