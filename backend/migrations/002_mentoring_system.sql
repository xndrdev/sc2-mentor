-- SC2 Analytics Mentoring System Schema
-- Migration 002

-- Benutzer-Accounts
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    sc2_player_name TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_login DATETIME
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_player_name ON users(sc2_player_name);

-- Benutzer-Replays Verknüpfung
CREATE TABLE IF NOT EXISTS user_replays (
    user_id INTEGER NOT NULL,
    replay_id INTEGER NOT NULL,
    PRIMARY KEY (user_id, replay_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (replay_id) REFERENCES replays(id) ON DELETE CASCADE
);

-- Ziele
CREATE TABLE IF NOT EXISTS goals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    goal_type TEXT NOT NULL,        -- 'daily', 'weekly'
    metric_name TEXT NOT NULL,       -- 'apm', 'supply_block', 'games_played', 'win_rate', 'sq'
    target_value REAL NOT NULL,
    comparison TEXT DEFAULT '>=',    -- '>=', '<=', '>', '<', '='
    current_value REAL DEFAULT 0,
    status TEXT DEFAULT 'active',    -- 'active', 'completed', 'failed'
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deadline DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_goals_user_status ON goals(user_id, status);
CREATE INDEX IF NOT EXISTS idx_goals_deadline ON goals(deadline);

-- Tägliche Fortschritts-Snapshots
CREATE TABLE IF NOT EXISTS daily_progress (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    date DATE NOT NULL,
    games_played INTEGER DEFAULT 0,
    wins INTEGER DEFAULT 0,
    losses INTEGER DEFAULT 0,
    avg_apm REAL DEFAULT 0,
    avg_spending_quotient REAL DEFAULT 0,
    avg_supply_block_pct REAL DEFAULT 0,
    total_play_time INTEGER DEFAULT 0,  -- in Sekunden
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(user_id, date)
);

CREATE INDEX IF NOT EXISTS idx_daily_progress_user_date ON daily_progress(user_id, date);

-- Wochenberichte
CREATE TABLE IF NOT EXISTS weekly_reports (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    week_start DATE NOT NULL,
    week_end DATE NOT NULL,
    total_games INTEGER DEFAULT 0,
    wins INTEGER DEFAULT 0,
    losses INTEGER DEFAULT 0,
    win_rate REAL DEFAULT 0,
    avg_apm REAL DEFAULT 0,
    avg_sq REAL DEFAULT 0,
    avg_supply_block REAL DEFAULT 0,
    main_race TEXT,
    total_play_time INTEGER DEFAULT 0,
    improvements TEXT,              -- JSON: {"apm": "+5%", "supply_block": "-2%"}
    regressions TEXT,               -- JSON: {"win_rate": "-3%"}
    focus_suggestion TEXT,
    strengths TEXT,                 -- JSON: ["Good macro", "Fast expansion"]
    weaknesses TEXT,                -- JSON: ["Supply blocks", "Low APM"]
    generated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(user_id, week_start)
);

CREATE INDEX IF NOT EXISTS idx_weekly_reports_user_week ON weekly_reports(user_id, week_start);

-- Coaching Fokus (aktueller Fokusbereich für den Spieler)
CREATE TABLE IF NOT EXISTS coaching_focus (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    focus_area TEXT NOT NULL,       -- 'macro', 'micro', 'economy', 'army_control', 'scouting'
    description TEXT,
    started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    active INTEGER DEFAULT 1,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_coaching_focus_user_active ON coaching_focus(user_id, active);
