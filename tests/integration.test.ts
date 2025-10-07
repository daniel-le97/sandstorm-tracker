import { test, expect, describe, beforeAll, beforeEach } from 'bun:test';
import { unlinkSync } from 'fs';
import { Database } from 'bun:sqlite';
import { parseLogEvents } from '../src/events';

describe( 'Integration Tests', () => {
    let TrackerService: any;
    let CommandHandler: any;
    let db: Database;
    let testServerId: number;
    const testDbPath = 'tests/databases/test_integration.db';

    beforeAll( async () => {
        try
        {
            unlinkSync( testDbPath );
        } catch ( e )
        {
            // File doesn't exist
        }

        process.env.TEST_DB_PATH = testDbPath;

        const dbModule = await import( '../src/database.ts' );
        TrackerService = ( await import( '../src/trackerService.ts' ) ).default;
        CommandHandler = ( await import( '../src/command-handler' ) ).default;

        db = dbModule.default();
    } );

    beforeEach( () => {
        const dbModule = require( '../src/database' );
        db = dbModule.default();

        db.run( 'DELETE FROM kills' );
        db.run( 'DELETE FROM player_sessions' );
        db.run( 'DELETE FROM players' );
        db.run( 'DELETE FROM map_rounds' );
        db.run( 'DELETE FROM maps' );
        db.run( 'DELETE FROM servers' );

        const { upsertServer } = require( '../src/database' );
        testServerId = upsertServer(
            'integration-test-server',
            'Integration Test Server',
            'integration-config',
            '/test/integration/logs',
            'Server for integration tests'
        );

        TrackerService.endAllSessions( testServerId );
    } );

    test( 'Complete log processing pipeline works end-to-end', () => {
        const sampleLogContent = `
[2025.10.05-12.00.00:000][123]LogNet: Join succeeded: TestPlayer
[2025.10.05-12.00.01:000][124]LogGameplayEvents: Display: TestPlayer[76561198000000001, team 0] killed TestVictim[76561198000000002, team 1] with M16A4 rifle
`;

        const events = parseLogEvents( sampleLogContent );

        expect( events ).toBeDefined();
        expect( Array.isArray( events ) ).toBe( true );
        expect( events.length ).toBeGreaterThan( 0 );

        events.forEach( event => {
            TrackerService.processEvent( event, testServerId );
        } );

        const players = db.prepare( 'SELECT * FROM players WHERE server_id = ?' ).all( testServerId );
        expect( players.length ).toBeGreaterThan( 0 );

        const kills = db.prepare( 'SELECT * FROM kills WHERE server_id = ?' ).all( testServerId );
        expect( kills.length ).toBeGreaterThan( 0 );
    } );
} );
