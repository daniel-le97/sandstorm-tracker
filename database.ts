import { Database } from "bun:sqlite";

let db: Database | null = null;

// Get or create database connection
function getDatabase (): Database {
    if ( !db )
    {
        const dbPath = process.env.TEST_DB_PATH || "sandstorm_stats.db";
        db = new Database( dbPath );
        // Enable foreign key constraints
        db.run( "PRAGMA foreign_keys = ON;" );
    }
    return db;
}

// Create tables for tracking player statistics
export function initializeDatabase () {
    const database = getDatabase();
    console.log( "🗄️  Initializing database tables..." );

    // Players table - stores basic player information
    database.run( `
        CREATE TABLE IF NOT EXISTS players (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            steam_id TEXT UNIQUE NOT NULL,
            player_name TEXT NOT NULL,
            first_seen DATETIME DEFAULT CURRENT_TIMESTAMP,
            last_seen DATETIME DEFAULT CURRENT_TIMESTAMP,
            total_playtime_minutes INTEGER DEFAULT 0,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )
    `);

    // Player sessions table - tracks when players join/leave
    database.run( `
        CREATE TABLE IF NOT EXISTS player_sessions (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            player_id INTEGER NOT NULL,
            join_time DATETIME NOT NULL,
            leave_time DATETIME,
            duration_minutes INTEGER,
            map_name TEXT,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (player_id) REFERENCES players(id) ON DELETE CASCADE
        )
    `);

    // Kills table - tracks all kill events
    database.run( `
        CREATE TABLE IF NOT EXISTS kills (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            killer_id INTEGER,
            victim_id INTEGER,
            weapon TEXT NOT NULL,
            kill_type TEXT NOT NULL CHECK (kill_type IN ('player_kill', 'team_kill', 'suicide')),
            killer_team INTEGER,
            victim_team INTEGER,
            map_name TEXT,
            round_number INTEGER DEFAULT 1,
            timestamp DATETIME NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (killer_id) REFERENCES players(id) ON DELETE CASCADE,
            FOREIGN KEY (victim_id) REFERENCES players(id) ON DELETE CASCADE
        )
    `);

    // Maps table - stores information about maps/scenarios
    database.run( `
        CREATE TABLE IF NOT EXISTS maps (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            map_name TEXT UNIQUE NOT NULL,
            scenario TEXT,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )
    `);

    // Map rounds table - tracks round results
    database.run( `
        CREATE TABLE IF NOT EXISTS map_rounds (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            map_id INTEGER NOT NULL,
            round_number INTEGER NOT NULL,
            winning_team INTEGER,
            win_reason TEXT,
            end_time DATETIME,
            duration_seconds INTEGER,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (map_id) REFERENCES maps(id) ON DELETE CASCADE
        )
    `);

    // Weapon statistics - tracks per-player weapon performance
    database.run( `
        CREATE TABLE IF NOT EXISTS weapon_stats (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            player_id INTEGER NOT NULL,
            weapon_name TEXT NOT NULL,
            kills INTEGER DEFAULT 0,
            deaths INTEGER DEFAULT 0,
            team_kills INTEGER DEFAULT 0,
            suicides INTEGER DEFAULT 0,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (player_id) REFERENCES players(id) ON DELETE CASCADE,
            UNIQUE(player_id, weapon_name)
        )
    `);

    // Player round stats - per-round player performance
    database.run( `
        CREATE TABLE IF NOT EXISTS player_round_stats (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            player_id INTEGER NOT NULL,
            round_id INTEGER NOT NULL,
            kills INTEGER DEFAULT 0,
            deaths INTEGER DEFAULT 0,
            team_kills INTEGER DEFAULT 0,
            suicides INTEGER DEFAULT 0,
            score INTEGER DEFAULT 0,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (player_id) REFERENCES players(id) ON DELETE CASCADE,
            FOREIGN KEY (round_id) REFERENCES map_rounds(id) ON DELETE CASCADE,
            UNIQUE(player_id, round_id)
        )
    `);

    // Chat commands table - tracks !command usage
    database.run( `
        CREATE TABLE IF NOT EXISTS chat_commands (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            player_id INTEGER NOT NULL,
            command TEXT NOT NULL,
            arguments TEXT,
            timestamp DATETIME NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (player_id) REFERENCES players(id) ON DELETE CASCADE
        )
    `);

    // Create indexes for better query performance
    database.run( `CREATE INDEX IF NOT EXISTS idx_players_steam_id ON players(steam_id)` );
    database.run( `CREATE INDEX IF NOT EXISTS idx_kills_killer_id ON kills(killer_id)` );
    database.run( `CREATE INDEX IF NOT EXISTS idx_kills_victim_id ON kills(victim_id)` );
    database.run( `CREATE INDEX IF NOT EXISTS idx_kills_timestamp ON kills(timestamp)` );
    database.run( `CREATE INDEX IF NOT EXISTS idx_player_sessions_player_id ON player_sessions(player_id)` );
    database.run( `CREATE INDEX IF NOT EXISTS idx_weapon_stats_player_id ON weapon_stats(player_id)` );
    database.run( `CREATE INDEX IF NOT EXISTS idx_map_rounds_map_id ON map_rounds(map_id)` );
    database.run( `CREATE INDEX IF NOT EXISTS idx_player_round_stats_player_id ON player_round_stats(player_id)` );
    database.run( `CREATE INDEX IF NOT EXISTS idx_player_round_stats_round_id ON player_round_stats(round_id)` );

    console.log( "✅ Database tables created successfully!" );
}

