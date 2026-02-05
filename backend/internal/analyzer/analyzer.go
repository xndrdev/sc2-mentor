package analyzer

import (
	"encoding/json"
	"fmt"

	"sc2-analytics/internal/analyzer/builds"
	"sc2-analytics/internal/analyzer/macro"
	"sc2-analytics/internal/analyzer/micro"
	"sc2-analytics/internal/models"
	"sc2-analytics/internal/parser"
)

// Analyzer koordiniert alle Analyse-Module
type Analyzer struct {
	supplyAnalyzer   *macro.SupplyAnalyzer
	spendingAnalyzer *macro.SpendingAnalyzer
	injectAnalyzer   *macro.InjectAnalyzer
	apmAnalyzer      *micro.APMAnalyzer
	armyAnalyzer     *micro.ArmyAnalyzer
	buildAnalyzer    *builds.BuildOrderAnalyzer
}

// New erstellt einen neuen Analyzer
func New() *Analyzer {
	return &Analyzer{
		supplyAnalyzer:   macro.NewSupplyAnalyzer(),
		spendingAnalyzer: macro.NewSpendingAnalyzer(),
		injectAnalyzer:   macro.NewInjectAnalyzer(),
		apmAnalyzer:      micro.NewAPMAnalyzer(),
		armyAnalyzer:     micro.NewArmyAnalyzer(),
		buildAnalyzer:    builds.NewBuildOrderAnalyzer(),
	}
}

// AnalyzePlayer führt alle Analysen für einen Spieler durch
func (a *Analyzer) AnalyzePlayer(parsedReplay *parser.ParsedReplay, playerSlot int, race string) (*models.AnalysisData, error) {
	if parsedReplay == nil || parsedReplay.Events == nil {
		return nil, fmt.Errorf("keine Events zum Analysieren")
	}

	gameDuration := float64(parsedReplay.Duration)
	events := parsedReplay.Events

	data := &models.AnalysisData{
		Suggestions: []models.Suggestion{},
	}

	// Supply Analyse
	data.SupplyAnalysis = a.supplyAnalyzer.Analyze(events, playerSlot, gameDuration)
	if data.SupplyAnalysis != nil {
		suggestions := a.supplyAnalyzer.GenerateSuggestions(data.SupplyAnalysis)
		data.Suggestions = append(data.Suggestions, suggestions...)
	}

	// Spending Analyse
	data.SpendingAnalysis = a.spendingAnalyzer.Analyze(events, playerSlot, gameDuration)
	if data.SpendingAnalysis != nil {
		suggestions := a.spendingAnalyzer.GenerateSuggestions(data.SpendingAnalysis)
		data.Suggestions = append(data.Suggestions, suggestions...)
	}

	// APM Analyse
	data.APMAnalysis = a.apmAnalyzer.Analyze(events, playerSlot, gameDuration)
	if data.APMAnalysis != nil {
		suggestions := a.apmAnalyzer.GenerateSuggestions(data.APMAnalysis)
		data.Suggestions = append(data.Suggestions, suggestions...)
	}

	// Build Order
	data.BuildOrder = a.buildAnalyzer.Analyze(events, playerSlot)

	// Inject Analyse (nur für Zerg)
	data.InjectAnalysis = a.injectAnalyzer.Analyze(events, playerSlot, race, gameDuration)
	if data.InjectAnalysis != nil {
		suggestions := a.injectAnalyzer.GenerateSuggestions(data.InjectAnalysis)
		data.Suggestions = append(data.Suggestions, suggestions...)
	}

	// Army Analyse
	data.ArmyAnalysis = a.armyAnalyzer.Analyze(events, playerSlot, gameDuration)
	if data.ArmyAnalysis != nil {
		suggestions := a.armyAnalyzer.GenerateSuggestions(data.ArmyAnalysis)
		data.Suggestions = append(data.Suggestions, suggestions...)
	}

	// Sortiere Vorschläge nach Priorität
	sortSuggestions(data.Suggestions)

	return data, nil
}

// AnalyzeAndStore analysiert und speichert die Ergebnisse
func (a *Analyzer) AnalyzeAndStore(parsedReplay *parser.ParsedReplay, replayID int64, players []models.GamePlayer) (map[int64]*models.Analysis, error) {
	results := make(map[int64]*models.Analysis)

	for _, player := range players {
		if !player.IsHuman {
			continue
		}

		analysisData, err := a.AnalyzePlayer(parsedReplay, player.PlayerSlot, player.Race)
		if err != nil {
			continue // Überspringe fehlerhafte Analysen
		}

		// Serialisiere zu JSON
		jsonData, err := json.Marshal(analysisData)
		if err != nil {
			return nil, fmt.Errorf("konnte Analyse nicht serialisieren: %w", err)
		}

		analysis := &models.Analysis{
			ReplayID: replayID,
			PlayerID: player.PlayerID,
			Data:     jsonData,
		}

		results[player.PlayerID] = analysis
	}

	return results, nil
}

// GetPlayerMetrics extrahiert Metriken für game_players Tabelle
func (a *Analyzer) GetPlayerMetrics(parsedReplay *parser.ParsedReplay, playerSlot int, race string) (apm float64, sq float64) {
	if parsedReplay == nil || parsedReplay.Events == nil {
		return 0, 0
	}

	gameDuration := float64(parsedReplay.Duration)

	// APM
	apmAnalysis := a.apmAnalyzer.Analyze(parsedReplay.Events, playerSlot, gameDuration)
	if apmAnalysis != nil {
		apm = apmAnalysis.AverageAPM
	}

	// SQ
	spendingAnalysis := a.spendingAnalyzer.Analyze(parsedReplay.Events, playerSlot, gameDuration)
	if spendingAnalysis != nil {
		sq = spendingAnalysis.SpendingQuotient
	}

	return apm, sq
}

// sortSuggestions sortiert Vorschläge nach Priorität (high > medium > low)
func sortSuggestions(suggestions []models.Suggestion) {
	priorityOrder := map[string]int{
		"high":   0,
		"medium": 1,
		"low":    2,
	}

	for i := 0; i < len(suggestions)-1; i++ {
		for j := 0; j < len(suggestions)-i-1; j++ {
			if priorityOrder[suggestions[j].Priority] > priorityOrder[suggestions[j+1].Priority] {
				suggestions[j], suggestions[j+1] = suggestions[j+1], suggestions[j]
			}
		}
	}
}
