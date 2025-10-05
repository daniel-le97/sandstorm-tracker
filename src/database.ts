import { Database } from "bun:sqlite";
import { mkdirSync } from "fs";
import { dirname } from "path";

let db: Database | null = null;
let currentDbPath: string | null = null;

// Database connection errors
export class DatabaseError extends Error {
    public override cause?: Error;

    constructor(message: string, cause?: Error) {
        super(message);
        this.name = "DatabaseError";
        this.cause = cause;
    }
}

// Get or create database connection with comprehensive error handling
function getDatabase(): Database {
    try {
        let dbPath = process.env.TEST_DB_PATH || process.env.SANDSTORM_DB_PATH || "sandstorm_stats.db";

        // If it's a test database, ensure it goes in ./tests/databases directory
        if (process.env.TEST_DB_PATH) {
            dbPath = `tests/databases/${process.env.TEST_DB_PATH}`;

            // Create the tests/databases directory if it doesn't exist
            try {
                mkdirSync(dirname(dbPath), { recursive: true });
            } catch (e) {
                const error = e instanceof Error ? e : new Error(String(e));
                console.warn(`Could not create test database directory: ${error.message}`);
            }
        }

        // Reset connection if the database path has changed (for test isolation)
        if (db && currentDbPath !== dbPath) {
            try {
                db.close();
                console.log(`Database connection reset: ${currentDbPath} -> ${dbPath}`);
            } catch (e) {
                console.warn(`⚠️ Error closing previous database connection:`, e);
            }
            db = null;
            currentDbPath = null;
        }

        if (!db) {
            try {
                db = new Database(dbPath);
                currentDbPath = dbPath;
                console.log(`Database connected: ${dbPath}`);

                // Configure database for optimal performance and concurrency
                db.run("PRAGMA foreign_keys = ON;");
                db.run("PRAGMA journal_mode = WAL;");
                db.run("PRAGMA synchronous = NORMAL;");
                db.run("PRAGMA cache_size = 1000;");
                db.run(`
                    CREATE TABLE IF NOT EXISTS servers (
                        id INTEGER PRIMARY KEY AUTOINCREMENT,
                        server_id TEXT UNIQUE NOT NULL,
                        server_name TEXT NOT NULL,
                        config_id TEXT UNIQUE NOT NULL,
                        log_path TEXT NOT NULL,
                        enabled BOOLEAN DEFAULT 1,
                        description TEXT,
                        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP)`);
                return db;
            } catch (error) {
                const dbError = error instanceof Error ? error : new Error(String(error));
                throw new DatabaseError(`Failed to open database: ${dbPath}`, dbError);
            }
        }

        return db;
    } catch (error) {
        if (error instanceof DatabaseError) {
            throw error;
        }
        throw new DatabaseError(
            "Unexpected error in database connection",
            error instanceof Error ? error : new Error(String(error))
        );
    }
}