// Prepared statements for better performance
function createStatements () {
    const database = getDatabase();
    return {
        // Player operations
        upsertPlayer: database.prepare( `
        INSERT INTO players (steam_id, player_name, last_seen)
        VALUES (?, ?, CURRENT_TIMESTAMP)
        ON CONFLICT(steam_id) DO UPDATE SET
            player_name = excluded.player_name,
            last_seen = CURRENT_TIMESTAMP,
            updated_at = CURRENT_TIMESTAMP
        RETURNING id
    `),

        getPlayer: database.prepare( `
        SELECT * FROM players WHERE steam_id = ?
    `),

        getPlayerByName: database.prepare( `
        SELECT * FROM players WHERE player_name = ?
    `),

        // Session operations
        startSession: database.prepare( `
        INSERT INTO player_sessions (player_id, join_time, map_name)
        VALUES (?, ?, ?)
        RETURNING id
    `),

        endSession: database.prepare( `
        UPDATE player_sessions 
        SET leave_time = ?, duration_minutes = ?
        WHERE id = ?
    `),

        updatePlayerPlaytime: database.prepare( `
        UPDATE players 
        SET total_playtime_minutes = total_playtime_minutes + ?, 
            last_seen = CURRENT_TIMESTAMP,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = ?
    `),

        // Kill operations
        insertKill: database.prepare( `
        INSERT INTO kills (killer_id, victim_id, weapon, kill_type, killer_team, victim_team, map_name, round_number, timestamp)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
    `),

        // Map operations
        upsertMap: database.prepare( `
        INSERT INTO maps (map_name, scenario)
        VALUES (?, ?)
        ON CONFLICT(map_name) DO UPDATE SET
            scenario = excluded.scenario,
            updated_at = CURRENT_TIMESTAMP
        RETURNING id
    `),

        // Skip insertRound for now due to schema issues
        insertRound: null, // database.prepare( `` ),

        // Weapon stats
        upsertWeaponStats: database.prepare( `
        INSERT INTO weapon_stats (player_id, weapon_name, kills, deaths, team_kills, suicides)
        VALUES (?, ?, ?, ?, ?, ?)
        ON CONFLICT(player_id, weapon_name) DO UPDATE SET
            kills = kills + excluded.kills,
            deaths = deaths + excluded.deaths,
            team_kills = team_kills + excluded.team_kills,
            suicides = suicides + excluded.suicides,
            updated_at = CURRENT_TIMESTAMP
    `),

        // Skip round stats for now due to schema issues
        upsertPlayerRoundStats: null, // database.prepare( `` ),

        // Chat commands
        insertChatCommand: database.prepare( `
        INSERT INTO chat_commands (player_id, command, arguments, timestamp)
        VALUES (?, ?, ?, ?)
    `),

        // Statistics queries
        getPlayerStats: database.prepare( `
        SELECT 
            p.player_name,
            p.steam_id,
            p.total_playtime_minutes,
            COALESCE(kills.total_kills, 0) as total_kills,
            COALESCE(deaths.total_deaths, 0) as total_deaths,
            COALESCE(kills.team_kills, 0) as team_kills,
            COALESCE(kills.suicides, 0) as suicides,
            CASE 
                WHEN COALESCE(deaths.total_deaths, 0) = 0 THEN COALESCE(kills.total_kills, 0)
                ELSE ROUND(CAST(COALESCE(kills.total_kills, 0) AS FLOAT) / CAST(COALESCE(deaths.total_deaths, 0) AS FLOAT), 2)
            END as kdr
        FROM players p
        LEFT JOIN (
            SELECT 
                killer_id,
                SUM(CASE WHEN kill_type = 'player_kill' THEN 1 ELSE 0 END) as total_kills,
                SUM(CASE WHEN kill_type = 'team_kill' THEN 1 ELSE 0 END) as team_kills,
                SUM(CASE WHEN kill_type = 'suicide' THEN 1 ELSE 0 END) as suicides
            FROM kills
            GROUP BY killer_id
        ) kills ON p.id = kills.killer_id
        LEFT JOIN (
            SELECT 
                victim_id,
                SUM(CASE WHEN kill_type IN ('player_kill', 'team_kill', 'suicide') THEN 1 ELSE 0 END) as total_deaths
            FROM kills
            GROUP BY victim_id
        ) deaths ON p.id = deaths.victim_id
        WHERE p.steam_id = ?
    `),

        getTopPlayers: database.prepare( `
        SELECT 
            p.player_name,
            COALESCE(kills.total_kills, 0) as total_kills,
            COALESCE(deaths.total_deaths, 0) as total_deaths,
            CASE 
                WHEN COALESCE(deaths.total_deaths, 0) = 0 THEN COALESCE(kills.total_kills, 0)
                ELSE ROUND(CAST(COALESCE(kills.total_kills, 0) AS FLOAT) / CAST(COALESCE(deaths.total_deaths, 0) AS FLOAT), 2)
            END as kdr
        FROM players p
        LEFT JOIN (
            SELECT 
                killer_id,
                SUM(CASE WHEN kill_type = 'player_kill' THEN 1 ELSE 0 END) as total_kills
            FROM kills
            GROUP BY killer_id
        ) kills ON p.id = kills.killer_id
        LEFT JOIN (
            SELECT 
                victim_id,
                SUM(CASE WHEN kill_type IN ('player_kill', 'team_kill', 'suicide') THEN 1 ELSE 0 END) as total_deaths
            FROM kills
            GROUP BY victim_id
        ) deaths ON p.id = deaths.victim_id
        WHERE COALESCE(kills.total_kills, 0) > 0
        ORDER BY total_kills DESC, kdr DESC
        LIMIT ?
    `),

        getPlayerWeapons: database.prepare( `
        SELECT 
            weapon_name,
            SUM(kills) as kills,
            SUM(deaths) as deaths,
            SUM(team_kills) as team_kills,
            SUM(suicides) as suicides
        FROM weapon_stats 
        WHERE player_id = (SELECT id FROM players WHERE steam_id = ?)
        GROUP BY weapon_name
        ORDER BY kills DESC
        LIMIT ?
    `)
    };
}

let statements: ReturnType<typeof createStatements> | null = null;

// Get statements, creating them if needed
export function getStatements () {
    if ( !statements )
    {
        statements = createStatements();
    }
    return statements;
}

// Initialize database on import
initializeDatabase();

// Export function to get database instance for direct queries in tests
export default function getDefaultDatabase () {
    return getDatabase();
}