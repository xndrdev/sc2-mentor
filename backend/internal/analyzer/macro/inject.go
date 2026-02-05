package macro

import (
	"sc2-analytics/internal/models"
	"sc2-analytics/internal/parser"
	"strings"
)

// InjectAnalyzer analysiert Zerg Injections
type InjectAnalyzer struct{}

// NewInjectAnalyzer erstellt einen neuen InjectAnalyzer
func NewInjectAnalyzer() *InjectAnalyzer {
	return &InjectAnalyzer{}
}

// hatcheryState speichert den Zustand einer Hatchery
type hatcheryState struct {
	ID             int
	LastInjectTime float64
	InjectEndTime  float64
	TotalInjects   int
	MissedInjects  int
}

// Analyze analysiert Inject-Effizienz für Zerg-Spieler
func (ia *InjectAnalyzer) Analyze(events *parser.ParsedEvents, playerID int, race string, gameDuration float64) *models.InjectAnalysis {
	// Nur für Zerg relevant
	if strings.ToLower(race) != "zerg" {
		return nil
	}

	if events == nil {
		return nil
	}

	analysis := &models.InjectAnalysis{
		InjectTimeline: []models.InjectPoint{},
	}

	// Tracke Hatcheries und ihre Inject-Zeiten
	hatcheries := make(map[int]*hatcheryState)

	// Inject-Dauer: ca. 29 Sekunden (bei Faster-Geschwindigkeit)
	const injectDuration = 29.0

	for _, evt := range events.TrackerEvents {
		timeSeconds := parser.LoopsToRealSeconds(evt.Loop)

		switch evt.EventType {
		case "UnitBorn", "UnitDone":
			// Neue Hatchery/Lair/Hive geboren
			unitPlayerID := getUnitPlayerIDInject(evt.Data)
			if unitPlayerID != playerID {
				continue
			}

			unitType := getUnitTypeNameInject(evt.Data)
			if isHatcheryType(unitType) {
				unitTag := getUnitTagInject(evt.Data)
				if _, exists := hatcheries[unitTag]; !exists {
					hatcheries[unitTag] = &hatcheryState{
						ID:             unitTag,
						LastInjectTime: timeSeconds,
						InjectEndTime:  timeSeconds, // Startet "injected" um Anfangs-Bias zu vermeiden
					}
				}
			}

		case "UnitDied":
			// Hatchery zerstört
			unitTag := getUnitTagInject(evt.Data)
			delete(hatcheries, unitTag)
		}
	}

	// Analysiere Game-Events für Inject-Befehle (Spawn Larva)
	for _, evt := range events.GameEvents {
		if evt.PlayerID != playerID {
			continue
		}

		timeSeconds := parser.LoopsToRealSeconds(evt.Loop)

		// Suche nach Spawn Larva Ability (Cmd Events)
		if evt.EventType == "Cmd" {
			abilityID := getAbilityID(evt.Data)
			if isInjectAbilityID(abilityID) {
				// Finde die Ziel-Hatchery (vereinfacht: nimm die mit ältestem Inject)
				var oldestHatch *hatcheryState
				for _, h := range hatcheries {
					if oldestHatch == nil || h.InjectEndTime < oldestHatch.InjectEndTime {
						oldestHatch = h
					}
				}

				if oldestHatch != nil {
					// Prüfe ob Inject verpasst wurde
					if timeSeconds > oldestHatch.InjectEndTime+injectDuration {
						// Zeit zwischen Ende des letzten Injects und neuem Inject
						missedTime := timeSeconds - oldestHatch.InjectEndTime
						missedInjects := int(missedTime / injectDuration)
						oldestHatch.MissedInjects += missedInjects
						analysis.MissedInjects += missedInjects
					}

					oldestHatch.LastInjectTime = timeSeconds
					oldestHatch.InjectEndTime = timeSeconds + injectDuration
					oldestHatch.TotalInjects++
					analysis.TotalInjects++

					analysis.InjectTimeline = append(analysis.InjectTimeline, models.InjectPoint{
						Time:       timeSeconds,
						HatcheryID: oldestHatch.ID,
						Injected:   true,
					})
				}
			}
		}
	}

	// Berechne Effizienz
	if analysis.TotalInjects+analysis.MissedInjects > 0 {
		analysis.Efficiency = float64(analysis.TotalInjects) / float64(analysis.TotalInjects+analysis.MissedInjects) * 100
	}

	return analysis
}

