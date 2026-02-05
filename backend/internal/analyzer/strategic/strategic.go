package strategic

import (
	"fmt"
	"sc2-analytics/internal/models"
	"strings"
)

// StrategicAnalyzer erstellt strategische Spielanalysen
type StrategicAnalyzer struct{}

// NewStrategicAnalyzer erstellt einen neuen StrategicAnalyzer
func NewStrategicAnalyzer() *StrategicAnalyzer {
	return &StrategicAnalyzer{}
}

// Analyze erstellt eine vollständige strategische Analyse
func (sa *StrategicAnalyzer) Analyze(
	loserName, winnerName string,
	loserRace, winnerRace string,
	loserAnalysis, winnerAnalysis *models.AnalysisData,
) *models.StrategicAnalysis {
	if loserAnalysis == nil || winnerAnalysis == nil {
		return nil
	}

	matchup := strings.ToUpper(string(loserRace[0])) + "v" + strings.ToUpper(string(winnerRace[0]))

	analysis := &models.StrategicAnalysis{
		Winner:     winnerName,
		Loser:      loserName,
		WinnerRace: winnerRace,
		LoserRace:  loserRace,
		Matchup:    matchup,
	}

	// Metriken vergleichen
	analysis.MetricsComparison = sa.compareMetrics(loserAnalysis, winnerAnalysis)

	// Supply Blocks extrahieren
	analysis.SupplyBlocks = sa.extractSupplyBlocks(loserAnalysis)

	// Kritische Momente finden
	analysis.CriticalMoments = sa.findCriticalMoments(loserAnalysis, winnerAnalysis)

	// Probleme identifizieren
	analysis.Problems = sa.identifyProblems(loserAnalysis, winnerAnalysis)

	// Matchup-Tipps
	analysis.MatchupTips = sa.getMatchupTips(loserRace, winnerRace)

	// Verbesserungsschritte
	analysis.ImprovementSteps = sa.generateImprovementSteps(analysis.Problems)

	// Zusammenfassung
	analysis.Summary = sa.generateSummary(analysis)

	return analysis
}

// compareMetrics vergleicht Metriken zwischen Spielern
func (sa *StrategicAnalyzer) compareMetrics(loser, winner *models.AnalysisData) []models.MetricComparison {
	var comparisons []models.MetricComparison

	// APM
	if loser.APMAnalysis != nil && winner.APMAnalysis != nil {
		comparisons = append(comparisons, models.MetricComparison{
			Metric:      "APM (Durchschnitt)",
			PlayerValue: loser.APMAnalysis.AverageAPM,
			EnemyValue:  winner.APMAnalysis.AverageAPM,
			IsWorse:     loser.APMAnalysis.AverageAPM < winner.APMAnalysis.AverageAPM*0.8,
		})
		comparisons = append(comparisons, models.MetricComparison{
			Metric:      "EAPM (Effektiv)",
			PlayerValue: loser.APMAnalysis.EAPM,
			EnemyValue:  winner.APMAnalysis.EAPM,
			IsWorse:     loser.APMAnalysis.EAPM < winner.APMAnalysis.EAPM*0.8,
		})
	}

	// Spending Quotient
	if loser.SpendingAnalysis != nil && winner.SpendingAnalysis != nil {
		comparisons = append(comparisons, models.MetricComparison{
			Metric:      "Spending Quotient",
			PlayerValue: loser.SpendingAnalysis.SpendingQuotient,
			EnemyValue:  winner.SpendingAnalysis.SpendingQuotient,
			IsWorse:     loser.SpendingAnalysis.SpendingQuotient < winner.SpendingAnalysis.SpendingQuotient-20,
		})
		comparisons = append(comparisons, models.MetricComparison{
			Metric:      "Ø Ungenutzte Mineralien",
			PlayerValue: loser.SpendingAnalysis.AverageUnspent.Minerals,
			EnemyValue:  winner.SpendingAnalysis.AverageUnspent.Minerals,
			IsWorse:     loser.SpendingAnalysis.AverageUnspent.Minerals > winner.SpendingAnalysis.AverageUnspent.Minerals*1.5,
		})
	}

	// Supply Block
	if loser.SupplyAnalysis != nil && winner.SupplyAnalysis != nil {
		comparisons = append(comparisons, models.MetricComparison{
			Metric:      "Supply Block Zeit",
			PlayerValue: loser.SupplyAnalysis.BlockPercentage,
			EnemyValue:  winner.SupplyAnalysis.BlockPercentage,
			IsWorse:     loser.SupplyAnalysis.BlockPercentage > winner.SupplyAnalysis.BlockPercentage*1.5,
		})
		comparisons = append(comparisons, models.MetricComparison{
			Metric:      "Anzahl Supply Blocks",
			PlayerValue: float64(len(loser.SupplyAnalysis.Blocks)),
			EnemyValue:  float64(len(winner.SupplyAnalysis.Blocks)),
			IsWorse:     len(loser.SupplyAnalysis.Blocks) > len(winner.SupplyAnalysis.Blocks)+2,
		})
	}

	// Army
	if loser.ArmyAnalysis != nil && winner.ArmyAnalysis != nil {
		comparisons = append(comparisons, models.MetricComparison{
			Metric:      "Peak Armeewert",
			PlayerValue: float64(loser.ArmyAnalysis.PeakArmyValue),
			EnemyValue:  float64(winner.ArmyAnalysis.PeakArmyValue),
			IsWorse:     loser.ArmyAnalysis.PeakArmyValue < winner.ArmyAnalysis.PeakArmyValue/2,
		})
	}

	return comparisons
}

