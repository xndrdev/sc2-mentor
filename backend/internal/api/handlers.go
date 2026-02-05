package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"sc2-analytics/internal/analyzer"
	"sc2-analytics/internal/analyzer/strategic"
	"sc2-analytics/internal/models"
	"sc2-analytics/internal/parser"
	"sc2-analytics/internal/repository"
)

// Handler verwaltet alle API-Requests
type Handler struct {
	repo       *repository.Repository
	parser     *parser.Parser
	analyzer   *analyzer.Analyzer
	uploadDir  string
}

// NewHandler erstellt einen neuen Handler
func NewHandler(repo *repository.Repository, uploadDir string) *Handler {
	return &Handler{
		repo:      repo,
		parser:    parser.New(),
		analyzer:  analyzer.New(),
		uploadDir: uploadDir,
	}
}

// Response-Hilfsfunktionen
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

// UploadReplay behandelt POST /api/v1/replays/upload
func (h *Handler) UploadReplay(w http.ResponseWriter, r *http.Request) {
	// Limitiere Upload-Größe auf 50MB
	r.ParseMultipartForm(50 << 20)

	file, header, err := r.FormFile("replay")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Keine Replay-Datei gefunden")
		return
	}
	defer file.Close()

	// Prüfe Dateierweiterung
	if filepath.Ext(header.Filename) != ".SC2Replay" {
		respondError(w, http.StatusBadRequest, "Nur .SC2Replay Dateien erlaubt")
		return
	}

	// Erstelle temporäre Datei
	tempDir := filepath.Join(h.uploadDir, "temp")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		respondError(w, http.StatusInternalServerError, "Konnte Upload-Verzeichnis nicht erstellen")
		return
	}

	tempFile, err := os.CreateTemp(tempDir, "replay-*.SC2Replay")
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Konnte temporäre Datei nicht erstellen")
		return
	}
	defer os.Remove(tempFile.Name())

	// Kopiere Upload in temporäre Datei
	if _, err := io.Copy(tempFile, file); err != nil {
		respondError(w, http.StatusInternalServerError, "Konnte Datei nicht speichern")
		return
	}
	tempFile.Close()

	// Parse Replay
	parsedReplay, err := h.parser.ParseFile(tempFile.Name())
	if err != nil {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("Konnte Replay nicht parsen: %v", err))
		return
	}

	// Prüfe ob Replay bereits existiert
	existing, err := h.repo.GetReplayByHash(parsedReplay.Hash)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Datenbankfehler")
		return
	}
	if existing != nil {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"message":   "Replay bereits vorhanden",
			"replay_id": existing.ID,
			"replay":    existing,
		})
		return
	}

	// Speichere Replay permanent
	finalPath := filepath.Join(h.uploadDir, "replays", parsedReplay.Hash+".SC2Replay")
	if err := os.MkdirAll(filepath.Dir(finalPath), 0755); err != nil {
		respondError(w, http.StatusInternalServerError, "Konnte Replay-Verzeichnis nicht erstellen")
		return
	}

	if err := copyFile(tempFile.Name(), finalPath); err != nil {
		respondError(w, http.StatusInternalServerError, "Konnte Replay nicht speichern")
		return
	}

	// Erstelle Replay-Eintrag
	replay := &models.Replay{
		Hash:        parsedReplay.Hash,
		Filename:    header.Filename,
		Map:         parsedReplay.Map,
		Duration:    parsedReplay.Duration,
		GameVersion: parsedReplay.GameVersion,
		PlayedAt:    parsedReplay.PlayedAt,
	}

	if err := h.repo.CreateReplay(replay); err != nil {
		respondError(w, http.StatusInternalServerError, "Konnte Replay nicht in DB speichern")
		return
	}

	// Erstelle Spieler-Einträge
	var gamePlayers []models.GamePlayer
	for _, p := range parsedReplay.Players {
		// Erstelle oder finde Spieler
		player, err := h.repo.CreatePlayer(p.ToonHandle, p.Name, p.Region)
		if err != nil {
			continue
		}

		// Berechne Metriken
		apm, sq := h.analyzer.GetPlayerMetrics(parsedReplay, p.Slot, p.Race)

		gp := models.GamePlayer{
			ReplayID:         replay.ID,
			PlayerID:         player.ID,
			PlayerSlot:       p.Slot,
			Name:             p.Name,
			Race:             p.Race,
			Result:           p.Result,
			APM:              apm,
			SpendingQuotient: sq,
			IsHuman:          p.IsHuman,
		}

		if err := h.repo.CreateGamePlayer(&gp); err != nil {
			continue
		}

		gamePlayers = append(gamePlayers, gp)
	}

	replay.GamePlayers = gamePlayers

	// Führe Analyse durch
	analyses, err := h.analyzer.AnalyzeAndStore(parsedReplay, replay.ID, gamePlayers)
	if err == nil {
		for _, analysis := range analyses {
			h.repo.SaveAnalysis(analysis)
		}
	}

	// Prüfe ob Benutzer authentifiziert ist
	user := GetUserFromContext(r.Context())
	needsPlayerSelection := user != nil

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"message":               "Replay erfolgreich hochgeladen",
		"replay_id":             replay.ID,
		"replay":                replay,
		"needs_player_selection": needsPlayerSelection,
	})
}

