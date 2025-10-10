-- Simplified schema for Sandstorm tracker (based on schema.md)

-- Servers
CREATE TABLE IF NOT EXISTS servers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    external_id TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    path TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Server logs
CREATE TABLE IF NOT EXISTS server_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    server_id INTEGER NOT NULL,
    open_time DATETIME NOT NULL,
    log_path TEXT NOT NULL,
    lines_processed INTEGER DEFAULT 0,
    file_size_bytes INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE
);

-- Players (global list)
CREATE TABLE IF NOT EXISTS players (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    external_id TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Matches
CREATE TABLE IF NOT EXISTS matches (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    server_id INTEGER NOT NULL,
    map_id INTEGER,
    winner_team INTEGER,
    start_time DATETIME,
    end_time DATETIME,
    mode TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE,
    FOREIGN KEY (map_id) REFERENCES maps(id) ON DELETE SET NULL
);

-- Match participants
CREATE TABLE IF NOT EXISTS match_participant (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    player_id INTEGER NOT NULL,
    match_id INTEGER NOT NULL,
    join_time DATETIME,
    leave_time DATETIME,
    team INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (player_id) REFERENCES players(id) ON DELETE CASCADE,
    FOREIGN KEY (match_id) REFERENCES matches(id) ON DELETE CASCADE
);

-- Kills
CREATE TABLE IF NOT EXISTS kills (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    server_id INTEGER NOT NULL,
    match_id INTEGER,
    match_participant_id INTEGER,
    weapon_name TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    killer_id INTEGER,
    victim_name TEXT,
    kill_type INTEGER NOT NULL CHECK (kill_type IN (0, 1, 2)), -- 0=regular, 1=suicide, 2=teamkill
    FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE,
    FOREIGN KEY (killer_id) REFERENCES players(id) ON DELETE SET NULL,
    FOREIGN KEY (match_id) REFERENCES matches(id) ON DELETE SET NULL,
    FOREIGN KEY (match_participant_id) REFERENCES match_participant(id) ON DELETE SET NULL
);

-- Player lives
CREATE TABLE IF NOT EXISTS player_lives (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    match_participant_id INTEGER NOT NULL,
    spawn_time DATETIME NOT NULL,
    death_time DATETIME, -- NULL if still alive or disconnected
    cause_of_death TEXT, -- optional: e.g., killed, disconnected, etc.
    FOREIGN KEY (match_participant_id) REFERENCES match_participant(id) ON DELETE CASCADE
);

-- Optional maps table (some queries may reference it)
CREATE TABLE IF NOT EXISTS maps (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    map_name TEXT NOT NULL,
    scenario TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Schema version
CREATE TABLE IF NOT EXISTS schema_version (
    id INTEGER PRIMARY KEY,
    version INTEGER NOT NULL,
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT OR IGNORE INTO schema_version (id, version) VALUES (1, 1);