// extractSupplyBlocks extrahiert Supply Blocks für Anzeige
func (sa *StrategicAnalyzer) extractSupplyBlocks(analysis *models.AnalysisData) []models.SupplyBlockSummary {
	var blocks []models.SupplyBlockSummary

	if analysis.SupplyAnalysis == nil {
		return blocks
	}

	for _, b := range analysis.SupplyAnalysis.Blocks {
		blocks = append(blocks, models.SupplyBlockSummary{
			Time:     b.StartTime,
			Duration: b.Duration,
			Severity: b.Severity,
		})
	}

	return blocks
}

// findCriticalMoments findet kritische Kampfmomente
func (sa *StrategicAnalyzer) findCriticalMoments(loser, winner *models.AnalysisData) []models.CriticalMoment {
	var moments []models.CriticalMoment

	if loser.ArmyAnalysis == nil || winner.ArmyAnalysis == nil {
		return moments
	}

	loserTimeline := loser.ArmyAnalysis.ArmyTimeline
	winnerTimeline := winner.ArmyAnalysis.ArmyTimeline

	// Erstelle eine Map für schnellen Zugriff auf Winner-Werte nach Zeit
	winnerValues := make(map[float64]int)
	for _, point := range winnerTimeline {
		winnerValues[point.Time] = point.UnitCount
	}

	var lastLoserCount, lastWinnerCount int
	for i, point := range loserTimeline {
		if i == 0 {
			lastLoserCount = point.UnitCount
			if wc, ok := winnerValues[point.Time]; ok {
				lastWinnerCount = wc
			}
			continue
		}

		loserLoss := lastLoserCount - point.UnitCount
		winnerLoss := 0
		if wc, ok := winnerValues[point.Time]; ok {
			winnerLoss = lastWinnerCount - wc
			lastWinnerCount = wc
		}

		// Signifikante Verluste
		if loserLoss > 2 || winnerLoss > 2 {
			assessment := ""
			isPositive := false

			if loserLoss > 0 && winnerLoss == 0 {
				assessment = "Einseitige Verluste"
			} else if loserLoss > winnerLoss*2 {
				assessment = "Schlechter Trade"
			} else if loserLoss < winnerLoss {
				assessment = "Guter Trade!"
				isPositive = true
			} else if loserLoss > winnerLoss {
				assessment = "Leichter Nachteil"
			} else {
				assessment = "Vorteil für dich"
				isPositive = true
			}

			if loserLoss > 0 || winnerLoss > 0 {
				moments = append(moments, models.CriticalMoment{
					Time:       point.Time,
					PlayerLoss: loserLoss,
					EnemyLoss:  winnerLoss,
					Assessment: assessment,
					IsPositive: isPositive,
				})
			}
		}

		lastLoserCount = point.UnitCount
	}

	// Limitiere auf die wichtigsten 10 Momente
	if len(moments) > 10 {
		moments = moments[:10]
	}

	return moments
}

