package api

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"sc2-analytics/internal/models"
	"sc2-analytics/internal/repository"
)

// Rate Limiter für Auth Endpoints
type rateLimiter struct {
	requests map[string][]time.Time
	mu       sync.Mutex
	limit    int
	window   time.Duration
}

var authRateLimiter = &rateLimiter{
	requests: make(map[string][]time.Time),
	limit:    5,              // 5 Versuche
	window:   time.Minute,    // pro Minute
}

func (rl *rateLimiter) isAllowed(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Entferne alte Einträge
	var recent []time.Time
	for _, t := range rl.requests[ip] {
		if t.After(windowStart) {
			recent = append(recent, t)
		}
	}
	rl.requests[ip] = recent

	// Prüfe Limit
	if len(recent) >= rl.limit {
		return false
	}

	// Füge neuen Request hinzu
	rl.requests[ip] = append(rl.requests[ip], now)
	return true
}

func getClientIP(r *http.Request) string {
	// X-Forwarded-For für Reverse Proxy
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	// Fallback auf RemoteAddr
	return strings.Split(r.RemoteAddr, ":")[0]
}

// JWT Konfiguration
var jwtSecret string

const tokenDuration = 7 * 24 * time.Hour // 7 Tage

func init() {
	jwtSecret = os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		// In Production muss JWT_SECRET gesetzt sein
		if os.Getenv("GO_ENV") == "production" {
			panic("JWT_SECRET environment variable is required in production")
		}
		// Für Development: generiere ein zufälliges Secret
		bytes := make([]byte, 32)
		if _, err := rand.Read(bytes); err != nil {
			panic("Failed to generate random JWT secret: " + err.Error())
		}
		jwtSecret = base64.StdEncoding.EncodeToString(bytes)
		fmt.Println("WARNING: Using random JWT secret. Set JWT_SECRET env var for persistent sessions.")
	}
}

// contextKey ist der Typ für Context-Keys
type contextKey string

const userContextKey contextKey = "user"

// Claims sind die JWT Claims
type Claims struct {
	UserID        int64  `json:"user_id"`
	Email         string `json:"email"`
	SC2PlayerName string `json:"sc2_player_name"`
	jwt.RegisteredClaims
}

// AuthHandler verwaltet Authentifizierungs-Endpoints
type AuthHandler struct {
	repo *repository.Repository
}

// NewAuthHandler erstellt einen neuen AuthHandler
func NewAuthHandler(repo *repository.Repository) *AuthHandler {
	return &AuthHandler{repo: repo}
}

// Register registriert einen neuen Benutzer
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// Rate Limiting
	clientIP := getClientIP(r)
	if !authRateLimiter.isAllowed(clientIP) {
		respondError(w, http.StatusTooManyRequests, "Zu viele Anfragen. Bitte warte eine Minute.")
		return
	}

	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Ungültige Anfrage: "+err.Error())
		return
	}

	// Validierung
	if req.Email == "" || req.Password == "" || req.SC2PlayerName == "" {
		respondError(w, http.StatusBadRequest, "Email, Passwort und SC2 Spielername sind erforderlich")
		return
	}

	if len(req.Password) < 8 {
		respondError(w, http.StatusBadRequest, "Passwort muss mindestens 8 Zeichen haben")
		return
	}

	// Prüfe ob Email bereits existiert
	existingUser, err := h.repo.GetUserByEmail(req.Email)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Datenbankfehler: "+err.Error())
		return
	}
	if existingUser != nil {
		respondError(w, http.StatusConflict, "Diese Email ist bereits registriert")
		return
	}

	// Hash das Passwort
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Fehler beim Hashen des Passworts")
		return
	}

	// Erstelle den Benutzer
	user, err := h.repo.CreateUser(req.Email, string(hashedPassword), req.SC2PlayerName)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Fehler beim Erstellen des Benutzers: "+err.Error())
		return
	}

	// Generiere JWT Token
	token, err := generateToken(user)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Fehler beim Generieren des Tokens")
		return
	}

	respondJSON(w, http.StatusCreated, models.AuthResponse{
		Token: token,
		User:  user.ToPublic(),
	})
}

// Login authentifiziert einen Benutzer
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// Rate Limiting
	clientIP := getClientIP(r)
	if !authRateLimiter.isAllowed(clientIP) {
		respondError(w, http.StatusTooManyRequests, "Zu viele Anfragen. Bitte warte eine Minute.")
		return
	}

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Ungültige Anfrage: "+err.Error())
		return
	}

	if req.Email == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "Email und Passwort sind erforderlich")
		return
	}

	// Finde den Benutzer
	user, err := h.repo.GetUserByEmail(req.Email)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Datenbankfehler")
		return
	}
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Ungültige Anmeldedaten")
		return
	}

	// Prüfe das Passwort
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		respondError(w, http.StatusUnauthorized, "Ungültige Anmeldedaten")
		return
	}

	// Aktualisiere letzten Login
	h.repo.UpdateUserLastLogin(user.ID)

	// Generiere JWT Token
	token, err := generateToken(user)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Fehler beim Generieren des Tokens")
		return
	}

	respondJSON(w, http.StatusOK, models.AuthResponse{
		Token: token,
		User:  user.ToPublic(),
	})
}

// Me gibt den aktuellen Benutzer zurück
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Nicht authentifiziert")
		return
	}

	respondJSON(w, http.StatusOK, user.ToPublic())
}

// Logout invalidiert den Token (nur client-seitig)
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Bei JWT-basierter Authentifizierung kann der Server den Token nicht invalidieren
	// Der Client muss den Token einfach löschen
	respondJSON(w, http.StatusOK, map[string]string{"message": "Erfolgreich abgemeldet"})
}

// AuthMiddleware prüft JWT Tokens
func AuthMiddleware(repo *repository.Repository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondError(w, http.StatusUnauthorized, "Authorization Header fehlt")
				return
			}

			// Bearer Token extrahieren
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				respondError(w, http.StatusUnauthorized, "Ungültiges Token-Format")
				return
			}

			tokenString := parts[1]

			// Token validieren
			claims := &Claims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("ungültige Signaturmethode")
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				respondError(w, http.StatusUnauthorized, "Ungültiger Token")
				return
			}

			// Benutzer aus DB laden
			user, err := repo.GetUserByID(claims.UserID)
			if err != nil || user == nil {
				respondError(w, http.StatusUnauthorized, "Benutzer nicht gefunden")
				return
			}

			// Benutzer zum Context hinzufügen
			ctx := context.WithValue(r.Context(), userContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuthMiddleware prüft JWT Tokens, aber erlaubt auch unauthentifizierte Anfragen
func OptionalAuthMiddleware(repo *repository.Repository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				next.ServeHTTP(w, r)
				return
			}

			tokenString := parts[1]
			claims := &Claims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("ungültige Signaturmethode")
				}
				return []byte(jwtSecret), nil
			})

			if err == nil && token.Valid {
				user, _ := repo.GetUserByID(claims.UserID)
				if user != nil {
					ctx := context.WithValue(r.Context(), userContextKey, user)
					r = r.WithContext(ctx)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserFromContext extrahiert den Benutzer aus dem Context
func GetUserFromContext(ctx context.Context) *models.User {
	user, ok := ctx.Value(userContextKey).(*models.User)
	if !ok {
		return nil
	}
	return user
}

// generateToken generiert einen JWT Token
func generateToken(user *models.User) (string, error) {
	expirationTime := time.Now().Add(tokenDuration)
	claims := &Claims{
		UserID:        user.ID,
		Email:         user.Email,
		SC2PlayerName: user.SC2PlayerName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "sc2-analytics",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}