// getUnitPlayerIDInject extrahiert die Spieler-ID (ohne m_ Präfix)
func getUnitPlayerIDInject(data map[string]interface{}) int {
	for _, key := range []string{"controlPlayerId", "upkeepPlayerId"} {
		if val, ok := data[key]; ok {
			switch v := val.(type) {
			case int:
				return v
			case int64:
				return int(v)
			case float64:
				return int(v)
			}
		}
	}
	return 0
}

// getUnitTypeNameInject extrahiert den Einheitentyp (ohne m_ Präfix)
func getUnitTypeNameInject(data map[string]interface{}) string {
	if unitType, ok := data["unitTypeName"]; ok {
		if s, ok := unitType.(string); ok {
			return s
		}
	}
	return ""
}

// getUnitTagInject extrahiert den Unit-Tag (ohne m_ Präfix)
func getUnitTagInject(data map[string]interface{}) int {
	if tag, ok := data["unitTagIndex"]; ok {
		switch v := tag.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		}
	}
	return 0
}

// getAbilityID extrahiert die Ability-ID aus einem Cmd-Event (ohne m_ Präfix)
func getAbilityID(data map[string]interface{}) int {
	if abil, ok := data["abil"]; ok {
		if abilMap, ok := abil.(map[string]interface{}); ok {
			if link, ok := abilMap["abilLink"]; ok {
				switch v := link.(type) {
				case int:
					return v
				case int64:
					return int(v)
				case float64:
					return int(v)
				}
			}
		}
	}
	return 0
}

// isHatcheryType prüft ob es eine Hatchery/Lair/Hive ist
func isHatcheryType(unitType string) bool {
	lowerType := strings.ToLower(unitType)
	return strings.Contains(lowerType, "hatchery") ||
		strings.Contains(lowerType, "lair") ||
		strings.Contains(lowerType, "hive")
}

// isInjectAbilityID prüft ob es eine Spawn Larva Ability ist (nach ID)
func isInjectAbilityID(id int) bool {
	// Spawn Larva hat verschiedene IDs je nach Patch
	switch id {
	case 2731, 2732, 2733, 183, 184, 185: // Bekannte Spawn Larva IDs
		return true
	}
	return false
}

// GenerateSuggestions erstellt Verbesserungsvorschläge für Injects
func (ia *InjectAnalyzer) GenerateSuggestions(analysis *models.InjectAnalysis) []models.Suggestion {
	var suggestions []models.Suggestion

	if analysis == nil {
		return suggestions
	}

	if analysis.Efficiency < 50 {
		suggestions = append(suggestions, models.Suggestion{
			Priority:    "high",
			Category:    "macro",
			Title:       "Inject-Effizienz verbessern",
			Description: "Deine Inject-Effizienz liegt bei nur %.0f%%. Nutze Hotkeys und regelmäßige Inject-Zyklen.",
			TargetValue: "> 80% Effizienz",
		})
	} else if analysis.Efficiency < 70 {
		suggestions = append(suggestions, models.Suggestion{
			Priority:    "medium",
			Category:    "macro",
			Title:       "Injects optimieren",
			Description: "Deine Inject-Effizienz von %.0f%% kann verbessert werden. Trainiere den Inject-Rhythmus.",
			TargetValue: "> 85% Effizienz",
		})
	}

	if analysis.MissedInjects > 10 {
		suggestions = append(suggestions, models.Suggestion{
			Priority:    "high",
			Category:    "macro",
			Title:       "Zu viele verpasste Injects",
			Description: "Du hast ca. %d Injects verpasst. Setze einen Timer oder nutze das Inject-Hotkey-System.",
			TargetValue: "< 5 verpasste Injects",
		})
	}

	return suggestions
}
