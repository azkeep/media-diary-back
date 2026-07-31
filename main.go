package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/azkeep/MediaDiary/backend-go/config"
	"github.com/azkeep/MediaDiary/backend-go/db"
	"github.com/azkeep/MediaDiary/backend-go/handler"
	"github.com/azkeep/MediaDiary/backend-go/repository"
	"github.com/azkeep/MediaDiary/backend-go/service"
)

func main() {
	log.Println("Starting Go Backend...")

	cfg := config.Load()

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

	repo := repository.NewMediaRepository(dbConn)
	svc := service.NewMediaService(repo)

	mediaHandler := handler.NewMediaHandler(svc)

	mux := http.NewServeMux()

	mediaHandler.RegisterRoutes(mux)
	handlerWithCORS := handler.CORSMiddleware(mux, cfg.AllowedOrigin)

	log.Printf("Server listening on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, handlerWithCORS); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
