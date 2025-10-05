// Type for detailed match information (getMatchDetails result)
export interface MatchDetails {
    match: Match;
    participants: MatchParticipant[];
    maps: {
        id: number;
        match_id: number;
        map_id: number;
        server_id: number;
        sequence_order: number;
        start_time: string | null;
        end_time: string | null;
        created_at: string;
        map_name: string;
        scenario: string | null;
    }[];
}
// Type for a single match record (matches table)
export interface Match {
    id: number;
    server_id: number;
    match_name: string | null;
    start_time: string;
    end_time: string | null;
    duration_minutes: number | null;
    status: "active" | "completed" | "aborted";
    total_players: number;
    max_players: number;
    created_at: string;
    updated_at: string;
}
// Type for match history (returned by getMatchHistory)
export interface MatchHistory {
    id: number;
    server_id: number;
    match_name: string | null;
    start_time: string;
    end_time: string | null;
    duration_minutes: number | null;
    status: string;
    total_players: number;
    max_players: number;
    created_at: string;
    updated_at: string;
    participant_count: number;
    maps_played: string | null;
}
// Type for match participants (joined with player info)
export interface MatchParticipant {
    id: number;
    match_id: number;
    player_id: number;
    server_id: number;
    join_time: string;
    leave_time: string | null;
    duration_minutes: number | null;
    final_kills: number;
    final_deaths: number;
    final_team_kills: number;
    final_suicides: number;
    final_score: number;
    created_at: string;
    player_name: string;
    steam_id: string;
}
import { Database } from "bun:sqlite";
import { mkdirSync } from "fs";
import { dirname } from "path";

// Database connection errors
export class DatabaseError extends Error {
    public override cause?: Error;

    constructor ( message: string, cause?: Error ) {
        super( message );
        this.name = "DatabaseError";
        this.cause = cause;
    }
}

// Database service singleton class
export class DatabaseService {
    private db: Database | null = null;
    private currentDbPath: string | null = null;
    private statements: ReturnType<typeof this.createStatements> | null = null;
    private statementsDbPath: string | null = null;
    private initialized = false;

    constructor () {
        // Don't initialize here to avoid circular dependency issues
    }

    private ensureInitialized () {
        if ( !this.initialized )
        {
            this.initializeDatabase();
            this.initialized = true;
        }
    }

    // Get or create database connection with comprehensive error handling
    private getDatabase (): Database {
        this.ensureInitialized();

        if ( !this.db )
        {
            throw new DatabaseError( "Database not initialized" );
        }

        return this.db;
    }