// identifyProblems identifiziert die Hauptprobleme
func (sa *StrategicAnalyzer) identifyProblems(loser, winner *models.AnalysisData) []models.IdentifiedProblem {
	var problems []models.IdentifiedProblem

	// Supply Blocks
	if loser.SupplyAnalysis != nil && loser.SupplyAnalysis.BlockPercentage > 10 {
		problems = append(problems, models.IdentifiedProblem{
			Title:       fmt.Sprintf("Supply Blocks (%.1f%% der Zeit)", loser.SupplyAnalysis.BlockPercentage),
			Description: "Du warst zu oft supply-blocked und konntest keine Einheiten produzieren.",
			Priority:    "high",
		})
	}

	// Spending
	if loser.SpendingAnalysis != nil && loser.SpendingAnalysis.SpendingQuotient < 50 {
		problems = append(problems, models.IdentifiedProblem{
			Title:       fmt.Sprintf("Niedriger Spending Quotient (%d)", int(loser.SpendingAnalysis.SpendingQuotient)),
			Description: "Du hast zu viele Ressourcen angesammelt ohne sie auszugeben.",
			Priority:    "high",
		})
	}

	// APM
	if loser.APMAnalysis != nil && winner.APMAnalysis != nil {
		if loser.APMAnalysis.AverageAPM < winner.APMAnalysis.AverageAPM*0.7 {
			problems = append(problems, models.IdentifiedProblem{
				Title:       "Deutlich niedrigere APM als Gegner",
				Description: fmt.Sprintf("Deine APM (%.0f) war deutlich niedriger als die des Gegners (%.0f).", loser.APMAnalysis.AverageAPM, winner.APMAnalysis.AverageAPM),
				Priority:    "medium",
			})
		}
	}

	// Army Value
	if loser.ArmyAnalysis != nil && winner.ArmyAnalysis != nil {
		if loser.ArmyAnalysis.PeakArmyValue < winner.ArmyAnalysis.PeakArmyValue/2 {
			problems = append(problems, models.IdentifiedProblem{
				Title:       "Zu wenig Armee produziert",
				Description: fmt.Sprintf("Dein Peak-Armeewert (%d) war deutlich niedriger als der des Gegners (%d).", loser.ArmyAnalysis.PeakArmyValue, winner.ArmyAnalysis.PeakArmyValue),
				Priority:    "high",
			})
		}
	}

	return problems
}