// Create database tables
function createTables() {
    const database = getDatabase();
    console.log("Initializing database tables...");

    // Servers table - tracks configured servers
    database.run(`
        CREATE TABLE IF NOT EXISTS servers (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            server_id TEXT UNIQUE NOT NULL,
            server_name TEXT NOT NULL,
            config_id TEXT UNIQUE NOT NULL,
            log_path TEXT NOT NULL,
            enabled BOOLEAN DEFAULT 1,
            description TEXT,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )
    `);

    // Players table - stores basic player information (now per-server)
    database.run(`
        CREATE TABLE IF NOT EXISTS players (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            steam_id TEXT NOT NULL,
            player_name TEXT NOT NULL,
            server_id INTEGER NOT NULL,
            first_seen DATETIME DEFAULT CURRENT_TIMESTAMP,
            last_seen DATETIME DEFAULT CURRENT_TIMESTAMP,
            total_playtime_minutes INTEGER DEFAULT 0,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE,
            UNIQUE(steam_id, server_id)
        )
    `);

    // Player sessions table - tracks when players join/leave
    database.run(`
        CREATE TABLE IF NOT EXISTS player_sessions (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            player_id INTEGER NOT NULL,
            server_id INTEGER NOT NULL,
            join_time DATETIME NOT NULL,
            leave_time DATETIME,
            duration_minutes INTEGER,
            map_name TEXT,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (player_id) REFERENCES players(id) ON DELETE CASCADE,
            FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE
        )
    `);

    // Kills table - tracks all kill events
    database.run(`
        CREATE TABLE IF NOT EXISTS kills (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            killer_id INTEGER,
            victim_id INTEGER,
            server_id INTEGER NOT NULL,
            weapon TEXT NOT NULL,
            kill_type TEXT NOT NULL CHECK (kill_type IN ('player_kill', 'team_kill', 'suicide')),
            killer_team INTEGER,
            victim_team INTEGER,
            map_name TEXT,
            round_number INTEGER DEFAULT 1,
            timestamp DATETIME NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (killer_id) REFERENCES players(id) ON DELETE CASCADE,
            FOREIGN KEY (victim_id) REFERENCES players(id) ON DELETE CASCADE,
            FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE
        )
    `);

    // Maps table - stores information about maps/scenarios per server
    database.run(`
        CREATE TABLE IF NOT EXISTS maps (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            map_name TEXT NOT NULL,
            scenario TEXT,
            server_id INTEGER NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE,
            UNIQUE(map_name, server_id)
        )
    `);

    // Map rounds table - tracks round results
    database.run(`
        CREATE TABLE IF NOT EXISTS map_rounds (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            map_id INTEGER NOT NULL,
            server_id INTEGER NOT NULL,
            round_number INTEGER NOT NULL,
            winning_team INTEGER,
            win_reason TEXT,
            end_time DATETIME,
            duration_seconds INTEGER,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (map_id) REFERENCES maps(id) ON DELETE CASCADE,
            FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE
        )
    `);

    // Weapon statistics - tracks per-player weapon performance
    database.run(`
        CREATE TABLE IF NOT EXISTS weapon_stats (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            player_id INTEGER NOT NULL,
            server_id INTEGER NOT NULL,
            weapon_name TEXT NOT NULL,
            kills INTEGER DEFAULT 0,
            deaths INTEGER DEFAULT 0,
            team_kills INTEGER DEFAULT 0,
            suicides INTEGER DEFAULT 0,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (player_id) REFERENCES players(id) ON DELETE CASCADE,
            FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE,
            UNIQUE(player_id, weapon_name, server_id)
        )
    `);

    // Player round stats - per-round player performance
    database.run(`
        CREATE TABLE IF NOT EXISTS player_round_stats (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            player_id INTEGER NOT NULL,
            round_id INTEGER NOT NULL,
            server_id INTEGER NOT NULL,
            kills INTEGER DEFAULT 0,
            deaths INTEGER DEFAULT 0,
            team_kills INTEGER DEFAULT 0,
            suicides INTEGER DEFAULT 0,
            score INTEGER DEFAULT 0,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (player_id) REFERENCES players(id) ON DELETE CASCADE,
            FOREIGN KEY (round_id) REFERENCES map_rounds(id) ON DELETE CASCADE,
            FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE,
            UNIQUE(player_id, round_id)
        )
    `);

    // Chat commands table - tracks !command usage
    database.run(`
        CREATE TABLE IF NOT EXISTS chat_commands (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            player_id INTEGER NOT NULL,
            server_id INTEGER NOT NULL,
            command TEXT NOT NULL,
            arguments TEXT,
            timestamp DATETIME NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (player_id) REFERENCES players(id) ON DELETE CASCADE,
            FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE
        )
    `);

    // Matches table - tracks complete game sessions/matches
    database.run(`
        CREATE TABLE IF NOT EXISTS matches (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            server_id INTEGER NOT NULL,
            match_name TEXT,
            start_time DATETIME NOT NULL,
            end_time DATETIME,
            duration_minutes INTEGER,
            status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'completed', 'aborted')),
            total_players INTEGER DEFAULT 0,
            max_players INTEGER DEFAULT 0,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE
        )
    `);

    // Match participants table - tracks which players participated in each match
    database.run(`
        CREATE TABLE IF NOT EXISTS match_participants (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            match_id INTEGER NOT NULL,
            player_id INTEGER NOT NULL,
            server_id INTEGER NOT NULL,
            join_time DATETIME NOT NULL,
            leave_time DATETIME,
            duration_minutes INTEGER,
            final_kills INTEGER DEFAULT 0,
            final_deaths INTEGER DEFAULT 0,
            final_team_kills INTEGER DEFAULT 0,
            final_suicides INTEGER DEFAULT 0,
            final_score INTEGER DEFAULT 0,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (match_id) REFERENCES matches(id) ON DELETE CASCADE,
            FOREIGN KEY (player_id) REFERENCES players(id) ON DELETE CASCADE,
            FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE,
            UNIQUE(match_id, player_id)
        )
    `);

    // Match maps table - links matches to the maps/scenarios played
    database.run(`
        CREATE TABLE IF NOT EXISTS match_maps (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            match_id INTEGER NOT NULL,
            map_id INTEGER NOT NULL,
            server_id INTEGER NOT NULL,
            sequence_order INTEGER DEFAULT 1,
            start_time DATETIME,
            end_time DATETIME,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (match_id) REFERENCES matches(id) ON DELETE CASCADE,
            FOREIGN KEY (map_id) REFERENCES maps(id) ON DELETE CASCADE,
            FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE,
            UNIQUE(match_id, map_id, sequence_order)
        )
    `);

    // Create indexes for better query performance
    database.run(`CREATE INDEX IF NOT EXISTS idx_servers_server_id ON servers(server_id)`);
    database.run(`CREATE INDEX IF NOT EXISTS idx_servers_config_id ON servers(config_id)`);
    database.run(`CREATE INDEX IF NOT EXISTS idx_players_steam_id ON players(steam_id)`);
    database.run(`CREATE INDEX IF NOT EXISTS idx_players_server_id ON players(server_id)`);
    database.run(`CREATE INDEX IF NOT EXISTS idx_players_steam_server ON players(steam_id, server_id)`);
    database.run(`CREATE INDEX IF NOT EXISTS idx_kills_killer_id ON kills(killer_id)`);
    database.run(`CREATE INDEX IF NOT EXISTS idx_kills_victim_id ON kills(victim_id)`);
    database.run(`CREATE INDEX IF NOT EXISTS idx_kills_server_id ON kills(server_id)`);
    database.run(`CREATE INDEX IF NOT EXISTS idx_kills_timestamp ON kills(timestamp)`);
    database.run(`CREATE INDEX IF NOT EXISTS idx_player_sessions_player_id ON player_sessions(player_id)`);
    database.run(`CREATE INDEX IF NOT EXISTS idx_player_sessions_server_id ON player_sessions(server_id)`);
    database.run(`CREATE INDEX IF NOT EXISTS idx_weapon_stats_player_id ON weapon_stats(player_id)`);
    database.run(`CREATE INDEX IF NOT EXISTS idx_weapon_stats_server_id ON weapon_stats(server_id)`);
    database.run(`CREATE INDEX IF NOT EXISTS idx_maps_server_id ON maps(server_id)`);
    database.run(`CREATE INDEX IF NOT EXISTS idx_map_rounds_map_id ON map_rounds(map_id)`);
    database.run(`CREATE INDEX IF NOT EXISTS idx_map_rounds_server_id ON map_rounds(server_id)`);
    database.run(`CREATE INDEX IF NOT EXISTS idx_player_round_stats_player_id ON player_round_stats(player_id)`);
    database.run(`CREATE INDEX IF NOT EXISTS idx_player_round_stats_round_id ON player_round_stats(round_id)`);
    database.run(`CREATE INDEX IF NOT EXISTS idx_player_round_stats_server_id ON player_round_stats(server_id)`);
    database.run(`CREATE INDEX IF NOT EXISTS idx_chat_commands_server_id ON chat_commands(server_id)`);
    database.run(`CREATE INDEX IF NOT EXISTS idx_matches_server_id ON matches(server_id)`);
    database.run(`CREATE INDEX IF NOT EXISTS idx_matches_status ON matches(status)`);
    database.run(`CREATE INDEX IF NOT EXISTS idx_matches_start_time ON matches(start_time)`);
    database.run(`CREATE INDEX IF NOT EXISTS idx_match_participants_match_id ON match_participants(match_id)`);
    database.run(`CREATE INDEX IF NOT EXISTS idx_match_participants_player_id ON match_participants(player_id)`);
    database.run(`CREATE INDEX IF NOT EXISTS idx_match_participants_server_id ON match_participants(server_id)`);
    database.run(`CREATE INDEX IF NOT EXISTS idx_match_maps_match_id ON match_maps(match_id)`);
    database.run(`CREATE INDEX IF NOT EXISTS idx_match_maps_map_id ON match_maps(map_id)`);
    database.run(`CREATE INDEX IF NOT EXISTS idx_match_maps_server_id ON match_maps(server_id)`);

    console.log("✅ Database tables created successfully!");
}

