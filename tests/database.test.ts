import { test, expect, describe, beforeAll, beforeEach } from 'bun:test';
import { Database } from 'bun:sqlite';
import { unlinkSync } from 'fs';

// Import after setting up test database path
const testDbPath = 'test_sandstorm_stats.db';

describe( 'Database Operations', () => {
    let db: Database;
    let TrackerService: any;
    let testServerId: number;

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
        TrackerService = ( await import( '../src/trackerService.ts' ) ).default;

        db = dbModule.default();
    } );

    beforeEach( () => {
        // Clear all data before each test

        db.run( 'DELETE FROM kills' );
        db.run( 'DELETE FROM player_sessions' );
        db.run( 'DELETE FROM players' );
        db.run( 'DELETE FROM map_rounds' );
        db.run( 'DELETE FROM maps' );
        db.run( 'DELETE FROM servers' );

        // Create a test server for each test
        const { upsertServer } = require( '../src/database' );
        testServerId = upsertServer(
            'test-server-uuid',
            'Test Server',
            'test-config',
            '/test/log/path',
            'Test server for database tests'
        );
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

    } );

    test( 'Player join creates session', () => {
        TrackerService.processEvent( {
            type: 'player_join',
            timestamp: '2025.10.05-12.00.00:000',
            data: {
                playerName: 'TestPlayer'
            },
            rawLine: '[2025.10.05-12.00.00:000][001]LogNet: Join succeeded: TestPlayer'
        }, testServerId );

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
        TrackerService.processEvent( {
            type: 'player_join',
            timestamp: '2025.10.05-12.00.00:000',
            data: { playerName: 'Killer' },
            rawLine: 'test'
        }, testServerId );

        TrackerService.processEvent( {
            type: 'player_join',
            timestamp: '2025.10.05-12.00.00:000',
            data: { playerName: 'Victim' },
            rawLine: 'test'
        }, testServerId );

        // Record kill
        TrackerService.processEvent( {
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
        }, testServerId );

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
        TrackerService.processEvent( {
            type: 'player_join',
            timestamp: '2025.10.05-12.00.00:000',
            data: {
                playerName: 'StatsTestPlayer'
            },
            rawLine: 'test join'
        }, testServerId );

        const player = TrackerService.getPlayerStatsByName( 'StatsTestPlayer', testServerId );
        expect( player ).toBeDefined();
        expect( player?.player_name ).toBe( 'StatsTestPlayer' );
    } );

    test( 'Top players query works', () => {
        // Create some test players first
        TrackerService.processEvent( {
            type: 'player_join',
            timestamp: '2025.10.05-12.00.00:000',
            data: { playerName: 'TopPlayer1' },
            rawLine: 'test'
        }, testServerId );

        TrackerService.processEvent( {
            type: 'player_join',
            timestamp: '2025.10.05-12.00.00:000',
            data: { playerName: 'TopPlayer2' },
            rawLine: 'test'
        }, testServerId );

        // Add some kills to make them show up in top players
        TrackerService.processEvent( {
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
        }, testServerId );

        const topPlayers = TrackerService.getTopPlayers( testServerId, 5 );

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
        const weapons = TrackerService.getPlayerWeapons( '76561198000000001', testServerId, 5 );

        expect( weapons ).toBeDefined();
        expect( Array.isArray( weapons ) ).toBe( true );

        if ( weapons.length > 0 )
        {
            expect( weapons[ 0 ] ).toHaveProperty( 'weapon_name' );
            expect( weapons[ 0 ] ).toHaveProperty( 'kills' );
            expect( weapons[ 0 ] ).toHaveProperty( 'deaths' );
        }
    } );

    test( 'K/D ratio calculation is correct', () => {
        // Create a fresh player for K/D testing
        TrackerService.processEvent( {
            type: 'player_join',
            timestamp: '2025.10.05-12.00.00:000',
            data: { playerName: 'KDRTest' },
            rawLine: 'test'
        }, testServerId );

        // Record multiple kills, no deaths
        for ( let i = 0; i < 3; i++ )
        {
            TrackerService.processEvent( {
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
            }, testServerId );
        }

        const stats = TrackerService.getPlayerStats( '76561198000000099', testServerId );
        expect( stats.total_kills ).toBe( 3 );
        expect( stats.total_deaths ).toBe( 0 );
        expect( stats.kdr ).toBe( 3 ); // Should be kills when no deaths
    } );

    test( 'Suicides and team kills are not counted as player kills', () => {
        // Create a test player
        TrackerService.processEvent( {
            type: 'player_join',
            timestamp: '2025.10.05-12.00.00:000',
            data: { playerName: 'KillTypeTest' },
            rawLine: 'test'
        }, testServerId );

        // Record 2 legitimate player kills
        for ( let i = 0; i < 2; i++ )
        {
            TrackerService.processEvent( {
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
            }, testServerId );
        }

        // Record a team kill (should NOT count as player kill)
        TrackerService.processEvent( {
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
        }, testServerId );

        // Record a suicide (should NOT count as player kill)
        TrackerService.processEvent( {
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
        }, testServerId );

        // Verify stats: only player kills count toward total_kills
        const stats = TrackerService.getPlayerStats( '76561198000000200', testServerId );
        expect( stats ).toBeDefined();
        expect( stats.total_kills ).toBe( 2 ); // Only the 2 player kills
        expect( stats.team_kills ).toBe( 1 ); // Team kill tracked separately
        expect( stats.suicides ).toBe( 1 ); // Suicide tracked separately
        expect( stats.total_deaths ).toBe( 1 ); // Death from suicide
        expect( stats.kdr ).toBe( 2 ); // 2 kills / 1 death = 2.0
    } );
} );
