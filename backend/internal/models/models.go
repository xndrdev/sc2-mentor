package models

import (
	"encoding/json"
	"time"
)

// Player repräsentiert einen SC2 Spieler
type Player struct {
	ID         int64  `json:"id"`
	ToonHandle string `json:"toon_handle"`
	Name       string `json:"name"`
	Region     string `json:"region"`
}

// Replay repräsentiert eine hochgeladene Replay-Datei
type Replay struct {
	ID           int64     `json:"id"`
	Hash         string    `json:"hash"`
	Filename     string    `json:"filename"`
	Map          string    `json:"map"`
	Duration     int       `json:"duration"` // in Sekunden
	GameVersion  string    `json:"game_version"`
	PlayedAt     time.Time `json:"played_at"`
	UploadedAt   time.Time `json:"uploaded_at"`
	GamePlayers  []GamePlayer `json:"players,omitempty"`
}

// GamePlayer verbindet Spieler mit Replays
type GamePlayer struct {
	ReplayID    int64   `json:"replay_id"`
	PlayerID    int64   `json:"player_id"`
	PlayerSlot  int     `json:"player_slot"`
	Name        string  `json:"name"`
	Race        string  `json:"race"`
	Result      string  `json:"result"` // Win, Loss, Undecided
	APM         float64 `json:"apm"`
	SpendingQuotient float64 `json:"spending_quotient"`
	IsHuman     bool    `json:"is_human"`
}

// Analysis enthält die vollständige Analyse eines Spielers in einem Replay
type Analysis struct {
	ID        int64           `json:"id"`
	ReplayID  int64           `json:"replay_id"`
	PlayerID  int64           `json:"player_id"`
	Data      json.RawMessage `json:"data"`
	CreatedAt time.Time       `json:"created_at"`
}

// AnalysisData ist die strukturierte Analyse
type AnalysisData struct {
	SupplyAnalysis     *SupplyAnalysis     `json:"supply_analysis"`
	SpendingAnalysis   *SpendingAnalysis   `json:"spending_analysis"`
	APMAnalysis        *APMAnalysis        `json:"apm_analysis"`
	BuildOrder         []BuildOrderItem    `json:"build_order"`
	InjectAnalysis     *InjectAnalysis     `json:"inject_analysis,omitempty"`
	ProductionAnalysis *ProductionAnalysis `json:"production_analysis,omitempty"`
	ArmyAnalysis       *ArmyAnalysis       `json:"army_analysis,omitempty"`
	Suggestions        []Suggestion        `json:"suggestions"`
}

// SupplyAnalysis enthält Supply Block Informationen
type SupplyAnalysis struct {
	TotalBlockTime    float64       `json:"total_block_time"` // Sekunden
	BlockPercentage   float64       `json:"block_percentage"`
	Blocks            []SupplyBlock `json:"blocks"`
	SupplyTimeline    []SupplyPoint `json:"supply_timeline"`
}

// SupplyBlock repräsentiert einen einzelnen Supply Block
type SupplyBlock struct {
	StartTime  float64 `json:"start_time"` // Sekunden
	EndTime    float64 `json:"end_time"`
	Duration   float64 `json:"duration"`
	Severity   string  `json:"severity"` // low, medium, high
	SupplyUsed int     `json:"supply_used"`
	SupplyMax  int     `json:"supply_max"`
}

// SupplyPoint für Timeline-Darstellung
type SupplyPoint struct {
	Time       float64 `json:"time"`
	SupplyUsed int     `json:"supply_used"`
	SupplyMax  int     `json:"supply_max"`
	IsBlocked  bool    `json:"is_blocked"`
}

// SpendingAnalysis enthält Ressourcen-Spending Informationen
type SpendingAnalysis struct {
	SpendingQuotient    float64         `json:"spending_quotient"`
	Rating              string          `json:"rating"` // poor, average, good, excellent
	AverageUnspent      ResourceValue   `json:"average_unspent"`
	AverageIncome       ResourceValue   `json:"average_income"`
	ResourceTimeline    []ResourcePoint `json:"resource_timeline"`
}

