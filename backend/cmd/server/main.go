package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"sc2-analytics/internal/api"
	"sc2-analytics/internal/repository"
)

func main() {
	// Konfiguration via Flags
	port := flag.Int("port", 8080, "Server Port")
	dbPath := flag.String("db", "./data/sc2analytics.db", "Pfad zur SQLite Datenbank")
	uploadDir := flag.String("uploads", "./data/uploads", "Upload-Verzeichnis")
	staticDir := flag.String("static", "./static", "Verzeichnis für statische Dateien (Frontend)")
	flag.Parse()

	// Stelle sicher, dass Verzeichnisse existieren
	if err := os.MkdirAll(filepath.Dir(*dbPath), 0755); err != nil {
		log.Fatalf("Konnte Datenbank-Verzeichnis nicht erstellen: %v", err)
	}
	if err := os.MkdirAll(*uploadDir, 0755); err != nil {
		log.Fatalf("Konnte Upload-Verzeichnis nicht erstellen: %v", err)
	}

	// Initialisiere Repository
	repo, err := repository.New(*dbPath)
	if err != nil {
		log.Fatalf("Konnte Datenbank nicht initialisieren: %v", err)
	}
	defer repo.Close()

	// Erstelle Handler und Router
	handler := api.NewHandler(repo, *uploadDir)
	router := api.NewRouter(handler, repo)

	// Statische Dateien servieren (für Production)
	if _, err := os.Stat(*staticDir); err == nil {
		log.Printf("Serving static files from %s", *staticDir)
		fileServer := http.FileServer(http.Dir(*staticDir))
		router.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// API-Requests durchlassen
			if strings.HasPrefix(r.URL.Path, "/api/") {
				http.NotFound(w, r)
				return
			}
			// Prüfe ob Datei existiert
			path := filepath.Join(*staticDir, r.URL.Path)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				// SPA Fallback: index.html für alle nicht-existierenden Pfade
				http.ServeFile(w, r, filepath.Join(*staticDir, "index.html"))
				return
			}
			fileServer.ServeHTTP(w, r)
		}))
	}

	// Starte Server
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("SC2 Analytics Server startet auf http://localhost%s", addr)
	log.Printf("Environment: %s", getEnv("GO_ENV", "development"))

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server-Fehler: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