// getMatchupTips gibt matchup-spezifische Tipps zurück
func (sa *StrategicAnalyzer) getMatchupTips(loserRace, winnerRace string) *models.MatchupTips {
	loserRaceLower := strings.ToLower(loserRace)
	winnerRaceLower := strings.ToLower(winnerRace)

	tips := &models.MatchupTips{}

	switch loserRaceLower + "v" + winnerRaceLower {
	case "protossvzerg":
		tips.Opening = []string{
			"Standard: Gate → Nexus → Cyber → Stargate/Robo",
			"Scout mit Adept Shade oder erstem Stalker",
			"Früh Wall-Off gegen Zerglings",
		}
		tips.MidGame = []string{
			"Gegen Roach/Ravager: Immortals + Chargelots",
			"Gegen Hydras: Storm ist ESSENTIELL",
			"Gegen Mutas: Phoenix oder schnell Archons",
		}
		tips.Timing = []string{
			"2-Base All-in mit Immortal/Archon ~7:30",
			"Oder 8-Gate Chargelot Timing ~6:00",
			"Greife VOR Hive-Tech an!",
		}
		tips.LateGame = []string{
			"Carrier/Tempest + Storm + Archons",
			"Braucht gute Upgrades (3/3)",
			"Vermeide große Kämpfe ohne Storm",
		}

	case "zergvprotoss":
		tips.Opening = []string{
			"Hatch first ist standard gegen Protoss",
			"Speedlings für frühen Druck möglich",
			"Overlord Scout bei 3:00-3:30",
		}
		tips.MidGame = []string{
			"Roach/Ravager gegen Immortal-Archon",
			"Lurkers gegen Ground-Armies",
			"Mutas wenn Protoss wenig Anti-Air hat",
		}
		tips.Timing = []string{
			"Roach/Ravager Timing bei ~5:00",
			"Ling/Bane All-in gegen greedy Builds",
		}
		tips.LateGame = []string{
			"Broodlord/Corruptor/Viper Deathball",
			"Infestors gegen Carrier",
			"Immer auf Flächeneffekte achten",
		}

	case "terranvzerg":
		tips.Opening = []string{
			"Reaper Scout + CC first üblich",
			"Hellion Harass zum Dronen-Killen",
			"Wall-Off gegen Zerglings",
		}
		tips.MidGame = []string{
			"Marine/Tank gegen Roach/Hydra",
			"Liberators gegen Mutas",
			"Hellbats gegen Ling/Bane",
		}
		tips.Timing = []string{
			"2-1-1 Push mit Medivacs ~5:30",
			"Oder 3CC Macro-Spiel",
		}
		tips.LateGame = []string{
			"Ghost/Lib/Thor Komposition",
			"Vikings + Thors gegen Broodlords",
		}

	case "zergvterran":
		tips.Opening = []string{
			"3 Hatch vor Pool oft gut",
			"Ling/Bane gegen Hellions",
			"Scout für Banshees/Liberators",
		}
		tips.MidGame = []string{
			"Ling/Bane/Muta klassisch",
			"Oder Roach/Ravager/Hydra",
		}
		tips.Timing = []string{
			"2-Base Ling/Bane Timing möglich",
			"Roach/Ravager All-in bei ~4:30",
		}
		tips.LateGame = []string{
			"Ultras + Vipers gegen Mech",
			"Infestors gegen Bio",
		}

	case "protossvterran":
		tips.Opening = []string{
			"Gate Expand Standard",
			"Robo oder Stargate Tech",
			"Scout für Proxy-Barracks",
		}
		tips.MidGame = []string{
			"Chargelot/Archon/Immortal core",
			"Colossus oder Disruptors hinzufügen",
			"Storm gegen Bio",
		}
		tips.Timing = []string{
			"Blink-Stalker Timing ~5:00",
			"Chargelot All-in gegen Mech",
		}
		tips.LateGame = []string{
			"Carrier/Tempest bei gutem Eco",
			"Feedback gegen Ghosts",
		}

	case "terranvprotoss":
		tips.Opening = []string{
			"Reaper Expand üblich",
			"Factory für Cyclone/Tank",
			"Scout für DTs/Oracles",
		}
		tips.MidGame = []string{
			"Bio + Medivac + Ghosts",
			"Widow Mines gegen Chargelots",
			"Liberators für Zonen-Control",
		}
		tips.Timing = []string{
			"Stim-Timing ~5:30",
			"Liberator Harass",
		}
		tips.LateGame = []string{
			"Ghosts sind ESSENTIELL",
			"Vikings gegen Carriers",
		}

	// Spiegel-Matchups
	default:
		tips.Opening = []string{
			"Nutze Standard-Openings für deine Rasse",
			"Scout früh um Cheese zu erkennen",
		}
		tips.MidGame = []string{
			"Konzentriere dich auf gute Macro",
			"Baue Einheiten kontinuierlich",
		}
		tips.Timing = []string{
			"Greife an wenn du einen Vorteil hast",
			"Timing-Attacks bei Tech-Switches",
		}
		tips.LateGame = []string{
			"Upgrades sind entscheidend",
			"Kontrolliere die Map",
		}
	}

	return tips
}

