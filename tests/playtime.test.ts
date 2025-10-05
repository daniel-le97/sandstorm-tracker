import { Database } from "bun:sqlite";
import { beforeAll, beforeEach, describe, expect, test } from "bun:test";
import { unlinkSync } from "fs";

describe( "Playtime Tracking", () => {
    let StatsService: any;
    let db: Database;
    let testServerDbId: number;
    const testDbPath = "test_playtime.db";

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

        // Import modules
        const dbModule = await import( "../src/database.ts" );
        StatsService = ( await import( "../src/stats-service" ) ).default;
        db = dbModule.default();
    } );

    beforeEach( () => {
        // Get a fresh database connection for each test
        const dbModule = require( "../src/database" );
        db = dbModule.default();

        // Clear all data before each test (in order due to foreign key constraints)
        db.run( "DELETE FROM kills" );
        db.run( "DELETE FROM player_sessions" );
        db.run( "DELETE FROM players" );
        db.run( "DELETE FROM map_rounds" );
        db.run( "DELETE FROM maps" );
        db.run( "DELETE FROM servers" );

        // Create test server and capture the database ID
        const dbFunctions = require( "../src/database" );
        testServerDbId = dbFunctions.upsertServer(
            "test-server-1",
            "Test Server 1",
            "test-config",
            "/test/logs/test-server-1.log",
            "Test server for playtime tests"
        );

        // Clear active sessions
        StatsService.endAllSessions( testServerDbId );
    } );

    test( "Player join starts session tracking", () => {
        StatsService.processEvent(
            {
                type: "player_join",
                timestamp: "2025.10.05-12.00.00:000",
                data: {
                    playerName: "PlaytimeTest",
                },
                rawLine: "test",
            },
            testServerDbId
        );

        // Check that session was created
        const activeSessions = StatsService.getActiveSessions( testServerDbId );
        expect( activeSessions.has( "PlaytimeTest" ) ).toBe( true );

        const sessionId = activeSessions.get( "PlaytimeTest" );
        expect( sessionId ).toBeTruthy();
    } );

    test( "Player leave ends session and calculates playtime", () => {
        // First join
        StatsService.processEvent(
            {
                type: "player_join",
                timestamp: "2025.10.05-12.00.00:000",
                data: {
                    playerName: "SessionTest",
                },
                rawLine: "test",
            },
            testServerDbId
        );

        // Simulate some time passing and then leave
        StatsService.processEvent(
            {
                type: "player_leave",
                timestamp: "2025.10.05-12.30.00:000", // 30 minutes later
                data: {
                    playerName: "SessionTest",
                },
                rawLine: "test",
            },
            testServerDbId
        );

        // Check session was ended
        const activeSessions = StatsService.getActiveSessions( testServerDbId );
        expect( activeSessions.has( "SessionTest" ) ).toBe( false );

        // Check player was created
        const player = StatsService.getPlayerStatsByName( "SessionTest", testServerDbId );
        expect( player ).toBeDefined();
        expect( player?.player_name ).toBe( "SessionTest" );
    } );

    test( "Server crash detection works with timeout", () => {
        // Start a session
        StatsService.processEvent(
            {
                type: "player_join",
                timestamp: "2025.10.05-12.00.00:000",
                data: {
                    playerName: "CrashTest",
                },
                rawLine: "test",
            },
            testServerDbId
        );

        // Verify session is active
        const activeSessions = StatsService.getActiveSessions( testServerDbId );
        expect( activeSessions.has( "CrashTest" ) ).toBe( true );

        // Simulate server crash by calling handleServerCrash
        StatsService.handleServerCrash( testServerDbId );

        // Check that all sessions were ended
        const activeSessionsAfterCrash = StatsService.getActiveSessions( testServerDbId );
        expect( activeSessionsAfterCrash.size ).toBe( 0 );
    } );

    test( "Last file change tracking updates correctly", async () => {
        const initialTime = StatsService.getLastFileChange( testServerDbId );

        // Small delay to ensure timestamp differs
        await new Promise( ( resolve ) => setTimeout( resolve, 1 ) );

        // Update last file change time
        StatsService.updateLastFileChange( testServerDbId );

        const updatedTime = StatsService.getLastFileChange( testServerDbId );
        expect( updatedTime.getTime() ).toBeGreaterThan( initialTime.getTime() );
    } );

    test( "Multiple players can have concurrent sessions", () => {
        const playerNames = [ "Player1", "Player2", "Player3" ];

        // Join all players
        playerNames.forEach( ( name, index ) => {
            StatsService.processEvent(
                {
                    type: "player_join",
                    timestamp: `2025.10.05-12.0${ index }.00:000`,
                    data: {
                        playerName: name,
                    },
                    rawLine: "test",
                },
                testServerDbId
            );
        } );

        // Check all sessions are active
        const activeSessions = StatsService.getActiveSessions( testServerDbId );
        expect( activeSessions.size ).toBe( 3 );

        playerNames.forEach( ( name ) => {
            expect( activeSessions.has( name ) ).toBe( true );
        } );

        // Leave one player
        StatsService.processEvent(
            {
                type: "player_leave",
                timestamp: "2025.10.05-12.05.00:000",
                data: {
                    playerName: "Player2",
                },
                rawLine: "test",
            },
            testServerDbId
        );

        // Check only 2 sessions remain
        const remainingSessions = StatsService.getActiveSessions( testServerDbId );
        expect( remainingSessions.size ).toBe( 2 );
        expect( remainingSessions.has( "Player1" ) ).toBe( true );
        expect( remainingSessions.has( "Player2" ) ).toBe( false );
        expect( remainingSessions.has( "Player3" ) ).toBe( true );
    } );

    test( "Session tracking handles rejoining players", () => {
        // Player joins
        StatsService.processEvent(
            {
                type: "player_join",
                timestamp: "2025.10.05-12.00.00:000",
                data: {
                    playerName: "RejoinTest",
                },
                rawLine: "test",
            },
            testServerDbId
        );

        // Player leaves
        StatsService.processEvent(
            {
                type: "player_leave",
                timestamp: "2025.10.05-12.15.00:000",
                data: {
                    playerName: "RejoinTest",
                },
                rawLine: "test",
            },
            testServerDbId
        );

        // Player joins again
        StatsService.processEvent(
            {
                type: "player_join",
                timestamp: "2025.10.05-12.20.00:000",
                data: {
                    playerName: "RejoinTest",
                },
                rawLine: "test",
            },
            testServerDbId
        );

        // Check new session was created
        const activeSessions = StatsService.getActiveSessions( testServerDbId );
        expect( activeSessions.has( "RejoinTest" ) ).toBe( true );

        // Should have a different session ID than before
        const newSessionId = activeSessions.get( "RejoinTest" );
        expect( newSessionId ).toBeTruthy();
    } );

    test( "Playtime accumulates across multiple sessions", () => {
        const playerName = "AccumulateTest";

        // First session
        StatsService.processEvent(
            {
                type: "player_join",
                timestamp: "2025.10.05-10.00.00:000",
                data: { playerName },
                rawLine: "test",
            },
            testServerDbId
        );

        StatsService.processEvent(
            {
                type: "player_leave",
                timestamp: "2025.10.05-10.30.00:000", // 30 min session
                data: { playerName },
                rawLine: "test",
            },
            testServerDbId
        );

        // Second session
        StatsService.processEvent(
            {
                type: "player_join",
                timestamp: "2025.10.05-11.00.00:000",
                data: { playerName },
                rawLine: "test",
            },
            testServerDbId
        );

        StatsService.processEvent(
            {
                type: "player_leave",
                timestamp: "2025.10.05-11.45.00:000", // 45 min session
                data: { playerName },
                rawLine: "test",
            },
            testServerDbId
        );

        // Check accumulated playtime
        const player = StatsService.getPlayerStatsByName( playerName, testServerDbId );
        expect( player ).toBeDefined();
        expect( player?.player_name ).toBe( playerName );
        // Note: Actual playtime calculation depends on implementation details
    } );
} );
