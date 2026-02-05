package macro

import (
	"sc2-analytics/internal/models"
	"sc2-analytics/internal/parser"
)

// SupplyAnalyzer erkennt und analysiert Supply Blocks
type SupplyAnalyzer struct{}

// NewSupplyAnalyzer erstellt einen neuen SupplyAnalyzer
func NewSupplyAnalyzer() *SupplyAnalyzer {
	return &SupplyAnalyzer{}
}

// Analyze analysiert Supply Blocks für einen Spieler
func (sa *SupplyAnalyzer) Analyze(events *parser.ParsedEvents, playerID int, gameDuration float64) *models.SupplyAnalysis {
	if events == nil {
		return nil
	}

	analysis := &models.SupplyAnalysis{
		Blocks:         []models.SupplyBlock{},
		SupplyTimeline: []models.SupplyPoint{},
	}

	var currentBlock *models.SupplyBlock
	var lastSupplyUsed, lastSupplyMax int

	for _, evt := range events.TrackerEvents {
		// Event-Typ prüfen (vereinfachter Name in neuer s2prot Version)
		if evt.EventType != "PlayerStats" {
			continue
		}
		if evt.PlayerID != playerID {
			continue
		}

		// Extrahiere Supply-Werte aus stats (ohne m_ Präfix)
		foodUsed := getScoreValue(evt.Data, "scoreValueFoodUsed")
		foodMade := getScoreValue(evt.Data, "scoreValueFoodMade")

		// Konvertiere von 4096er Einheiten
		supplyUsed := foodUsed / 4096
		supplyMax := foodMade / 4096

		// Zeitpunkt in Sekunden
		timeSeconds := parser.LoopsToRealSeconds(evt.Loop)

		// Supply Point für Timeline
		isBlocked := supplyUsed >= supplyMax && supplyMax > 0
		analysis.SupplyTimeline = append(analysis.SupplyTimeline, models.SupplyPoint{
			Time:       timeSeconds,
			SupplyUsed: supplyUsed,
			SupplyMax:  supplyMax,
			IsBlocked:  isBlocked,
		})

		// Supply Block Erkennung
		if isBlocked {
			if currentBlock == nil {
				// Neuer Block beginnt
				currentBlock = &models.SupplyBlock{
					StartTime:  timeSeconds,
					SupplyUsed: supplyUsed,
					SupplyMax:  supplyMax,
				}
			}
		} else {
			if currentBlock != nil {
				// Block endet
				currentBlock.EndTime = timeSeconds
				currentBlock.Duration = currentBlock.EndTime - currentBlock.StartTime
				currentBlock.Severity = classifyBlockSeverity(currentBlock.Duration)
				analysis.Blocks = append(analysis.Blocks, *currentBlock)
				analysis.TotalBlockTime += currentBlock.Duration
				currentBlock = nil
			}
		}

		lastSupplyUsed = supplyUsed
		lastSupplyMax = supplyMax
	}

	// Falls ein Block am Ende des Spiels noch offen ist
	if currentBlock != nil {
		currentBlock.EndTime = gameDuration
		currentBlock.Duration = currentBlock.EndTime - currentBlock.StartTime
		currentBlock.Severity = classifyBlockSeverity(currentBlock.Duration)
		analysis.Blocks = append(analysis.Blocks, *currentBlock)
		analysis.TotalBlockTime += currentBlock.Duration
	}

	// Berechne Prozentsatz der blockierten Zeit
	if gameDuration > 0 {
		analysis.BlockPercentage = (analysis.TotalBlockTime / gameDuration) * 100
	}

	// Dummy-Check für Compiler-Warnung
	_ = lastSupplyUsed
	_ = lastSupplyMax

	return analysis
}

// getScoreValue extrahiert einen Score-Wert aus den Event-Daten
func getScoreValue(data map[string]interface{}, key string) int {
	// Die Stats sind in "stats" verschachtelt (ohne m_ Präfix)
	statsRaw, ok := data["stats"]
	if !ok {
		return 0
	}

	stats, ok := statsRaw.(map[string]interface{})
	if !ok {
		return 0
	}

	val, ok := stats[key]
	if !ok {
		return 0
	}

	switch v := val.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	}

	return 0
}

// classifyBlockSeverity klassifiziert die Schwere eines Supply Blocks
func classifyBlockSeverity(duration float64) string {
	switch {
	case duration < 5:
		return "low"
	case duration < 15:
		return "medium"
	default:
		return "high"
	}
}

// GenerateSuggestions erstellt Verbesserungsvorschläge basierend auf der Analyse
func (sa *SupplyAnalyzer) GenerateSuggestions(analysis *models.SupplyAnalysis) []models.Suggestion {
	var suggestions []models.Suggestion

	if analysis == nil {
		return suggestions
	}

	// Vorschlag basierend auf Gesamtblockzeit
	if analysis.BlockPercentage > 10 {
		suggestions = append(suggestions, models.Suggestion{
			Priority:    "high",
			Category:    "macro",
			Title:       "Zu viele Supply Blocks",
			Description: "Du warst %.1f%% der Spielzeit Supply-blockiert. Baue präventiv Supply-Gebäude.",
			TargetValue: "< 5% Blockzeit",
		})
	} else if analysis.BlockPercentage > 5 {
		suggestions = append(suggestions, models.Suggestion{
			Priority:    "medium",
			Category:    "macro",
			Title:       "Supply Blocks reduzieren",
			Description: "Einige Supply Blocks könnten vermieden werden. Achte auf dein Supply-Limit.",
			TargetValue: "< 5% Blockzeit",
		})
	}

	// Vorschläge für schwere Blocks
	for _, block := range analysis.Blocks {
		if block.Severity == "high" {
			suggestions = append(suggestions, models.Suggestion{
				Priority:    "high",
				Category:    "macro",
				Title:       "Schwerer Supply Block",
				Description: "Supply Block von %.0f Sekunden bei %.0f:%02.0f",
				Timestamp:   block.StartTime,
				TargetValue: "< 5s",
			})
		}
	}

	return suggestions
}
