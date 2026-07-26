package main

import (
	"database/sql"
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"time"

	"github.com/azkeep/MediaDiary/backend-go/config"
	"github.com/azkeep/MediaDiary/backend-go/db"
	"github.com/azkeep/MediaDiary/backend-go/handler"
	"github.com/azkeep/MediaDiary/backend-go/model"
	"github.com/azkeep/MediaDiary/backend-go/repository"
	"github.com/azkeep/MediaDiary/backend-go/service"
)

//go:embed templates/* static/*
var resourcesFS embed.FS

func main() {
	log.Println("Starting Go Backend...")

	// 1. Load configuration
	cfg := config.Load()

	// 2. Connect to database
	dbConn, err := db.Connect(cfg)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer func(dbConn *sql.DB) {
		err := dbConn.Close()
		if err != nil {

		}
	}(dbConn)
	log.Println("Database connection established.")

	// 3. Initialize layers
	repo := repository.NewMediaRepository(dbConn)
	svc := service.NewMediaService(repo)

	// 4. Parse templates from embedded filesystem
	tmpl, err := template.New("view-recent.html").Funcs(template.FuncMap{
		"formatDate": func(ld model.LocalDate) string {
			return time.Time(ld).Format("2006-01-02")
		},
		"derefString": func(s *string) string {
			if s == nil {
				return ""
			}
			return *s
		},
	}).ParseFS(resourcesFS, "templates/view-recent.html")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}

	// 5. Initialize handlers
	mediaHandler := handler.NewMediaHandler(svc)
	uiHandler := handler.NewUIHandler(svc, tmpl)

	// 6. Setup router
	mux := http.NewServeMux()

	// Register API and UI routes
	mediaHandler.RegisterRoutes(mux)
	uiHandler.RegisterRoutes(mux)

	// Serve static files from embedded FS
	staticFS, err := fs.Sub(resourcesFS, "static")
	if err != nil {
		log.Fatalf("Failed to get static sub filesystem: %v", err)
	}

	// Map static files to /css/, /js/ and /static/
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	mux.Handle("GET /css/", http.StripPrefix("/css/", http.FileServer(http.FS(subDirFS(staticFS, "css")))))
	mux.Handle("GET /js/", http.StripPrefix("/js/", http.FileServer(http.FS(subDirFS(staticFS, "js")))))

	// Apply CORS middleware
	handlerWithCORS := handler.CORSMiddleware(mux, cfg.AllowedOrigin)

	// 7. Start HTTP server
	log.Printf("Server listening on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, handlerWithCORS); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// Helper function to safely get a sub filesystem, returning the parent if subdirectory doesn't exist
func subDirFS(parent fs.FS, dir string) fs.FS {
	sub, err := fs.Sub(parent, dir)
	if err != nil {
		return parent
	}
	return sub
}
