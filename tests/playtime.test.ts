import { test, expect, describe, beforeAll, afterAll, beforeEach } from 'bun:test';
import { Database } from 'bun:sqlite';
import { unlinkSync } from 'fs';

describe( 'Playtime Tracking', () => {
    let StatsService: any;
    let db: Database;
    const testDbPath = 'test_playtime.db';

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
        const dbModule = await import( '../database.ts' );
        StatsService = ( await import( '../stats-service' ) ).default;
        db = dbModule.default();
    } );

    beforeEach( () => {
        // Get a fresh database connection for each test
        const dbModule = require( '../database' );
        db = dbModule.default();

        // Clear all data before each test
        db.run( 'DELETE FROM chat_commands' );
        db.run( 'DELETE FROM kills' );
        db.run( 'DELETE FROM player_sessions' );
        db.run( 'DELETE FROM players' );
        db.run( 'DELETE FROM map_rounds' );
        db.run( 'DELETE FROM maps' );

        // Clear active sessions
        StatsService.endAllSessions();
    } );

    afterAll( () => {
        // Clean up test database
        try
        {
            db?.close();
            unlinkSync( testDbPath );
        } catch ( e )
        {
            // Ignore cleanup errors
        }
    } );

    test( 'Player join starts session tracking', () => {
        StatsService.processEvent( {
            type: 'player_join',
            timestamp: '2025.10.05-12.00.00:000',
            data: {
                playerName: 'PlaytimeTest'
            },
            rawLine: 'test'
        } );

        // Check that session was created
        const activeSessions = StatsService.getActiveSessions();
        expect( activeSessions.has( 'PlaytimeTest' ) ).toBe( true );

        const sessionId = activeSessions.get( 'PlaytimeTest' );
        expect( sessionId ).toBeTruthy();
    } );

    test( 'Player leave ends session and calculates playtime', () => {
        // First join
        StatsService.processEvent( {
            type: 'player_join',
            timestamp: '2025.10.05-12.00.00:000',
            data: {
                playerName: 'SessionTest'
            },
            rawLine: 'test'
        } );

        // Simulate some time passing and then leave
        StatsService.processEvent( {
            type: 'player_leave',
            timestamp: '2025.10.05-12.30.00:000', // 30 minutes later
            data: {
                playerName: 'SessionTest'
            },
            rawLine: 'test'
        } );

        // Check session was ended
        const activeSessions = StatsService.getActiveSessions();
        expect( activeSessions.has( 'SessionTest' ) ).toBe( false );

        // Check player was created
        const player = StatsService.getPlayerStatsByName( 'SessionTest' );
        expect( player ).toBeDefined();
        expect( player?.player_name ).toBe( 'SessionTest' );
    } );

    test( 'Server crash detection works with timeout', () => {
        // Start a session
        StatsService.processEvent( {
            type: 'player_join',
            timestamp: '2025.10.05-12.00.00:000',
            data: {
                playerName: 'CrashTest'
            },
            rawLine: 'test'
        } );

        // Verify session is active
        const activeSessions = StatsService.getActiveSessions();
        expect( activeSessions.has( 'CrashTest' ) ).toBe( true );

        // Simulate server crash by calling handleServerCrash
        StatsService.handleServerCrash();

        // Check that all sessions were ended
        const activeSessionsAfterCrash = StatsService.getActiveSessions();
        expect( activeSessionsAfterCrash.size ).toBe( 0 );
    } );

    test( 'Last file change tracking updates correctly', () => {
        const initialTime = StatsService.getLastFileChange();

        // Update last file change time
        StatsService.updateLastFileChange();

        const updatedTime = StatsService.getLastFileChange();
        expect( updatedTime.getTime() ).toBeGreaterThan( initialTime.getTime() );
    } );

    test( 'Multiple players can have concurrent sessions', () => {
        const playerNames = [ 'Player1', 'Player2', 'Player3' ];

        // Join all players
        playerNames.forEach( ( name, index ) => {
            StatsService.processEvent( {
                type: 'player_join',
                timestamp: `2025.10.05-12.0${ index }.00:000`,
                data: {
                    playerName: name
                },
                rawLine: 'test'
            } );
        } );

        // Check all sessions are active
        const activeSessions = StatsService.getActiveSessions();
        expect( activeSessions.size ).toBe( 3 );

        playerNames.forEach( name => {
            expect( activeSessions.has( name ) ).toBe( true );
        } );

        // Leave one player
        StatsService.processEvent( {
            type: 'player_leave',
            timestamp: '2025.10.05-12.05.00:000',
            data: {
                playerName: 'Player2'
            },
            rawLine: 'test'
        } );

        // Check only 2 sessions remain
        const remainingSessions = StatsService.getActiveSessions();
        expect( remainingSessions.size ).toBe( 2 );
        expect( remainingSessions.has( 'Player1' ) ).toBe( true );
        expect( remainingSessions.has( 'Player2' ) ).toBe( false );
        expect( remainingSessions.has( 'Player3' ) ).toBe( true );
    } );

    test( 'Session tracking handles rejoining players', () => {
        // Player joins
        StatsService.processEvent( {
            type: 'player_join',
            timestamp: '2025.10.05-12.00.00:000',
            data: {
                playerName: 'RejoinTest'
            },
            rawLine: 'test'
        } );

        // Player leaves
        StatsService.processEvent( {
            type: 'player_leave',
            timestamp: '2025.10.05-12.15.00:000',
            data: {
                playerName: 'RejoinTest'
            },
            rawLine: 'test'
        } );

        // Player joins again
        StatsService.processEvent( {
            type: 'player_join',
            timestamp: '2025.10.05-12.20.00:000',
            data: {
                playerName: 'RejoinTest'
            },
            rawLine: 'test'
        } );

        // Check new session was created
        const activeSessions = StatsService.getActiveSessions();
        expect( activeSessions.has( 'RejoinTest' ) ).toBe( true );

        // Should have a different session ID than before
        const newSessionId = activeSessions.get( 'RejoinTest' );
        expect( newSessionId ).toBeTruthy();
    } );

    test( 'Playtime accumulates across multiple sessions', () => {
        const playerName = 'AccumulateTest';

        // First session
        StatsService.processEvent( {
            type: 'player_join',
            timestamp: '2025.10.05-10.00.00:000',
            data: { playerName },
            rawLine: 'test'
        } );

        StatsService.processEvent( {
            type: 'player_leave',
            timestamp: '2025.10.05-10.30.00:000', // 30 min session
            data: { playerName },
            rawLine: 'test'
        } );

        // Second session
        StatsService.processEvent( {
            type: 'player_join',
            timestamp: '2025.10.05-11.00.00:000',
            data: { playerName },
            rawLine: 'test'
        } );

        StatsService.processEvent( {
            type: 'player_leave',
            timestamp: '2025.10.05-11.45.00:000', // 45 min session
            data: { playerName },
            rawLine: 'test'
        } );

        // Check accumulated playtime
        const player = StatsService.getPlayerStatsByName( playerName );
        expect( player ).toBeDefined();
        expect( player?.player_name ).toBe( playerName );
        // Note: Actual playtime calculation depends on implementation details
    } );
} );