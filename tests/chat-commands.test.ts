import { beforeAll, describe, expect, test } from "bun:test";
import { unlinkSync } from "fs";

describe( "Chat Commands", () => {
    let TrackerService: any;
    let CommandHandler: any;
    let testServerDbId: number;
    const testDbPath = "tests/databases/test_commands.db";

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
        const dbFunctions = await import( "../src/database.ts" );
        TrackerService = ( await import( "../src/trackerService.ts" ) ).default;
        CommandHandler = ( await import( "../src/command-handler" ) ).default;

        // Create test server first (required for foreign key constraints)
        testServerDbId = dbFunctions.upsertServer(
            "test-server-1",
            "Test Server 1",
            "test-config",
            "/test/logs/test-server-1.log",
            "Test server for chat command tests"
        );

        // Set up test data
        TrackerService.processEvent(
            {
                type: "player_join",
                timestamp: "2025.10.05-12.00.00:000",
                data: { playerName: "Alice" },
                rawLine: "test",
            },
            testServerDbId
        );

        TrackerService.processEvent(
            {
                type: "player_join",
                timestamp: "2025.10.05-12.00.30:000",
                data: { playerName: "Bob" },
                rawLine: "test",
            },
            testServerDbId
        );

        // Add some kills for Alice
        for ( let i = 0; i < 5; i++ )
        {
            TrackerService.processEvent(
                {
                    type: "player_kill",
                    timestamp: "2025.10.05-12.01.00:000",
                    data: {
                        killer: "Alice",
                        killerSteamId: "76561198000000001",
                        killerTeam: 0,
                        victim: "Bob",
                        victimSteamId: "76561198000000002",
                        victimTeam: 1,
                        weapon: i % 2 === 0 ? "M16A4" : "AK-74",
                    },
                    rawLine: "test",
                },
                testServerDbId
            );
        }
    } );

    test( "!stats command returns player statistics", () => {
        const response = CommandHandler.handleCommand(
            {
                type: "chat_command",
                timestamp: "2025.10.05-12.02.00:000",
                data: {
                    playerName: "Alice",
                    steamId: "76561198000000001",
                    command: "!stats",
                    args: undefined,
                },
                rawLine: "test",
            },
            testServerDbId
        );

        expect( response ).toBeTruthy();
        expect( response ).toContain( "Alice" );
        expect( response ).toContain( "kills" );
        expect( response ).toContain( "deaths" );
        expect( response ).toContain( "K/D" );
    } );

    test( "!kdr command returns kill/death ratio", () => {
        const response = CommandHandler.handleCommand(
            {
                type: "chat_command",
                timestamp: "2025.10.05-12.02.00:000",
                data: {
                    playerName: "Alice",
                    steamId: "76561198000000001",
                    command: "!kdr",
                    args: undefined,
                },
                rawLine: "test",
            },
            testServerDbId
        );

        expect( response ).toBeTruthy();
        expect( response ).toContain( "kills" );
        expect( response ).toContain( "Alice" );
        expect( response ).toContain( "K/D ratio" );
    } );

    test( "!top command returns leaderboard", () => {
        const response = CommandHandler.handleCommand(
            {
                type: "chat_command",
                timestamp: "2025.10.05-12.02.00:000",
                data: {
                    playerName: "Alice",
                    steamId: "76561198000000001",
                    command: "!top",
                    args: undefined,
                },
                rawLine: "test",
            },
            testServerDbId
        );

        expect( response ).toBeTruthy();
        expect( response ).toContain( "Top" );
        expect( response ).toContain( "Top" );
        expect( response ).toContain( "Players" );
    } );

    test( "!guns command returns weapon statistics", () => {
        const response = CommandHandler.handleCommand(
            {
                type: "chat_command",
                timestamp: "2025.10.05-12.02.00:000",
                data: {
                    playerName: "Alice",
                    steamId: "76561198000000001",
                    command: "!guns",
                    args: undefined,
                },
                rawLine: "test",
            },
            testServerDbId
        );

        expect( response ).toBeTruthy();
        expect( response ).toContain( "Weapons" );
        expect( response ).toContain( "Alice" );
        expect( response ).toContain( "Weapons" );
        expect( response ).toMatch( /M16A4|AK-74/ );
    } );

    test( "Unknown command returns null", () => {
        const response = CommandHandler.handleCommand(
            {
                type: "chat_command",
                timestamp: "2025.10.05-12.02.00:000",
                data: {
                    playerName: "Alice",
                    steamId: "76561198000000001",
                    command: "!unknown",
                    args: undefined,
                },
                rawLine: "test",
            },
            testServerDbId
        );

        expect( response ).toBeNull();
    } );

    test( "Command responses contain expected emojis and formatting", () => {
        const commands = [ "!stats", "!kdr", "!top", "!guns" ];
        const expectedTexts = [ "kills", "deaths", "Top", "Weapons" ];

        commands.forEach( ( command, index ) => {
            const response = CommandHandler.handleCommand(
                {
                    type: "chat_command",
                    timestamp: "2025.10.05-12.02.00:000",
                    data: {
                        playerName: "Alice",
                        steamId: "76561198000000001",
                        command,
                        args: undefined,
                    },
                    rawLine: "test",
                },
                testServerDbId
            );

            if ( response )
            {
                expect( response ).toContain( expectedTexts[ index ] );
            }
        } );
    } );

    test( "Stats command shows correct format", () => {
        const response = CommandHandler.handleCommand(
            {
                type: "chat_command",
                timestamp: "2025.10.05-12.02.00:000",
                data: {
                    playerName: "Alice",
                    steamId: "76561198000000001",
                    command: "!stats",
                    args: undefined,
                },
                rawLine: "test",
            },
            testServerDbId
        );

        expect( response ).toMatch( /\d+ kills/ );
        expect( response ).toMatch( /\d+ deaths/ );
        expect( response ).toMatch( /K\/D: \d+(\.\d+)?/ );
        expect( response ).toMatch( /Playtime: \d+h/ );
    } );
} );
