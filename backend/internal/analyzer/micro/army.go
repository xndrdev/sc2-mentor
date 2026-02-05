package micro

import (
	"sc2-analytics/internal/models"
	"sc2-analytics/internal/parser"
	"strings"
)

// ArmyAnalyzer trackt Armeewert und Einheitenkomposition
type ArmyAnalyzer struct{}

// NewArmyAnalyzer erstellt einen neuen ArmyAnalyzer
func NewArmyAnalyzer() *ArmyAnalyzer {
	return &ArmyAnalyzer{}
}

// unitInfo speichert Informationen über eine Einheit
type unitInfo struct {
	UnitType    string
	MineralCost int
	GasCost     int
	IsArmy      bool
}

// Analyze analysiert Armeewert und Komposition
func (aa *ArmyAnalyzer) Analyze(events *parser.ParsedEvents, playerID int, gameDuration float64) *models.ArmyAnalysis {
	if events == nil {
		return nil
	}

	analysis := &models.ArmyAnalysis{
		ArmyTimeline:    []models.ArmyPoint{},
		UnitComposition: []models.UnitCount{},
	}

	// Tracke lebende Einheiten
	livingUnits := make(map[int]*unitInfo) // unitTag -> info
	unitCounts := make(map[string]int)     // unitType -> count

	// Samplen des Armeewerts alle 30 Sekunden
	const sampleInterval = 30.0
	var lastSampleTime float64
	var peakArmyValue int

	for _, evt := range events.TrackerEvents {
		timeSeconds := parser.LoopsToRealSeconds(evt.Loop)

		switch evt.EventType {
		case "UnitBorn", "UnitDone":
			// Prüfe ob die Einheit zum Spieler gehört
			unitPlayerID := getUnitPlayerID(evt.Data)
			if unitPlayerID != playerID {
				continue
			}

			unitType := getUnitTypeFromEvent(evt.Data)
			unitTag := getUnitTagFromEvent(evt.Data)

			if unitType != "" && isArmyUnit(unitType) {
				info := &unitInfo{
					UnitType:    unitType,
					MineralCost: getUnitMineralCost(unitType),
					GasCost:     getUnitGasCost(unitType),
					IsArmy:      true,
				}
				livingUnits[unitTag] = info
				unitCounts[unitType]++
			}

		case "UnitDied":
			unitTag := getUnitTagFromEvent(evt.Data)
			if info, exists := livingUnits[unitTag]; exists {
				unitCounts[info.UnitType]--
				delete(livingUnits, unitTag)
			}
		}

		// Sample Armeewert
		if timeSeconds >= lastSampleTime+sampleInterval {
			armyValue := calculateArmyValue(livingUnits)
			unitCount := len(livingUnits)

			if armyValue > peakArmyValue {
				peakArmyValue = armyValue
			}

			analysis.ArmyTimeline = append(analysis.ArmyTimeline, models.ArmyPoint{
				Time:      timeSeconds,
				Value:     armyValue,
				UnitCount: unitCount,
			})

			lastSampleTime = timeSeconds
		}
	}

	analysis.PeakArmyValue = peakArmyValue

	// Erstelle finale Einheitenkomposition
	for unitType, count := range unitCounts {
		if count > 0 {
			analysis.UnitComposition = append(analysis.UnitComposition, models.UnitCount{
				UnitType: unitType,
				Count:    count,
				Value:    count * (getUnitMineralCost(unitType) + getUnitGasCost(unitType)),
			})
		}
	}

	return analysis
}