// Database version tracking
function getSchemaVersion(): number {
    const database = getDatabase();
    try {
        const result = database.prepare("SELECT version FROM schema_version ORDER BY id DESC LIMIT 1").get() as
            | { version: number }
            | undefined;
        return result?.version || 0;
    } catch (error) {
        // Table doesn't exist, create it
        database.run(`
            CREATE TABLE IF NOT EXISTS schema_version (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                version INTEGER NOT NULL,
                migration_date DATETIME DEFAULT CURRENT_TIMESTAMP
            )
        `);
        database.run("INSERT INTO schema_version (version) VALUES (0)");
        return 0;
    }
}

function setSchemaVersion(version: number): void {
    const database = getDatabase();
    database.run("INSERT INTO schema_version (version) VALUES (?)", [version]);
}

// Migration functions
function migrateToVersion1(): void {
    const database = getDatabase();
    console.log("Migrating database to version 1 (multi-server support)...");

    // Check if this is a fresh database or needs migration
    const tables = database.prepare("SELECT name FROM sqlite_master WHERE type='table'").all() as { name: string }[];
    const tableNames = tables.map((t) => t.name);

    if (!tableNames.includes("servers")) {
        // This is either a fresh database or needs migration
        if (tableNames.includes("players")) {
            // Need to migrate existing data
            console.log("Backing up existing data...");

            // Create backup tables
            database.run("ALTER TABLE players RENAME TO players_backup");
            database.run("ALTER TABLE kills RENAME TO kills_backup");
            database.run("ALTER TABLE player_sessions RENAME TO player_sessions_backup");
            database.run("ALTER TABLE maps RENAME TO maps_backup");
            database.run("ALTER TABLE map_rounds RENAME TO map_rounds_backup");
            database.run("ALTER TABLE weapon_stats RENAME TO weapon_stats_backup");
            database.run("ALTER TABLE player_round_stats RENAME TO player_round_stats_backup");
            database.run("ALTER TABLE chat_commands RENAME TO chat_commands_backup");

            // Create new schema
            createTables();

            // Create default server entry for existing data
            const defaultServerId = database
                .prepare(
                    `
                INSERT INTO servers (server_id, server_name, config_id, log_path, description)
                VALUES (?, ?, ?, ?, ?)
                RETURNING id
            `
                )
                .get(
                    "default-server",
                    "Legacy Server",
                    "legacy",
                    "legacy-path",
                    "Migrated from single-server setup"
                ) as { id: number };

            // Migrate existing data
            migrateExistingData(database, defaultServerId.id);

            console.log("✅ Data migration completed!");
        } else {
            // Fresh database, just create new schema
            createTables();
        }
    }

    setSchemaVersion(1);
}