// ResourceValue für Mineralien und Gas
type ResourceValue struct {
	Minerals float64 `json:"minerals"`
	Gas      float64 `json:"gas"`
}

// ResourcePoint für Timeline
type ResourcePoint struct {
	Time      float64 `json:"time"`
	Minerals  int     `json:"minerals"`
	Gas       int     `json:"gas"`
	Income    ResourceValue `json:"income"`
}

// APMAnalysis enthält APM-bezogene Metriken
type APMAnalysis struct {
	AverageAPM  float64    `json:"average_apm"`
	PeakAPM     float64    `json:"peak_apm"`
	EAPM        float64    `json:"eapm"` // Effective APM
	APMTimeline []APMPoint `json:"apm_timeline"`
}

// APMPoint für Timeline
type APMPoint struct {
	Time float64 `json:"time"`
	APM  float64 `json:"apm"`
}

// BuildOrderItem ist ein einzelner Build Order Eintrag
type BuildOrderItem struct {
	Time        float64 `json:"time"`
	Supply      int     `json:"supply"`
	Action      string  `json:"action"`
	UnitOrBuilding string `json:"unit_or_building"`
}

// InjectAnalysis für Zerg
type InjectAnalysis struct {
	Efficiency      float64       `json:"efficiency"`
	TotalInjects    int           `json:"total_injects"`
	MissedInjects   int           `json:"missed_injects"`
	InjectTimeline  []InjectPoint `json:"inject_timeline"`
}

// InjectPoint für Timeline
type InjectPoint struct {
	Time       float64 `json:"time"`
	HatcheryID int     `json:"hatchery_id"`
	Injected   bool    `json:"injected"`
}

// ProductionAnalysis für Produktionsgebäude
type ProductionAnalysis struct {
	Efficiency       float64              `json:"efficiency"`
	IdleTime         float64              `json:"idle_time"`
	IdlePeriods      []ProductionIdlePeriod `json:"idle_periods"`
}

// ProductionIdlePeriod repräsentiert eine Idle-Phase
type ProductionIdlePeriod struct {
	BuildingType string  `json:"building_type"`
	BuildingID   int     `json:"building_id"`
	StartTime    float64 `json:"start_time"`
	EndTime      float64 `json:"end_time"`
	Duration     float64 `json:"duration"`
}

// ArmyAnalysis für Armeewert-Tracking
type ArmyAnalysis struct {
	PeakArmyValue    int              `json:"peak_army_value"`
	ArmyTimeline     []ArmyPoint      `json:"army_timeline"`
	UnitComposition  []UnitCount      `json:"unit_composition"`
}

// ArmyPoint für Timeline
type ArmyPoint struct {
	Time      float64 `json:"time"`
	Value     int     `json:"value"`
	UnitCount int     `json:"unit_count"`
}

// UnitCount für Einheiten-Zusammensetzung
type UnitCount struct {
	UnitType string `json:"unit_type"`
	Count    int    `json:"count"`
	Value    int    `json:"value"`
}

// Suggestion ist ein Verbesserungsvorschlag
type Suggestion struct {
	Priority    string  `json:"priority"` // high, medium, low
	Category    string  `json:"category"` // macro, micro, strategy
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Timestamp   float64 `json:"timestamp,omitempty"` // optionaler Zeitpunkt
	TargetValue string  `json:"target_value,omitempty"` // Zielwert
}

// TrendData für Verbesserungstrends
type TrendData struct {
	Metric     string       `json:"metric"`
	DataPoints []TrendPoint `json:"data_points"`
	Trend      string       `json:"trend"` // improving, stable, declining
	Change     float64      `json:"change"` // prozentuale Änderung
}

// TrendPoint für einzelne Datenpunkte
type TrendPoint struct {
	ReplayID  int64     `json:"replay_id"`
	PlayedAt  time.Time `json:"played_at"`
	Value     float64   `json:"value"`
}

