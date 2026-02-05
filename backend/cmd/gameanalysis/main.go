package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"sc2-analytics/internal/analyzer"
	"sc2-analytics/internal/models"
	"sc2-analytics/internal/parser"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run ./cmd/gameanalysis <replay-file>")
	}

	filepath := os.Args[1]
	fmt.Printf("╔══════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║           SC2 STRATEGISCHE SPIELANALYSE                      ║\n")
	fmt.Printf("╚══════════════════════════════════════════════════════════════╝\n\n")

	p := parser.New()
	replay, err := p.ParseFile(filepath)
	if err != nil {
		log.Fatalf("Parse error: %v", err)
	}

	fmt.Printf("Map: %s\n", replay.Map)
	fmt.Printf("Dauer: %d:%02d\n", replay.Duration/60, replay.Duration%60)
	fmt.Printf("Datum: %s\n\n", replay.PlayedAt.Format("02.01.2006 15:04"))

	// Spieler-Info
	fmt.Printf("━━━ SPIELER ━━━\n")
	var winner, loser *parser.ParsedPlayer
	for i := range replay.Players {
		pl := &replay.Players[i]
		status := ""
		if pl.Result == "Win" {
			status = " ★ GEWINNER"
			winner = pl
		} else if pl.Result == "Loss" {
			loser = pl
		}
		fmt.Printf("  %s (%s)%s\n", pl.Name, pl.Race, status)
	}
	fmt.Println()

	if loser == nil || winner == nil {
		fmt.Println("Konnte Gewinner/Verlierer nicht bestimmen")
		return
	}

	// Analyse durchführen
	a := analyzer.New()
	loserAnalysis, _ := a.AnalyzePlayer(replay, loser.Slot, loser.Race)
	winnerAnalysis, _ := a.AnalyzePlayer(replay, winner.Slot, winner.Race)

	// Metriken-Vergleich
	fmt.Printf("━━━ METRIKEN-VERGLEICH ━━━\n\n")
	fmt.Printf("%-30s %12s %12s\n", "Metrik", loser.Name, winner.Name)
	fmt.Printf("%s\n", strings.Repeat("─", 56))

	if loserAnalysis.APMAnalysis != nil && winnerAnalysis.APMAnalysis != nil {
		diff := winnerAnalysis.APMAnalysis.AverageAPM - loserAnalysis.APMAnalysis.AverageAPM
		indicator := ""
		if diff > 20 {
			indicator = " ⚠️"
		}
		fmt.Printf("%-30s %12.0f %12.0f%s\n", "APM (Durchschnitt)",
			loserAnalysis.APMAnalysis.AverageAPM, winnerAnalysis.APMAnalysis.AverageAPM, indicator)
		fmt.Printf("%-30s %12.0f %12.0f\n", "EAPM (Effektiv)",
			loserAnalysis.APMAnalysis.EAPM, winnerAnalysis.APMAnalysis.EAPM)
	}

	if loserAnalysis.SpendingAnalysis != nil && winnerAnalysis.SpendingAnalysis != nil {
		fmt.Printf("%-30s %12.0f %12.0f\n", "Spending Quotient",
			loserAnalysis.SpendingAnalysis.SpendingQuotient, winnerAnalysis.SpendingAnalysis.SpendingQuotient)
		fmt.Printf("%-30s %12.0f %12.0f\n", "Ø Ungenutzte Mineralien",
			loserAnalysis.SpendingAnalysis.AverageUnspent.Minerals, winnerAnalysis.SpendingAnalysis.AverageUnspent.Minerals)
		fmt.Printf("%-30s %12.0f %12.0f\n", "Ø Ungenutztes Gas",
			loserAnalysis.SpendingAnalysis.AverageUnspent.Gas, winnerAnalysis.SpendingAnalysis.AverageUnspent.Gas)
	}

	if loserAnalysis.SupplyAnalysis != nil && winnerAnalysis.SupplyAnalysis != nil {
		indicator := ""
		if loserAnalysis.SupplyAnalysis.BlockPercentage > 10 {
			indicator = " ⚠️"
		}
		fmt.Printf("%-30s %11.1f%% %11.1f%%%s\n", "Supply Block Zeit",
			loserAnalysis.SupplyAnalysis.BlockPercentage, winnerAnalysis.SupplyAnalysis.BlockPercentage, indicator)
		fmt.Printf("%-30s %12d %12d\n", "Anzahl Supply Blocks",
			len(loserAnalysis.SupplyAnalysis.Blocks), len(winnerAnalysis.SupplyAnalysis.Blocks))
	}

	if loserAnalysis.ArmyAnalysis != nil && winnerAnalysis.ArmyAnalysis != nil {
		fmt.Printf("%-30s %12d %12d\n", "Peak Armeewert",
			loserAnalysis.ArmyAnalysis.PeakArmyValue, winnerAnalysis.ArmyAnalysis.PeakArmyValue)
	}
	fmt.Println()

	// Build Order Analyse
	fmt.Printf("━━━ BUILD ORDER (erste 5 Min) ━━━\n\n")
	fmt.Printf("▸ %s (%s):\n", loser.Name, loser.Race)
	printBuildOrder(loserAnalysis.BuildOrder, 15)
	fmt.Println()

	fmt.Printf("▸ %s (%s):\n", winner.Name, winner.Race)
	printBuildOrder(winnerAnalysis.BuildOrder, 15)
	fmt.Println()

	// Einheiten-Analyse
	fmt.Printf("━━━ EINHEITEN PRODUZIERT ━━━\n\n")
	analyzeUnits(replay, loser.Slot, winner.Slot, loser.Name, winner.Name)

	// Supply Block Details
	if loserAnalysis.SupplyAnalysis != nil && len(loserAnalysis.SupplyAnalysis.Blocks) > 0 {
		fmt.Printf("━━━ DEINE SUPPLY BLOCKS ━━━\n\n")
		for _, block := range loserAnalysis.SupplyAnalysis.Blocks {
			severity := "leicht"
			if block.Severity == "medium" {
				severity = "mittel"
			} else if block.Severity == "high" {
				severity = "SCHWER ⚠️"
			}
			fmt.Printf("  %s - %.0f Sekunden (%s)\n", formatTime(block.StartTime), block.Duration, severity)
		}
		fmt.Println()
	}

	// Kritische Momente
	fmt.Printf("━━━ KRITISCHE MOMENTE / KÄMPFE ━━━\n\n")
	findCriticalMoments(replay, loser.Slot, winner.Slot)

	// Strategische Empfehlungen
	fmt.Printf("╔══════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║               WAS DU BESSER MACHEN KANNST                    ║\n")
	fmt.Printf("╚══════════════════════════════════════════════════════════════╝\n\n")
	generateStrategicAdvice(replay, loserAnalysis, winnerAnalysis, loser, winner)
}

