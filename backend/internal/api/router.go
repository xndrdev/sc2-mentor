package api

import (
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"sc2-analytics/internal/repository"
)

// NewRouter erstellt den API-Router
func NewRouter(handler *Handler, repo *repository.Repository) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)

	// CORS Konfiguration aus Environment oder Defaults
	allowedOrigins := getAllowedOrigins()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Handler erstellen
	authHandler := NewAuthHandler(repo)
	mentorHandler := NewMentorHandler(repo)

	// API v1 Routes
	r.Route("/api/v1", func(r chi.Router) {
		// Auth Routes (öffentlich)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
			r.Post("/logout", authHandler.Logout)
			// /me benötigt Authentifizierung
			r.Group(func(r chi.Router) {
				r.Use(AuthMiddleware(repo))
				r.Get("/me", authHandler.Me)
			})
		})

		// Replays (authentifiziert - nur eigene Replays sichtbar)
		r.Route("/replays", func(r chi.Router) {
			// Upload erlaubt optional authentifiziert (für User-Zuordnung)
			r.Group(func(r chi.Router) {
				r.Use(OptionalAuthMiddleware(repo))
				r.Post("/upload", handler.UploadReplay)
			})
			// Alle anderen Replay-Operationen erfordern Authentifizierung
			r.Group(func(r chi.Router) {
				r.Use(AuthMiddleware(repo))
				r.Get("/", handler.ListReplays)
				r.Get("/{id}", handler.GetReplay)
				r.Get("/{id}/analysis", handler.GetReplayAnalysis)
				r.Get("/{id}/strategic", handler.GetStrategicAnalysis)
				r.Delete("/{id}", handler.DeleteReplay)
				r.Post("/{id}/claim", handler.ClaimReplay)
			})
		})

		// Stats
		r.Route("/stats", func(r chi.Router) {
			r.Get("/trends", handler.GetTrends)
		})

		// Mentor Routes (authentifiziert)
		r.Route("/mentor", func(r chi.Router) {
			r.Use(AuthMiddleware(repo))
			r.Get("/dashboard", mentorHandler.GetDashboard)
			r.Get("/goals", mentorHandler.GetGoals)
			r.Post("/goals", mentorHandler.CreateGoal)
			r.Delete("/goals/{id}", mentorHandler.DeleteGoal)
			r.Get("/progress", mentorHandler.GetProgress)
			r.Get("/weekly-report", mentorHandler.GetWeeklyReport)
			r.Post("/focus", mentorHandler.SetCoachingFocus)
			r.Get("/goal-templates", mentorHandler.GetGoalTemplates)
		})

		// Health Check
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			respondJSON(w, 200, map[string]string{"status": "ok"})
		})
	})

	return r
}

// getAllowedOrigins gibt die erlaubten CORS Origins zurück
func getAllowedOrigins() []string {
	// Aus Environment Variable lesen (kommasepariert)
	originsEnv := os.Getenv("CORS_ORIGINS")
	if originsEnv != "" {
		origins := strings.Split(originsEnv, ",")
		// Trim whitespace
		for i, origin := range origins {
			origins[i] = strings.TrimSpace(origin)
		}
		return origins
	}

	// Defaults für Development
	return []string{
		"http://localhost:5173",
		"http://localhost:5174",
		"http://localhost:5175",
		"http://localhost:3000",
		"http://127.0.0.1:5173",
	}
}
