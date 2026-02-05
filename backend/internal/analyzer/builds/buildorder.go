package builds

import (
	"sc2-analytics/internal/models"
	"sc2-analytics/internal/parser"
	"sort"
	"strings"
)

// BuildOrderAnalyzer extrahiert Build Orders
type BuildOrderAnalyzer struct{}

// NewBuildOrderAnalyzer erstellt einen neuen BuildOrderAnalyzer
func NewBuildOrderAnalyzer() *BuildOrderAnalyzer {
	return &BuildOrderAnalyzer{}
}

// buildEvent speichert ein Build-Event mit Zeitstempel
type buildEvent struct {
	Time           float64
	Supply         int
	Action         string
	UnitOrBuilding string
}

// Analyze extrahiert die Build Order eines Spielers
func (ba *BuildOrderAnalyzer) Analyze(events *parser.ParsedEvents, playerID int) []models.BuildOrderItem {
	if events == nil {
		return nil
	}

	var buildEvents []buildEvent
	currentSupply := make(map[int]int) // playerID -> supply

	for _, evt := range events.TrackerEvents {
		timeSeconds := parser.LoopsToRealSeconds(evt.Loop)

		switch evt.EventType {
		case "PlayerStats":
			// Update Supply für alle Spieler
			pid := getPlayerID(evt.Data)
			supply := getSupplyUsed(evt.Data)
			currentSupply[pid] = supply

		case "UnitInit":
			// Gebäude beginnt zu bauen
			if getPlayerIDFromUnit(evt.Data) != playerID {
				continue
			}

			unitType := getUnitTypeName(evt.Data)
			if isBuilding(unitType) {
				buildEvents = append(buildEvents, buildEvent{
					Time:           timeSeconds,
					Supply:         currentSupply[playerID],
					Action:         "Build",
					UnitOrBuilding: formatUnitName(unitType),
				})
			}

		case "UnitBorn":
			// Einheit geboren (für Worker und Morphs)
			if getPlayerIDFromUnit(evt.Data) != playerID {
				continue
			}

			unitType := getUnitTypeName(evt.Data)

			// Nur relevante Einheiten für Build Order
			if isBuildOrderUnit(unitType) {
				action := "Train"
				if isWorker(unitType) {
					action = "Train Worker"
				}

				buildEvents = append(buildEvents, buildEvent{
					Time:           timeSeconds,
					Supply:         currentSupply[playerID],
					Action:         action,
					UnitOrBuilding: formatUnitName(unitType),
				})
			}

		case "Upgrade":
			// Upgrade erforscht
			if evt.PlayerID != playerID {
				continue
			}

			upgradeName := getUpgradeName(evt.Data)
			if upgradeName != "" && !isCosmetic(upgradeName) {
				buildEvents = append(buildEvents, buildEvent{
					Time:           timeSeconds,
					Supply:         currentSupply[playerID],
					Action:         "Upgrade",
					UnitOrBuilding: formatUpgradeName(upgradeName),
				})
			}
		}
	}

	// Sortiere nach Zeit
	sort.Slice(buildEvents, func(i, j int) bool {
		return buildEvents[i].Time < buildEvents[j].Time
	})

	// Konvertiere zu BuildOrderItems (begrenzt auf erste 8 Minuten)
	var result []models.BuildOrderItem
	for _, be := range buildEvents {
		if be.Time > 480 { // 8 Minuten
			break
		}

		result = append(result, models.BuildOrderItem{
			Time:           be.Time,
			Supply:         be.Supply,
			Action:         be.Action,
			UnitOrBuilding: be.UnitOrBuilding,
		})
	}

	return result
}

