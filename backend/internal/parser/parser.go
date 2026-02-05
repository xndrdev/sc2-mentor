package parser

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/icza/s2prot"
	"github.com/icza/s2prot/rep"
)

// ParsedReplay enthält die extrahierten Replay-Daten
type ParsedReplay struct {
	Hash        string
	Filename    string
	Map         string
	Duration    int // Sekunden (Spielzeit)
	GameVersion string
	PlayedAt    time.Time
	Players     []ParsedPlayer
	Events      *ParsedEvents
}

// ParsedPlayer enthält Spielerinformationen
type ParsedPlayer struct {
	Slot       int
	Name       string
	ToonHandle string
	Race       string
	Result     string // Win, Loss, Undecided
	IsHuman    bool
	Region     string
}

// ParsedEvents enthält die relevanten Events für die Analyse
type ParsedEvents struct {
	TrackerEvents []TrackerEvent
	GameEvents    []GameEvent
	MessageEvents []MessageEvent
}

// TrackerEvent repräsentiert ein Tracker-Event
type TrackerEvent struct {
	Loop      int
	EventType string
	PlayerID  int
	Data      map[string]interface{}
}

// GameEvent repräsentiert ein Spieler-Action-Event
type GameEvent struct {
	Loop      int
	EventType string
	PlayerID  int
	Data      map[string]interface{}
}

// MessageEvent repräsentiert eine Chat-Nachricht
type MessageEvent struct {
	Loop     int
	PlayerID int
	Message  string
}

// Parser ist der Replay-Parser
type Parser struct{}

// New erstellt einen neuen Parser
func New() *Parser {
	return &Parser{}
}

// ParseFile parst eine SC2Replay-Datei
func (p *Parser) ParseFile(filepath string) (*ParsedReplay, error) {
	// Berechne Hash
	hash, err := hashFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("konnte Hash nicht berechnen: %w", err)
	}

	// Öffne Replay mit allen Event-Typen (game, message, tracker)
	r, err := rep.NewFromFileEvts(filepath, true, true, true)
	if err != nil {
		return nil, fmt.Errorf("konnte Replay nicht öffnen: %w", err)
	}
	defer r.Close()

	// Extrahiere Metadaten
	parsed := &ParsedReplay{
		Hash:     hash,
		Filename: filepath,
	}

	// Header-Informationen
	header := r.Header
	// Konvertiere Loops zu Sekunden (bei "Faster" Geschwindigkeit)
	loops := header.Loops()
	parsed.Duration = loopsToSeconds(int(loops))

	// Version
	version := header.Version()
	parsed.GameVersion = fmt.Sprintf("%d.%d.%d.%d",
		version.Major(),
		version.Minor(),
		version.Revision(),
		version.Build())

	// Details (Map, Spieler, Zeitpunkt)
	details := r.Details
	parsed.Map = details.Title()
	parsed.PlayedAt = details.TimeUTC()
	parsed.Players = parseDetailPlayers(details)

	// Initialisiere Events
	parsed.Events = &ParsedEvents{
		TrackerEvents: []TrackerEvent{},
		GameEvents:    []GameEvent{},
		MessageEvents: []MessageEvent{},
	}

	// Lade Tracker-Events für Analyse
	if r.TrackerEvts != nil && len(r.TrackerEvts.Evts) > 0 {
		parsed.Events.TrackerEvents = parseTrackerEvents(r.TrackerEvts.Evts)
	}

	// Lade Game-Events für APM-Berechnung
	if len(r.GameEvts) > 0 {
		parsed.Events.GameEvents = parseGameEvents(r.GameEvts)
	}

	// Lade Message-Events
	if len(r.MessageEvts) > 0 {
		parsed.Events.MessageEvents = parseMessageEvents(r.MessageEvts)
	}

	return parsed, nil
}

// parseDetailPlayers extrahiert Spielerinformationen aus Details
func parseDetailPlayers(details rep.Details) []ParsedPlayer {
	players := details.Players()
	result := make([]ParsedPlayer, 0, len(players))

	for i, p := range players {
		toon := p.Toon
		toonHandle := ""
		region := ""
		if toon.RegionID() > 0 {
			toonHandle = fmt.Sprintf("%d-%s-%d-%d",
				toon.RegionID(), toon.ProgramID(), toon.RealmID(), toon.ID())
			region = fmt.Sprintf("%d", toon.RegionID())
		}

		isHuman := false
		if p.Control() != nil && p.Control().Name == "Human" {
			isHuman = true
		}

		resultStr := "Undecided"
		if p.Result() != nil {
			resultStr = parseResult(p.Result())
		}

		raceName := "Unknown"
		if p.Race() != nil {
			raceName = p.Race().Name
		}

		result = append(result, ParsedPlayer{
			Slot:       i + 1,
			Name:       p.Name,
			ToonHandle: toonHandle,
			Race:       raceName,
			Result:     resultStr,
			IsHuman:    isHuman,
			Region:     region,
		})
	}

	return result
}

