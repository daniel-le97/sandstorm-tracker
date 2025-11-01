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

-- Matches
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

-- Performance indexes for matches
CREATE INDEX IF NOT EXISTS idx_matches_server ON matches(server_id);
CREATE INDEX IF NOT EXISTS idx_matches_start_time ON matches(start_time);
CREATE INDEX IF NOT EXISTS idx_matches_end_time ON matches(end_time);
-- Composite index for active match queries (WHERE server_id AND end_time IS NULL)
CREATE INDEX IF NOT EXISTS idx_matches_server_end_time ON matches(server_id, end_time);
-- Composite index for time-based queries per server
CREATE INDEX IF NOT EXISTS idx_matches_server_start_time ON matches(server_id, start_time DESC);

-- Match player stats (one row per player per match, reused across reconnects)
CREATE TABLE IF NOT EXISTS match_player_stats (
    match_id INTEGER NOT NULL,
    player_id INTEGER NOT NULL,
    team INTEGER,
    
    -- Combat stats (aggregated across all sessions if player reconnects)
    kills INTEGER DEFAULT 0,
    assists INTEGER DEFAULT 0,
    deaths INTEGER DEFAULT 0,
    friendly_fire_kills INTEGER DEFAULT 0,
    score INTEGER DEFAULT 0,
    objectives_captured INTEGER DEFAULT 0,
    objectives_destroyed INTEGER DEFAULT 0,
    
    -- Session tracking
    total_play_time INTEGER DEFAULT 0,  -- Total seconds in match (sum of all sessions)
    session_count INTEGER DEFAULT 1,     -- How many times they joined this match
    first_joined_at DATETIME,
    last_left_at DATETIME,
    is_currently_connected INTEGER DEFAULT 1,  -- 0 = disconnected, 1 = connected
    
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    PRIMARY KEY (match_id, player_id),
    FOREIGN KEY (match_id) REFERENCES matches(id) ON DELETE CASCADE,
    FOREIGN KEY (player_id) REFERENCES players(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_match_player_stats_match ON match_player_stats(match_id);
CREATE INDEX IF NOT EXISTS idx_match_player_stats_player ON match_player_stats(player_id);
-- Index for connection status queries
CREATE INDEX IF NOT EXISTS idx_match_player_stats_connected ON match_player_stats(is_currently_connected);
-- Composite index for stale connection detection (WHERE is_currently_connected = 1)
CREATE INDEX IF NOT EXISTS idx_match_player_stats_match_connected ON match_player_stats(match_id, is_currently_connected);

-- Match weapon stats (weapon usage per player per match)
CREATE TABLE IF NOT EXISTS match_weapon_stats (
    match_id INTEGER NOT NULL,
    player_id INTEGER NOT NULL,
    weapon_name TEXT NOT NULL,
    kills INTEGER DEFAULT 0,
    assists INTEGER DEFAULT 0,
    PRIMARY KEY (match_id, player_id, weapon_name),
    FOREIGN KEY (match_id, player_id) REFERENCES match_player_stats(match_id, player_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_match_weapon_stats_match ON match_weapon_stats(match_id);
CREATE INDEX IF NOT EXISTS idx_match_weapon_stats_player ON match_weapon_stats(player_id);
-- Index for weapon-specific queries
CREATE INDEX IF NOT EXISTS idx_match_weapon_stats_weapon ON match_weapon_stats(weapon_name);
-- Composite index for player weapon stats across matches
CREATE INDEX IF NOT EXISTS idx_match_weapon_stats_player_weapon ON match_weapon_stats(player_id, weapon_name);