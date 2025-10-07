import { beforeAll, beforeEach, afterAll, afterEach, describe, test, expect } from "bun:test";
import { existsSync, mkdirSync, rmSync, writeFileSync, appendFileSync } from "fs";
import { join } from "path";

describe( "WatchersService Integration", () => {
    const testLogDir = "./test-watchers";
    const testDbPath = "test_watchers_service.db";
    const serverConfigId = `watcher-test-${ Date.now() }`;
    // Use a fixed, valid UUID for the serverId (used for the log filename)
    const serverUuid = "60844f66-b93b-4fe1-afc4-a0a91b493865";

    let TrackerServiceRef: any;
    let watchersService: any;
    let processor: any;

    beforeAll( async () => {
        process.env.TEST_DB_PATH = testDbPath;

        // Import app services
        const dbModule = await import( "../src/database" );
        TrackerServiceRef = ( await import( "../src/trackerService" ) ).default;
        watchersService = ( await import( "../src/lib/watcher/watcher" ) ).watchersService;
        processor = ( await import( "../src/lib/watcher/processor" ) );

        // Initialize DB
        dbModule.default();

        // Ensure log dir
        if ( !existsSync( testLogDir ) ) mkdirSync( testLogDir, { recursive: true } );

        TrackerServiceRef.initialize();
    } );

    beforeEach( () => {
        // Reset TrackerService state when available
        if ( TrackerServiceRef.reset ) TrackerServiceRef.reset();
    } );

    afterEach( () => {
        // Stop watchers if running
        if ( watchersService && typeof watchersService.stopAll === "function" )
        {
            watchersService.stopAll();
        }
        // Clean up log dir contents
        try { rmSync( join( testLogDir ), { recursive: true, force: true } ); } catch { }
    } );

    afterAll( () => {
        try { rmSync( testDbPath, { force: true } ); } catch { }
    } );

    test( "should process log line via watchersService and processor", async () => {
        // Create server record
        const { getStatements } = await import( "../src/database" );
        const statements = getStatements();
        const server = statements.upsertServer.get(
            serverConfigId,
            "Watcher Test Server",
            `cfg-${ Date.now() }`,
            testLogDir,
            "Test server"
        ) as any;

        const serverId = server?.id || 1;

        // Add server watcher via service
        const { createServerConfig } = await import( "../src/config" );
        const cfg = createServerConfig( {
            id: serverConfigId,
            name: "Watcher Test Server",
            logPath: testLogDir,
            serverId: serverUuid, // valid UUID used for filename
            enabled: true,
        } as any );

        await watchersService.addServerWatcher( cfg );

        // Spy TrackerService.processEvent
        const originalProcessEvent = TrackerServiceRef.processEvent;
        let processed = 0;
        TrackerServiceRef.processEvent = function ( ev: any, sid: number ) {
            processed++; // debug counter
            if ( originalProcessEvent ) return originalProcessEvent.call( this, ev, sid );
        };

        // Create the log file BEFORE starting the watchers to guarantee at least one change event
        const logFilePath = join( testLogDir, `${ cfg.serverId }.log` );
        writeFileSync( logFilePath, "" );

        // Start watchers (non-blocking) — pass the processor callback
        const startPromise = watchersService.startWatching( processor.processServerLogChange );

        // Small delay to ensure fs watcher is established
        await new Promise( r => setTimeout( r, 150 ) );

        // Append a kill log line that parseLogEvents recognizes
        const line = "[2025.10.05-12.01.00:000][002]LogGameplayEvents: Display: TestPlayer[76561198000000001, team 0] killed Enemy[76561198000000002, team 1] with BP_Firearm_M16A4_C_2147481419\n";
        appendFileSync( logFilePath, line );

        // Allow some time for watcher to pick up (use Promise.race to avoid test harness timeout)
        const overallDeadline = Date.now() + 30000; // extended to ensure processing window
        while ( Date.now() < overallDeadline )
        {
            if ( processed > 0 ) break;
            await new Promise( r => setTimeout( r, 250 ) );
        }

        // Stop watchers
        await watchersService.stopAll();
        // Await start promise resolution if it ended
        try { await startPromise; } catch { }

        // Restore original
        TrackerServiceRef.processEvent = originalProcessEvent;

        expect( processed ).toBeGreaterThan( 0 );
    } );
} );