// parseResult konvertiert das Spielergebnis
func parseResult(result *rep.Result) string {
	if result == nil {
		return "Undecided"
	}
	switch result.Name {
	case "Victory":
		return "Win"
	case "Defeat":
		return "Loss"
	default:
		return "Undecided"
	}
}

// parseTrackerEvents extrahiert relevante Tracker-Events
func parseTrackerEvents(evts []s2prot.Event) []TrackerEvent {
	var result []TrackerEvent

	for _, evt := range evts {
		// Konvertiere Struct zu Map für flexiblen Zugriff
		data := structToMap(evt.Struct)

		// PlayerID kann in verschiedenen Feldern sein
		playerID := getIntFromMap(data, "playerId", 0)

		te := TrackerEvent{
			Loop:      int(evt.Loop()),
			EventType: evt.EvtType.Name,
			PlayerID:  playerID,
			Data:      data,
		}

		// Filtere relevante Events (vereinfachte Namen in neueren s2prot Versionen)
		switch evt.EvtType.Name {
		case "PlayerStats", "UnitBorn", "UnitDied", "UnitTypeChange",
			"Upgrade", "UnitInit", "UnitDone":
			result = append(result, te)
		}
	}

	return result
}

// parseGameEvents extrahiert Spieler-Aktionen für APM
func parseGameEvents(gameEvts []s2prot.Event) []GameEvent {
	var result []GameEvent

	for _, evt := range gameEvts {
		data := structToMap(evt.Struct)

		// UserID ist verschachtelt in "userid" -> "userId"
		userID := 0
		if useridRaw, ok := data["userid"]; ok {
			if useridMap, ok := useridRaw.(map[string]interface{}); ok {
				userID = getIntFromMap(useridMap, "userId", 0)
			}
		}

		ge := GameEvent{
			Loop:      int(evt.Loop()),
			EventType: evt.EvtType.Name,
			PlayerID:  userID + 1, // 0-indexed zu 1-indexed
			Data:      data,
		}

		// Relevante Aktionen (vereinfachte Namen)
		switch evt.EvtType.Name {
		case "Cmd", "CmdUpdateTargetPoint", "CmdUpdateTargetUnit",
			"SelectionDelta", "ControlGroupUpdate",
			"CameraUpdate", "CommandManagerState":
			result = append(result, ge)
		}
	}

	return result
}

// parseMessageEvents extrahiert Chat-Nachrichten
func parseMessageEvents(msgEvts []s2prot.Event) []MessageEvent {
	var result []MessageEvent

	for _, evt := range msgEvts {
		// Chat-Nachrichten können verschiedene Namen haben
		if evt.EvtType.Name == "Chat" || evt.EvtType.Name == "ChatMessage" {
			data := structToMap(evt.Struct)
			msg := getStringFromMap(data, "string", "")
			if msg == "" {
				msg = getStringFromMap(data, "message", "")
			}

			userID := 0
			if useridRaw, ok := data["userid"]; ok {
				if useridMap, ok := useridRaw.(map[string]interface{}); ok {
					userID = getIntFromMap(useridMap, "userId", 0)
				}
			}

			if msg != "" {
				result = append(result, MessageEvent{
					Loop:     int(evt.Loop()),
					PlayerID: userID + 1,
					Message:  msg,
				})
			}
		}
	}

	return result
}

// structToMap konvertiert ein s2prot.Struct zu einer Map
func structToMap(s s2prot.Struct) map[string]interface{} {
	result := make(map[string]interface{})
	if s == nil {
		return result
	}

	for key, val := range s {
		switch v := val.(type) {
		case s2prot.Struct:
			result[key] = structToMap(v)
		case []interface{}:
			result[key] = v
		default:
			result[key] = v
		}
	}

	return result
}

// getIntFromMap extrahiert einen Int-Wert aus einer Map
func getIntFromMap(m map[string]interface{}, key string, defaultVal int) int {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		}
	}
	return defaultVal
}

// getStringFromMap extrahiert einen String-Wert aus einer Map
func getStringFromMap(m map[string]interface{}, key string, defaultVal string) string {
	if val, ok := m[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return defaultVal
}

// hashFile berechnet den SHA256-Hash einer Datei
func hashFile(filepath string) (string, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// loopsToSeconds konvertiert Game-Loops zu Sekunden
// Bei "Faster" Geschwindigkeit: 16 Loops = 1 Game-Sekunde
// Zusätzlich: Game-Zeit zu Echtzeit = / 1.4
func loopsToSeconds(loops int) int {
	gameSeconds := float64(loops) / 16.0
	realSeconds := gameSeconds / 1.4
	return int(realSeconds)
}

// LoopsToGameSeconds konvertiert Loops zu Spielsekunden (ohne Speedfaktor)
func LoopsToGameSeconds(loops int) float64 {
	return float64(loops) / 16.0
}

// LoopsToRealSeconds konvertiert Loops zu Echtzeitsekunden
func LoopsToRealSeconds(loops int) float64 {
	return float64(loops) / 16.0 / 1.4
}