// StrategicAnalysis enthält die vollständige strategische Spielanalyse
type StrategicAnalysis struct {
	Winner            string                   `json:"winner"`
	Loser             string                   `json:"loser"`
	WinnerRace        string                   `json:"winner_race"`
	LoserRace         string                   `json:"loser_race"`
	Matchup           string                   `json:"matchup"`
	MetricsComparison []MetricComparison       `json:"metrics_comparison"`
	SupplyBlocks      []SupplyBlockSummary     `json:"supply_blocks"`
	CriticalMoments   []CriticalMoment         `json:"critical_moments"`
	Problems          []IdentifiedProblem      `json:"problems"`
	MatchupTips       *MatchupTips             `json:"matchup_tips"`
	ImprovementSteps  []ImprovementStep        `json:"improvement_steps"`
	Summary           string                   `json:"summary"`
}

// MetricComparison vergleicht eine Metrik zwischen Spielern
type MetricComparison struct {
	Metric      string  `json:"metric"`
	PlayerValue float64 `json:"player_value"`
	EnemyValue  float64 `json:"enemy_value"`
	IsWorse     bool    `json:"is_worse"`
}

// SupplyBlockSummary für Anzeige
type SupplyBlockSummary struct {
	Time     float64 `json:"time"`
	Duration float64 `json:"duration"`
	Severity string  `json:"severity"`
}

// CriticalMoment repräsentiert einen kritischen Spielmoment
type CriticalMoment struct {
	Time        float64 `json:"time"`
	PlayerLoss  int     `json:"player_loss"`
	EnemyLoss   int     `json:"enemy_loss"`
	Assessment  string  `json:"assessment"`
	IsPositive  bool    `json:"is_positive"`
}

// IdentifiedProblem ist ein erkanntes Problem
type IdentifiedProblem struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
}

// MatchupTips enthält matchup-spezifische Tipps
type MatchupTips struct {
	Opening  []string `json:"opening"`
	MidGame  []string `json:"mid_game"`
	Timing   []string `json:"timing"`
	LateGame []string `json:"late_game"`
}