func printBuildOrder(items []models.BuildOrderItem, limit int) {
	count := 0
	for _, item := range items {
		if item.Time > 300 { // Nur erste 5 Minuten
			break
		}
		if count >= limit {
			fmt.Printf("  ... (und weitere)\n")
			break
		}
		fmt.Printf("  %s [%d] %s %s\n", formatTime(item.Time), item.Supply, item.Action, item.UnitOrBuilding)
		count++
	}
}

func formatTime(seconds float64) string {
	mins := int(seconds) / 60
	secs := int(seconds) % 60
	return fmt.Sprintf("%d:%02d", mins, secs)
}

func analyzeUnits(replay *parser.ParsedReplay, loserSlot, winnerSlot int, loserName, winnerName string) {
	loserUnits := make(map[string]int)
	winnerUnits := make(map[string]int)

	for _, evt := range replay.Events.TrackerEvents {
		if evt.EventType != "UnitBorn" && evt.EventType != "UnitDone" {
			continue
		}

		playerID := getPlayerIDFromData(evt.Data)
		unitType := getUnitTypeFromData(evt.Data)

		if unitType == "" || isWorkerOrBuilding(unitType) {
			continue
		}

		if playerID == loserSlot {
			loserUnits[unitType]++
		} else if playerID == winnerSlot {
			winnerUnits[unitType]++
		}
	}

	fmt.Printf("%-25s %10s %10s\n", "Einheit", loserName, winnerName)
	fmt.Printf("%s\n", strings.Repeat("─", 47))

	// Alle Einheiten sammeln und sortieren
	allUnits := make(map[string]bool)
	for u := range loserUnits {
		allUnits[u] = true
	}
	for u := range winnerUnits {
		allUnits[u] = true
	}

	var unitList []string
	for u := range allUnits {
		unitList = append(unitList, u)
	}
	sort.Strings(unitList)

	for _, u := range unitList {
		l := loserUnits[u]
		w := winnerUnits[u]
		if l > 0 || w > 0 {
			fmt.Printf("%-25s %10d %10d\n", u, l, w)
		}
	}
	fmt.Println()
}

