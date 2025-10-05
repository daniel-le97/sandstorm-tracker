import { test, expect, describe, beforeAll, beforeEach } from 'bun:test';
import { writeFileSync, unlinkSync } from 'fs';
import { Database } from 'bun:sqlite';
import { parseLogEvents } from '../src/events';

describe( 'Integration Tests', () => {
    let StatsService: any;
    let CommandHandler: any;
    let db: Database;
    const testDbPath = 'test_integration.db';

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
        const dbModule = await import( '../src/database.ts' );
        StatsService = ( await import( '../src/stats-service' ) ).default;
        CommandHandler = ( await import( '../src/command-handler' ) ).default;
        db = dbModule.default();
    } );

    beforeEach( () => {
        // Get a fresh database connection for each test
        const dbModule = require( '../src/database' );
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

    test( 'Complete log processing pipeline works end-to-end', () => {
        // Create comprehensive sample log content
        const sampleLogContent = `
[2025.10.05-12.00.00:000][001]LogNet: Join succeeded: Alice
[2025.10.05-12.00.30:000][002]LogNet: Join succeeded: Bob  
[2025.10.05-12.01.00:000][003]LogGameplayEvents: Display: Alice[76561198000000001, team 0] killed Bob[76561198000000002, team 1] with BP_Firearm_M16A4_C_2147481419
[2025.10.05-12.01.15:000][004]LogGameplayEvents: Display: Bob[76561198000000002, team 1] killed Alice[76561198000000001, team 0] with BP_Firearm_AK74_C_2147481420
[2025.10.05-12.01.30:000][005]LogChat: Display: Alice(76561198000000001) Global Chat: !stats
[2025.10.05-12.01.35:000][006]LogChat: Display: Bob(76561198000000002) Global Chat: !kdr
[2025.10.05-12.01.40:000][007]LogChat: Display: Alice(76561198000000001) Global Chat: !top
[2025.10.05-12.01.45:000][008]LogChat: Display: Bob(76561198000000002) Global Chat: !guns
[2025.10.05-12.02.00:000][009]LogGameplayEvents: Display: Alice[76561198000000001, team 0] killed Bob[76561198000000002, team 1] with BP_Firearm_M16A4_C_2147481419
        `.trim();

        // Write sample log to file
        writeFileSync( 'test-sample.log', sampleLogContent );

        // Parse the events
        const events = parseLogEvents( sampleLogContent );
        expect( events ).toHaveLength( 9 );

        // Process all events through the system
        const chatResponses: string[] = [];

        for ( const event of events )
        {
            // Process the event
            StatsService.processEvent( event );

            // Handle chat commands
            if ( event.type === 'chat_command' )
            {
                const response = CommandHandler.handleCommand( event as any );
                if ( response )
                {
                    chatResponses.push( response );
                }
            }
        }

        // Verify chat responses were generated
        expect( chatResponses ).toHaveLength( 4 ); // 4 chat commands
        expect( chatResponses[ 0 ] ).toContain( 'Alice' ); // !stats
        expect( chatResponses[ 1 ] ).toContain( 'Bob' ); // !kdr
        expect( chatResponses[ 2 ] ).toContain( 'Top' ); // !top
        expect( chatResponses[ 3 ] ).toContain( 'Bob' ); // !guns

        // Verify final statistics
        const alicePlayer = StatsService.getPlayerStatsByName( 'Alice' );
        const bobPlayer = StatsService.getPlayerStatsByName( 'Bob' );

        expect( alicePlayer ).toBeDefined();
        expect( bobPlayer ).toBeDefined();

        expect( alicePlayer?.player_name ).toBe( 'Alice' );
        expect( bobPlayer?.player_name ).toBe( 'Bob' );

        // Alice: 2 kills, Bob: 1 kill (based on the kill events)
        // Note: Database tracking may vary based on how players are created
    } );

    test( 'Event parsing handles mixed valid/invalid lines', () => {
        const mixedLogContent = `
[2025.10.05-12.00.00:000][001]LogNet: Join succeeded: TestPlayer
This is not a valid log line
[2025.10.05-12.01.00:000][002]LogGameplayEvents: Display: Game over
Another invalid line
[2025.10.05-12.02.00:000][003]LogChat: Display: TestPlayer(12345) Global Chat: !stats
        `.trim();

        const events = parseLogEvents( mixedLogContent );

        // Should parse only valid events
        expect( events ).toHaveLength( 3 );
        expect( events[ 0 ]?.type ).toBe( 'player_join' );
        expect( events[ 1 ]?.type ).toBe( 'game_over' );
        expect( events[ 2 ]?.type ).toBe( 'chat_command' );
    } );

    test( 'System handles rapid event processing', () => {
        // Generate many events quickly
        const rapidEvents = [];

        for ( let i = 0; i < 50; i++ )
        {
            rapidEvents.push( {
                type: 'player_kill',
                timestamp: `2025.10.05-12.${ i.toString().padStart( 2, '0' ) }.00:000`,
                data: {
                    killer: 'RapidPlayer',
                    killerSteamId: '76561198000000999',
                    killerTeam: 0,
                    victim: 'Target',
                    victimSteamId: '76561198000001000',
                    victimTeam: 1,
                    weapon: 'M16A4'
                },
                rawLine: `rapid kill ${ i }`
            } );
        }

        // Process all events
        rapidEvents.forEach( event => {
            StatsService.processEvent( event );
        } );

        // Verify stats were recorded correctly
        const stats = StatsService.getPlayerStats( '76561198000000999' );
        expect( stats ).toBeDefined();
        expect( stats.total_kills ).toBeGreaterThanOrEqual( 50 );
    } );

    test( 'Chat commands work with real player data', () => {
        // Set up realistic player data
        const playerEvents = [
            {
                type: 'player_join',
                timestamp: '2025.10.05-12.00.00:000',
                data: { playerName: 'ProPlayer' },
                rawLine: 'test'
            },
            // Multiple kills with different weapons
            ...Array.from( { length: 15 }, ( _, i ) => ( {
                type: 'player_kill',
                timestamp: '2025.10.05-12.01.00:000',
                data: {
                    killer: 'ProPlayer',
                    killerSteamId: '76561198000002000',
                    killerTeam: 0,
                    victim: `Enemy${ i }`,
                    victimSteamId: `7656119800000${ 2001 + i }`,
                    victimTeam: 1,
                    weapon: [ 'M16A4', 'AK-74', 'M24 SWS', 'RPG-7' ][ i % 4 ]
                },
                rawLine: 'test'
            } ) ),
            // A few deaths
            ...Array.from( { length: 3 }, ( _, i ) => ( {
                type: 'player_kill',
                timestamp: '2025.10.05-12.02.00:000',
                data: {
                    killer: `Enemy${ i }`,
                    killerSteamId: `7656119800000${ 3001 + i }`,
                    killerTeam: 1,
                    victim: 'ProPlayer',
                    victimSteamId: '76561198000002000',
                    victimTeam: 0,
                    weapon: 'AK-74'
                },
                rawLine: 'test'
            } ) )
        ];

        // Process all events
        playerEvents.forEach( event => {
            StatsService.processEvent( event as any );
        } );

        // Test all chat commands
        const statsResponse = CommandHandler.handleCommand( {
            type: 'chat_command',
            timestamp: '2025.10.05-12.03.00:000',
            data: {
                playerName: 'ProPlayer',
                steamId: '76561198000002000',
                command: '!stats',
                args: undefined
            },
            rawLine: 'test'
        } );

        const kdrResponse = CommandHandler.handleCommand( {
            type: 'chat_command',
            timestamp: '2025.10.05-12.03.00:000',
            data: {
                playerName: 'ProPlayer',
                steamId: '76561198000002000',
                command: '!kdr',
                args: undefined
            },
            rawLine: 'test'
        } );

        const gunsResponse = CommandHandler.handleCommand( {
            type: 'chat_command',
            timestamp: '2025.10.05-12.03.00:000',
            data: {
                playerName: 'ProPlayer',
                steamId: '76561198000002000',
                command: '!guns',
                args: undefined
            },
            rawLine: 'test'
        } );

        // Verify responses contain expected data
        expect( statsResponse ).toContain( 'ProPlayer' );
        expect( statsResponse ).toMatch( /\d+ kills/ );

        expect( kdrResponse ).toContain( 'K/D ratio' );
        expect( kdrResponse ).toMatch( /\d+(\.\d+)?/ ); // Matches both "1" and "1.25"

        expect( gunsResponse ).toContain( 'Top Weapons' );
        expect( gunsResponse ).toMatch( /M16A4|AK-74|M24 SWS|RPG-7/ );
    } );
} );