// DeleteReplay behandelt DELETE /api/v1/replays/:id
func (h *Handler) DeleteReplay(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Ungültige Replay-ID")
		return
	}

	// Lade Replay um Hash für Dateilöschung zu bekommen
	replay, err := h.repo.GetReplayByID(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Datenbankfehler")
		return
	}
	if replay == nil {
		respondError(w, http.StatusNotFound, "Replay nicht gefunden")
		return
	}

	// Lösche aus Datenbank (CASCADE löscht auch game_players, analyses, user_replays)
	if err := h.repo.DeleteReplay(id); err != nil {
		respondError(w, http.StatusInternalServerError, "Fehler beim Löschen: "+err.Error())
		return
	}

	// Lösche Replay-Datei
	filePath := filepath.Join(h.uploadDir, "replays", replay.Hash+".SC2Replay")
	os.Remove(filePath) // Fehler ignorieren, Datei könnte bereits gelöscht sein

	respondJSON(w, http.StatusOK, map[string]string{"message": "Replay gelöscht"})
}

// ClaimReplay behandelt POST /api/v1/replays/:id/claim
// Ordnet ein Replay einem Benutzer zu und aktualisiert den Fortschritt
func (h *Handler) ClaimReplay(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Nicht authentifiziert")
		return
	}

	// Replay ID aus URL
	idStr := chi.URLParam(r, "id")
	replayID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Ungültige Replay-ID")
		return
	}

	// Request Body parsen
	var req struct {
		PlayerID int64 `json:"player_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Ungültige Anfrage")
		return
	}

	if req.PlayerID == 0 {
		respondError(w, http.StatusBadRequest, "player_id ist erforderlich")
		return
	}

	// Lade Replay
	replay, err := h.repo.GetReplayByID(replayID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Datenbankfehler")
		return
	}
	if replay == nil {
		respondError(w, http.StatusNotFound, "Replay nicht gefunden")
		return
	}

	// Prüfe ob der Spieler im Replay ist
	var selectedPlayer *models.GamePlayer
	for _, gp := range replay.GamePlayers {
		if gp.PlayerID == req.PlayerID {
			selectedPlayer = &gp
			break
		}
	}
	if selectedPlayer == nil {
		respondError(w, http.StatusBadRequest, "Spieler nicht in diesem Replay")
		return
	}

	// Verknüpfe Replay mit Benutzer
	if err := h.repo.LinkReplayToUser(user.ID, replayID); err != nil {
		respondError(w, http.StatusInternalServerError, "Fehler beim Verknüpfen")
		return
	}

	// Hole Supply Block Prozent aus der Analyse
	var supplyBlockPct float64
	analysis, err := h.repo.GetAnalysis(replayID, req.PlayerID)
	if err == nil && analysis != nil {
		var data models.AnalysisData
		if err := json.Unmarshal(analysis.Data, &data); err == nil {
			if data.SupplyAnalysis != nil {
				supplyBlockPct = data.SupplyAnalysis.BlockPercentage
			}
		}
	}

	// Aktualisiere den täglichen Fortschritt
	if err := h.repo.UpdateProgressFromReplay(user.ID, replay, selectedPlayer, supplyBlockPct); err != nil {
		// Nicht kritisch, logge nur
		fmt.Printf("Fehler beim Aktualisieren des Fortschritts: %v\n", err)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":    "Replay erfolgreich zugeordnet",
		"replay_id":  replayID,
		"player_id":  req.PlayerID,
		"player_name": selectedPlayer.Name,
	})
}

// ListReplays behandelt GET /api/v1/replays
func (h *Handler) ListReplays(w http.ResponseWriter, r *http.Request) {
	// Paginierung
	limit := 20
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	replays, err := h.repo.ListReplays(limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Konnte Replays nicht laden")
		return
	}

	total, err := h.repo.CountReplays()
	if err != nil {
		total = len(replays)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"replays": replays,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// GetReplay behandelt GET /api/v1/replays/:id
func (h *Handler) GetReplay(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Ungültige Replay-ID")
		return
	}

	replay, err := h.repo.GetReplayByID(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Datenbankfehler")
		return
	}
	if replay == nil {
		respondError(w, http.StatusNotFound, "Replay nicht gefunden")
		return
	}

	respondJSON(w, http.StatusOK, replay)
}

// GetReplayAnalysis behandelt GET /api/v1/replays/:id/analysis
func (h *Handler) GetReplayAnalysis(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Ungültige Replay-ID")
		return
	}

	replay, err := h.repo.GetReplayByID(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Datenbankfehler")
		return
	}
	if replay == nil {
		respondError(w, http.StatusNotFound, "Replay nicht gefunden")
		return
	}

	// Lade Analysen
	analyses, err := h.repo.GetAnalysesByReplayID(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Konnte Analysen nicht laden")
		return
	}

	// Parse JSON-Daten
	playerAnalyses := make(map[int64]interface{})
	for _, a := range analyses {
		var data models.AnalysisData
		if err := json.Unmarshal(a.Data, &data); err == nil {
			playerAnalyses[a.PlayerID] = data
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"replay":   replay,
		"analyses": playerAnalyses,
	})
}

// GetTrends behandelt GET /api/v1/stats/trends
func (h *Handler) GetTrends(w http.ResponseWriter, r *http.Request) {
	// Player-ID aus Query
	playerIDStr := r.URL.Query().Get("player_id")
	if playerIDStr == "" {
		respondError(w, http.StatusBadRequest, "player_id Parameter erforderlich")
		return
	}

	playerID, err := strconv.ParseInt(playerIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Ungültige player_id")
		return
	}

	// Limit für Datenpunkte
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 50 {
			limit = parsed
		}
	}

	// Lade alle Metriken
	metrics, err := h.repo.GetAllPlayerMetrics(playerID, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Konnte Trends nicht laden")
		return
	}

	// Berechne Trends
	trends := calculateTrends(metrics)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"player_id": playerID,
		"trends":    trends,
		"data":      metrics,
	})
}

// calculateTrends berechnet Trend-Daten aus Metriken
func calculateTrends(metrics []map[string]interface{}) map[string]models.TrendData {
	trends := make(map[string]models.TrendData)

	if len(metrics) < 2 {
		return trends
	}

	// APM Trend
	trends["apm"] = calculateSingleTrend(metrics, "apm")
	trends["spending_quotient"] = calculateSingleTrend(metrics, "spending_quotient")

	return trends
}

// calculateSingleTrend berechnet den Trend für eine einzelne Metrik
func calculateSingleTrend(metrics []map[string]interface{}, key string) models.TrendData {
	trend := models.TrendData{
		Metric:     key,
		DataPoints: []models.TrendPoint{},
	}

	var values []float64
	for _, m := range metrics {
		if val, ok := m[key].(float64); ok {
			values = append(values, val)
		}
	}

	if len(values) < 2 {
		trend.Trend = "insufficient_data"
		return trend
	}

	// Berechne Änderung
	first := values[len(values)-1] // Ältester Wert
	last := values[0]              // Neuester Wert

	if first > 0 {
		trend.Change = ((last - first) / first) * 100
	}

	// Bestimme Trend-Richtung
	if trend.Change > 5 {
		trend.Trend = "improving"
	} else if trend.Change < -5 {
		trend.Trend = "declining"
	} else {
		trend.Trend = "stable"
	}

	return trend
}

// GetStrategicAnalysis behandelt GET /api/v1/replays/:id/strategic
func (h *Handler) GetStrategicAnalysis(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Ungültige Replay-ID")
		return
	}

	// Lade Replay mit Spielern
	replay, err := h.repo.GetReplayByID(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Datenbankfehler")
		return
	}
	if replay == nil {
		respondError(w, http.StatusNotFound, "Replay nicht gefunden")
		return
	}

	// Lade Analysen
	analyses, err := h.repo.GetAnalysesByReplayID(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Konnte Analysen nicht laden")
		return
	}

	// Parse JSON-Daten und finde Winner/Loser
	var winnerAnalysis, loserAnalysis *models.AnalysisData
	var winnerPlayer, loserPlayer *models.GamePlayer

	for _, gp := range replay.GamePlayers {
		var data models.AnalysisData
		for _, a := range analyses {
			if a.PlayerID == gp.PlayerID {
				if err := json.Unmarshal(a.Data, &data); err == nil {
					if gp.Result == "Win" {
						winnerAnalysis = &data
						winnerPlayer = &gp
					} else if gp.Result == "Loss" {
						loserAnalysis = &data
						loserPlayer = &gp
					}
				}
				break
			}
		}
	}

	if winnerPlayer == nil || loserPlayer == nil {
		respondError(w, http.StatusBadRequest, "Kein eindeutiger Winner/Loser gefunden")
		return
	}

	// Erstelle strategische Analyse
	sa := strategic.NewStrategicAnalyzer()
	strategicAnalysis := sa.Analyze(
		loserPlayer.Name, winnerPlayer.Name,
		loserPlayer.Race, winnerPlayer.Race,
		loserAnalysis, winnerAnalysis,
	)

	if strategicAnalysis == nil {
		respondError(w, http.StatusInternalServerError, "Konnte strategische Analyse nicht erstellen")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"replay":   replay,
		"analysis": strategicAnalysis,
	})
}

// copyFile kopiert eine Datei
func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

// ============== Mentor Handlers ==============

// MentorHandler verwaltet Mentor-bezogene API-Requests
type MentorHandler struct {
	repo *repository.Repository
}

// NewMentorHandler erstellt einen neuen MentorHandler
func NewMentorHandler(repo *repository.Repository) *MentorHandler {
	return &MentorHandler{repo: repo}
}

// GetDashboard behandelt GET /api/v1/mentor/dashboard
func (h *MentorHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Nicht authentifiziert")
		return
	}

	// Prüfe und aktualisiere abgelaufene Ziele
	h.repo.CheckAndUpdateExpiredGoals(user.ID)

	// Hole heutige Statistiken
	todayStats, err := h.repo.GetOrCreateDailyProgress(user.ID, time.Now())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Fehler beim Laden der Tagesstatistiken")
		return
	}

	// Hole Wochenstatistiken
	weekStats, err := h.repo.GetWeekStats(user.ID)
	if err != nil {
		weekStats = &models.WeekStats{}
	}

	// Hole aktive Ziele
	activeGoals, err := h.repo.GetActiveGoals(user.ID)
	if err != nil {
		activeGoals = []models.Goal{}
	}

	// Hole letzte Spiele
	recentGames, err := h.repo.GetRecentGames(user.ID, 5)
	if err != nil {
		recentGames = []models.RecentGame{}
	}

	// Hole Coaching-Fokus
	currentFocus, _ := h.repo.GetActiveCoachingFocus(user.ID)

	// Hole Fortschritts-Trend (letzte 14 Tage)
	progressTrend, err := h.repo.GetProgressHistory(user.ID, 14)
	if err != nil {
		progressTrend = []models.DailyProgress{}
	}

	// Hole Wochenbericht (aktuelle Woche)
	now := time.Now()
	weekStart := now.AddDate(0, 0, -int(now.Weekday()))
	weeklyReport, _ := h.repo.GetWeeklyReport(user.ID, weekStart)

	dashboard := models.MentorDashboard{
		User:          user.ToPublic(),
		TodayStats:    todayStats,
		WeekStats:     weekStats,
		ActiveGoals:   activeGoals,
		RecentGames:   recentGames,
		CurrentFocus:  currentFocus,
		WeeklyReport:  weeklyReport,
		ProgressTrend: progressTrend,
	}

	respondJSON(w, http.StatusOK, dashboard)
}

// GetGoals behandelt GET /api/v1/mentor/goals
func (h *MentorHandler) GetGoals(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Nicht authentifiziert")
		return
	}

	// Prüfe und aktualisiere abgelaufene Ziele
	h.repo.CheckAndUpdateExpiredGoals(user.ID)

	goals, err := h.repo.GetActiveGoals(user.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Fehler beim Laden der Ziele")
		return
	}

	// Füge Zielvorlagen hinzu
	templates := models.GetGoalTemplates()

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"goals":     goals,
		"templates": templates,
	})
}

// CreateGoal behandelt POST /api/v1/mentor/goals
func (h *MentorHandler) CreateGoal(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Nicht authentifiziert")
		return
	}

	var req models.CreateGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Ungültige Anfrage: "+err.Error())
		return
	}

	// Validierung
	if req.GoalType == "" || req.MetricName == "" || req.TargetValue <= 0 {
		respondError(w, http.StatusBadRequest, "goal_type, metric_name und target_value sind erforderlich")
		return
	}

	// Validiere goal_type
	if req.GoalType != "daily" && req.GoalType != "weekly" {
		respondError(w, http.StatusBadRequest, "goal_type muss 'daily' oder 'weekly' sein")
		return
	}

	// Validiere metric_name
	validMetrics := map[string]bool{
		"games_played": true, "apm": true, "supply_block": true,
		"win_rate": true, "sq": true,
	}
	if !validMetrics[req.MetricName] {
		respondError(w, http.StatusBadRequest, "Ungültiger metric_name")
		return
	}

	// Setze Standard-Comparison
	if req.Comparison == "" {
		if req.MetricName == "supply_block" {
			req.Comparison = "<="
		} else {
			req.Comparison = ">="
		}
	}

	// Berechne Deadline
	var deadline time.Time
	now := time.Now()
	if req.GoalType == "daily" {
		// Ende des heutigen Tages
		deadline = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
	} else {
		// Ende der Woche (Sonntag 23:59)
		daysUntilSunday := 7 - int(now.Weekday())
		if daysUntilSunday == 7 {
			daysUntilSunday = 0
		}
		endOfWeek := now.AddDate(0, 0, daysUntilSunday)
		deadline = time.Date(endOfWeek.Year(), endOfWeek.Month(), endOfWeek.Day(), 23, 59, 59, 0, now.Location())
	}

	goal := &models.Goal{
		UserID:      user.ID,
		GoalType:    req.GoalType,
		MetricName:  req.MetricName,
		TargetValue: req.TargetValue,
		Comparison:  req.Comparison,
		Deadline:    deadline,
	}

	if err := h.repo.CreateGoal(goal); err != nil {
		respondError(w, http.StatusInternalServerError, "Fehler beim Erstellen des Ziels: "+err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, goal)
}

// DeleteGoal behandelt DELETE /api/v1/mentor/goals/:id
func (h *MentorHandler) DeleteGoal(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Nicht authentifiziert")
		return
	}

	idStr := chi.URLParam(r, "id")
	goalID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Ungültige Ziel-ID")
		return
	}

	// Setze Status auf 'deleted' (soft delete)
	if err := h.repo.UpdateGoalStatus(goalID, "deleted"); err != nil {
		respondError(w, http.StatusInternalServerError, "Fehler beim Löschen des Ziels")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Ziel gelöscht"})
}

// GetProgress behandelt GET /api/v1/mentor/progress
func (h *MentorHandler) GetProgress(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Nicht authentifiziert")
		return
	}

	// Standard: 14 Tage
	days := 14
	if d := r.URL.Query().Get("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 && parsed <= 90 {
			days = parsed
		}
	}

	progress, err := h.repo.GetProgressHistory(user.ID, days)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Fehler beim Laden des Fortschritts")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"progress": progress,
		"days":     days,
	})
}

// GetWeeklyReport behandelt GET /api/v1/mentor/weekly-report
func (h *MentorHandler) GetWeeklyReport(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Nicht authentifiziert")
		return
	}

	// Berechne Wochenstart (Montag)
	now := time.Now()
	weekStart := now.AddDate(0, 0, -int(now.Weekday())+1) // +1 für Montag statt Sonntag
	if now.Weekday() == time.Sunday {
		weekStart = weekStart.AddDate(0, 0, -7)
	}
	weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, now.Location())

	// Prüfe ob "generate" Parameter gesetzt ist
	generate := r.URL.Query().Get("generate") == "true"

	report, err := h.repo.GetWeeklyReport(user.ID, weekStart)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Fehler beim Laden des Wochenberichts")
		return
	}

	// Generiere neuen Bericht wenn keiner existiert oder wenn generate=true
	if report == nil || generate {
		report, err = h.repo.GenerateWeeklyReport(user.ID, weekStart)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Fehler beim Generieren des Wochenberichts: "+err.Error())
			return
		}
	}

	respondJSON(w, http.StatusOK, report)
}

// SetCoachingFocus behandelt POST /api/v1/mentor/focus
func (h *MentorHandler) SetCoachingFocus(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Nicht authentifiziert")
		return
	}

	var req struct {
		FocusArea   string `json:"focus_area"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Ungültige Anfrage")
		return
	}

	// Validiere focus_area
	validAreas := map[string]bool{
		"macro": true, "micro": true, "economy": true,
		"army_control": true, "scouting": true,
	}
	if !validAreas[req.FocusArea] {
		respondError(w, http.StatusBadRequest, "Ungültiger Fokusbereich")
		return
	}

	focus, err := h.repo.SetCoachingFocus(user.ID, req.FocusArea, req.Description)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Fehler beim Setzen des Fokus")
		return
	}

	respondJSON(w, http.StatusOK, focus)
}

// GetGoalTemplates behandelt GET /api/v1/mentor/goal-templates
func (h *MentorHandler) GetGoalTemplates(w http.ResponseWriter, r *http.Request) {
	templates := models.GetGoalTemplates()
	respondJSON(w, http.StatusOK, templates)
}
