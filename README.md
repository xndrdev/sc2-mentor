# SC2 Replay Analytics Tool

Ein Tool zur Analyse von StarCraft 2 Replay-Dateien mit Go-Backend und Vue.js-Frontend.

## Features

- **Replay Upload**: Hochladen und Parsen von .SC2Replay Dateien
- **Supply Block Analyse**: Erkennung und Visualisierung von Supply Blocks
- **Spending Quotient**: Berechnung und Bewertung des Resource Management
- **APM Tracking**: Aktionen pro Minute im Zeitverlauf
- **Build Order Extraktion**: Automatische Extraktion der ersten 8 Minuten
- **Inject Analyse** (Zerg): Effizienz der Spawn Larva Injects
- **Armeewert Tracking**: Entwicklung des Armeewerts über Zeit
- **Verbesserungsvorschläge**: Priorisierte Tipps basierend auf der Analyse
- **Trends**: Verbesserungstrends über mehrere Spiele

## Tech Stack

- **Backend**: Go mit [s2prot](https://github.com/icza/s2prot), chi Router, SQLite
- **Frontend**: Vue 3, TypeScript, TailwindCSS, ApexCharts
- **Datenbank**: SQLite

## Projektstruktur

```
sc2-analytics/
├── backend/
│   ├── cmd/server/main.go           # Entry Point
│   ├── internal/
│   │   ├── api/                     # REST API Handler
│   │   ├── parser/                  # s2prot Replay Parser
│   │   ├── analyzer/                # Analyse-Module
│   │   │   ├── macro/              # Supply, Inject, Spending
│   │   │   ├── micro/              # APM, Army
│   │   │   └── builds/             # Build Order
│   │   ├── models/                  # Datenmodelle
│   │   └── repository/              # SQLite Repository
│   └── go.mod
├── frontend/
│   ├── src/
│   │   ├── views/                   # Seiten
│   │   ├── components/              # Vue Komponenten
│   │   ├── stores/                  # Pinia State
│   │   └── api/                     # API Client
│   └── package.json
└── Makefile
```

## Installation

### Backend

```bash
cd backend
go mod download
go build -o bin/server ./cmd/server
```

### Frontend

```bash
cd frontend
npm install
npm run build
```

## Verwendung

### Development

Terminal 1 - Backend:
```bash
make backend-dev
# oder
cd backend && go run ./cmd/server
```

Terminal 2 - Frontend:
```bash
make frontend-dev
# oder
cd frontend && npm run dev
```

### Production

```bash
make build
make run
```

Das Backend läuft auf `http://localhost:8080`, das Frontend auf `http://localhost:5173`.

## API Endpoints

| Methode | Endpoint | Beschreibung |
|---------|----------|--------------|
| POST | `/api/v1/replays/upload` | Replay hochladen |
| GET | `/api/v1/replays` | Alle Replays auflisten |
| GET | `/api/v1/replays/:id` | Replay Details |
| GET | `/api/v1/replays/:id/analysis` | Vollständige Analyse |
| GET | `/api/v1/stats/trends` | Verbesserungstrends |

## Analyse-Metriken

### Supply Block
- Erkennt Zeiträume, in denen Supply = Max Supply
- Klassifiziert nach Schweregrad: <5s (leicht), 5-15s (mittel), >15s (schwer)

### Spending Quotient (SQ)
Formel: `SQ = 35 * (0.00137 * avgIncome - ln(avgUnspent + 1)) + 240`

Bewertung:
- < 70: Verbesserungswürdig
- 70-90: Unterdurchschnittlich
- 90-110: Durchschnittlich
- 110-130: Gut
- > 130: Exzellent

### APM
- Durchschnitts-APM über gesamtes Spiel
- Peak-APM
- EAPM (Effective APM, filtert Spam)

## Konfiguration

Der Server akzeptiert folgende Flags:

```
-port int     Server Port (default 8080)
-db string    Pfad zur SQLite Datenbank (default "./data/sc2analytics.db")
-uploads string    Upload-Verzeichnis (default "./data/uploads")
```

## Lizenz

MIT
