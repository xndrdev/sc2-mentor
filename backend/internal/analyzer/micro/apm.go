package micro

import (
	"sc2-analytics/internal/models"
	"sc2-analytics/internal/parser"
)

// APMAnalyzer berechnet APM-bezogene Metriken
type APMAnalyzer struct{}

// NewAPMAnalyzer erstellt einen neuen APMAnalyzer
func NewAPMAnalyzer() *APMAnalyzer {
	return &APMAnalyzer{}
}

// Analyze analysiert APM für einen Spieler
func (aa *APMAnalyzer) Analyze(events *parser.ParsedEvents, playerID int, gameDuration float64) *models.APMAnalysis {
	if events == nil || gameDuration <= 0 {
		return nil
	}

	analysis := &models.APMAnalysis{
		APMTimeline: []models.APMPoint{},
	}

	// Sammle Aktionen pro Zeitfenster (30 Sekunden)
	const windowSize = 30.0
	actionWindows := make(map[int]int) // window index -> action count
	var totalActions int
	var effectiveActions int

	// Für EAPM: Tracke letzte Aktion um Spam zu filtern
	var lastActionLoop int
	const minLoopsBetweenActions = 8 // Mindestens 0.5 Sekunden zwischen "echten" Aktionen

	for _, evt := range events.GameEvents {
		if evt.PlayerID != playerID {
			continue
		}

		// Zähle nur relevante Aktionen (vereinfachte Namen)
		if !isCountableAction(evt.EventType) {
			continue
		}

		totalActions++

		// EAPM: Filtere Spam-Aktionen
		if evt.Loop-lastActionLoop >= minLoopsBetweenActions {
			effectiveActions++
		}
		lastActionLoop = evt.Loop

		// Ordne Aktion einem Zeitfenster zu
		timeSeconds := parser.LoopsToRealSeconds(evt.Loop)
		windowIndex := int(timeSeconds / windowSize)
		actionWindows[windowIndex]++
	}

	// Berechne Durchschnitts-APM
	gameDurationMinutes := gameDuration / 60.0
	if gameDurationMinutes > 0 {
		analysis.AverageAPM = float64(totalActions) / gameDurationMinutes
		analysis.EAPM = float64(effectiveActions) / gameDurationMinutes
	}

	// Erstelle Timeline und finde Peak-APM
	var peakAPM float64
	for windowIndex, actions := range actionWindows {
		// Konvertiere zu APM für dieses Fenster
		apm := float64(actions) * (60.0 / windowSize)
		if apm > peakAPM {
			peakAPM = apm
		}

		analysis.APMTimeline = append(analysis.APMTimeline, models.APMPoint{
			Time: float64(windowIndex) * windowSize,
			APM:  apm,
		})
	}
	analysis.PeakAPM = peakAPM

	// Sortiere Timeline nach Zeit
	sortAPMTimeline(analysis.APMTimeline)

	return analysis
}

// isCountableAction prüft ob ein Event als APM-Aktion zählt (vereinfachte Namen)
func isCountableAction(eventType string) bool {
	switch eventType {
	case "Cmd", "CmdUpdateTargetPoint", "CmdUpdateTargetUnit",
		"SelectionDelta", "ControlGroupUpdate":
		return true
	}
	return false
}

// sortAPMTimeline sortiert die Timeline nach Zeit
func sortAPMTimeline(timeline []models.APMPoint) {
	// Simple Bubble Sort (Liste ist klein)
	for i := 0; i < len(timeline)-1; i++ {
		for j := 0; j < len(timeline)-i-1; j++ {
			if timeline[j].Time > timeline[j+1].Time {
				timeline[j], timeline[j+1] = timeline[j+1], timeline[j]
			}
		}
	}
}

// GenerateSuggestions erstellt Verbesserungsvorschläge
func (aa *APMAnalyzer) GenerateSuggestions(analysis *models.APMAnalysis) []models.Suggestion {
	var suggestions []models.Suggestion

	if analysis == nil {
		return suggestions
	}

	// APM-basierte Vorschläge
	if analysis.AverageAPM < 50 {
		suggestions = append(suggestions, models.Suggestion{
			Priority:    "medium",
			Category:    "micro",
			Title:       "APM steigern",
			Description: "Deine APM von %.0f ist niedrig. Übe schnellere Eingaben und Hotkey-Nutzung.",
			TargetValue: "> 80 APM",
		})
	}

	// EAPM vs APM Verhältnis
	if analysis.AverageAPM > 0 {
		eapmRatio := analysis.EAPM / analysis.AverageAPM
		if eapmRatio < 0.6 {
			suggestions = append(suggestions, models.Suggestion{
				Priority:    "low",
				Category:    "micro",
				Title:       "Weniger Spam-Aktionen",
				Description: "Dein EAPM/APM-Verhältnis zeigt viele ineffektive Aktionen. Fokussiere auf sinnvolle Befehle.",
				TargetValue: "> 70% EAPM/APM",
			})
		}
	}

	// APM-Einbrüche erkennen
	if len(analysis.APMTimeline) > 2 {
		avgAPM := analysis.AverageAPM
		for _, point := range analysis.APMTimeline {
			if point.APM < avgAPM*0.3 && point.Time > 120 { // Ignoriere erste 2 Minuten
				suggestions = append(suggestions, models.Suggestion{
					Priority:    "low",
					Category:    "micro",
					Title:       "APM-Einbruch erkannt",
					Description: "Bei %.0f:%02.0f fiel deine APM deutlich ab. Versuche, gleichmäßig aktiv zu bleiben.",
					Timestamp:   point.Time,
				})
				break // Nur einen Vorschlag pro Typ
			}
		}
	}

	return suggestions
}