    // Create database tables
    private createTables () {
        if ( !this.db )
        {
            throw new DatabaseError( "Database connection not established" );
        }
        console.log( "Initializing database tables..." );

        // Servers table - tracks configured servers
        this.db.run( `
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
        this.db.run( `
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
        this.db.run( `
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
        this.db.run( `
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
        this.db.run( `
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
        this.db.run( `
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
        this.db.run( `
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
        this.db.run( `
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

        // Matches table - tracks complete game sessions/matches
        this.db.run( `
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
        this.db.run( `
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
        this.db.run( `
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
        this.db.run( `CREATE INDEX IF NOT EXISTS idx_servers_server_id ON servers(server_id)` );
        this.db.run( `CREATE INDEX IF NOT EXISTS idx_servers_config_id ON servers(config_id)` );
        this.db.run( `CREATE INDEX IF NOT EXISTS idx_players_steam_id ON players(steam_id)` );
        this.db.run( `CREATE INDEX IF NOT EXISTS idx_players_server_id ON players(server_id)` );
        this.db.run( `CREATE INDEX IF NOT EXISTS idx_players_steam_server ON players(steam_id, server_id)` );
        this.db.run( `CREATE INDEX IF NOT EXISTS idx_kills_killer_id ON kills(killer_id)` );
        this.db.run( `CREATE INDEX IF NOT EXISTS idx_kills_victim_id ON kills(victim_id)` );
        this.db.run( `CREATE INDEX IF NOT EXISTS idx_kills_server_id ON kills(server_id)` );
        this.db.run( `CREATE INDEX IF NOT EXISTS idx_kills_timestamp ON kills(timestamp)` );
        this.db.run( `CREATE INDEX IF NOT EXISTS idx_player_sessions_player_id ON player_sessions(player_id)` );
        this.db.run( `CREATE INDEX IF NOT EXISTS idx_player_sessions_server_id ON player_sessions(server_id)` );
        this.db.run( `CREATE INDEX IF NOT EXISTS idx_weapon_stats_player_id ON weapon_stats(player_id)` );
        this.db.run( `CREATE INDEX IF NOT EXISTS idx_weapon_stats_server_id ON weapon_stats(server_id)` );
        this.db.run( `CREATE INDEX IF NOT EXISTS idx_maps_server_id ON maps(server_id)` );
        this.db.run( `CREATE INDEX IF NOT EXISTS idx_map_rounds_map_id ON map_rounds(map_id)` );
        this.db.run( `CREATE INDEX IF NOT EXISTS idx_map_rounds_server_id ON map_rounds(server_id)` );
        this.db.run( `CREATE INDEX IF NOT EXISTS idx_player_round_stats_player_id ON player_round_stats(player_id)` );
        this.db.run( `CREATE INDEX IF NOT EXISTS idx_player_round_stats_round_id ON player_round_stats(round_id)` );
        this.db.run( `CREATE INDEX IF NOT EXISTS idx_player_round_stats_server_id ON player_round_stats(server_id)` );

        this.db.run( `CREATE INDEX IF NOT EXISTS idx_matches_server_id ON matches(server_id)` );
        this.db.run( `CREATE INDEX IF NOT EXISTS idx_matches_status ON matches(status)` );
        this.db.run( `CREATE INDEX IF NOT EXISTS idx_matches_start_time ON matches(start_time)` );
        this.db.run( `CREATE INDEX IF NOT EXISTS idx_match_participants_match_id ON match_participants(match_id)` );
        this.db.run( `CREATE INDEX IF NOT EXISTS idx_match_participants_player_id ON match_participants(player_id)` );
        this.db.run( `CREATE INDEX IF NOT EXISTS idx_match_participants_server_id ON match_participants(server_id)` );
        this.db.run( `CREATE INDEX IF NOT EXISTS idx_match_maps_match_id ON match_maps(match_id)` );
        this.db.run( `CREATE INDEX IF NOT EXISTS idx_match_maps_map_id ON match_maps(map_id)` );
        this.db.run( `CREATE INDEX IF NOT EXISTS idx_match_maps_server_id ON match_maps(server_id)` );

        // Schema version table - tracks database schema version
        this.db.run( `
        CREATE TABLE IF NOT EXISTS schema_version (
            id INTEGER PRIMARY KEY,
            version INTEGER NOT NULL,
            applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )
    `);

        // Insert initial schema version if not exists
        const versionExists = this.db.prepare( "SELECT COUNT(*) as count FROM schema_version" ).get() as { count: number; };
        if ( versionExists.count === 0 )
        {
            this.db.prepare( "INSERT INTO schema_version (version) VALUES (?)" ).run( 1 );
        }

        console.log( "Database tables created successfully!" );
    }



    // Migration functions - simplified for development
    private ensureCurrentSchema (): void {
        console.log( "Ensuring current database schema..." );

        // Just create the current schema - no migrations needed for development
        this.createTables();
    }





    // Server management functions
    // Upsert server and return database ID with error handling
    upsertServer (
        serverId: string,
        serverName: string,
        configId: string,
        logPath: string,
        description?: string
    ): number {
        this.ensureInitialized();
        try
        {
            // Validate required parameters
            if ( !serverId || !serverName || !configId || !logPath )
            {
                throw new DatabaseError( "Missing required parameters for upsertServer" );
            }

            const result = this.getStatements().upsertServer.get(
                serverId,
                serverName,
                configId,
                logPath,
                description || null
            ) as { id: number; } | undefined;

            if ( !result || typeof result.id !== "number" )
            {
                throw new DatabaseError( `Failed to upsert server: ${ serverName } (${ serverId })` );
            }

            return result.id;
        } catch ( error )
        {
            if ( error instanceof DatabaseError )
            {
                throw error;
            }
            throw new DatabaseError(
                `Error upserting server ${ serverName }`,
                error instanceof Error ? error : new Error( String( error ) )
            );
        }
    }

    getServerByUuid (
        serverUuid: string
    ): { id: number; server_id: string; server_name: string; config_id: string; } | undefined {
        this.ensureInitialized();
        if ( !this.db ) throw new DatabaseError( "Database not initialized" );
        return this.db
            .prepare( "SELECT id, server_id, server_name, config_id FROM servers WHERE server_id = ?" )
            .get( serverUuid ) as any;
    }

    getServerByConfigId (
        configId: string
    ): { id: number; server_id: string; server_name: string; config_id: string; } | undefined {
        this.ensureInitialized();
        if ( !this.db ) throw new DatabaseError( "Database not initialized" );
        return this.db
            .prepare( "SELECT id, server_id, server_name, config_id FROM servers WHERE config_id = ?" )
            .get( configId ) as any;
    }

    getAllServers (): {
        id: number;
        server_id: string;
        server_name: string;
        config_id: string;
        enabled: boolean;
    }[] {
        this.ensureInitialized();
        if ( !this.db ) throw new DatabaseError( "Database not initialized" );
        return this.db
            .prepare( "SELECT id, server_id, server_name, config_id, enabled FROM servers ORDER BY server_name" )
            .all() as any[];
    }

    // Public function for manual database initialization with error handling
    private initializeDatabase () {
        try
        {
            console.log( "Initializing database..." );

            // Create the database connection without calling ensureInitialized
            this.createDatabaseConnection();

            // Create tables if they don't exist
            try
            {
                this.createTables();
                console.log( "✓ Database tables created/verified" );
            } catch ( error )
            {
                throw new DatabaseError(
                    "Failed to create database tables",
                    error instanceof Error ? error : new Error( String( error ) )
                );
            }

            // Ensure current schema
            try
            {
                this.ensureCurrentSchema();
                console.log( "✓ Database schema ensured" );
            } catch ( error )
            {
                throw new DatabaseError(
                    "Failed to ensure database schema",
                    error instanceof Error ? error : new Error( String( error ) )
                );
            }

            // Prepare statements
            try
            {
                this.statements = this.createStatements();
                console.log( "✓ Database statements prepared" );
            } catch ( error )
            {
                throw new DatabaseError(
                    "Failed to prepare database statements",
                    error instanceof Error ? error : new Error( String( error ) )
                );
            }

            console.log( "✅ Database initialization complete" );
        } catch ( error )
        {
            console.error( "❌ Critical database initialization error:", error );
            throw error;
        }
    }

    // Create database connection without calling ensureInitialized (to avoid circular dependency)
    private createDatabaseConnection () {
        let dbPath = process.env.TEST_DB_PATH || process.env.SANDSTORM_DB_PATH || "sandstorm_stats.db";

        // If it's a test database, ensure it goes in ./tests/databases directory
        if ( process.env.TEST_DB_PATH )
        {
            dbPath = `${ process.env.TEST_DB_PATH }`;

            // Create the tests/databases directory if it doesn't exist
            try
            {
                mkdirSync( dirname( dbPath ), { recursive: true } );
            } catch ( e )
            {
                const error = e instanceof Error ? e : new Error( String( e ) );
                console.warn( `Could not create test database directory: ${ error.message }` );
            }
        }

        // Reset connection if the database path has changed (for test isolation)
        if ( this.db && this.currentDbPath !== dbPath )
        {
            try
            {
                this.db.close();
                console.log( `Database connection reset: ${ this.currentDbPath } -> ${ dbPath }` );
            } catch ( e )
            {
                console.warn( `⚠️ Error closing previous database connection:`, e );
            }
            this.db = null;
            this.currentDbPath = null;
        }

        if ( !this.db )
        {
            try
            {
                this.db = new Database( dbPath );
                this.currentDbPath = dbPath;
                console.log( `Database connected: ${ dbPath }` );

                // Configure database for optimal performance and concurrency
                this.db.run( "PRAGMA foreign_keys = ON;" );
                this.db.run( "PRAGMA journal_mode = WAL;" );
                this.db.run( "PRAGMA synchronous = NORMAL;" );
                this.db.run( "PRAGMA cache_size = 1000;" );
            } catch ( error )
            {
                const dbError = error instanceof Error ? error : new Error( String( error ) );
                throw new DatabaseError( `Failed to open database: ${ dbPath }`, dbError );
            }
        }
    }

    // Prepared statements for better performance
    private createStatements () {
        if ( !this.db )
        {
            throw new DatabaseError( "Database connection not established" );
        }
        const database = this.db;
        return {
            // Server operations
            upsertServer: database.prepare( `
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
            upsertPlayer: database.prepare( `
        INSERT INTO players (steam_id, player_name, server_id, last_seen)
        VALUES (?, ?, ?, CURRENT_TIMESTAMP)
        ON CONFLICT(steam_id, server_id) DO UPDATE SET
            player_name = excluded.player_name,
            last_seen = CURRENT_TIMESTAMP,
            updated_at = CURRENT_TIMESTAMP
        RETURNING id
    `),

            getPlayer: database.prepare( `
        SELECT * FROM players WHERE steam_id = ? AND server_id = ?
    `),

            getPlayerByName: database.prepare( `
        SELECT * FROM players WHERE player_name = ? AND server_id = ?
    `),

            // Get player across all servers
            getPlayerGlobal: database.prepare( `
        SELECT * FROM players WHERE steam_id = ?
    `),

            // Update player Steam ID (for when we learn the real Steam ID after initial unknown join)
            updatePlayerSteamId: database.prepare( `
        UPDATE players SET steam_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?
    `),

            // Session operations
            startSession: database.prepare( `
        INSERT INTO player_sessions (player_id, server_id, join_time, map_name)
        VALUES (?, ?, ?, ?)
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
        INSERT INTO kills (killer_id, victim_id, server_id, weapon, kill_type, killer_team, victim_team, map_name, round_number, timestamp)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `),

            // Map operations
            upsertMap: database.prepare( `
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
            upsertWeaponStats: database.prepare( `
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



            // Statistics queries (server-specific)
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

            getPlayerWeapons: database.prepare( `
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
            getPlayerStatsGlobal: database.prepare( `
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
            startMatch: database.prepare( `
        INSERT INTO matches (server_id, match_name, start_time, total_players, max_players)
        VALUES (?, ?, ?, ?, ?)
        RETURNING id
    `),

            endMatch: database.prepare( `
        UPDATE matches
        SET end_time = ?, duration_minutes = ?, status = 'completed', updated_at = CURRENT_TIMESTAMP
        WHERE id = ? AND server_id = ?
    `),

            abortMatch: database.prepare( `
        UPDATE matches
        SET end_time = ?, status = 'aborted', updated_at = CURRENT_TIMESTAMP
        WHERE id = ? AND server_id = ?
    `),

            updateMatchPlayerCount: database.prepare( `
        UPDATE matches
        SET total_players = ?, max_players = MAX(max_players, ?), updated_at = CURRENT_TIMESTAMP
        WHERE id = ? AND server_id = ?
    `),

            addMatchParticipant: database.prepare( `
        INSERT INTO match_participants (match_id, player_id, server_id, join_time)
        VALUES (?, ?, ?, ?)
        ON CONFLICT(match_id, player_id) DO UPDATE SET
            join_time = excluded.join_time
        RETURNING id
    `),

            endMatchParticipant: database.prepare( `
        UPDATE match_participants
        SET leave_time = ?, duration_minutes = ?, final_kills = ?, final_deaths = ?,
            final_team_kills = ?, final_suicides = ?, final_score = ?
        WHERE match_id = ? AND player_id = ?
    `),

            addMatchMap: database.prepare( `
        INSERT INTO match_maps (match_id, map_id, server_id, sequence_order, start_time)
        VALUES (?, ?, ?, ?, ?)
        ON CONFLICT(match_id, map_id, sequence_order) DO UPDATE SET
            start_time = excluded.start_time
        RETURNING id
    `),

            endMatchMap: database.prepare( `
        UPDATE match_maps
        SET end_time = ?
        WHERE match_id = ? AND map_id = ? AND sequence_order = ?
    `),

            // Match queries
            getActiveMatches: database.prepare( `
        SELECT * FROM matches
        WHERE server_id = ? AND status = 'active'
        ORDER BY start_time DESC
    `),

            getMatchHistory: database.prepare( `
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

            getMatchDetails: database.prepare( `
        SELECT
            m.*,
            (SELECT COUNT(*) FROM match_participants WHERE match_id = m.id) as participant_count
        FROM matches m
        WHERE m.id = ? AND m.server_id = ?
    `),

            getMatchParticipants: database.prepare( `
        SELECT
            mp.*,
            p.player_name,
            p.steam_id
        FROM match_participants mp
        JOIN players p ON mp.player_id = p.id
        WHERE mp.match_id = ? AND mp.server_id = ?
        ORDER BY mp.join_time ASC
    `),

            getMatchMaps: database.prepare( `
        SELECT
            mm.*,
            m.map_name,
            m.scenario
        FROM match_maps mm
        JOIN maps m ON mm.map_id = m.id
        WHERE mm.match_id = ? AND mm.server_id = ?
        ORDER BY mm.sequence_order ASC
    `),

            getPlayerMatchHistory: database.prepare( `
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

    // Get statements, creating them if needed
    getStatements () {
        this.ensureInitialized();
        const dbPath = process.env.TEST_DB_PATH || "sandstorm_stats.db";

        // Reset statements if the database path has changed
        if ( this.statements && this.statementsDbPath !== dbPath )
        {
            this.statements = null;
            this.statementsDbPath = null;
        }

        if ( !this.statements )
        {
            this.statements = this.createStatements();
            this.statementsDbPath = dbPath;
        }
        return this.statements;
    }

    // Export function to get database instance for direct queries in tests
    getDefaultDatabase () {
        this.ensureInitialized();
        if ( !this.db ) throw new DatabaseError( "Database not initialized" );
        return this.db;
    }
}

// Create and export singleton instance
export const DB = new DatabaseService();

// Export individual functions for backward compatibility
export const upsertServer = ( serverId: string, serverName: string, configId: string, logPath: string, description?: string ) =>
    DB.upsertServer( serverId, serverName, configId, logPath, description );

export const getServerByUuid = ( serverUuid: string ) =>
    DB.getServerByUuid( serverUuid );

export const getServerByConfigId = ( configId: string ) =>
    DB.getServerByConfigId( configId );

export const getAllServers = () =>
    DB.getAllServers();

export const getStatements = () =>
    DB.getStatements();

export const initializeDatabase = () =>
    DB;

export default function getDefaultDatabase () {
    return DB.getDefaultDatabase();
}