function migrateToVersion2(): void {
    const database = getDatabase();
    console.log("Migrating database to version 2 (match history support)...");

    // Add match history tables
    database.run(`
        CREATE TABLE IF NOT EXISTS matches (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            server_id INTEGER NOT NULL,
            match_name TEXT,
            start_time DATETIME NOT NULL,
            end_time DATETIME,
            duration_minutes INTEGER,
            status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'completed', 'aborted')),
            total_players INTEGER DEFAULT 0,
            max_players INTEGER DEFAULT 0,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE
        )
    `);

    database.run(`
        CREATE TABLE IF NOT EXISTS match_participants (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            match_id INTEGER NOT NULL,
            player_id INTEGER NOT NULL,
            server_id INTEGER NOT NULL,
            join_time DATETIME NOT NULL,
            leave_time DATETIME,
            duration_minutes INTEGER,
            final_kills INTEGER DEFAULT 0,
            final_deaths INTEGER DEFAULT 0,
            final_team_kills INTEGER DEFAULT 0,
            final_suicides INTEGER DEFAULT 0,
            final_score INTEGER DEFAULT 0,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (match_id) REFERENCES matches(id) ON DELETE CASCADE,
            FOREIGN KEY (player_id) REFERENCES players(id) ON DELETE CASCADE,
            FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE,
            UNIQUE(match_id, player_id)
        )
    `);

    database.run(`
        CREATE TABLE IF NOT EXISTS match_maps (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            match_id INTEGER NOT NULL,
            map_id INTEGER NOT NULL,
            server_id INTEGER NOT NULL,
            sequence_order INTEGER DEFAULT 1,
            start_time DATETIME,
            end_time DATETIME,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (match_id) REFERENCES matches(id) ON DELETE CASCADE,
            FOREIGN KEY (map_id) REFERENCES maps(id) ON DELETE CASCADE,
            FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE,
            UNIQUE(match_id, map_id, sequence_order)
        )
    `);

    // Add indexes
    database.run(`CREATE INDEX IF NOT EXISTS idx_matches_server_id ON matches(server_id)`);
    database.run(`CREATE INDEX IF NOT EXISTS idx_matches_status ON matches(status)`);
    database.run(`CREATE INDEX IF NOT EXISTS idx_matches_start_time ON matches(start_time)`);
    database.run(`CREATE INDEX IF NOT EXISTS idx_match_participants_match_id ON match_participants(match_id)`);
    database.run(`CREATE INDEX IF NOT EXISTS idx_match_participants_player_id ON match_participants(player_id)`);
    database.run(`CREATE INDEX IF NOT EXISTS idx_match_participants_server_id ON match_participants(server_id)`);
    database.run(`CREATE INDEX IF NOT EXISTS idx_match_maps_match_id ON match_maps(match_id)`);
    database.run(`CREATE INDEX IF NOT EXISTS idx_match_maps_map_id ON match_maps(map_id)`);
    database.run(`CREATE INDEX IF NOT EXISTS idx_match_maps_server_id ON match_maps(server_id)`);

    console.log("✅ Match history tables created successfully!");
    setSchemaVersion(2);
}

