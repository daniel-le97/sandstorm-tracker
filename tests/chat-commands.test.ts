import { test, expect, describe, beforeAll } from 'bun:test';

describe( 'Chat Commands', () => {
    let StatsService: any;
    let CommandHandler: any;

    beforeAll( async () => {
        // Set test environment
        process.env.TEST_DB_PATH = 'test_commands.db';

        // Import modules
        await import( '../src/database.ts' );
        StatsService = ( await import( '../src/stats-service' ) ).default;
        CommandHandler = ( await import( '../src/command-handler' ) ).default;

        // Set up test data
        StatsService.processEvent( {
            type: 'player_join',
            timestamp: '2025.10.05-12.00.00:000',
            data: { playerName: 'Alice' },
            rawLine: 'test'
        } );

        StatsService.processEvent( {
            type: 'player_join',
            timestamp: '2025.10.05-12.00.30:000',
            data: { playerName: 'Bob' },
            rawLine: 'test'
        } );

        // Add some kills for Alice
        for ( let i = 0; i < 5; i++ )
        {
            StatsService.processEvent( {
                type: 'player_kill',
                timestamp: '2025.10.05-12.01.00:000',
                data: {
                    killer: 'Alice',
                    killerSteamId: '76561198000000001',
                    killerTeam: 0,
                    victim: 'Bob',
                    victimSteamId: '76561198000000002',
                    victimTeam: 1,
                    weapon: i % 2 === 0 ? 'M16A4' : 'AK-74'
                },
                rawLine: 'test'
            } );
        }
    } );

    test( '!stats command returns player statistics', () => {
        const response = CommandHandler.handleCommand( {
            type: 'chat_command',
            timestamp: '2025.10.05-12.02.00:000',
            data: {
                playerName: 'Alice',
                steamId: '76561198000000001',
                command: '!stats',
                args: undefined
            },
            rawLine: 'test'
        } );

        expect( response ).toBeTruthy();
        expect( response ).toContain( 'Alice' );
        expect( response ).toContain( 'kills' );
        expect( response ).toContain( 'deaths' );
        expect( response ).toContain( 'K/D' );
    } );

    test( '!kdr command returns kill/death ratio', () => {
        const response = CommandHandler.handleCommand( {
            type: 'chat_command',
            timestamp: '2025.10.05-12.02.00:000',
            data: {
                playerName: 'Alice',
                steamId: '76561198000000001',
                command: '!kdr',
                args: undefined
            },
            rawLine: 'test'
        } );

        expect( response ).toBeTruthy();
        expect( response ).toContain( '💀' );
        expect( response ).toContain( 'Alice' );
        expect( response ).toContain( 'K/D ratio' );
    } );

    test( '!top command returns leaderboard', () => {
        const response = CommandHandler.handleCommand( {
            type: 'chat_command',
            timestamp: '2025.10.05-12.02.00:000',
            data: {
                playerName: 'Alice',
                steamId: '76561198000000001',
                command: '!top',
                args: undefined
            },
            rawLine: 'test'
        } );

        expect( response ).toBeTruthy();
        expect( response ).toContain( '🏆' );
        expect( response ).toContain( 'Top' );
        expect( response ).toContain( 'Players' );
    } );

    test( '!guns command returns weapon statistics', () => {
        const response = CommandHandler.handleCommand( {
            type: 'chat_command',
            timestamp: '2025.10.05-12.02.00:000',
            data: {
                playerName: 'Alice',
                steamId: '76561198000000001',
                command: '!guns',
                args: undefined
            },
            rawLine: 'test'
        } );

        expect( response ).toBeTruthy();
        expect( response ).toContain( '🔫' );
        expect( response ).toContain( 'Alice' );
        expect( response ).toContain( 'Weapons' );
        expect( response ).toMatch( /M16A4|AK-74/ );
    } );

    test( 'Unknown command returns null', () => {
        const response = CommandHandler.handleCommand( {
            type: 'chat_command',
            timestamp: '2025.10.05-12.02.00:000',
            data: {
                playerName: 'Alice',
                steamId: '76561198000000001',
                command: '!unknown',
                args: undefined
            },
            rawLine: 'test'
        } );

        expect( response ).toBeNull();
    } );

    test( 'Command responses contain expected emojis and formatting', () => {
        const commands = [ '!stats', '!kdr', '!top', '!guns' ];
        const expectedEmojis = [ '📊', '💀', '🏆', '🔫' ];

        commands.forEach( ( command, index ) => {
            const response = CommandHandler.handleCommand( {
                type: 'chat_command',
                timestamp: '2025.10.05-12.02.00:000',
                data: {
                    playerName: 'Alice',
                    steamId: '76561198000000001',
                    command,
                    args: undefined
                },
                rawLine: 'test'
            } );

            if ( response )
            {
                expect( response ).toContain( expectedEmojis[ index ] );
            }
        } );
    } );

    test( 'Stats command shows correct format', () => {
        const response = CommandHandler.handleCommand( {
            type: 'chat_command',
            timestamp: '2025.10.05-12.02.00:000',
            data: {
                playerName: 'Alice',
                steamId: '76561198000000001',
                command: '!stats',
                args: undefined
            },
            rawLine: 'test'
        } );

        expect( response ).toMatch( /\d+ kills/ );
        expect( response ).toMatch( /\d+ deaths/ );
        expect( response ).toMatch( /K\/D: \d+(\.\d+)?/ );
        expect( response ).toMatch( /Playtime: \d+h/ );
    } );
} );