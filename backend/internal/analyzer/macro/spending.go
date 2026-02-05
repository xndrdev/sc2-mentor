package macro

import (
	"math"

	"sc2-analytics/internal/models"
	"sc2-analytics/internal/parser"
)

// SpendingAnalyzer berechnet den Spending Quotient und Ressourcenmetriken
type SpendingAnalyzer struct{}

// NewSpendingAnalyzer erstellt einen neuen SpendingAnalyzer
func NewSpendingAnalyzer() *SpendingAnalyzer {
	return &SpendingAnalyzer{}
}

// Analyze analysiert das Ressourcen-Spending eines Spielers
func (sa *SpendingAnalyzer) Analyze(events *parser.ParsedEvents, playerID int, gameDuration float64) *models.SpendingAnalysis {
	if events == nil {
		return nil
	}

	analysis := &models.SpendingAnalysis{
		ResourceTimeline: []models.ResourcePoint{},
	}

	var totalMinerals, totalGas float64
	var totalIncome float64
	var dataPoints int

	var lastMineralsRate, lastGasRate float64

	for _, evt := range events.TrackerEvents {
		// Event-Typ prüfen (vereinfachter Name)
		if evt.EventType != "PlayerStats" {
			continue
		}
		if evt.PlayerID != playerID {
			continue
		}

		timeSeconds := parser.LoopsToRealSeconds(evt.Loop)

		// Aktuelle Ressourcen (Werte sind bereits in normalen Einheiten)
		minerals := getScoreValueFloat(evt.Data, "scoreValueMineralsCurrent")
		gas := getScoreValueFloat(evt.Data, "scoreValueVespeneCurrent")

		// Einkommensrate (durch 4096 teilen)
		mineralsRate := getScoreValueFloat(evt.Data, "scoreValueMineralsCollectionRate") / 4096
		gasRate := getScoreValueFloat(evt.Data, "scoreValueVespeneCollectionRate") / 4096

		// Sammle Daten für Durchschnittsberechnung
		totalMinerals += minerals
		totalGas += gas
		totalIncome += mineralsRate + gasRate
		dataPoints++

		// Timeline-Punkt
		analysis.ResourceTimeline = append(analysis.ResourceTimeline, models.ResourcePoint{
			Time:     timeSeconds,
			Minerals: int(minerals),
			Gas:      int(gas),
			Income: models.ResourceValue{
				Minerals: mineralsRate,
				Gas:      gasRate,
			},
		})

		lastMineralsRate = mineralsRate
		lastGasRate = gasRate
	}

	// Dummy-Check für Compiler-Warnung
	_ = lastMineralsRate
	_ = lastGasRate

	// Berechne Durchschnittswerte
	if dataPoints > 0 {
		avgUnspentMinerals := totalMinerals / float64(dataPoints)
		avgUnspentGas := totalGas / float64(dataPoints)
		avgIncome := totalIncome / float64(dataPoints)

		analysis.AverageUnspent = models.ResourceValue{
			Minerals: avgUnspentMinerals,
			Gas:      avgUnspentGas,
		}
		analysis.AverageIncome = models.ResourceValue{
			Minerals: avgIncome * 0.7, // Ungefähre Aufteilung
			Gas:      avgIncome * 0.3,
		}

		// Berechne Spending Quotient
		// SQ = 35 * (0.00137 * avgIncome - ln(avgUnspent + 1)) + 240
		avgUnspent := avgUnspentMinerals + avgUnspentGas
		analysis.SpendingQuotient = calculateSQ(avgIncome, avgUnspent)
		analysis.Rating = rateSQ(analysis.SpendingQuotient)
	}

	return analysis
}

// getScoreValueFloat extrahiert einen Float-Wert aus den Event-Daten
func getScoreValueFloat(data map[string]interface{}, key string) float64 {
	// Stats sind in "stats" verschachtelt (ohne m_ Präfix)
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
		return float64(v)
	case int64:
		return float64(v)
	case float64:
		return v
	}

	return 0
}

// calculateSQ berechnet den Spending Quotient
// Formel: SQ = 35 * (0.00137 * avgIncome - ln(avgUnspent + 1)) + 240
func calculateSQ(avgIncome, avgUnspent float64) float64 {
	if avgIncome <= 0 {
		return 0
	}

	sq := 35*(0.00137*avgIncome-math.Log(avgUnspent+1)) + 240

	// Begrenze auf sinnvollen Bereich
	if sq < 0 {
		sq = 0
	}
	if sq > 200 {
		sq = 200
	}

	return sq
}

// rateSQ bewertet den Spending Quotient
func rateSQ(sq float64) string {
	switch {
	case sq < 70:
		return "poor"
	case sq < 90:
		return "below_average"
	case sq < 110:
		return "average"
	case sq < 130:
		return "good"
	default:
		return "excellent"
	}
}

// GenerateSuggestions erstellt Verbesserungsvorschläge
func (sa *SpendingAnalyzer) GenerateSuggestions(analysis *models.SpendingAnalysis) []models.Suggestion {
	var suggestions []models.Suggestion

	if analysis == nil {
		return suggestions
	}

	// Vorschlag basierend auf SQ
	switch analysis.Rating {
	case "poor":
		suggestions = append(suggestions, models.Suggestion{
			Priority:    "high",
			Category:    "macro",
			Title:       "Ressourcen besser ausgeben",
			Description: "Dein Spending Quotient von %.0f ist niedrig. Baue mehr Produktionsgebäude oder Einheiten.",
			TargetValue: "> 90 SQ",
		})
	case "below_average":
		suggestions = append(suggestions, models.Suggestion{
			Priority:    "medium",
			Category:    "macro",
			Title:       "Spending verbessern",
			Description: "Du hast oft zu viele ungenutzte Ressourcen. Versuche, kontinuierlich zu produzieren.",
			TargetValue: "> 100 SQ",
		})
	}

	// Spezifische Ressourcen-Vorschläge
	if analysis.AverageUnspent.Minerals > 1000 {
		suggestions = append(suggestions, models.Suggestion{
			Priority:    "high",
			Category:    "macro",
			Title:       "Zu viele ungenutzte Mineralien",
			Description: "Du hattest durchschnittlich %.0f ungenutzte Mineralien. Baue mehr Produktionsgebäude.",
			TargetValue: "< 500 Mineralien",
		})
	}

	if analysis.AverageUnspent.Gas > 500 {
		suggestions = append(suggestions, models.Suggestion{
			Priority:    "medium",
			Category:    "macro",
			Title:       "Zu viel ungenutztes Gas",
			Description: "Du hattest durchschnittlich %.0f ungenutztes Gas. Baue mehr gas-intensive Einheiten.",
			TargetValue: "< 300 Gas",
		})
	}

	return suggestions
}