// generateImprovementSteps erstellt konkrete Verbesserungsschritte
func (sa *StrategicAnalyzer) generateImprovementSteps(problems []models.IdentifiedProblem) []models.ImprovementStep {
	var steps []models.ImprovementStep

	for _, p := range problems {
		switch {
		case strings.Contains(p.Title, "Supply Block"):
			steps = append(steps, models.ImprovementStep{
				Category:    "MACRO",
				Title:       "Pylons/Depots/Overlords früher bauen",
				Description: "Baue Supply-Gebäude BEVOR du Supply brauchst. Regel: Bei 75% Supply, baue das nächste Supply-Gebäude.",
			})
		case strings.Contains(p.Title, "Spending"):
			steps = append(steps, models.ImprovementStep{
				Category:    "MACRO",
				Title:       "Ressourcen schneller ausgeben",
				Description: "Füge mehr Produktionsgebäude hinzu oder starte Upgrades. Ressourcen auf der Bank gewinnen keine Spiele.",
			})
		case strings.Contains(p.Title, "APM"):
			steps = append(steps, models.ImprovementStep{
				Category:    "MECHANICS",
				Title:       "Hotkeys und Kamera-Shortcuts üben",
				Description: "Nutze Control-Groups für Armee und Produktionsgebäude. Übe schnelle Camera-Location Hotkeys.",
			})
		case strings.Contains(p.Title, "Armee"):
			steps = append(steps, models.ImprovementStep{
				Category:    "PRODUCTION",
				Title:       "Kontinuierlich Einheiten produzieren",
				Description: "Halte deine Produktionsgebäude aktiv. Füge mehr Barracks/Gates/Hatcheries hinzu wenn du Ressourcen ansammelst.",
			})
		}
	}

	// Allgemeine Tipps hinzufügen
	steps = append(steps, models.ImprovementStep{
		Category:    "BUILD ORDER",
		Title:       "Übe einen Standard-Build",
		Description: "Wähle einen Build und übe ihn im Custom Game bis du ihn blind ausführen kannst. Nutze spawningtool.com für Referenz-Builds.",
	})

	steps = append(steps, models.ImprovementStep{
		Category:    "SCOUTING",
		Title:       "Regelmäßig scouten",
		Description: "Scout bei 3:30-4:00 für Tech-Gebäude. Reagiere auf das was du siehst statt blind zu spielen.",
	})

	return steps
}

// generateSummary erstellt eine Zusammenfassung
func (sa *StrategicAnalyzer) generateSummary(analysis *models.StrategicAnalysis) string {
	var mainReasons []string

	for _, p := range analysis.Problems {
		if p.Priority == "high" {
			switch {
			case strings.Contains(p.Title, "Supply"):
				mainReasons = append(mainReasons, "Zu viele Supply Blocks → weniger Einheiten produziert")
			case strings.Contains(p.Title, "Spending"):
				mainReasons = append(mainReasons, "Ressourcen nicht ausgegeben → schwächere Armee")
			case strings.Contains(p.Title, "APM"):
				mainReasons = append(mainReasons, "Niedrigere APM → langsamere Reaktionen")
			case strings.Contains(p.Title, "Armee"):
				mainReasons = append(mainReasons, "Zu wenig Armee produziert → konnte nicht verteidigen")
			}
		}
	}

	if len(mainReasons) == 0 {
		for _, p := range analysis.Problems {
			if p.Priority == "medium" {
				mainReasons = append(mainReasons, p.Title)
			}
		}
	}

	summary := fmt.Sprintf("Du hast als %s gegen %s verloren.\n\nDie HAUPTGRÜNDE waren wahrscheinlich:\n", analysis.LoserRace, analysis.WinnerRace)
	for _, reason := range mainReasons {
		summary += fmt.Sprintf("• %s\n", reason)
	}

	return summary
}