function migrateExistingData(database: any, serverId: number): void {
    console.log("Migrating existing data to new schema...");

    // Migrate players
    database.run(
        `
                INSERT INTO players (steam_id, player_name, server_id, first_seen, last_seen, total_playtime_minutes, created_at, updated_at)
                SELECT steam_id, player_name, ?, first_seen, last_seen, total_playtime_minutes, created_at, updated_at
                FROM players_backup
            `,
        [serverId]
    ); // Create player ID mapping
    const playerMapping = database
        .prepare(
            `
        SELECT p_new.id as new_id, p_old.id as old_id
        FROM players p_new
        JOIN players_backup p_old ON p_new.steam_id = p_old.steam_id
        WHERE p_new.server_id = ?
    `
        )
        .all([serverId]) as { new_id: number; old_id: number }[];

    const playerMap = new Map(playerMapping.map((p) => [p.old_id, p.new_id]));

    // Migrate maps
    database.run(
        `
        INSERT INTO maps (map_name, scenario, server_id, created_at, updated_at)
        SELECT map_name, scenario, ?, created_at, updated_at
        FROM maps_backup
    `,
        [serverId]
    );

    // Create map ID mapping
    const mapMapping = database
        .prepare(
            `
        SELECT m_new.id as new_id, m_old.id as old_id
        FROM maps m_new
        JOIN maps_backup m_old ON m_new.map_name = m_old.map_name
        WHERE m_new.server_id = ?
    `
        )
        .all([serverId]) as { new_id: number; old_id: number }[];

    const mapMap = new Map(mapMapping.map((m) => [m.old_id, m.new_id]));

    // Migrate sessions
    for (const session of database.prepare("SELECT * FROM player_sessions_backup").all()) {
        const newPlayerId = playerMap.get(session.player_id);
        if (newPlayerId) {
            database.run(
                `
                INSERT INTO player_sessions (player_id, server_id, join_time, leave_time, duration_minutes, map_name, created_at)
                VALUES (?, ?, ?, ?, ?, ?, ?)
            `,
                newPlayerId,
                serverId,
                session.join_time,
                session.leave_time,
                session.duration_minutes,
                session.map_name,
                session.created_at
            );
        }
    }

    // Migrate kills
    for (const kill of database.prepare("SELECT * FROM kills_backup").all()) {
        const newKillerId = kill.killer_id ? playerMap.get(kill.killer_id) : null;
        const newVictimId = kill.victim_id ? playerMap.get(kill.victim_id) : null;

        database.run(
            `
            INSERT INTO kills (killer_id, victim_id, server_id, weapon, kill_type, killer_team, victim_team, map_name, round_number, timestamp, created_at)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        `,
            newKillerId,
            newVictimId,
            serverId,
            kill.weapon,
            kill.kill_type,
            kill.killer_team,
            kill.victim_team,
            kill.map_name,
            kill.round_number,
            kill.timestamp,
            kill.created_at
        );
    }

    // Migrate weapon stats
    for (const weapon of database.prepare("SELECT * FROM weapon_stats_backup").all()) {
        const newPlayerId = playerMap.get(weapon.player_id);
        if (newPlayerId) {
            database.run(
                `
                INSERT INTO weapon_stats (player_id, server_id, weapon_name, kills, deaths, team_kills, suicides, created_at, updated_at)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
            `,
                newPlayerId,
                serverId,
                weapon.weapon_name,
                weapon.kills,
                weapon.deaths,
                weapon.team_kills,
                weapon.suicides,
                weapon.created_at,
                weapon.updated_at
            );
        }
    }

    // Migrate map rounds
    for (const round of database.prepare("SELECT * FROM map_rounds_backup").all()) {
        const newMapId = mapMap.get(round.map_id);
        if (newMapId) {
            const newRoundResult = database.run(
                `
                INSERT INTO map_rounds (map_id, server_id, round_number, winning_team, win_reason, end_time, duration_seconds, created_at)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?)
                RETURNING id
            `,
                newMapId,
                serverId,
                round.round_number,
                round.winning_team,
                round.win_reason,
                round.end_time,
                round.duration_seconds,
                round.created_at
            );

            // Migrate player round stats for this round
            for (const playerRound of database
                .prepare("SELECT * FROM player_round_stats_backup WHERE round_id = ?")
                .all(round.id)) {
                const newPlayerId = playerMap.get(playerRound.player_id);
                if (newPlayerId) {
                    database.run(
                        `
                        INSERT INTO player_round_stats (player_id, round_id, server_id, kills, deaths, team_kills, suicides, score, created_at)
                        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
                    `,
                        newPlayerId,
                        newRoundResult.lastInsertRowid,
                        serverId,
                        playerRound.kills,
                        playerRound.deaths,
                        playerRound.team_kills,
                        playerRound.suicides,
                        playerRound.score,
                        playerRound.created_at
                    );
                }
            }
        }
    }

    // Migrate chat commands
    for (const command of database.prepare("SELECT * FROM chat_commands_backup").all()) {
        const newPlayerId = playerMap.get(command.player_id);
        if (newPlayerId) {
            database.run(
                `
                INSERT INTO chat_commands (player_id, server_id, command, arguments, timestamp, created_at)
                VALUES (?, ?, ?, ?, ?, ?)
            `,
                newPlayerId,
                serverId,
                command.command,
                command.arguments,
                command.timestamp,
                command.created_at
            );
        }
    }

    console.log(`Migrated data for ${playerMapping.length} players and ${mapMapping.length} maps`);
}

