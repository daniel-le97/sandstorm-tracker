import { test, expect, describe, beforeAll, beforeEach } from 'bun:test';
import { Database } from 'bun:sqlite';
import { unlinkSync } from 'fs';

// Import after setting up test database path
const testDbPath = 'test_sandstorm_stats.db';

describe( 'Database Operations', () => {
    let db: Database;
    let StatsService: any;

    beforeAll( async () => {
        // Clean up any existing test database
        try
        {
            unlinkSync( testDbPath );
        } catch ( e )
        {
            // File doesn't exist, that's fine
        }

        // Set up test database path
        process.env.TEST_DB_PATH = testDbPath;

        // Import modules after setting test path
        const dbModule = await import( '../src/database.ts' );
        StatsService = ( await import( '../src/stats-service' ) ).default;

        db = dbModule.default();
    } );

    beforeEach( () => {
        // Clear all data before each test
        db.run( 'DELETE FROM chat_commands' );
        db.run( 'DELETE FROM kills' );
        db.run( 'DELETE FROM player_sessions' );
        db.run( 'DELETE FROM players' );
        db.run( 'DELETE FROM map_rounds' );
        db.run( 'DELETE FROM maps' );
    } );

    test( 'Database initializes correctly', () => {
        expect( db ).toBeDefined();

        // Check that tables exist
        const tables = db.prepare( "SELECT name FROM sqlite_master WHERE type='table'" ).all();
        const tableNames = tables.map( ( t: any ) => t.name );

        expect( tableNames ).toContain( 'players' );
        expect( tableNames ).toContain( 'player_sessions' );
        expect( tableNames ).toContain( 'kills' );
        expect( tableNames ).toContain( 'weapon_stats' );
        expect( tableNames ).toContain( 'chat_commands' );
    } );

    test( 'Player join creates session', () => {
        StatsService.processEvent( {
            type: 'player_join',
            timestamp: '2025.10.05-12.00.00:000',
            data: {
                playerName: 'TestPlayer'
            },
            rawLine: '[2025.10.05-12.00.00:000][001]LogNet: Join succeeded: TestPlayer'
        } );

        // Check player was created
        const player = db.prepare( 'SELECT * FROM players WHERE player_name = ?' ).get( 'TestPlayer' ) as any;
        expect( player ).toBeDefined();
        expect( player.player_name ).toBe( 'TestPlayer' );

        // Check session was started
        const sessions = db.prepare( 'SELECT * FROM player_sessions WHERE player_id = ? AND leave_time IS NULL' ).all( player.id );
        expect( sessions ).toHaveLength( 1 );
    } );

    test( 'Kill event records correctly', () => {
        // First ensure players exist
        StatsService.processEvent( {
            type: 'player_join',
            timestamp: '2025.10.05-12.00.00:000',
            data: { playerName: 'Killer' },
            rawLine: 'test'
        } );

        StatsService.processEvent( {
            type: 'player_join',
            timestamp: '2025.10.05-12.00.00:000',
            data: { playerName: 'Victim' },
            rawLine: 'test'
        } );

        // Record kill
        StatsService.processEvent( {
            type: 'player_kill',
            timestamp: '2025.10.05-12.01.00:000',
            data: {
                killer: 'Killer',
                killerSteamId: '76561198000000001',
                killerTeam: 0,
                victim: 'Victim',
                victimSteamId: '76561198000000002',
                victimTeam: 1,
                weapon: 'M16A4'
            },
            rawLine: 'test kill event'
        } );

        // Check kill was recorded
        const kills = db.prepare( 'SELECT * FROM kills' ).all() as any[];
        expect( kills ).toHaveLength( 1 );
        expect( kills[ 0 ].weapon ).toBe( 'M16A4' );
        expect( kills[ 0 ].kill_type ).toBe( 'player_kill' );

        // Check weapon stats
        const weaponStats = db.prepare( 'SELECT * FROM weapon_stats WHERE weapon_name = ?' ).get( 'M16A4' ) as any;
        expect( weaponStats ).toBeDefined();
        expect( weaponStats.kills ).toBe( 1 );
    } );

    test( 'Player stats retrieval works', () => {
        // First create a player and some data
        StatsService.processEvent( {
            type: 'player_join',
            timestamp: '2025.10.05-12.00.00:000',
            data: {
                playerName: 'StatsTestPlayer'
            },
            rawLine: 'test join'
        } );

        const player = StatsService.getPlayerStatsByName( 'StatsTestPlayer' );
        expect( player ).toBeDefined();
        expect( player?.player_name ).toBe( 'StatsTestPlayer' );
    } );

    test( 'Top players query works', () => {
        // Create some test players first
        StatsService.processEvent( {
            type: 'player_join',
            timestamp: '2025.10.05-12.00.00:000',
            data: { playerName: 'TopPlayer1' },
            rawLine: 'test'
        } );

        StatsService.processEvent( {
            type: 'player_join',
            timestamp: '2025.10.05-12.00.00:000',
            data: { playerName: 'TopPlayer2' },
            rawLine: 'test'
        } );

        // Add some kills to make them show up in top players
        StatsService.processEvent( {
            type: 'player_kill',
            timestamp: '2025.10.05-12.01.00:000',
            data: {
                killer: 'TopPlayer1',
                killerSteamId: '76561198000000101',
                killerTeam: 0,
                victim: 'TopPlayer2',
                victimSteamId: '76561198000000102',
                victimTeam: 1,
                weapon: 'M16A4'
            },
            rawLine: 'test kill'
        } );

        const topPlayers = StatsService.getTopPlayers( 5 );

        expect( topPlayers ).toBeDefined();
        expect( Array.isArray( topPlayers ) ).toBe( true );
        expect( topPlayers.length ).toBeGreaterThan( 0 );

        // Check structure
        if ( topPlayers.length > 0 )
        {
            expect( topPlayers[ 0 ] ).toHaveProperty( 'player_name' );
            expect( topPlayers[ 0 ] ).toHaveProperty( 'total_kills' );
            expect( topPlayers[ 0 ] ).toHaveProperty( 'total_deaths' );
            expect( topPlayers[ 0 ] ).toHaveProperty( 'kdr' );
        }
    } );

    test( 'Player weapons query works', () => {
        const weapons = StatsService.getPlayerWeapons( '76561198000000001', 5 );

        expect( weapons ).toBeDefined();
        expect( Array.isArray( weapons ) ).toBe( true );

        if ( weapons.length > 0 )
        {
            expect( weapons[ 0 ] ).toHaveProperty( 'weapon_name' );
            expect( weapons[ 0 ] ).toHaveProperty( 'kills' );
            expect( weapons[ 0 ] ).toHaveProperty( 'deaths' );
        }
    } );

    test( 'Chat command logging works', () => {
        StatsService.processEvent( {
            type: 'chat_command',
            timestamp: '2025.10.05-12.02.00:000',
            data: {
                playerName: 'TestPlayer',
                steamId: '76561198000000001',
                command: '!stats',
                args: undefined
            },
            rawLine: 'test chat command'
        } );

        const commands = db.prepare( 'SELECT * FROM chat_commands WHERE command = ?' ).all( '!stats' ) as any[];
        expect( commands.length ).toBeGreaterThan( 0 );
        expect( commands[ 0 ].command ).toBe( '!stats' );
    } );

    test( 'K/D ratio calculation is correct', () => {
        // Create a fresh player for K/D testing
        StatsService.processEvent( {
            type: 'player_join',
            timestamp: '2025.10.05-12.00.00:000',
            data: { playerName: 'KDRTest' },
            rawLine: 'test'
        } );

        // Record multiple kills, no deaths
        for ( let i = 0; i < 3; i++ )
        {
            StatsService.processEvent( {
                type: 'player_kill',
                timestamp: '2025.10.05-12.01.00:000',
                data: {
                    killer: 'KDRTest',
                    killerSteamId: '76561198000000099',
                    killerTeam: 0,
                    victim: 'Enemy',
                    victimSteamId: '76561198000000100',
                    victimTeam: 1,
                    weapon: 'AK-74'
                },
                rawLine: 'test kill'
            } );
        }

        const stats = StatsService.getPlayerStats( '76561198000000099' );
        expect( stats.total_kills ).toBe( 3 );
        expect( stats.total_deaths ).toBe( 0 );
        expect( stats.kdr ).toBe( 3 ); // Should be kills when no deaths
    } );

    test( 'Suicides and team kills are not counted as player kills', () => {
        // Create a test player
        StatsService.processEvent( {
            type: 'player_join',
            timestamp: '2025.10.05-12.00.00:000',
            data: { playerName: 'KillTypeTest' },
            rawLine: 'test'
        } );

        // Record 2 legitimate player kills
        for ( let i = 0; i < 2; i++ )
        {
            StatsService.processEvent( {
                type: 'player_kill',
                timestamp: '2025.10.05-12.01.00:000',
                data: {
                    killer: 'KillTypeTest',
                    killerSteamId: '76561198000000200',
                    killerTeam: 0,
                    victim: `Enemy${ i }`,
                    victimSteamId: `7656119800000021${ i }`,
                    victimTeam: 1,
                    weapon: 'M16A4'
                },
                rawLine: 'test player kill'
            } );
        }

        // Record a team kill (should NOT count as player kill)
        StatsService.processEvent( {
            type: 'team_kill',
            timestamp: '2025.10.05-12.02.00:000',
            data: {
                killer: 'KillTypeTest',
                killerSteamId: '76561198000000200',
                killerTeam: 0,
                victim: 'Teammate',
                victimSteamId: '76561198000000201',
                victimTeam: 0,
                weapon: 'M16A4'
            },
            rawLine: 'test team kill'
        } );

        // Record a suicide (should NOT count as player kill)
        StatsService.processEvent( {
            type: 'suicide',
            timestamp: '2025.10.05-12.03.00:000',
            data: {
                killer: 'KillTypeTest',
                killerSteamId: '76561198000000200',
                killerTeam: 0,
                victim: 'KillTypeTest',
                victimSteamId: '76561198000000200',
                victimTeam: 0,
                weapon: 'Fall Damage'
            },
            rawLine: 'test suicide'
        } );

        // Verify stats: only player kills count toward total_kills
        const stats = StatsService.getPlayerStats( '76561198000000200' );
        expect( stats ).toBeDefined();
        expect( stats.total_kills ).toBe( 2 ); // Only the 2 player kills
        expect( stats.team_kills ).toBe( 1 ); // Team kill tracked separately
        expect( stats.suicides ).toBe( 1 ); // Suicide tracked separately
        expect( stats.total_deaths ).toBe( 1 ); // Death from suicide
        expect( stats.kdr ).toBe( 2 ); // 2 kills / 1 death = 2.0
    } );
} );