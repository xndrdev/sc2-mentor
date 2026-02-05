package main

import (
	"fmt"
	"log"
	"os"

	"github.com/icza/s2prot/rep"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run ./cmd/rawtest <replay-file>")
	}

	filepath := os.Args[1]
	fmt.Printf("Raw s2prot test: %s\n\n", filepath)

	// Lade mit allen Events
	r, err := rep.NewFromFileEvts(filepath, true, true, true)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	defer r.Close()

	fmt.Printf("=== GAME EVENTS ===\n")
	fmt.Printf("GameEvts length: %d\n", len(r.GameEvts))

	// Zähle Game Event Typen
	gameEvtCounts := make(map[string]int)
	for _, evt := range r.GameEvts {
		gameEvtCounts[evt.EvtType.Name]++
	}
	fmt.Printf("\nGame Event Types:\n")
	for name, count := range gameEvtCounts {
		fmt.Printf("  %s: %d\n", name, count)
	}

	// Zeige ein Beispiel CmdEvent
	fmt.Printf("\n=== SAMPLE GAME EVENTS ===\n")
	for i, evt := range r.GameEvts {
		if i >= 3 {
			break
		}
		fmt.Printf("[%d] %s: %+v\n\n", i, evt.EvtType.Name, evt.Struct)
	}
}
