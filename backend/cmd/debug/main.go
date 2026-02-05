package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"sc2-analytics/internal/analyzer"
	"sc2-analytics/internal/parser"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run ./cmd/debug <replay-file>")
	}

	filepath := os.Args[1]
	fmt.Printf("Parsing: %s\n\n", filepath)

	p := parser.New()
	replay, err := p.ParseFile(filepath)
	if err != nil {
		log.Fatalf("Parse error: %v", err)
	}

	fmt.Printf("=== REPLAY INFO ===\n")
	fmt.Printf("Map: %s\n", replay.Map)
	fmt.Printf("Duration: %d seconds\n", replay.Duration)
	fmt.Printf("Version: %s\n", replay.GameVersion)
	fmt.Printf("Played At: %s\n", replay.PlayedAt)

	fmt.Printf("\n=== PLAYERS ===\n")
	for _, p := range replay.Players {
		fmt.Printf("  [%d] %s (%s) - %s (Human: %v)\n",
			p.Slot, p.Name, p.Race, p.Result, p.IsHuman)
	}

	fmt.Printf("\n=== EVENTS ===\n")
	if replay.Events != nil {
		fmt.Printf("Tracker Events: %d\n", len(replay.Events.TrackerEvents))
		fmt.Printf("Game Events: %d\n", len(replay.Events.GameEvents))
		fmt.Printf("Message Events: %d\n", len(replay.Events.MessageEvents))

		// Zeige erste paar Tracker Events
		fmt.Printf("\n=== SAMPLE TRACKER EVENTS (first 10) ===\n")
		for i, evt := range replay.Events.TrackerEvents {
			if i >= 10 {
				break
			}
			fmt.Printf("  [%d] Loop=%d Type=%s PlayerID=%d\n",
				i, evt.Loop, evt.EventType, evt.PlayerID)
		}

		// Zähle Event-Typen
		eventCounts := make(map[string]int)
		for _, evt := range replay.Events.TrackerEvents {
			eventCounts[evt.EventType]++
		}
		fmt.Printf("\n=== EVENT TYPE COUNTS ===\n")
		for evtType, count := range eventCounts {
			fmt.Printf("  %s: %d\n", evtType, count)
		}

		// Prüfe PlayerStats Events
		fmt.Printf("\n=== PLAYER STATS EVENTS (first 3) ===\n")
		count := 0
		for _, evt := range replay.Events.TrackerEvents {
			if evt.EventType == "NNet.Replay.Tracker.SPlayerStatsEvent" && count < 3 {
				data, _ := json.MarshalIndent(evt.Data, "  ", "  ")
				fmt.Printf("  PlayerID=%d Loop=%d\n%s\n\n", evt.PlayerID, evt.Loop, string(data))
				count++
			}
		}
	} else {
		fmt.Printf("NO EVENTS FOUND!\n")
	}

	// Teste Analyse
	fmt.Printf("\n=== ANALYSIS TEST ===\n")
	a := analyzer.New()
	for _, player := range replay.Players {
		if !player.IsHuman {
			continue
		}
		fmt.Printf("\nAnalyzing player %d (%s)...\n", player.Slot, player.Name)
		analysis, err := a.AnalyzePlayer(replay, player.Slot, player.Race)
		if err != nil {
			fmt.Printf("  ERROR: %v\n", err)
			continue
		}

		if analysis.SupplyAnalysis != nil {
			fmt.Printf("  Supply: %.1f%% blocked, %d blocks\n",
				analysis.SupplyAnalysis.BlockPercentage,
				len(analysis.SupplyAnalysis.Blocks))
			fmt.Printf("  Supply Timeline Points: %d\n",
				len(analysis.SupplyAnalysis.SupplyTimeline))
		} else {
			fmt.Printf("  Supply: NO DATA\n")
		}

		if analysis.SpendingAnalysis != nil {
			fmt.Printf("  SQ: %.1f (%s)\n",
				analysis.SpendingAnalysis.SpendingQuotient,
				analysis.SpendingAnalysis.Rating)
			fmt.Printf("  Resource Timeline Points: %d\n",
				len(analysis.SpendingAnalysis.ResourceTimeline))
		} else {
			fmt.Printf("  Spending: NO DATA\n")
		}

		if analysis.APMAnalysis != nil {
			fmt.Printf("  APM: %.1f avg, %.1f peak, %.1f eapm\n",
				analysis.APMAnalysis.AverageAPM,
				analysis.APMAnalysis.PeakAPM,
				analysis.APMAnalysis.EAPM)
			fmt.Printf("  APM Timeline Points: %d\n",
				len(analysis.APMAnalysis.APMTimeline))
		} else {
			fmt.Printf("  APM: NO DATA\n")
		}

		if analysis.BuildOrder != nil {
			fmt.Printf("  Build Order: %d items\n", len(analysis.BuildOrder))
		} else {
			fmt.Printf("  Build Order: NO DATA\n")
		}

		fmt.Printf("  Suggestions: %d\n", len(analysis.Suggestions))
	}
}
