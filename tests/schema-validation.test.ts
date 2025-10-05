import { Database } from "bun:sqlite";
import { beforeAll, describe, expect, test } from "bun:test";
import { unlinkSync } from "fs";

describe( "Database Schema Validation", () => {
    let db: Database;
    const testDbPath = "tests/databases/test_schema_validation.db";

    beforeAll( async () => {
        // Clean up any existing test database
        try
        {
            unlinkSync( testDbPath );
        } catch ( e )
        {
            // File doesn't exist, that's fine
        }

        // Set test environment
        process.env.TEST_DB_PATH = testDbPath;

        // Import and initialize database
        const dbModule = await import( "../src/database.ts" );
        db = dbModule.default();
    } );

    test( "All expected tables exist", () => {
        const expectedTables = [
            "servers",
            "players",
            "player_sessions",
            "kills",
            "weapon_stats",
            "maps",
            "map_rounds",
            "player_round_stats",
            "schema_version",
        ];

        const tables = db.query( 'SELECT name FROM sqlite_master WHERE type="table" ORDER BY name' ).all() as {
            name: string;
        }[];
        const tableNames = tables.map( ( t ) => t.name ).filter( ( name ) => name !== "sqlite_sequence" );

        expectedTables.forEach( ( expectedTable ) => {
            expect( tableNames ).toContain( expectedTable );
        } );
    } );

    test( "Servers table has correct structure", () => {
        const columns = db.query( "PRAGMA table_info(servers)" ).all() as { name: string; pk: number; }[];
        const columnNames = columns.map( ( col ) => col.name );

        expect( columnNames ).toContain( "id" );
        expect( columnNames ).toContain( "server_id" );
        expect( columnNames ).toContain( "server_name" );
        expect( columnNames ).toContain( "config_id" );
        expect( columnNames ).toContain( "log_path" );
        expect( columnNames ).toContain( "enabled" );
        expect( columnNames ).toContain( "description" );

        // Check primary key
        const pkColumn = columns.find( ( col ) => col.pk === 1 );
        expect( pkColumn?.name ).toBe( "id" );
    } );

    test( "Players table has correct foreign key to servers", () => {
        const foreignKeys = db.query( "PRAGMA foreign_key_list(players)" ).all() as {
            table: string;
            from: string;
            to: string;
        }[];
        const serverFk = foreignKeys.find( ( fk ) => fk.table === "servers" );

        expect( serverFk ).toBeDefined();
        expect( serverFk?.from ).toBe( "server_id" );
        expect( serverFk?.to ).toBe( "id" );
    } );

    test( "All multi-server tables have server_id foreign key", () => {
        const multiServerTables = [
            "players",
            "player_sessions",
            "kills",
            "weapon_stats",
            "maps",
            "map_rounds",
            "player_round_stats",
        ];

        multiServerTables.forEach( ( tableName ) => {
            const foreignKeys = db.query( `PRAGMA foreign_key_list(${ tableName })` ).all() as {
                table: string;
                from: string;
                to: string;
            }[];
            const serverFk = foreignKeys.find( ( fk ) => fk.table === "servers" );

            expect( serverFk, `${ tableName } should have foreign key to servers table` ).toBeDefined();
            expect( serverFk?.from ).toBe( "server_id" );
        } );
    } );

    test( "Foreign key constraints are enabled", () => {
        const pragmaResult = db.query( "PRAGMA foreign_keys" ).get() as { foreign_keys: number; };
        expect( pragmaResult.foreign_keys ).toBe( 1 );
    } );

    test( "Database schema version is tracked", () => {
        const version = db.query( "PRAGMA user_version" ).get() as { user_version: number; };
        expect( typeof version.user_version ).toBe( "number" );
        expect( version.user_version ).toBeGreaterThanOrEqual( 0 );
    } );

    test( "Can insert server and query it back", () => {
        // Insert a test server
        const uniqueId = Date.now().toString();
        const serverId = `test-server-schema-${ uniqueId }`;
        const serverName = "Test Server Schema";
        const configId = `test-config-schema-${ uniqueId }`;
        const logPath = "/test/logs/schema.log";

        const insertResult = db
            .query(
                `
            INSERT INTO servers (server_id, server_name, config_id, log_path, description)
            VALUES (?, ?, ?, ?, ?)
            RETURNING id
        `
            )
            .get( serverId, serverName, configId, logPath, "Schema validation test server" ) as { id: number; };

        expect( insertResult ).toBeDefined();
        expect( typeof insertResult.id ).toBe( "number" );

        // Query it back
        const server = db
            .query(
                `
            SELECT * FROM servers WHERE server_id = ?
        `
            )
            .get( serverId ) as { server_id: string; server_name: string; config_id: string; log_path: string; };

        expect( server ).toBeDefined();
        expect( server.server_id ).toBe( serverId );
        expect( server.server_name ).toBe( serverName );
        expect( server.config_id ).toBe( configId );
        expect( server.log_path ).toBe( logPath );
    } );

    test( "Foreign key constraints work correctly", () => {
        // First create a server
        const fkUniqueId = Date.now().toString() + "-fk";
        const serverResult = db
            .query(
                `
            INSERT INTO servers (server_id, server_name, config_id, log_path)
            VALUES (?, ?, ?, ?)
            RETURNING id
        `
            )
            .get( `fk-test-server-${ fkUniqueId }`, "FK Test Server", `fk-test-config-${ fkUniqueId }`, "/test/fk.log" ) as { id: number; };

        const serverDbId = serverResult.id;

        // Create a player linked to this server
        const playerResult = db
            .query(
                `
            INSERT INTO players (steam_id, player_name, server_id)
            VALUES (?, ?, ?)
            RETURNING id
        `
            )
            .get( "76561198000000001", "Test Player", serverDbId ) as { id: number; };

        expect( playerResult ).toBeDefined();
        expect( typeof playerResult.id ).toBe( "number" );

        // Verify the player is linked to the correct server
        const player = db
            .query(
                `
            SELECT p.*, s.server_id, s.server_name
            FROM players p
            JOIN servers s ON p.server_id = s.id
            WHERE p.id = ?
        `
            )
            .get( playerResult.id ) as { server_id: string; server_name: string; };

        expect( player.server_id ).toBe( `fk-test-server-${ fkUniqueId }` );
        expect( player.server_name ).toBe( "FK Test Server" );

        // Try to create a player with invalid server_id (should fail)
        expect( () => {
            db.query(
                `
                INSERT INTO players (steam_id, player_name, server_id)
                VALUES (?, ?, ?)
            `
            ).run( "76561198000000002", "Invalid Player", 99999 );
        } ).toThrow();
    } );
} );
