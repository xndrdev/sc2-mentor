package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"sc2-analytics/internal/models"
)

// Repository verwaltet die Datenbankoperationen
type Repository struct {
	db *sql.DB
}

// New erstellt ein neues Repository und initialisiert die DB
func New(dbPath string) (*Repository, error) {
	// Stelle sicher, dass das Verzeichnis existiert
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("konnte DB-Verzeichnis nicht erstellen: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("konnte Datenbank nicht öffnen: %w", err)
	}

	// Teste die Verbindung
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("konnte keine DB-Verbindung herstellen: %w", err)
	}

	repo := &Repository{db: db}

	// Führe Migrationen aus
	if err := repo.migrate(); err != nil {
		return nil, fmt.Errorf("Migration fehlgeschlagen: %w", err)
	}

	return repo, nil
}

// Close schließt die Datenbankverbindung
func (r *Repository) Close() error {
	return r.db.Close()
}

// migrate führt die SQL-Migrationen aus
func (r *Repository) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS players (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		toon_handle TEXT UNIQUE NOT NULL,
		name TEXT NOT NULL,
		region TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS replays (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		hash TEXT UNIQUE NOT NULL,
		filename TEXT NOT NULL,
		map TEXT NOT NULL,
		duration INTEGER NOT NULL,
		game_version TEXT DEFAULT '',
		played_at DATETIME,
		uploaded_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS game_players (
		replay_id INTEGER NOT NULL,
		player_id INTEGER NOT NULL,
		player_slot INTEGER NOT NULL,
		name TEXT NOT NULL,
		race TEXT NOT NULL,
		result TEXT NOT NULL,
		apm REAL DEFAULT 0,
		spending_quotient REAL DEFAULT 0,
		is_human INTEGER DEFAULT 1,
		PRIMARY KEY (replay_id, player_id),
		FOREIGN KEY (replay_id) REFERENCES replays(id) ON DELETE CASCADE,
		FOREIGN KEY (player_id) REFERENCES players(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS analyses (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		replay_id INTEGER NOT NULL,
		player_id INTEGER NOT NULL,
		data TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(replay_id, player_id),
		FOREIGN KEY (replay_id) REFERENCES replays(id) ON DELETE CASCADE,
		FOREIGN KEY (player_id) REFERENCES players(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_replays_played_at ON replays(played_at);
	CREATE INDEX IF NOT EXISTS idx_replays_hash ON replays(hash);
	CREATE INDEX IF NOT EXISTS idx_game_players_replay ON game_players(replay_id);
	CREATE INDEX IF NOT EXISTS idx_game_players_player ON game_players(player_id);
	CREATE INDEX IF NOT EXISTS idx_analyses_replay ON analyses(replay_id);

	-- Mentoring System Tables
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		sc2_player_name TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_login DATETIME
	);

	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	CREATE INDEX IF NOT EXISTS idx_users_player_name ON users(sc2_player_name);

	CREATE TABLE IF NOT EXISTS user_replays (
		user_id INTEGER NOT NULL,
		replay_id INTEGER NOT NULL,
		PRIMARY KEY (user_id, replay_id),
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (replay_id) REFERENCES replays(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS goals (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		goal_type TEXT NOT NULL,
		metric_name TEXT NOT NULL,
		target_value REAL NOT NULL,
		comparison TEXT DEFAULT '>=',
		current_value REAL DEFAULT 0,
		status TEXT DEFAULT 'active',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		deadline DATETIME NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_goals_user_status ON goals(user_id, status);
	CREATE INDEX IF NOT EXISTS idx_goals_deadline ON goals(deadline);

	CREATE TABLE IF NOT EXISTS daily_progress (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		date DATE NOT NULL,
		games_played INTEGER DEFAULT 0,
		wins INTEGER DEFAULT 0,
		losses INTEGER DEFAULT 0,
		avg_apm REAL DEFAULT 0,
		avg_spending_quotient REAL DEFAULT 0,
		avg_supply_block_pct REAL DEFAULT 0,
		total_play_time INTEGER DEFAULT 0,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		UNIQUE(user_id, date)
	);

	CREATE INDEX IF NOT EXISTS idx_daily_progress_user_date ON daily_progress(user_id, date);

	CREATE TABLE IF NOT EXISTS weekly_reports (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		week_start DATE NOT NULL,
		week_end DATE NOT NULL,
		total_games INTEGER DEFAULT 0,
		wins INTEGER DEFAULT 0,
		losses INTEGER DEFAULT 0,
		win_rate REAL DEFAULT 0,
		avg_apm REAL DEFAULT 0,
		avg_sq REAL DEFAULT 0,
		avg_supply_block REAL DEFAULT 0,
		main_race TEXT,
		total_play_time INTEGER DEFAULT 0,
		improvements TEXT,
		regressions TEXT,
		focus_suggestion TEXT,
		strengths TEXT,
		weaknesses TEXT,
		generated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		UNIQUE(user_id, week_start)
	);

	CREATE INDEX IF NOT EXISTS idx_weekly_reports_user_week ON weekly_reports(user_id, week_start);

	CREATE TABLE IF NOT EXISTS coaching_focus (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		focus_area TEXT NOT NULL,
		description TEXT,
		started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		active INTEGER DEFAULT 1,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_coaching_focus_user_active ON coaching_focus(user_id, active);
	`

	_, err := r.db.Exec(schema)
	return err
}

// CreatePlayer erstellt oder findet einen Spieler
func (r *Repository) CreatePlayer(toonHandle, name, region string) (*models.Player, error) {
	// Versuche zuerst, den Spieler zu finden
	var player models.Player
	err := r.db.QueryRow(
		"SELECT id, toon_handle, name, region FROM players WHERE toon_handle = ?",
		toonHandle,
	).Scan(&player.ID, &player.ToonHandle, &player.Name, &player.Region)

	if err == nil {
		return &player, nil
	}

	if err != sql.ErrNoRows {
		return nil, err
	}

	// Spieler existiert nicht, erstelle ihn
	result, err := r.db.Exec(
		"INSERT INTO players (toon_handle, name, region) VALUES (?, ?, ?)",
		toonHandle, name, region,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &models.Player{
		ID:         id,
		ToonHandle: toonHandle,
		Name:       name,
		Region:     region,
	}, nil
}

// GetPlayerByToonHandle findet einen Spieler anhand des ToonHandle
func (r *Repository) GetPlayerByToonHandle(toonHandle string) (*models.Player, error) {
	var player models.Player
	err := r.db.QueryRow(
		"SELECT id, toon_handle, name, region FROM players WHERE toon_handle = ?",
		toonHandle,
	).Scan(&player.ID, &player.ToonHandle, &player.Name, &player.Region)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &player, nil
}

// CreateReplay speichert ein neues Replay
func (r *Repository) CreateReplay(replay *models.Replay) error {
	result, err := r.db.Exec(
		`INSERT INTO replays (hash, filename, map, duration, game_version, played_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		replay.Hash, replay.Filename, replay.Map, replay.Duration,
		replay.GameVersion, replay.PlayedAt,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	replay.ID = id
	return nil
}

// GetReplayByHash findet ein Replay anhand des Hashes
func (r *Repository) GetReplayByHash(hash string) (*models.Replay, error) {
	var replay models.Replay
	err := r.db.QueryRow(
		`SELECT id, hash, filename, map, duration, game_version, played_at, uploaded_at
		 FROM replays WHERE hash = ?`,
		hash,
	).Scan(&replay.ID, &replay.Hash, &replay.Filename, &replay.Map,
		&replay.Duration, &replay.GameVersion, &replay.PlayedAt, &replay.UploadedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &replay, nil
}

// GetReplayByID findet ein Replay anhand der ID
func (r *Repository) GetReplayByID(id int64) (*models.Replay, error) {
	var replay models.Replay
	err := r.db.QueryRow(
		`SELECT id, hash, filename, map, duration, game_version, played_at, uploaded_at
		 FROM replays WHERE id = ?`,
		id,
	).Scan(&replay.ID, &replay.Hash, &replay.Filename, &replay.Map,
		&replay.Duration, &replay.GameVersion, &replay.PlayedAt, &replay.UploadedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Lade Spieler
	players, err := r.GetGamePlayersByReplayID(id)
	if err != nil {
		return nil, err
	}
	replay.GamePlayers = players

	return &replay, nil
}

// ListReplays gibt alle Replays zurück (neueste zuerst)
func (r *Repository) ListReplays(limit, offset int) ([]models.Replay, error) {
	rows, err := r.db.Query(
		`SELECT id, hash, filename, map, duration, game_version, played_at, uploaded_at
		 FROM replays ORDER BY played_at DESC LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var replays []models.Replay
	for rows.Next() {
		var replay models.Replay
		err := rows.Scan(&replay.ID, &replay.Hash, &replay.Filename, &replay.Map,
			&replay.Duration, &replay.GameVersion, &replay.PlayedAt, &replay.UploadedAt)
		if err != nil {
			return nil, err
		}

		// Lade Spieler für dieses Replay
		players, err := r.GetGamePlayersByReplayID(replay.ID)
		if err != nil {
			return nil, err
		}
		replay.GamePlayers = players

		replays = append(replays, replay)
	}

	return replays, rows.Err()
}

// CountReplays zählt alle Replays
func (r *Repository) CountReplays() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM replays").Scan(&count)
	return count, err
}

// DeleteReplay löscht ein Replay und alle zugehörigen Daten
func (r *Repository) DeleteReplay(id int64) error {
	_, err := r.db.Exec("DELETE FROM replays WHERE id = ?", id)
	return err
}

// CreateGamePlayer speichert einen Spieler für ein Replay
func (r *Repository) CreateGamePlayer(gp *models.GamePlayer) error {
	_, err := r.db.Exec(
		`INSERT INTO game_players (replay_id, player_id, player_slot, name, race, result, apm, spending_quotient, is_human)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		gp.ReplayID, gp.PlayerID, gp.PlayerSlot, gp.Name, gp.Race, gp.Result,
		gp.APM, gp.SpendingQuotient, gp.IsHuman,
	)
	return err
}

// GetGamePlayersByReplayID gibt alle Spieler eines Replays zurück
func (r *Repository) GetGamePlayersByReplayID(replayID int64) ([]models.GamePlayer, error) {
	rows, err := r.db.Query(
		`SELECT replay_id, player_id, player_slot, name, race, result, apm, spending_quotient, is_human
		 FROM game_players WHERE replay_id = ? ORDER BY player_slot`,
		replayID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []models.GamePlayer
	for rows.Next() {
		var gp models.GamePlayer
		err := rows.Scan(&gp.ReplayID, &gp.PlayerID, &gp.PlayerSlot, &gp.Name,
			&gp.Race, &gp.Result, &gp.APM, &gp.SpendingQuotient, &gp.IsHuman)
		if err != nil {
			return nil, err
		}
		players = append(players, gp)
	}

	return players, rows.Err()
}

// UpdateGamePlayerMetrics aktualisiert APM und SQ eines Spielers
func (r *Repository) UpdateGamePlayerMetrics(replayID, playerID int64, apm, sq float64) error {
	_, err := r.db.Exec(
		`UPDATE game_players SET apm = ?, spending_quotient = ?
		 WHERE replay_id = ? AND player_id = ?`,
		apm, sq, replayID, playerID,
	)
	return err
}

// SaveAnalysis speichert eine Analyse
func (r *Repository) SaveAnalysis(analysis *models.Analysis) error {
	result, err := r.db.Exec(
		`INSERT OR REPLACE INTO analyses (replay_id, player_id, data)
		 VALUES (?, ?, ?)`,
		analysis.ReplayID, analysis.PlayerID, analysis.Data,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	analysis.ID = id
	return nil
}

// GetAnalysis lädt eine Analyse für ein Replay und einen Spieler
func (r *Repository) GetAnalysis(replayID, playerID int64) (*models.Analysis, error) {
	var analysis models.Analysis
	err := r.db.QueryRow(
		`SELECT id, replay_id, player_id, data, created_at
		 FROM analyses WHERE replay_id = ? AND player_id = ?`,
		replayID, playerID,
	).Scan(&analysis.ID, &analysis.ReplayID, &analysis.PlayerID,
		&analysis.Data, &analysis.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &analysis, nil
}

// GetAnalysesByReplayID lädt alle Analysen für ein Replay
func (r *Repository) GetAnalysesByReplayID(replayID int64) ([]models.Analysis, error) {
	rows, err := r.db.Query(
		`SELECT id, replay_id, player_id, data, created_at
		 FROM analyses WHERE replay_id = ?`,
		replayID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var analyses []models.Analysis
	for rows.Next() {
		var a models.Analysis
		err := rows.Scan(&a.ID, &a.ReplayID, &a.PlayerID, &a.Data, &a.CreatedAt)
		if err != nil {
			return nil, err
		}
		analyses = append(analyses, a)
	}

	return analyses, rows.Err()
}

// GetPlayerTrends gibt Trend-Daten für einen Spieler zurück
func (r *Repository) GetPlayerTrends(playerID int64, limit int) ([]models.TrendPoint, error) {
	rows, err := r.db.Query(
		`SELECT gp.replay_id, r.played_at, gp.apm, gp.spending_quotient
		 FROM game_players gp
		 JOIN replays r ON r.id = gp.replay_id
		 WHERE gp.player_id = ?
		 ORDER BY r.played_at DESC
		 LIMIT ?`,
		playerID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []models.TrendPoint
	for rows.Next() {
		var tp models.TrendPoint
		var apm, sq float64
		err := rows.Scan(&tp.ReplayID, &tp.PlayedAt, &apm, &sq)
		if err != nil {
			return nil, err
		}
		// Hier könnte man unterschiedliche Metriken extrahieren
		tp.Value = sq // oder apm, je nach Kontext
		points = append(points, tp)
	}

	return points, rows.Err()
}

// GetAllPlayerMetrics gibt alle Metriken für Trend-Analyse zurück
func (r *Repository) GetAllPlayerMetrics(playerID int64, limit int) ([]map[string]interface{}, error) {
	rows, err := r.db.Query(
		`SELECT gp.replay_id, r.played_at, gp.apm, gp.spending_quotient, a.data
		 FROM game_players gp
		 JOIN replays r ON r.id = gp.replay_id
		 LEFT JOIN analyses a ON a.replay_id = gp.replay_id AND a.player_id = gp.player_id
		 WHERE gp.player_id = ?
		 ORDER BY r.played_at DESC
		 LIMIT ?`,
		playerID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var replayID int64
		var playedAt string
		var apm, sq float64
		var analysisData sql.NullString

		err := rows.Scan(&replayID, &playedAt, &apm, &sq, &analysisData)
		if err != nil {
			return nil, err
		}

		result := map[string]interface{}{
			"replay_id":        replayID,
			"played_at":        playedAt,
			"apm":              apm,
			"spending_quotient": sq,
		}

		// Parse Analyse-Daten wenn vorhanden
		if analysisData.Valid {
			var data models.AnalysisData
			if err := json.Unmarshal([]byte(analysisData.String), &data); err == nil {
				if data.SupplyAnalysis != nil {
					result["supply_block_percentage"] = data.SupplyAnalysis.BlockPercentage
				}
				if data.InjectAnalysis != nil {
					result["inject_efficiency"] = data.InjectAnalysis.Efficiency
				}
			}
		}

		results = append(results, result)
	}

	return results, rows.Err()
}

// ============== User Methods ==============

// CreateUser erstellt einen neuen Benutzer
func (r *Repository) CreateUser(email, passwordHash, sc2PlayerName string) (*models.User, error) {
	result, err := r.db.Exec(
		`INSERT INTO users (email, password_hash, sc2_player_name) VALUES (?, ?, ?)`,
		email, passwordHash, sc2PlayerName,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &models.User{
		ID:            id,
		Email:         email,
		PasswordHash:  passwordHash,
		SC2PlayerName: sc2PlayerName,
		CreatedAt:     time.Now(),
	}, nil
}

// GetUserByEmail findet einen Benutzer anhand der Email
func (r *Repository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	var lastLogin sql.NullTime
	err := r.db.QueryRow(
		`SELECT id, email, password_hash, sc2_player_name, created_at, last_login
		 FROM users WHERE email = ?`,
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.SC2PlayerName, &user.CreatedAt, &lastLogin)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	}
	return &user, nil
}

// GetUserByID findet einen Benutzer anhand der ID
func (r *Repository) GetUserByID(id int64) (*models.User, error) {
	var user models.User
	var lastLogin sql.NullTime
	err := r.db.QueryRow(
		`SELECT id, email, password_hash, sc2_player_name, created_at, last_login
		 FROM users WHERE id = ?`,
		id,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.SC2PlayerName, &user.CreatedAt, &lastLogin)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	}
	return &user, nil
}

// UpdateUserLastLogin aktualisiert den letzten Login
func (r *Repository) UpdateUserLastLogin(userID int64) error {
	_, err := r.db.Exec(
		`UPDATE users SET last_login = CURRENT_TIMESTAMP WHERE id = ?`,
		userID,
	)
	return err
}

// LinkReplayToUser verknüpft ein Replay mit einem Benutzer
func (r *Repository) LinkReplayToUser(userID, replayID int64) error {
	_, err := r.db.Exec(
		`INSERT OR IGNORE INTO user_replays (user_id, replay_id) VALUES (?, ?)`,
		userID, replayID,
	)
	return err
}

// GetUserReplays gibt alle Replays eines Benutzers zurück
func (r *Repository) GetUserReplays(userID int64, limit, offset int) ([]models.Replay, error) {
	rows, err := r.db.Query(
		`SELECT r.id, r.hash, r.filename, r.map, r.duration, r.game_version, r.played_at, r.uploaded_at
		 FROM replays r
		 JOIN user_replays ur ON r.id = ur.replay_id
		 WHERE ur.user_id = ?
		 ORDER BY r.played_at DESC
		 LIMIT ? OFFSET ?`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var replays []models.Replay
	for rows.Next() {
		var replay models.Replay
		err := rows.Scan(&replay.ID, &replay.Hash, &replay.Filename, &replay.Map,
			&replay.Duration, &replay.GameVersion, &replay.PlayedAt, &replay.UploadedAt)
		if err != nil {
			return nil, err
		}
		players, err := r.GetGamePlayersByReplayID(replay.ID)
		if err != nil {
			return nil, err
		}
		replay.GamePlayers = players
		replays = append(replays, replay)
	}
	return replays, rows.Err()
}

// ============== Goal Methods ==============

// CreateGoal erstellt ein neues Ziel
func (r *Repository) CreateGoal(goal *models.Goal) error {
	result, err := r.db.Exec(
		`INSERT INTO goals (user_id, goal_type, metric_name, target_value, comparison, deadline)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		goal.UserID, goal.GoalType, goal.MetricName, goal.TargetValue, goal.Comparison, goal.Deadline,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	goal.ID = id
	goal.Status = "active"
	goal.CreatedAt = time.Now()
	return nil
}

// GetActiveGoals gibt alle aktiven Ziele eines Benutzers zurück
func (r *Repository) GetActiveGoals(userID int64) ([]models.Goal, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, goal_type, metric_name, target_value, comparison, current_value, status, created_at, deadline
		 FROM goals WHERE user_id = ? AND status = 'active'
		 ORDER BY deadline ASC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var goals []models.Goal
	for rows.Next() {
		var g models.Goal
		err := rows.Scan(&g.ID, &g.UserID, &g.GoalType, &g.MetricName, &g.TargetValue,
			&g.Comparison, &g.CurrentValue, &g.Status, &g.CreatedAt, &g.Deadline)
		if err != nil {
			return nil, err
		}
		goals = append(goals, g)
	}
	return goals, rows.Err()
}

// UpdateGoalProgress aktualisiert den aktuellen Wert eines Ziels
func (r *Repository) UpdateGoalProgress(goalID int64, currentValue float64) error {
	_, err := r.db.Exec(
		`UPDATE goals SET current_value = ? WHERE id = ?`,
		currentValue, goalID,
	)
	return err
}

// UpdateGoalStatus aktualisiert den Status eines Ziels
func (r *Repository) UpdateGoalStatus(goalID int64, status string) error {
	_, err := r.db.Exec(
		`UPDATE goals SET status = ? WHERE id = ?`,
		status, goalID,
	)
	return err
}

// CheckAndUpdateExpiredGoals prüft und aktualisiert abgelaufene Ziele
func (r *Repository) CheckAndUpdateExpiredGoals(userID int64) error {
	// Hole alle aktiven Ziele, deren Deadline vorbei ist
	rows, err := r.db.Query(
		`SELECT id, target_value, comparison, current_value
		 FROM goals WHERE user_id = ? AND status = 'active' AND deadline < CURRENT_TIMESTAMP`,
		userID,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var g models.Goal
		err := rows.Scan(&g.ID, &g.TargetValue, &g.Comparison, &g.CurrentValue)
		if err != nil {
			return err
		}

		// Prüfe ob das Ziel erreicht wurde
		status := "failed"
		if g.IsAchieved() {
			status = "completed"
		}
		if err := r.UpdateGoalStatus(g.ID, status); err != nil {
			return err
		}
	}
	return rows.Err()
}

// ============== Daily Progress Methods ==============

// GetOrCreateDailyProgress holt oder erstellt den täglichen Fortschritt
func (r *Repository) GetOrCreateDailyProgress(userID int64, date time.Time) (*models.DailyProgress, error) {
	dateStr := date.Format("2006-01-02")

	var dp models.DailyProgress
	err := r.db.QueryRow(
		`SELECT id, user_id, date, games_played, wins, losses, avg_apm, avg_spending_quotient, avg_supply_block_pct, total_play_time
		 FROM daily_progress WHERE user_id = ? AND date = ?`,
		userID, dateStr,
	).Scan(&dp.ID, &dp.UserID, &dp.Date, &dp.GamesPlayed, &dp.Wins, &dp.Losses,
		&dp.AvgAPM, &dp.AvgSpendingQuotient, &dp.AvgSupplyBlockPct, &dp.TotalPlayTime)

	if err == sql.ErrNoRows {
		// Erstelle neuen Eintrag
		result, err := r.db.Exec(
			`INSERT INTO daily_progress (user_id, date) VALUES (?, ?)`,
			userID, dateStr,
		)
		if err != nil {
			return nil, err
		}
		id, err := result.LastInsertId()
		if err != nil {
			return nil, err
		}
		return &models.DailyProgress{
			ID:     id,
			UserID: userID,
			Date:   date,
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return &dp, nil
}

// UpdateDailyProgress aktualisiert den täglichen Fortschritt
func (r *Repository) UpdateDailyProgress(dp *models.DailyProgress) error {
	_, err := r.db.Exec(
		`UPDATE daily_progress SET
			games_played = ?, wins = ?, losses = ?,
			avg_apm = ?, avg_spending_quotient = ?, avg_supply_block_pct = ?,
			total_play_time = ?
		 WHERE id = ?`,
		dp.GamesPlayed, dp.Wins, dp.Losses,
		dp.AvgAPM, dp.AvgSpendingQuotient, dp.AvgSupplyBlockPct,
		dp.TotalPlayTime, dp.ID,
	)
	return err
}

// GetProgressHistory gibt die Fortschrittshistorie zurück
func (r *Repository) GetProgressHistory(userID int64, days int) ([]models.DailyProgress, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, date, games_played, wins, losses, avg_apm, avg_spending_quotient, avg_supply_block_pct, total_play_time
		 FROM daily_progress
		 WHERE user_id = ? AND date >= date('now', '-' || ? || ' days')
		 ORDER BY date ASC`,
		userID, days,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var progress []models.DailyProgress
	for rows.Next() {
		var dp models.DailyProgress
		err := rows.Scan(&dp.ID, &dp.UserID, &dp.Date, &dp.GamesPlayed, &dp.Wins, &dp.Losses,
			&dp.AvgAPM, &dp.AvgSpendingQuotient, &dp.AvgSupplyBlockPct, &dp.TotalPlayTime)
		if err != nil {
			return nil, err
		}
		progress = append(progress, dp)
	}
	return progress, rows.Err()
}

// ============== Weekly Report Methods ==============

// GetWeeklyReport holt den Wochenbericht
func (r *Repository) GetWeeklyReport(userID int64, weekStart time.Time) (*models.WeeklyReport, error) {
	var wr models.WeeklyReport
	var improvements, regressions, strengths, weaknesses sql.NullString
	var mainRace sql.NullString

	err := r.db.QueryRow(
		`SELECT id, user_id, week_start, week_end, total_games, wins, losses, win_rate,
		        avg_apm, avg_sq, avg_supply_block, main_race, total_play_time,
		        improvements, regressions, focus_suggestion, strengths, weaknesses, generated_at
		 FROM weekly_reports WHERE user_id = ? AND week_start = ?`,
		userID, weekStart.Format("2006-01-02"),
	).Scan(&wr.ID, &wr.UserID, &wr.WeekStart, &wr.WeekEnd, &wr.TotalGames, &wr.Wins, &wr.Losses,
		&wr.WinRate, &wr.AvgAPM, &wr.AvgSQ, &wr.AvgSupplyBlock, &mainRace, &wr.TotalPlayTime,
		&improvements, &regressions, &wr.FocusSuggestion, &strengths, &weaknesses, &wr.GeneratedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if mainRace.Valid {
		wr.MainRace = mainRace.String
	}
	if improvements.Valid {
		wr.Improvements = json.RawMessage(improvements.String)
	}
	if regressions.Valid {
		wr.Regressions = json.RawMessage(regressions.String)
	}
	if strengths.Valid {
		wr.Strengths = json.RawMessage(strengths.String)
	}
	if weaknesses.Valid {
		wr.Weaknesses = json.RawMessage(weaknesses.String)
	}

	return &wr, nil
}

// SaveWeeklyReport speichert einen Wochenbericht
func (r *Repository) SaveWeeklyReport(wr *models.WeeklyReport) error {
	var improvements, regressions, strengths, weaknesses *string
	if wr.Improvements != nil {
		s := string(wr.Improvements)
		improvements = &s
	}
	if wr.Regressions != nil {
		s := string(wr.Regressions)
		regressions = &s
	}
	if wr.Strengths != nil {
		s := string(wr.Strengths)
		strengths = &s
	}
	if wr.Weaknesses != nil {
		s := string(wr.Weaknesses)
		weaknesses = &s
	}

	result, err := r.db.Exec(
		`INSERT OR REPLACE INTO weekly_reports
		 (user_id, week_start, week_end, total_games, wins, losses, win_rate,
		  avg_apm, avg_sq, avg_supply_block, main_race, total_play_time,
		  improvements, regressions, focus_suggestion, strengths, weaknesses)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		wr.UserID, wr.WeekStart.Format("2006-01-02"), wr.WeekEnd.Format("2006-01-02"),
		wr.TotalGames, wr.Wins, wr.Losses, wr.WinRate,
		wr.AvgAPM, wr.AvgSQ, wr.AvgSupplyBlock, wr.MainRace, wr.TotalPlayTime,
		improvements, regressions, wr.FocusSuggestion, strengths, weaknesses,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	wr.ID = id
	return nil
}

// GenerateWeeklyReport generiert einen neuen Wochenbericht aus den täglichen Daten
func (r *Repository) GenerateWeeklyReport(userID int64, weekStart time.Time) (*models.WeeklyReport, error) {
	weekEnd := weekStart.AddDate(0, 0, 6)

	// Aggregiere die täglichen Fortschrittsdaten
	var totalGames, wins, losses, totalPlayTime int
	var sumAPM, sumSQ, sumSupplyBlock float64
	var gamesWithMetrics int

	rows, err := r.db.Query(
		`SELECT games_played, wins, losses, avg_apm, avg_spending_quotient, avg_supply_block_pct, total_play_time
		 FROM daily_progress
		 WHERE user_id = ? AND date >= ? AND date <= ?`,
		userID, weekStart.Format("2006-01-02"), weekEnd.Format("2006-01-02"),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var dp models.DailyProgress
		err := rows.Scan(&dp.GamesPlayed, &dp.Wins, &dp.Losses, &dp.AvgAPM,
			&dp.AvgSpendingQuotient, &dp.AvgSupplyBlockPct, &dp.TotalPlayTime)
		if err != nil {
			return nil, err
		}
		totalGames += dp.GamesPlayed
		wins += dp.Wins
		losses += dp.Losses
		totalPlayTime += dp.TotalPlayTime
		if dp.GamesPlayed > 0 {
			sumAPM += dp.AvgAPM * float64(dp.GamesPlayed)
			sumSQ += dp.AvgSpendingQuotient * float64(dp.GamesPlayed)
			sumSupplyBlock += dp.AvgSupplyBlockPct * float64(dp.GamesPlayed)
			gamesWithMetrics += dp.GamesPlayed
		}
	}

	var winRate, avgAPM, avgSQ, avgSupplyBlock float64
	if totalGames > 0 {
		winRate = float64(wins) / float64(totalGames) * 100
	}
	if gamesWithMetrics > 0 {
		avgAPM = sumAPM / float64(gamesWithMetrics)
		avgSQ = sumSQ / float64(gamesWithMetrics)
		avgSupplyBlock = sumSupplyBlock / float64(gamesWithMetrics)
	}

	// Ermittle die Hauptrasse
	mainRace := r.getMainRaceForPeriod(userID, weekStart, weekEnd)

	// Berechne Verbesserungen/Verschlechterungen gegenüber Vorwoche
	prevWeekStart := weekStart.AddDate(0, 0, -7)
	prevReport, _ := r.GetWeeklyReport(userID, prevWeekStart)

	var improvements, regressions map[string]string
	var focusSuggestion string
	var strengthsList, weaknessesList []string

	if prevReport != nil && prevReport.TotalGames > 0 {
		improvements = make(map[string]string)
		regressions = make(map[string]string)

		// APM Vergleich
		if avgAPM > prevReport.AvgAPM*1.05 {
			change := ((avgAPM - prevReport.AvgAPM) / prevReport.AvgAPM) * 100
			improvements["apm"] = fmt.Sprintf("+%.1f%%", change)
		} else if avgAPM < prevReport.AvgAPM*0.95 {
			change := ((prevReport.AvgAPM - avgAPM) / prevReport.AvgAPM) * 100
			regressions["apm"] = fmt.Sprintf("-%.1f%%", change)
		}

		// SQ Vergleich
		if avgSQ > prevReport.AvgSQ*1.05 {
			change := ((avgSQ - prevReport.AvgSQ) / prevReport.AvgSQ) * 100
			improvements["sq"] = fmt.Sprintf("+%.1f%%", change)
		} else if avgSQ < prevReport.AvgSQ*0.95 {
			change := ((prevReport.AvgSQ - avgSQ) / prevReport.AvgSQ) * 100
			regressions["sq"] = fmt.Sprintf("-%.1f%%", change)
		}

		// Supply Block Vergleich (weniger ist besser)
		if avgSupplyBlock < prevReport.AvgSupplyBlock*0.95 {
			change := ((prevReport.AvgSupplyBlock - avgSupplyBlock) / prevReport.AvgSupplyBlock) * 100
			improvements["supply_block"] = fmt.Sprintf("-%.1f%%", change)
		} else if avgSupplyBlock > prevReport.AvgSupplyBlock*1.05 {
			change := ((avgSupplyBlock - prevReport.AvgSupplyBlock) / prevReport.AvgSupplyBlock) * 100
			regressions["supply_block"] = fmt.Sprintf("+%.1f%%", change)
		}

		// Win Rate Vergleich
		if winRate > prevReport.WinRate+5 {
			improvements["win_rate"] = fmt.Sprintf("+%.1f%%", winRate-prevReport.WinRate)
		} else if winRate < prevReport.WinRate-5 {
			regressions["win_rate"] = fmt.Sprintf("-%.1f%%", prevReport.WinRate-winRate)
		}
	}

	// Bestimme Stärken und Schwächen
	if avgAPM > 100 {
		strengthsList = append(strengthsList, "Gute APM")
	} else if avgAPM < 60 {
		weaknessesList = append(weaknessesList, "APM verbessern")
	}

	if avgSQ > 70 {
		strengthsList = append(strengthsList, "Gutes Ressourcen-Management")
	} else if avgSQ < 50 {
		weaknessesList = append(weaknessesList, "Ressourcen schneller ausgeben")
	}

	if avgSupplyBlock < 10 {
		strengthsList = append(strengthsList, "Wenige Supply Blocks")
	} else if avgSupplyBlock > 20 {
		weaknessesList = append(weaknessesList, "Supply Blocks reduzieren")
	}

	// Fokus-Empfehlung basierend auf größter Schwäche
	if avgSupplyBlock > 15 {
		focusSuggestion = "Fokussiere dich diese Woche auf das Vermeiden von Supply Blocks. Baue frühzeitig Supply-Gebäude."
	} else if avgSQ < 60 {
		focusSuggestion = "Fokussiere dich auf dein Macro: Halte deine Ressourcen niedrig und produziere kontinuierlich."
	} else if avgAPM < 80 {
		focusSuggestion = "Arbeite an deiner Spielgeschwindigkeit. Nutze Hotkeys und übe deine Makro-Zyklen."
	} else {
		focusSuggestion = "Du machst gute Fortschritte! Konzentriere dich auf Timing-Angriffe und strategische Entscheidungen."
	}

	wr := &models.WeeklyReport{
		UserID:          userID,
		WeekStart:       weekStart,
		WeekEnd:         weekEnd,
		TotalGames:      totalGames,
		Wins:            wins,
		Losses:          losses,
		WinRate:         winRate,
		AvgAPM:          avgAPM,
		AvgSQ:           avgSQ,
		AvgSupplyBlock:  avgSupplyBlock,
		MainRace:        mainRace,
		TotalPlayTime:   totalPlayTime,
		FocusSuggestion: focusSuggestion,
		GeneratedAt:     time.Now(),
	}

	// Konvertiere Maps zu JSON
	if len(improvements) > 0 {
		if data, err := json.Marshal(improvements); err == nil {
			wr.Improvements = data
		}
	}
	if len(regressions) > 0 {
		if data, err := json.Marshal(regressions); err == nil {
			wr.Regressions = data
		}
	}
	if len(strengthsList) > 0 {
		if data, err := json.Marshal(strengthsList); err == nil {
			wr.Strengths = data
		}
	}
	if len(weaknessesList) > 0 {
		if data, err := json.Marshal(weaknessesList); err == nil {
			wr.Weaknesses = data
		}
	}

	// Speichere den Bericht
	if err := r.SaveWeeklyReport(wr); err != nil {
		return nil, err
	}

	return wr, nil
}

// getMainRaceForPeriod ermittelt die meistgespielte Rasse in einem Zeitraum
func (r *Repository) getMainRaceForPeriod(userID int64, start, end time.Time) string {
	var race sql.NullString
	r.db.QueryRow(
		`SELECT gp.race
		 FROM game_players gp
		 JOIN replays r ON r.id = gp.replay_id
		 JOIN user_replays ur ON ur.replay_id = r.id
		 JOIN players p ON p.id = gp.player_id
		 JOIN users u ON u.sc2_player_name = p.name
		 WHERE ur.user_id = ? AND r.played_at >= ? AND r.played_at <= ?
		 GROUP BY gp.race
		 ORDER BY COUNT(*) DESC
		 LIMIT 1`,
		userID, start.Format("2006-01-02"), end.Format("2006-01-02"),
	).Scan(&race)

	if race.Valid {
		return race.String
	}
	return ""
}

// ============== Coaching Focus Methods ==============

// GetActiveCoachingFocus holt den aktiven Coaching-Fokus
func (r *Repository) GetActiveCoachingFocus(userID int64) (*models.CoachingFocus, error) {
	var cf models.CoachingFocus
	var description sql.NullString

	err := r.db.QueryRow(
		`SELECT id, user_id, focus_area, description, started_at, active
		 FROM coaching_focus WHERE user_id = ? AND active = 1
		 ORDER BY started_at DESC LIMIT 1`,
		userID,
	).Scan(&cf.ID, &cf.UserID, &cf.FocusArea, &description, &cf.StartedAt, &cf.Active)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if description.Valid {
		cf.Description = description.String
	}
	return &cf, nil
}

// SetCoachingFocus setzt einen neuen Coaching-Fokus
func (r *Repository) SetCoachingFocus(userID int64, focusArea, description string) (*models.CoachingFocus, error) {
	// Deaktiviere alte Fokus-Einträge
	_, err := r.db.Exec(
		`UPDATE coaching_focus SET active = 0 WHERE user_id = ?`,
		userID,
	)
	if err != nil {
		return nil, err
	}

	// Erstelle neuen Fokus
	result, err := r.db.Exec(
		`INSERT INTO coaching_focus (user_id, focus_area, description) VALUES (?, ?, ?)`,
		userID, focusArea, description,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &models.CoachingFocus{
		ID:          id,
		UserID:      userID,
		FocusArea:   focusArea,
		Description: description,
		StartedAt:   time.Now(),
		Active:      true,
	}, nil
}

// ============== Dashboard Data Methods ==============

// GetWeekStats berechnet die Wochenstatistiken
func (r *Repository) GetWeekStats(userID int64) (*models.WeekStats, error) {
	now := time.Now()
	weekStart := now.AddDate(0, 0, -int(now.Weekday()))
	prevWeekStart := weekStart.AddDate(0, 0, -7)

	// Aktuelle Woche
	current := &models.WeekStats{}
	rows, err := r.db.Query(
		`SELECT COALESCE(SUM(games_played), 0), COALESCE(SUM(wins), 0), COALESCE(SUM(losses), 0),
		        COALESCE(AVG(CASE WHEN games_played > 0 THEN avg_apm END), 0),
		        COALESCE(AVG(CASE WHEN games_played > 0 THEN avg_spending_quotient END), 0),
		        COALESCE(AVG(CASE WHEN games_played > 0 THEN avg_supply_block_pct END), 0),
		        COALESCE(SUM(total_play_time), 0)
		 FROM daily_progress
		 WHERE user_id = ? AND date >= ?`,
		userID, weekStart.Format("2006-01-02"),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&current.GamesPlayed, &current.Wins, &current.Losses,
			&current.AvgAPM, &current.AvgSQ, &current.AvgSupplyBlock, &current.TotalPlayTime)
		if err != nil {
			return nil, err
		}
	}

	if current.GamesPlayed > 0 {
		current.WinRate = float64(current.Wins) / float64(current.GamesPlayed) * 100
	}

	// Vorwoche für Vergleich
	var prevGames, prevWins int
	var prevAPM, prevSQ, prevSupplyBlock float64
	r.db.QueryRow(
		`SELECT COALESCE(SUM(games_played), 0), COALESCE(SUM(wins), 0),
		        COALESCE(AVG(CASE WHEN games_played > 0 THEN avg_apm END), 0),
		        COALESCE(AVG(CASE WHEN games_played > 0 THEN avg_spending_quotient END), 0),
		        COALESCE(AVG(CASE WHEN games_played > 0 THEN avg_supply_block_pct END), 0)
		 FROM daily_progress
		 WHERE user_id = ? AND date >= ? AND date < ?`,
		userID, prevWeekStart.Format("2006-01-02"), weekStart.Format("2006-01-02"),
	).Scan(&prevGames, &prevWins, &prevAPM, &prevSQ, &prevSupplyBlock)

	// Berechne Änderungen
	if prevAPM > 0 {
		current.APMChange = ((current.AvgAPM - prevAPM) / prevAPM) * 100
	}
	if prevSQ > 0 {
		current.SQChange = ((current.AvgSQ - prevSQ) / prevSQ) * 100
	}
	if prevSupplyBlock > 0 {
		current.SupplyBlockChange = ((current.AvgSupplyBlock - prevSupplyBlock) / prevSupplyBlock) * 100
	}
	if prevGames > 0 {
		prevWinRate := float64(prevWins) / float64(prevGames) * 100
		current.WinRateChange = current.WinRate - prevWinRate
	}

	return current, nil
}

// GetRecentGames holt die letzten Spiele eines Benutzers
func (r *Repository) GetRecentGames(userID int64, limit int) ([]models.RecentGame, error) {
	rows, err := r.db.Query(
		`SELECT r.id, r.map, gp.result, gp.race, gp.apm, gp.spending_quotient, r.duration, r.played_at
		 FROM replays r
		 JOIN user_replays ur ON ur.replay_id = r.id
		 JOIN game_players gp ON gp.replay_id = r.id
		 JOIN players p ON p.id = gp.player_id
		 JOIN users u ON u.sc2_player_name = p.name AND u.id = ur.user_id
		 WHERE ur.user_id = ?
		 ORDER BY r.played_at DESC
		 LIMIT ?`,
		userID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []models.RecentGame
	for rows.Next() {
		var g models.RecentGame
		err := rows.Scan(&g.ReplayID, &g.Map, &g.Result, &g.Race, &g.APM, &g.SQ, &g.Duration, &g.PlayedAt)
		if err != nil {
			return nil, err
		}

		// Hole Gegner-Rasse
		r.db.QueryRow(
			`SELECT gp.race FROM game_players gp
			 JOIN players p ON p.id = gp.player_id
			 WHERE gp.replay_id = ? AND p.name != (SELECT sc2_player_name FROM users WHERE id = ?)
			 LIMIT 1`,
			g.ReplayID, userID,
		).Scan(&g.EnemyRace)

		games = append(games, g)
	}
	return games, rows.Err()
}

// UpdateProgressFromReplay aktualisiert den Fortschritt basierend auf einem neuen Replay
func (r *Repository) UpdateProgressFromReplay(userID int64, replay *models.Replay, playerMetrics *models.GamePlayer, supplyBlockPct float64) error {
	date := replay.PlayedAt
	dp, err := r.GetOrCreateDailyProgress(userID, date)
	if err != nil {
		return err
	}

	// Inkrementiere Spiele
	dp.GamesPlayed++
	if playerMetrics.Result == "Win" {
		dp.Wins++
	} else if playerMetrics.Result == "Loss" {
		dp.Losses++
	}

	// Berechne neue Durchschnitte
	oldTotal := float64(dp.GamesPlayed - 1)
	newTotal := float64(dp.GamesPlayed)

	if oldTotal > 0 {
		dp.AvgAPM = (dp.AvgAPM*oldTotal + playerMetrics.APM) / newTotal
		dp.AvgSpendingQuotient = (dp.AvgSpendingQuotient*oldTotal + playerMetrics.SpendingQuotient) / newTotal
		dp.AvgSupplyBlockPct = (dp.AvgSupplyBlockPct*oldTotal + supplyBlockPct) / newTotal
	} else {
		dp.AvgAPM = playerMetrics.APM
		dp.AvgSpendingQuotient = playerMetrics.SpendingQuotient
		dp.AvgSupplyBlockPct = supplyBlockPct
	}

	dp.TotalPlayTime += replay.Duration

	if err := r.UpdateDailyProgress(dp); err != nil {
		return err
	}

	// Aktualisiere aktive Ziele
	return r.updateGoalsFromProgress(userID, dp)
}

// updateGoalsFromProgress aktualisiert Ziele basierend auf Fortschritt
func (r *Repository) updateGoalsFromProgress(userID int64, dp *models.DailyProgress) error {
	goals, err := r.GetActiveGoals(userID)
	if err != nil {
		return err
	}

	for _, goal := range goals {
		var currentValue float64

		switch goal.MetricName {
		case "games_played":
			if goal.GoalType == "daily" {
				currentValue = float64(dp.GamesPlayed)
			} else {
				// Wöchentliche Spiele summieren
				weekStats, _ := r.GetWeekStats(userID)
				if weekStats != nil {
					currentValue = float64(weekStats.GamesPlayed)
				}
			}
		case "apm":
			currentValue = dp.AvgAPM
		case "supply_block":
			currentValue = dp.AvgSupplyBlockPct
		case "sq":
			currentValue = dp.AvgSpendingQuotient
		case "win_rate":
			currentValue = dp.WinRate()
		}

		if err := r.UpdateGoalProgress(goal.ID, currentValue); err != nil {
			return err
		}

		// Prüfe ob Ziel erreicht
		goal.CurrentValue = currentValue
		if goal.IsAchieved() {
			r.UpdateGoalStatus(goal.ID, "completed")
		}
	}

	return nil
}