// getUnitPlayerID extrahiert die Spieler-ID aus Unit-Events
func getUnitPlayerID(data map[string]interface{}) int {
	// Prüfe verschiedene Feldnamen (ohne m_ Präfix)
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

// getUnitTypeFromEvent extrahiert den Einheitentyp (ohne m_ Präfix)
func getUnitTypeFromEvent(data map[string]interface{}) string {
	if unitType, ok := data["unitTypeName"]; ok {
		if s, ok := unitType.(string); ok {
			return s
		}
	}
	return ""
}

// getUnitTagFromEvent extrahiert den Unit-Tag (ohne m_ Präfix)
func getUnitTagFromEvent(data map[string]interface{}) int {
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

// isArmyUnit prüft ob eine Einheit zur Armee zählt
func isArmyUnit(unitType string) bool {
	// Gebäude und Worker ausschließen
	lowerType := strings.ToLower(unitType)

	// Worker
	workers := []string{"scv", "probe", "drone"}
	for _, w := range workers {
		if strings.Contains(lowerType, w) {
			return false
		}
	}

	// Gebäude haben typischerweise bestimmte Suffixe oder Namen
	buildings := []string{
		"commandcenter", "orbitalcommand", "planetaryfortress",
		"supplydepot", "barracks", "factory", "starport",
		"engineeringbay", "armory", "ghostacademy", "fusioncore",
		"bunker", "missileturret", "sensortower", "refinery",
		"nexus", "pylon", "gateway", "forge", "cyberneticscore",
		"roboticsfacility", "roboticsbay", "stargate", "fleetbeacon",
		"twilightcouncil", "templararchive", "darkshrine",
		"photoncannon", "shieldbattery", "assimilator", "warpgate",
		"hatchery", "lair", "hive", "spawningpool", "evolutionchamber",
		"roachwarren", "banelingnest", "hydraliskden", "lurkerden",
		"spire", "greaterspire", "infestationpit", "ultraliscavern",
		"nydusnetwork", "nyduscanal", "spinecrawler", "sporecrawler",
		"extractor", "creeptumor",
	}

	for _, b := range buildings {
		if strings.Contains(lowerType, b) {
			return false
		}
	}

	// Larva und Eggs ausschließen
	excluded := []string{"larva", "egg", "cocoon", "locust", "broodling", "interceptor"}
	for _, e := range excluded {
		if strings.Contains(lowerType, e) {
			return false
		}
	}

	return true
}

// calculateArmyValue berechnet den Gesamtwert der Armee
func calculateArmyValue(units map[int]*unitInfo) int {
	total := 0
	for _, info := range units {
		total += info.MineralCost + info.GasCost
	}
	return total
}

// getUnitMineralCost gibt die Mineralkosten einer Einheit zurück
func getUnitMineralCost(unitType string) int {
	costs := map[string]int{
		// Terran
		"marine": 50, "marauder": 100, "reaper": 50, "ghost": 150,
		"hellion": 100, "hellbat": 100, "widowmine": 75, "siegetank": 150,
		"cyclone": 150, "thor": 300, "viking": 150, "medivac": 100,
		"liberator": 150, "banshee": 150, "raven": 100, "battlecruiser": 400,
		// Protoss
		"zealot": 100, "stalker": 125, "sentry": 50, "adept": 100,
		"hightemplar": 50, "darktemplar": 125, "archon": 0,
		"observer": 25, "immortal": 275, "colossus": 300, "disruptor": 150,
		"warpprism": 200, "phoenix": 150, "voidray": 250, "oracle": 150,
		"tempest": 250, "carrier": 350, "mothership": 400,
		// Zerg
		"zergling": 25, "baneling": 25, "roach": 75, "ravager": 75,
		"hydralisk": 100, "lurker": 100, "infestor": 100, "swarmhost": 100,
		"ultralisk": 300, "mutalisk": 100, "corruptor": 150,
		"broodlord": 150, "viper": 100, "queen": 150, "overseer": 50,
	}

	lowerType := strings.ToLower(unitType)
	for key, cost := range costs {
		if strings.Contains(lowerType, key) {
			return cost
		}
	}
	return 50 // Default
}

// getUnitGasCost gibt die Gaskosten einer Einheit zurück
func getUnitGasCost(unitType string) int {
	costs := map[string]int{
		// Terran
		"marine": 0, "marauder": 25, "reaper": 50, "ghost": 125,
		"hellion": 0, "hellbat": 0, "widowmine": 25, "siegetank": 125,
		"cyclone": 100, "thor": 200, "viking": 75, "medivac": 100,
		"liberator": 150, "banshee": 100, "raven": 200, "battlecruiser": 300,
		// Protoss
		"zealot": 0, "stalker": 50, "sentry": 100, "adept": 25,
		"hightemplar": 150, "darktemplar": 125, "archon": 0,
		"observer": 75, "immortal": 100, "colossus": 200, "disruptor": 150,
		"warpprism": 0, "phoenix": 100, "voidray": 150, "oracle": 150,
		"tempest": 175, "carrier": 250, "mothership": 300,
		// Zerg
		"zergling": 0, "baneling": 25, "roach": 25, "ravager": 75,
		"hydralisk": 50, "lurker": 100, "infestor": 150, "swarmhost": 75,
		"ultralisk": 200, "mutalisk": 100, "corruptor": 100,
		"broodlord": 150, "viper": 200, "queen": 0, "overseer": 50,
	}

	lowerType := strings.ToLower(unitType)
	for key, cost := range costs {
		if strings.Contains(lowerType, key) {
			return cost
		}
	}
	return 25 // Default
}

// GenerateSuggestions erstellt Verbesserungsvorschläge
func (aa *ArmyAnalyzer) GenerateSuggestions(analysis *models.ArmyAnalysis) []models.Suggestion {
	var suggestions []models.Suggestion

	if analysis == nil {
		return suggestions
	}

	// Prüfe auf Armee-Verluste ohne Gegenwert
	if len(analysis.ArmyTimeline) > 2 {
		for i := 1; i < len(analysis.ArmyTimeline); i++ {
			current := analysis.ArmyTimeline[i]
			previous := analysis.ArmyTimeline[i-1]

			// Signifikanter Wertverlust (> 50% in einem Intervall)
			if previous.Value > 0 {
				lossRatio := float64(previous.Value-current.Value) / float64(previous.Value)
				if lossRatio > 0.5 && previous.Value > 500 {
					suggestions = append(suggestions, models.Suggestion{
						Priority:    "high",
						Category:    "micro",
						Title:       "Große Armeeverluste",
						Description: "Bei %.0f:%02.0f hast du über 50%% deiner Armee verloren. Achte auf besseres Engagement.",
						Timestamp:   current.Time,
					})
					break // Nur einen Vorschlag
				}
			}
		}
	}

	return suggestions
}
