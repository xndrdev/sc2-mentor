package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"sc2-analytics/internal/api"
	"sc2-analytics/internal/repository"
)

func main() {
	// Konfiguration via Flags
	port := flag.Int("port", 8080, "Server Port")
	dbPath := flag.String("db", "./data/sc2analytics.db", "Pfad zur SQLite Datenbank")
	uploadDir := flag.String("uploads", "./data/uploads", "Upload-Verzeichnis")
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

	// Starte Server
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("SC2 Analytics Server startet auf http://localhost%s", addr)
	log.Printf("API-Dokumentation: http://localhost%s/api/v1/health", addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server-Fehler: %v", err)
	}
}