// getPlayerID extrahiert die Spieler-ID aus Stats-Events (ohne m_ Präfix)
func getPlayerID(data map[string]interface{}) int {
	if pid, ok := data["playerId"]; ok {
		switch v := pid.(type) {
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

// getPlayerIDFromUnit extrahiert die Spieler-ID aus Unit-Events (ohne m_ Präfix)
func getPlayerIDFromUnit(data map[string]interface{}) int {
	for _, key := range []string{"controlPlayerId", "upkeepPlayerId"} {
		if pid, ok := data[key]; ok {
			switch v := pid.(type) {
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

// getSupplyUsed extrahiert das aktuelle Supply (ohne m_ Präfix)
func getSupplyUsed(data map[string]interface{}) int {
	statsRaw, ok := data["stats"]
	if !ok {
		return 0
	}

	stats, ok := statsRaw.(map[string]interface{})
	if !ok {
		return 0
	}

	if val, ok := stats["scoreValueFoodUsed"]; ok {
		switch v := val.(type) {
		case int:
			return v / 4096
		case int64:
			return int(v) / 4096
		case float64:
			return int(v) / 4096
		}
	}
	return 0
}

// getUnitTypeName extrahiert den Einheitentyp (ohne m_ Präfix)
func getUnitTypeName(data map[string]interface{}) string {
	if unitType, ok := data["unitTypeName"]; ok {
		if s, ok := unitType.(string); ok {
			return s
		}
	}
	return ""
}

// getUpgradeName extrahiert den Upgrade-Namen (ohne m_ Präfix)
func getUpgradeName(data map[string]interface{}) string {
	if name, ok := data["upgradeTypeName"]; ok {
		if s, ok := name.(string); ok {
			return s
		}
	}
	return ""
}

// isCosmetic prüft ob es ein kosmetisches Upgrade ist
func isCosmetic(upgradeName string) bool {
	lowerName := strings.ToLower(upgradeName)
	cosmetics := []string{"reward", "dance", "skin", "spray", "voice", "emote"}
	for _, c := range cosmetics {
		if strings.Contains(lowerName, c) {
			return true
		}
	}
	return false
}

// isBuilding prüft ob es ein Gebäude ist
func isBuilding(unitType string) bool {
	lowerType := strings.ToLower(unitType)

	buildings := []string{
		// Terran
		"commandcenter", "orbitalcommand", "planetaryfortress",
		"supplydepot", "barracks", "factory", "starport",
		"engineeringbay", "armory", "ghostacademy", "fusioncore",
		"bunker", "missileturret", "sensortower", "refinery",
		"techlab", "reactor",
		// Protoss
		"nexus", "pylon", "gateway", "forge", "cyberneticscore",
		"roboticsfacility", "roboticsbay", "stargate", "fleetbeacon",
		"twilightcouncil", "templararchive", "darkshrine",
		"photoncannon", "shieldbattery", "assimilator",
		// Zerg
		"hatchery", "lair", "hive", "spawningpool", "evolutionchamber",
		"roachwarren", "banelingnest", "hydraliskden", "lurkerden",
		"spire", "greaterspire", "infestationpit", "ultraliscavern",
		"nydusnetwork", "spinecrawler", "sporecrawler", "extractor",
	}

	for _, b := range buildings {
		if strings.Contains(lowerType, b) {
			return true
		}
	}
	return false
}

// isWorker prüft ob es ein Worker ist
func isWorker(unitType string) bool {
	lowerType := strings.ToLower(unitType)
	return strings.Contains(lowerType, "scv") ||
		strings.Contains(lowerType, "probe") ||
		strings.Contains(lowerType, "drone")
}

// isBuildOrderUnit prüft ob die Einheit für Build Order relevant ist
func isBuildOrderUnit(unitType string) bool {
	lowerType := strings.ToLower(unitType)

	// Exclude unwichtige Einheiten
	excluded := []string{
		"larva", "locust", "broodling", "interceptor", "autoturret",
		"creeptumor", "mule", "changeling", "infested",
	}

	for _, e := range excluded {
		if strings.Contains(lowerType, e) {
			return false
		}
	}

	// Include wichtige Einheiten
	important := []string{
		// Workers
		"scv", "probe", "drone",
		// Key units
		"queen", "zergling", "marine", "zealot",
		"overlord", "overseer",
	}

	for _, i := range important {
		if strings.Contains(lowerType, i) {
			return true
		}
	}

	// Gebäude wurden bereits separat behandelt
	return false
}

// formatUnitName formatiert den Einheitennamen lesbar
func formatUnitName(name string) string {
	// Entferne Präfixe und formatiere
	name = strings.TrimPrefix(name, "Terran")
	name = strings.TrimPrefix(name, "Protoss")
	name = strings.TrimPrefix(name, "Zerg")

	// CamelCase zu Leerzeichen
	var result strings.Builder
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune(' ')
		}
		result.WriteRune(r)
	}

	return strings.TrimSpace(result.String())
}

// formatUpgradeName formatiert den Upgrade-Namen
func formatUpgradeName(name string) string {
	// Typische Upgrade-Namen bereinigen
	name = strings.ReplaceAll(name, "terran", "")
	name = strings.ReplaceAll(name, "protoss", "")
	name = strings.ReplaceAll(name, "zerg", "")

	return formatUnitName(name)
}