// Server management functions
// Upsert server and return database ID with error handling
export function upsertServer(
    serverId: string,
    serverName: string,
    configId: string,
    logPath: string,
    description?: string
): number {
    try {
        // Validate required parameters
        if (!serverId || !serverName || !configId || !logPath) {
            throw new DatabaseError("Missing required parameters for upsertServer");
        }

        const result = getStatements().upsertServer.get(
            serverId,
            serverName,
            configId,
            logPath,
            description || null
        ) as { id: number } | undefined;

        if (!result || typeof result.id !== "number") {
            throw new DatabaseError(`Failed to upsert server: ${serverName} (${serverId})`);
        }

        return result.id;
    } catch (error) {
        if (error instanceof DatabaseError) {
            throw error;
        }
        throw new DatabaseError(
            `Error upserting server ${serverName}`,
            error instanceof Error ? error : new Error(String(error))
        );
    }
}

export function getServerByUuid(
    serverUuid: string
): { id: number; server_id: string; server_name: string; config_id: string } | undefined {
    const database = getDatabase();
    return database
        .prepare("SELECT id, server_id, server_name, config_id FROM servers WHERE server_id = ?")
        .get(serverUuid) as any;
}

export function getServerByConfigId(
    configId: string
): { id: number; server_id: string; server_name: string; config_id: string } | undefined {
    const database = getDatabase();
    return database
        .prepare("SELECT id, server_id, server_name, config_id FROM servers WHERE config_id = ?")
        .get(configId) as any;
}

export function getAllServers(): {
    id: number;
    server_id: string;
    server_name: string;
    config_id: string;
    enabled: boolean;
}[] {
    const database = getDatabase();
    return database
        .prepare("SELECT id, server_id, server_name, config_id, enabled FROM servers ORDER BY server_name")
        .all() as any[];
}

// Public function for manual database initialization with error handling
export function initializeDatabase() {
    try {
        console.log("Initializing database...");

        // Ensure database connection is established
        const database = getDatabase();

        // Create tables if they don't exist
        try {
            createTables();
            console.log("✓ Database tables created/verified");
        } catch (error) {
            throw new DatabaseError(
                "Failed to create database tables",
                error instanceof Error ? error : new Error(String(error))
            );
        }

        // Run migrations
        try {
            const currentVersion = getSchemaVersion();

            if (currentVersion === 0) {
                migrateToVersion1();
                console.log("✓ Database migrated to version 1");
            }
            if (currentVersion <= 1) {
                migrateToVersion2();
                console.log("✓ Database migrated to version 2 (match history)");
            }
        } catch (error) {
            throw new DatabaseError(
                "Failed to run database migrations",
                error instanceof Error ? error : new Error(String(error))
            );
        }

        // Prepare statements
        try {
            statements = createStatements();
            console.log("✓ Database statements prepared");
        } catch (error) {
            throw new DatabaseError(
                "Failed to prepare database statements",
                error instanceof Error ? error : new Error(String(error))
            );
        }

        console.log(`✅ Database initialization complete (schema version: ${getSchemaVersion()})`);
    } catch (error) {
        console.error("❌ Critical database initialization error:", error);
        throw error;
    }
}