func getPlayerIDFromData(data map[string]interface{}) int {
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

func getUnitTypeFromData(data map[string]interface{}) string {
	if ut, ok := data["unitTypeName"]; ok {
		if s, ok := ut.(string); ok {
			return s
		}
	}
	return ""
}

func isWorkerOrBuilding(unitType string) bool {
	lower := strings.ToLower(unitType)
	excluded := []string{
		"scv", "probe", "drone", "larva", "egg", "cocoon", "mule",
		"commandcenter", "nexus", "hatchery", "lair", "hive",
		"pylon", "supplydepot", "overlord", "overseer",
		"gateway", "barracks", "spawningpool",
		"assimilator", "refinery", "extractor",
		"forge", "engineeringbay", "evolutionchamber",
		"cyberneticscore", "factory", "roachwarren",
		"stargate", "starport", "spire", "greaterspire",
		"roboticsfacility", "roboticsbay", "armory", "hydraliskden",
		"twilightcouncil", "ghostacademy", "banelingnest",
		"templararchive", "fusioncore", "infestationpit",
		"darkshrine", "fleetbeacon", "ultraliscavern",
		"photoncannon", "missileturret", "spinecrawler", "sporecrawler",
		"bunker", "shieldbattery", "nydusnetwork", "lurkerden",
		"techlab", "reactor", "warpgate", "sensortower",
		"creeptumor", "locust", "broodling", "interceptor",
		"orbitalcommand", "planetaryfortress",
	}

	for _, e := range excluded {
		if strings.Contains(lower, e) {
			return true
		}
	}
	return false
}

func findCriticalMoments(replay *parser.ParsedReplay, loserSlot, winnerSlot int) {
	type battleInterval struct {
		startTime    float64
		loserDeaths  int
		winnerDeaths int
	}

	battles := make(map[int]*battleInterval)

	for _, evt := range replay.Events.TrackerEvents {
		if evt.EventType != "UnitDied" {
			continue
		}

		timeSeconds := float64(evt.Loop) / 16.0 / 1.4
		interval := int(timeSeconds / 20) // 20-Sekunden-Intervalle

		if battles[interval] == nil {
			battles[interval] = &battleInterval{startTime: float64(interval) * 20}
		}

		// killerPlayerId zeigt wer getötet hat
		killerID := 0
		if kp, ok := evt.Data["killerPlayerId"]; ok {
			switch v := kp.(type) {
			case int:
				killerID = v
			case int64:
				killerID = int(v)
			case float64:
				killerID = int(v)
			}
		}

		// Wenn der Gewinner getötet hat, hat der Verlierer eine Einheit verloren
		if killerID == winnerSlot {
			battles[interval].loserDeaths++
		} else if killerID == loserSlot {
			battles[interval].winnerDeaths++
		}
	}

	// Signifikante Kämpfe finden
	type significantBattle struct {
		time         float64
		loserDeaths  int
		winnerDeaths int
	}

	var significant []significantBattle
	for _, b := range battles {
		total := b.loserDeaths + b.winnerDeaths
		if total >= 5 {
			significant = append(significant, significantBattle{
				time:         b.startTime,
				loserDeaths:  b.loserDeaths,
				winnerDeaths: b.winnerDeaths,
			})
		}
	}

	sort.Slice(significant, func(i, j int) bool {
		return significant[i].time < significant[j].time
	})

	if len(significant) == 0 {
		fmt.Printf("Keine großen Kämpfe gefunden (wenige Einheitenverluste).\n\n")
		return
	}

	for _, b := range significant {
		outcome := "📈 Vorteil für dich"
		if b.loserDeaths > b.winnerDeaths*2 {
			outcome = "📉 SCHLECHTER TRADE"
		} else if b.loserDeaths > b.winnerDeaths {
			outcome = "📉 Leichter Nachteil"
		} else if b.loserDeaths < b.winnerDeaths {
			outcome = "📈 Guter Trade!"
		}

		fmt.Printf("  %s: Du verlierst %d, Gegner verliert %d → %s\n",
			formatTime(b.time), b.loserDeaths, b.winnerDeaths, outcome)
	}
	fmt.Println()
}

func generateStrategicAdvice(replay *parser.ParsedReplay, loserAnalysis, winnerAnalysis *models.AnalysisData, loser, winner *parser.ParsedPlayer) {
	matchup := strings.ToLower(loser.Race) + "v" + strings.ToLower(winner.Race)

	// Allgemeine Probleme identifizieren
	var problems []string

	if loserAnalysis.SupplyAnalysis != nil && loserAnalysis.SupplyAnalysis.BlockPercentage > 10 {
		problems = append(problems, fmt.Sprintf("Supply Blocks (%.1f%% der Zeit)", loserAnalysis.SupplyAnalysis.BlockPercentage))
	}

	if loserAnalysis.SpendingAnalysis != nil && loserAnalysis.SpendingAnalysis.SpendingQuotient < 50 {
		problems = append(problems, fmt.Sprintf("Niedriger Spending Quotient (%.0f)", loserAnalysis.SpendingAnalysis.SpendingQuotient))
	}

	if loserAnalysis.APMAnalysis != nil && winnerAnalysis.APMAnalysis != nil {
		if winnerAnalysis.APMAnalysis.AverageAPM > loserAnalysis.APMAnalysis.AverageAPM*1.3 {
			problems = append(problems, "Deutlich niedrigere APM als Gegner")
		}
	}

	fmt.Printf("🔍 IDENTIFIZIERTE PROBLEME:\n\n")
	for i, p := range problems {
		fmt.Printf("   %d. %s\n", i+1, p)
	}
	fmt.Println()

	// Matchup-spezifische Tipps
	fmt.Printf("📚 %s vs %s TIPPS:\n\n", strings.ToUpper(loser.Race), strings.ToUpper(winner.Race))

	switch matchup {
	case "protossvzerg":
		fmt.Printf("   OPENING:\n")
		fmt.Printf("   • Standard: Gate → Nexus → Cyber → Stargate/Robo\n")
		fmt.Printf("   • Scout mit Adept Shade oder erstem Stalker\n")
		fmt.Printf("   • Früh Wall-Off gegen Zerglings\n\n")

		fmt.Printf("   MID GAME:\n")
		fmt.Printf("   • Gegen Roach/Ravager: Immortals + Chargelots\n")
		fmt.Printf("   • Gegen Hydras: Storm ist ESSENTIELL\n")
		fmt.Printf("   • Gegen Mutas: Phoenix oder schnell Archons\n\n")

		fmt.Printf("   TIMING ATTACKS:\n")
		fmt.Printf("   • 2-Base All-in mit Immortal/Archon ~7:30\n")
		fmt.Printf("   • Oder 8-Gate Chargelot Timing ~6:00\n")
		fmt.Printf("   • Greife VOR Hive-Tech an!\n\n")

		fmt.Printf("   LATE GAME:\n")
		fmt.Printf("   • Carrier/Tempest + Storm + Archons\n")
		fmt.Printf("   • Braucht gute Upgrades (3/3)\n")
		fmt.Printf("   • Vermeide große Kämpfe ohne Storm\n\n")

	case "zergvprotoss":
		fmt.Printf("   • Drohnen-Zählung ist key - nicht zu gierig\n")
		fmt.Printf("   • Scout für Cannon Rush und Proxy Gates\n")
		fmt.Printf("   • Roach/Ravager gut gegen Immortal-Push\n")
		fmt.Printf("   • Hydras + Lurker gegen Ground-Armies\n")
		fmt.Printf("   • Corruptors wenn Carrier kommen\n\n")

	case "terranvzerg":
		fmt.Printf("   • Bio + Medivacs ist der Standard\n")
		fmt.Printf("   • Siege Tanks gegen Roach/Ravager\n")
		fmt.Printf("   • Liberators gegen Hydras\n")
		fmt.Printf("   • Hellbats gegen Zerglings\n")
		fmt.Printf("   • Früh scout für Timing-Attacks\n\n")

	default:
		fmt.Printf("   • Fokus auf sauberes Macro\n")
		fmt.Printf("   • Scout regelmäßig\n")
		fmt.Printf("   • Passe Komposition an\n\n")
	}

	// Konkrete Verbesserungsvorschläge
	fmt.Printf("✅ KONKRETE SCHRITTE ZUR VERBESSERUNG:\n\n")

	fmt.Printf("   1. MACRO (Wichtigster Faktor!):\n")
	if loserAnalysis.SupplyAnalysis != nil && loserAnalysis.SupplyAnalysis.BlockPercentage > 5 {
		fmt.Printf("      → Baue Pylons BEVOR du Supply brauchst\n")
		fmt.Printf("      → Regel: Bei 75%% Supply, baue Supply-Gebäude\n")
	}
	if loserAnalysis.SpendingAnalysis != nil {
		if loserAnalysis.SpendingAnalysis.AverageUnspent.Minerals > 500 {
			fmt.Printf("      → Du hattest Ø%.0f ungenutzte Mineralien!\n", loserAnalysis.SpendingAnalysis.AverageUnspent.Minerals)
			fmt.Printf("      → Baue mehr Produktionsgebäude oder Expansions\n")
		}
	}
	fmt.Println()

	fmt.Printf("   2. BUILD ORDER:\n")
	fmt.Printf("      → Übe einen Standard-Build im Custom Game\n")
	fmt.Printf("      → Nutze spawningtool.com für Referenz-Builds\n")
	fmt.Printf("      → Ziel: Die ersten 5 Minuten perfekt ausführen\n\n")

	fmt.Printf("   3. SCOUTING:\n")
	fmt.Printf("      → Scout bei 3:30-4:00 für Tech-Gebäude\n")
	fmt.Printf("      → Reagiere auf das was du siehst\n\n")

	fmt.Printf("   4. ARMY CONTROL:\n")
	fmt.Printf("      → Kämpfe nur wenn du einen Vorteil hast\n")
	fmt.Printf("      → Nutze Concaves (Bogenformation)\n")
	fmt.Printf("      → F2 (Select All Army) nur im Notfall\n\n")

	// Zusammenfassung
	fmt.Printf("╔══════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║                      ZUSAMMENFASSUNG                         ║\n")
	fmt.Printf("╚══════════════════════════════════════════════════════════════╝\n\n")

	fmt.Printf("Du hast als %s gegen %s verloren.\n\n", loser.Race, winner.Race)
	fmt.Printf("Die HAUPTGRÜNDE waren wahrscheinlich:\n")

	if loserAnalysis.SupplyAnalysis != nil && loserAnalysis.SupplyAnalysis.BlockPercentage > 10 {
		fmt.Printf("  • Zu viele Supply Blocks → weniger Einheiten produziert\n")
	}
	if loserAnalysis.SpendingAnalysis != nil && loserAnalysis.SpendingAnalysis.AverageUnspent.Minerals > 500 {
		fmt.Printf("  • Zu viele ungenutzte Ressourcen → schwächere Armee\n")
	}
	if loserAnalysis.APMAnalysis != nil && winnerAnalysis.APMAnalysis != nil {
		if winnerAnalysis.APMAnalysis.AverageAPM > loserAnalysis.APMAnalysis.AverageAPM+30 {
			fmt.Printf("  • Niedrigere APM → langsamere Reaktionen\n")
		}
	}

	fmt.Printf("\n💡 TIPP: Fokussiere dich auf EIN Problem pro Woche.\n")
	fmt.Printf("   Diese Woche: Supply Blocks vermeiden!\n")
}
