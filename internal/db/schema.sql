-- Servers
CREATE TABLE IF NOT EXISTS servers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    external_id TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    path TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Players (global list)
CREATE TABLE IF NOT EXISTS players (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    external_id TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Player stats table
CREATE TABLE IF NOT EXISTS player_stats (
    id TEXT PRIMARY KEY,
    player_id INTEGER NOT NULL UNIQUE,
    server_id INTEGER NOT NULL,
    games_played INTEGER DEFAULT 0,
    wins INTEGER DEFAULT 0,
    losses INTEGER DEFAULT 0,
    total_score INTEGER DEFAULT 0,
    total_play_time INTEGER DEFAULT 0, -- store as seconds
    last_login TEXT, -- ISO8601 string
    total_deaths INTEGER DEFAULT 0,
    friendly_fire_kills INTEGER DEFAULT 0,
    highest_score INTEGER DEFAULT 0,
    FOREIGN KEY (player_id) REFERENCES players(id) ON DELETE CASCADE,
    FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE
);

-- Weapon stats (one row per player per weapon - lifetime totals)
CREATE TABLE IF NOT EXISTS weapon_stats (
    player_stats_id TEXT NOT NULL,
    weapon_name TEXT NOT NULL,
    kills INTEGER DEFAULT 0,
    assists INTEGER DEFAULT 0,
    PRIMARY KEY (player_stats_id, weapon_name),
    FOREIGN KEY (player_stats_id) REFERENCES player_stats(id) ON DELETE CASCADE
);

-- Daily player stats (for rolling time windows like "last 30 days")
CREATE TABLE IF NOT EXISTS daily_player_stats (
    player_id INTEGER NOT NULL,
    server_id INTEGER NOT NULL,
    date DATE NOT NULL,
    kills INTEGER DEFAULT 0,
    assists INTEGER DEFAULT 0,
    deaths INTEGER DEFAULT 0,
    games_played INTEGER DEFAULT 0,
    total_score INTEGER DEFAULT 0,
    PRIMARY KEY (player_id, server_id, date),
    FOREIGN KEY (player_id) REFERENCES players(id) ON DELETE CASCADE,
    FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE
);

CREATE INDEX idx_daily_player_stats_date ON daily_player_stats(date);
CREATE INDEX idx_daily_player_stats_player ON daily_player_stats(player_id, date);

-- Daily weapon stats (for rolling time windows like "last 30 days")
CREATE TABLE IF NOT EXISTS daily_weapon_stats (
    player_id INTEGER NOT NULL,
    server_id INTEGER NOT NULL,
    date DATE NOT NULL,
    weapon_name TEXT NOT NULL,
    kills INTEGER DEFAULT 0,
    assists INTEGER DEFAULT 0,
    PRIMARY KEY (player_id, server_id, date, weapon_name),
    FOREIGN KEY (player_id) REFERENCES players(id) ON DELETE CASCADE,
    FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE
);

CREATE INDEX idx_daily_weapon_stats_date ON daily_weapon_stats(date);
CREATE INDEX idx_daily_weapon_stats_player ON daily_weapon_stats(player_id, date);

CREATE TABLE IF NOT EXISTS matches (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    server_id INTEGER NOT NULL,
    map TEXT,
    winner_team INTEGER,
    start_time DATETIME,
    end_time DATETIME,
    mode TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS match_players (
    match_id INTEGER NOT NULL,
    player_id INTEGER NOT NULL,
    PRIMARY KEY (match_id, player_id),
    FOREIGN KEY (match_id) REFERENCES matches(id) ON DELETE CASCADE,
    FOREIGN KEY (player_id) REFERENCES players(id) ON DELETE CASCADE
);