// Prepared statements for better performance
function createStatements() {
    const database = getDatabase();
    return {
        // Server operations
        upsertServer: database.prepare(`
        INSERT INTO servers (server_id, server_name, config_id, log_path, description)
        VALUES (?, ?, ?, ?, ?)
        ON CONFLICT(server_id) DO UPDATE SET
            server_name = excluded.server_name,
            log_path = excluded.log_path,
            description = excluded.description,
            updated_at = CURRENT_TIMESTAMP
        RETURNING id
    `),

        // Player operations (now server-specific)
        upsertPlayer: database.prepare(`
        INSERT INTO players (steam_id, player_name, server_id, last_seen)
        VALUES (?, ?, ?, CURRENT_TIMESTAMP)
        ON CONFLICT(steam_id, server_id) DO UPDATE SET
            player_name = excluded.player_name,
            last_seen = CURRENT_TIMESTAMP,
            updated_at = CURRENT_TIMESTAMP
        RETURNING id
    `),

        getPlayer: database.prepare(`
        SELECT * FROM players WHERE steam_id = ? AND server_id = ?
    `),

        getPlayerByName: database.prepare(`
        SELECT * FROM players WHERE player_name = ? AND server_id = ?
    `),

        // Get player across all servers
        getPlayerGlobal: database.prepare(`
        SELECT * FROM players WHERE steam_id = ?
    `),

        // Update player Steam ID (for when we learn the real Steam ID after initial unknown join)
        updatePlayerSteamId: database.prepare(`
        UPDATE players SET steam_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?
    `),

        // Session operations
        startSession: database.prepare(`
        INSERT INTO player_sessions (player_id, server_id, join_time, map_name)
        VALUES (?, ?, ?, ?)
        RETURNING id
    `),

        endSession: database.prepare(`
        UPDATE player_sessions
        SET leave_time = ?, duration_minutes = ?
        WHERE id = ?
    `),

        updatePlayerPlaytime: database.prepare(`
        UPDATE players
        SET total_playtime_minutes = total_playtime_minutes + ?,
            last_seen = CURRENT_TIMESTAMP,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = ?
    `),

        // Kill operations
        insertKill: database.prepare(`
        INSERT INTO kills (killer_id, victim_id, server_id, weapon, kill_type, killer_team, victim_team, map_name, round_number, timestamp)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `),

        // Map operations
        upsertMap: database.prepare(`
        INSERT INTO maps (map_name, scenario, server_id)
        VALUES (?, ?, ?)
        ON CONFLICT(map_name, server_id) DO UPDATE SET
            scenario = excluded.scenario,
            updated_at = CURRENT_TIMESTAMP
        RETURNING id
    `),

        // Skip insertRound for now due to schema issues
        insertRound: null, // database.prepare( `` ),

        // Weapon stats
        upsertWeaponStats: database.prepare(`
        INSERT INTO weapon_stats (player_id, server_id, weapon_name, kills, deaths, team_kills, suicides)
        VALUES (?, ?, ?, ?, ?, ?, ?)
        ON CONFLICT(player_id, weapon_name, server_id) DO UPDATE SET
            kills = kills + excluded.kills,
            deaths = deaths + excluded.deaths,
            team_kills = team_kills + excluded.team_kills,
            suicides = suicides + excluded.suicides,
            updated_at = CURRENT_TIMESTAMP
    `),

        // Skip round stats for now due to schema issues
        upsertPlayerRoundStats: null, // database.prepare( `` ),

        // Chat commands
        insertChatCommand: database.prepare(`
        INSERT INTO chat_commands (player_id, server_id, command, arguments, timestamp)
        VALUES (?, ?, ?, ?, ?)
    `),

        // Statistics queries (server-specific)
        getPlayerStats: database.prepare(`
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
            WHERE server_id = ?
            GROUP BY killer_id
        ) kills ON p.id = kills.killer_id
        LEFT JOIN (
            SELECT
                victim_id,
                SUM(CASE WHEN kill_type IN ('player_kill', 'team_kill', 'suicide') THEN 1 ELSE 0 END) as total_deaths
            FROM kills
            WHERE server_id = ?
            GROUP BY victim_id
        ) deaths ON p.id = deaths.victim_id
        WHERE p.steam_id = ? AND p.server_id = ?
    `),

        getTopPlayers: database.prepare(`
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
            WHERE server_id = ?
            GROUP BY killer_id
        ) kills ON p.id = kills.killer_id
        LEFT JOIN (
            SELECT
                victim_id,
                SUM(CASE WHEN kill_type IN ('player_kill', 'team_kill', 'suicide') THEN 1 ELSE 0 END) as total_deaths
            FROM kills
            WHERE server_id = ?
            GROUP BY victim_id
        ) deaths ON p.id = deaths.victim_id
        WHERE p.server_id = ? AND COALESCE(kills.total_kills, 0) > 0
        ORDER BY total_kills DESC, kdr DESC
        LIMIT ?
    `),

        getPlayerWeapons: database.prepare(`
        SELECT
            weapon_name,
            SUM(kills) as kills,
            SUM(deaths) as deaths,
            SUM(team_kills) as team_kills,
            SUM(suicides) as suicides
        FROM weapon_stats
        WHERE player_id = (SELECT id FROM players WHERE steam_id = ? AND server_id = ?) AND server_id = ?
        GROUP BY weapon_name
        ORDER BY kills DESC
        LIMIT ?
    `),

        // Global statistics queries (across all servers)
        getPlayerStatsGlobal: database.prepare(`
        SELECT
            p.steam_id,
            MIN(p.player_name) as player_name,
            SUM(p.total_playtime_minutes) as total_playtime_minutes,
            COALESCE(SUM(kills.total_kills), 0) as total_kills,
            COALESCE(SUM(deaths.total_deaths), 0) as total_deaths,
            COALESCE(SUM(kills.team_kills), 0) as team_kills,
            COALESCE(SUM(kills.suicides), 0) as suicides,
            CASE
                WHEN COALESCE(SUM(deaths.total_deaths), 0) = 0 THEN COALESCE(SUM(kills.total_kills), 0)
                ELSE ROUND(CAST(COALESCE(SUM(kills.total_kills), 0) AS FLOAT) / CAST(COALESCE(SUM(deaths.total_deaths), 0) AS FLOAT), 2)
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
        GROUP BY p.steam_id
    `),

        // Match history operations
        startMatch: database.prepare(`
        INSERT INTO matches (server_id, match_name, start_time, total_players, max_players)
        VALUES (?, ?, ?, ?, ?)
        RETURNING id
    `),

        endMatch: database.prepare(`
        UPDATE matches
        SET end_time = ?, duration_minutes = ?, status = 'completed', updated_at = CURRENT_TIMESTAMP
        WHERE id = ? AND server_id = ?
    `),

        abortMatch: database.prepare(`
        UPDATE matches
        SET end_time = ?, status = 'aborted', updated_at = CURRENT_TIMESTAMP
        WHERE id = ? AND server_id = ?
    `),

        updateMatchPlayerCount: database.prepare(`
        UPDATE matches
        SET total_players = ?, max_players = MAX(max_players, ?), updated_at = CURRENT_TIMESTAMP
        WHERE id = ? AND server_id = ?
    `),

        addMatchParticipant: database.prepare(`
        INSERT INTO match_participants (match_id, player_id, server_id, join_time)
        VALUES (?, ?, ?, ?)
        ON CONFLICT(match_id, player_id) DO UPDATE SET
            join_time = excluded.join_time
        RETURNING id
    `),

        endMatchParticipant: database.prepare(`
        UPDATE match_participants
        SET leave_time = ?, duration_minutes = ?, final_kills = ?, final_deaths = ?,
            final_team_kills = ?, final_suicides = ?, final_score = ?
        WHERE match_id = ? AND player_id = ?
    `),

        addMatchMap: database.prepare(`
        INSERT INTO match_maps (match_id, map_id, server_id, sequence_order, start_time)
        VALUES (?, ?, ?, ?, ?)
        ON CONFLICT(match_id, map_id, sequence_order) DO UPDATE SET
            start_time = excluded.start_time
        RETURNING id
    `),

        endMatchMap: database.prepare(`
        UPDATE match_maps
        SET end_time = ?
        WHERE match_id = ? AND map_id = ? AND sequence_order = ?
    `),

        // Match queries
        getActiveMatches: database.prepare(`
        SELECT * FROM matches
        WHERE server_id = ? AND status = 'active'
        ORDER BY start_time DESC
    `),

        getMatchHistory: database.prepare(`
        SELECT
            m.*,
            COUNT(DISTINCT mp.player_id) as participant_count,
            GROUP_CONCAT(DISTINCT maps.map_name) as maps_played
        FROM matches m
        LEFT JOIN match_participants mp ON m.id = mp.match_id
        LEFT JOIN match_maps mm ON m.id = mm.match_id
        LEFT JOIN maps ON mm.map_id = maps.id
        WHERE m.server_id = ?
        GROUP BY m.id
        ORDER BY m.start_time DESC
        LIMIT ?
    `),

        getMatchDetails: database.prepare(`
        SELECT
            m.*,
            (SELECT COUNT(*) FROM match_participants WHERE match_id = m.id) as participant_count
        FROM matches m
        WHERE m.id = ? AND m.server_id = ?
    `),

        getMatchParticipants: database.prepare(`
        SELECT
            mp.*,
            p.player_name,
            p.steam_id
        FROM match_participants mp
        JOIN players p ON mp.player_id = p.id
        WHERE mp.match_id = ? AND mp.server_id = ?
        ORDER BY mp.join_time ASC
    `),

        getMatchMaps: database.prepare(`
        SELECT
            mm.*,
            m.map_name,
            m.scenario
        FROM match_maps mm
        JOIN maps m ON mm.map_id = m.id
        WHERE mm.match_id = ? AND mm.server_id = ?
        ORDER BY mm.sequence_order ASC
    `),

        getPlayerMatchHistory: database.prepare(`
        SELECT
            m.id as match_id,
            m.match_name,
            m.start_time,
            m.end_time,
            m.duration_minutes as match_duration,
            mp.join_time,
            mp.leave_time,
            mp.duration_minutes as participation_duration,
            mp.final_kills,
            mp.final_deaths,
            mp.final_team_kills,
            mp.final_suicides,
            mp.final_score,
            GROUP_CONCAT(DISTINCT maps.map_name) as maps_played
        FROM matches m
        JOIN match_participants mp ON m.id = mp.match_id
        LEFT JOIN match_maps mm ON m.id = mm.match_id
        LEFT JOIN maps ON mm.map_id = maps.id
        WHERE mp.player_id = (SELECT id FROM players WHERE steam_id = ? AND server_id = ?)
        AND m.server_id = ?
        GROUP BY m.id, mp.id
        ORDER BY m.start_time DESC
        LIMIT ?
    `),
    };
}

let statements: ReturnType<typeof createStatements> | null = null;
let statementsDbPath: string | null = null;

// Get statements, creating them if needed
export function getStatements() {
    const dbPath = process.env.TEST_DB_PATH || "sandstorm_stats.db";

    // Reset statements if the database path has changed
    if (statements && statementsDbPath !== dbPath) {
        statements = null;
        statementsDbPath = null;
    }

    if (!statements) {
        statements = createStatements();
        statementsDbPath = dbPath;
    }
    return statements;
}

// Initialize database on import
initializeDatabase();

// Export function to get database instance for direct queries in tests
export default function getDefaultDatabase() {
    return getDatabase();
}
