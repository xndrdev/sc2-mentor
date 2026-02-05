-- SC2 Analytics Initial Schema

-- Spieler-Tabelle
CREATE TABLE IF NOT EXISTS players (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    toon_handle TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    region TEXT DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Replays-Tabelle
CREATE TABLE IF NOT EXISTS replays (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    hash TEXT UNIQUE NOT NULL,
    filename TEXT NOT NULL,
    map TEXT NOT NULL,
    duration INTEGER NOT NULL, -- in Sekunden
    game_version TEXT DEFAULT '',
    played_at DATETIME,
    uploaded_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Spieler-Replay Verknüpfung
CREATE TABLE IF NOT EXISTS game_players (
    replay_id INTEGER NOT NULL,
    player_id INTEGER NOT NULL,
    player_slot INTEGER NOT NULL,
    name TEXT NOT NULL,
    race TEXT NOT NULL,
    result TEXT NOT NULL, -- Win, Loss, Undecided
    apm REAL DEFAULT 0,
    spending_quotient REAL DEFAULT 0,
    is_human INTEGER DEFAULT 1,
    PRIMARY KEY (replay_id, player_id),
    FOREIGN KEY (replay_id) REFERENCES replays(id) ON DELETE CASCADE,
    FOREIGN KEY (player_id) REFERENCES players(id) ON DELETE CASCADE
);

-- Analysen-Tabelle (JSON-basiert)
CREATE TABLE IF NOT EXISTS analyses (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    replay_id INTEGER NOT NULL,
    player_id INTEGER NOT NULL,
    data TEXT NOT NULL, -- JSON
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(replay_id, player_id),
    FOREIGN KEY (replay_id) REFERENCES replays(id) ON DELETE CASCADE,
    FOREIGN KEY (player_id) REFERENCES players(id) ON DELETE CASCADE
);

-- Indices für bessere Performance
CREATE INDEX IF NOT EXISTS idx_replays_played_at ON replays(played_at);
CREATE INDEX IF NOT EXISTS idx_replays_hash ON replays(hash);
CREATE INDEX IF NOT EXISTS idx_game_players_replay ON game_players(replay_id);
CREATE INDEX IF NOT EXISTS idx_game_players_player ON game_players(player_id);
CREATE INDEX IF NOT EXISTS idx_analyses_replay ON analyses(replay_id);