// ImprovementStep ist ein konkreter Verbesserungsschritt
type ImprovementStep struct {
	Category    string `json:"category"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// ============== Mentoring System Models ==============

// User repräsentiert einen registrierten Benutzer
type User struct {
	ID            int64      `json:"id"`
	Email         string     `json:"email"`
	PasswordHash  string     `json:"-"` // Nie im JSON ausgeben
	SC2PlayerName string     `json:"sc2_player_name"`
	CreatedAt     time.Time  `json:"created_at"`
	LastLogin     *time.Time `json:"last_login,omitempty"`
}

// UserPublic ist die öffentliche Ansicht eines Benutzers (ohne sensible Daten)
type UserPublic struct {
	ID            int64      `json:"id"`
	Email         string     `json:"email"`
	SC2PlayerName string     `json:"sc2_player_name"`
	CreatedAt     time.Time  `json:"created_at"`
	LastLogin     *time.Time `json:"last_login,omitempty"`
}

// ToPublic konvertiert User zu UserPublic
func (u *User) ToPublic() UserPublic {
	return UserPublic{
		ID:            u.ID,
		Email:         u.Email,
		SC2PlayerName: u.SC2PlayerName,
		CreatedAt:     u.CreatedAt,
		LastLogin:     u.LastLogin,
	}
}

// Goal repräsentiert ein Spielerziel
type Goal struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	GoalType     string    `json:"goal_type"`     // 'daily', 'weekly'
	MetricName   string    `json:"metric_name"`   // 'apm', 'supply_block', 'games_played', 'win_rate', 'sq'
	TargetValue  float64   `json:"target_value"`
	Comparison   string    `json:"comparison"`    // '>=', '<=', '>', '<', '='
	CurrentValue float64   `json:"current_value"`
	Status       string    `json:"status"`        // 'active', 'completed', 'failed'
	CreatedAt    time.Time `json:"created_at"`
	Deadline     time.Time `json:"deadline"`
}

// GoalProgress berechnet den Fortschritt in Prozent
func (g *Goal) Progress() float64 {
	if g.TargetValue == 0 {
		return 0
	}
	// Für "kleiner als" Ziele invertieren wir die Logik
	if g.Comparison == "<=" || g.Comparison == "<" {
		if g.CurrentValue <= g.TargetValue {
			return 100
		}
		// Wenn aktuell > target, berechne umgekehrt
		return (g.TargetValue / g.CurrentValue) * 100
	}
	return (g.CurrentValue / g.TargetValue) * 100
}

// IsAchieved prüft ob das Ziel erreicht ist
func (g *Goal) IsAchieved() bool {
	switch g.Comparison {
	case ">=":
		return g.CurrentValue >= g.TargetValue
	case "<=":
		return g.CurrentValue <= g.TargetValue
	case ">":
		return g.CurrentValue > g.TargetValue
	case "<":
		return g.CurrentValue < g.TargetValue
	case "=":
		return g.CurrentValue == g.TargetValue
	default:
		return g.CurrentValue >= g.TargetValue
	}
}

// DailyProgress enthält tägliche Fortschrittsdaten
type DailyProgress struct {
	ID                   int64     `json:"id"`
	UserID               int64     `json:"user_id"`
	Date                 time.Time `json:"date"`
	GamesPlayed          int       `json:"games_played"`
	Wins                 int       `json:"wins"`
	Losses               int       `json:"losses"`
	AvgAPM               float64   `json:"avg_apm"`
	AvgSpendingQuotient  float64   `json:"avg_spending_quotient"`
	AvgSupplyBlockPct    float64   `json:"avg_supply_block_pct"`
	TotalPlayTime        int       `json:"total_play_time"` // in Sekunden
}

// WinRate berechnet die Gewinnrate
func (dp *DailyProgress) WinRate() float64 {
	total := dp.Wins + dp.Losses
	if total == 0 {
		return 0
	}
	return float64(dp.Wins) / float64(total) * 100
}

// WeeklyReport enthält einen Wochenbericht
type WeeklyReport struct {
	ID              int64           `json:"id"`
	UserID          int64           `json:"user_id"`
	WeekStart       time.Time       `json:"week_start"`
	WeekEnd         time.Time       `json:"week_end"`
	TotalGames      int             `json:"total_games"`
	Wins            int             `json:"wins"`
	Losses          int             `json:"losses"`
	WinRate         float64         `json:"win_rate"`
	AvgAPM          float64         `json:"avg_apm"`
	AvgSQ           float64         `json:"avg_sq"`
	AvgSupplyBlock  float64         `json:"avg_supply_block"`
	MainRace        string          `json:"main_race"`
	TotalPlayTime   int             `json:"total_play_time"`
	Improvements    json.RawMessage `json:"improvements,omitempty"`
	Regressions     json.RawMessage `json:"regressions,omitempty"`
	FocusSuggestion string          `json:"focus_suggestion"`
	Strengths       json.RawMessage `json:"strengths,omitempty"`
	Weaknesses      json.RawMessage `json:"weaknesses,omitempty"`
	GeneratedAt     time.Time       `json:"generated_at"`
}

// CoachingFocus repräsentiert den aktuellen Fokusbereich
type CoachingFocus struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	FocusArea   string    `json:"focus_area"` // 'macro', 'micro', 'economy', 'army_control', 'scouting'
	Description string    `json:"description"`
	StartedAt   time.Time `json:"started_at"`
	Active      bool      `json:"active"`
}

// MentorDashboard aggregiert alle Daten für das Frontend-Dashboard
type MentorDashboard struct {
	User          UserPublic      `json:"user"`
	TodayStats    *DailyProgress  `json:"today_stats"`
	WeekStats     *WeekStats      `json:"week_stats"`
	ActiveGoals   []Goal          `json:"active_goals"`
	RecentGames   []RecentGame    `json:"recent_games"`
	CurrentFocus  *CoachingFocus  `json:"current_focus,omitempty"`
	WeeklyReport  *WeeklyReport   `json:"weekly_report,omitempty"`
	ProgressTrend []DailyProgress `json:"progress_trend"` // Letzte 14 Tage
}

// WeekStats enthält aggregierte Wochendaten
type WeekStats struct {
	GamesPlayed     int     `json:"games_played"`
	Wins            int     `json:"wins"`
	Losses          int     `json:"losses"`
	WinRate         float64 `json:"win_rate"`
	AvgAPM          float64 `json:"avg_apm"`
	AvgSQ           float64 `json:"avg_sq"`
	AvgSupplyBlock  float64 `json:"avg_supply_block"`
	TotalPlayTime   int     `json:"total_play_time"`
	// Vergleich zur Vorwoche
	APMChange         float64 `json:"apm_change"`
	SQChange          float64 `json:"sq_change"`
	WinRateChange     float64 `json:"win_rate_change"`
	SupplyBlockChange float64 `json:"supply_block_change"`
}

// RecentGame enthält Kurzinfos zu einem kürzlichen Spiel
type RecentGame struct {
	ReplayID    int64     `json:"replay_id"`
	Map         string    `json:"map"`
	Result      string    `json:"result"`
	Race        string    `json:"race"`
	EnemyRace   string    `json:"enemy_race"`
	APM         float64   `json:"apm"`
	SQ          float64   `json:"sq"`
	Duration    int       `json:"duration"`
	PlayedAt    time.Time `json:"played_at"`
}

// GoalTemplate ist eine vordefinierte Zielvorlage
type GoalTemplate struct {
	Name        string  `json:"name"`
	GoalType    string  `json:"goal_type"`
	MetricName  string  `json:"metric_name"`
	Comparison  string  `json:"comparison"`
	Beginner    float64 `json:"beginner"`
	Advanced    float64 `json:"advanced"`
	Description string  `json:"description"`
}

// GetGoalTemplates gibt vordefinierte Zielvorlagen zurück
func GetGoalTemplates() []GoalTemplate {
	return []GoalTemplate{
		// Tägliche Ziele
		{Name: "Spiele täglich spielen", GoalType: "daily", MetricName: "games_played", Comparison: ">=", Beginner: 3, Advanced: 5, Description: "Spielregelmäßigkeit aufbauen"},
		{Name: "APM halten", GoalType: "daily", MetricName: "apm", Comparison: ">=", Beginner: 80, Advanced: 120, Description: "Durchschnittliche APM pro Tag"},
		{Name: "Supply Blocks minimieren", GoalType: "daily", MetricName: "supply_block", Comparison: "<=", Beginner: 15, Advanced: 8, Description: "Supply Block Prozent unter Zielwert halten"},
		// Wöchentliche Ziele
		{Name: "Win Rate", GoalType: "weekly", MetricName: "win_rate", Comparison: ">=", Beginner: 45, Advanced: 55, Description: "Gewinnrate über der Woche"},
		{Name: "Gesamtspiele", GoalType: "weekly", MetricName: "games_played", Comparison: ">=", Beginner: 15, Advanced: 30, Description: "Anzahl Spiele pro Woche"},
		{Name: "Spending Quotient", GoalType: "weekly", MetricName: "sq", Comparison: ">=", Beginner: 60, Advanced: 80, Description: "Durchschnittlicher SQ über die Woche"},
	}
}

// CreateGoalRequest ist der Request zum Erstellen eines Ziels
type CreateGoalRequest struct {
	GoalType    string  `json:"goal_type"`
	MetricName  string  `json:"metric_name"`
	TargetValue float64 `json:"target_value"`
	Comparison  string  `json:"comparison,omitempty"`
}

// RegisterRequest ist der Request zur Registrierung
type RegisterRequest struct {
	Email         string `json:"email"`
	Password      string `json:"password"`
	SC2PlayerName string `json:"sc2_player_name"`
}

// LoginRequest ist der Request zum Login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse ist die Antwort bei erfolgreicher Authentifizierung
type AuthResponse struct {
	Token string     `json:"token"`
	User  UserPublic `json:"user"`
}
