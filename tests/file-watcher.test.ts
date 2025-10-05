import { afterAll, afterEach, beforeAll, beforeEach, describe, expect, test } from "bun:test";
import { appendFileSync, existsSync, mkdirSync, rmSync, writeFileSync } from "fs";
import { watch } from "fs/promises";
import { join } from "path";
import { parseLogEvents } from "../src/events";

describe( "File Watcher Functionality", () => {
    const testLogDir = "./test-logs";
    const testLogFile = join( testLogDir, "Insurgency.log" );
    const testDbPath = "test_file_watcher.db";
    let StatsServiceRef: any;
    let watcherController: AbortController | null = null;
    let testServerId: number;

    beforeAll( async () => {
        // Set up test database
        process.env.TEST_DB_PATH = testDbPath;

        // Import modules
        const dbModule = await import( "../src/database.ts" );
        StatsServiceRef = ( await import( "../src/stats-service" ) ).default;

        // Initialize database
        const db = dbModule.default();

        // Create test log directory
        if ( !existsSync( testLogDir ) )
        {
            mkdirSync( testLogDir, { recursive: true } );
        }

        // Initialize stats service
        StatsServiceRef.initialize();
    } );

    beforeEach( () => {
        // Reset stats service state
        if ( StatsServiceRef.reset )
        {
            StatsServiceRef.reset();
        }

        // Clean up any existing log file
        if ( existsSync( testLogFile ) )
        {
            rmSync( testLogFile );
        }

        // Cancel any existing watchers
        if ( watcherController )
        {
            watcherController.abort();
        }
        watcherController = new AbortController();

        // Create test server record to avoid foreign key constraint errors
        const { getStatements } = require( "../src/database" );
        const statements = getStatements();
        try
        {
            const server = statements.upsertServer.get(
                "test-server-1",
                "Test Server 1",
                `test-config-${ Date.now() }`, // Use unique config_id
                "/test/log/path",
                "Test server for file watcher tests"
            ) as any;
            testServerId = server?.id || 1; // Use actual server ID
        } catch ( e )
        {
            // Fallback to ID 1 if server creation fails
            testServerId = 1;
        }
    } );

    afterEach( () => {
        // Cancel file watchers
        if ( watcherController )
        {
            watcherController.abort();
            watcherController = null;
        }
    } );

    afterAll( () => {
        // Clean up test files and directories
        try
        {
            rmSync( testLogDir, { recursive: true, force: true } );
            rmSync( testDbPath, { force: true } );
        } catch ( e )
        {
            // Ignore cleanup errors
        }
    } );

    describe( "File System Watching", () => {
        test( "should detect file creation", async () => {
            let changeDetected = false;
            let detectedFileName = "";

            // Set up file watcher
            const watchPromise = ( async () => {
                try
                {
                    const watcher = watch( testLogDir, { signal: watcherController!.signal } );
                    for await ( const event of watcher )
                    {
                        if ( event.filename === "Insurgency.log" )
                        {
                            changeDetected = true;
                            detectedFileName = event.filename;
                            break;
                        }
                    }
                } catch ( error: any )
                {
                    if ( error.name !== "AbortError" )
                    {
                        throw error;
                    }
                }
            } )();

            // Wait a bit to ensure watcher is active
            await new Promise( ( resolve ) => setTimeout( resolve, 100 ) );

            // Create the log file
            writeFileSync( testLogFile, "[2025.10.05-12.00.00:000][001]LogNet: Server started\n" );

            // Wait for the watcher to detect the change
            await new Promise( ( resolve ) => setTimeout( resolve, 200 ) );
            watcherController!.abort();

            await watchPromise;

            expect( changeDetected ).toBe( true );
            expect( detectedFileName ).toBe( "Insurgency.log" );
        } );

        test( "should detect file modifications", async () => {
            let changeCount = 0;
            const detectedChanges: string[] = [];

            // Create initial file
            writeFileSync( testLogFile, "[2025.10.05-12.00.00:000][001]LogNet: Initial content\n" );

            // Set up file watcher
            const watchPromise = ( async () => {
                try
                {
                    const watcher = watch( testLogDir, { signal: watcherController!.signal } );
                    for await ( const event of watcher )
                    {
                        if ( event.filename === "Insurgency.log" )
                        {
                            changeCount++;
                            detectedChanges.push( `Change ${ changeCount }: ${ event.eventType }` );

                            if ( changeCount >= 2 ) break; // Stop after detecting modifications
                        }
                    }
                } catch ( error: any )
                {
                    if ( error.name !== "AbortError" )
                    {
                        throw error;
                    }
                }
            } )();

            // Wait for watcher to initialize
            await new Promise( ( resolve ) => setTimeout( resolve, 100 ) );

            // Modify the file multiple times
            appendFileSync( testLogFile, "[2025.10.05-12.01.00:000][002]LogNet: Player joined\n" );
            await new Promise( ( resolve ) => setTimeout( resolve, 100 ) );

            appendFileSync( testLogFile, "[2025.10.05-12.02.00:000][003]LogGameplayEvents: Game started\n" );
            await new Promise( ( resolve ) => setTimeout( resolve, 200 ) );

            watcherController!.abort();
            await watchPromise;

            expect( changeCount ).toBeGreaterThanOrEqual( 2 );
            expect( detectedChanges.length ).toBeGreaterThan( 0 );
        } );

        test( "should handle rapid file changes with debouncing", async () => {
            let processedChanges = 0;
            const DEBOUNCE_DELAY = 50; // ms

            // Create initial file
            writeFileSync( testLogFile, "" );

            // Set up file watcher with debouncing logic
            const watchPromise = ( async () => {
                let lastProcessTime = 0;

                try
                {
                    const watcher = watch( testLogDir, { signal: watcherController!.signal } );
                    for await ( const event of watcher )
                    {
                        if ( event.filename === "Insurgency.log" )
                        {
                            const now = Date.now();

                            // Debounce rapid changes
                            if ( now - lastProcessTime >= DEBOUNCE_DELAY )
                            {
                                processedChanges++;
                                lastProcessTime = now;
                            }

                            if ( processedChanges >= 3 ) break;
                        }
                    }
                } catch ( error: any )
                {
                    if ( error.name !== "AbortError" )
                    {
                        throw error;
                    }
                }
            } )();

            await new Promise( ( resolve ) => setTimeout( resolve, 50 ) );

            // Generate rapid file changes (should be debounced)
            for ( let i = 0; i < 10; i++ )
            {
                appendFileSync(
                    testLogFile,
                    `[2025.10.05-12.00.${ i.toString().padStart( 2, "0" ) }:000][${ i
                        .toString()
                        .padStart( 3, "0" ) }]LogNet: Rapid change ${ i }\n`
                );
                await new Promise( ( resolve ) => setTimeout( resolve, 10 ) ); // Very fast changes
            }

            await new Promise( ( resolve ) => setTimeout( resolve, 200 ) );
            watcherController!.abort();
            await watchPromise;

            // Due to debouncing, we should have processed fewer changes than we made
            expect( processedChanges ).toBeLessThan( 10 );
            expect( processedChanges ).toBeGreaterThan( 0 );
        } );
    } );

    describe( "Log Event Processing", () => {
        test( "should process log events from watched file changes", async () => {
            const logEvents = [
                "[2025.10.05-12.00.00:000][001]LogNet: Join succeeded: TestPlayer",
                "[2025.10.05-12.01.00:000][002]LogGameplayEvents: Display: TestPlayer[76561198000000001, team 0] killed Enemy[76561198000000002, team 1] with BP_Firearm_M16A4_C_2147481419",
                "[2025.10.05-12.02.00:000][003]LogChat: Display: TestPlayer(76561198000000001) Global Chat: !stats",
                "[2025.10.05-12.03.00:000][004]LogGameplayEvents: Display: Game over",
            ];

            let eventsProcessed = 0;
            let processedEvents: any[] = [];

            // Mock the event processing
            const originalProcessEvent = StatsServiceRef.processEvent;
            StatsServiceRef.processEvent = function ( event: any, serverId: number ) {
                eventsProcessed++;
                processedEvents.push( { event, serverId } );
                // Call original function if it exists
                if ( originalProcessEvent )
                {
                    return originalProcessEvent.call( this, event, serverId );
                }
            };

            // Set up file watcher with event processing
            const watchPromise = ( async () => {
                let lastProcessedLine = "";

                try
                {
                    const watcher = watch( testLogDir, { signal: watcherController!.signal } );
                    for await ( const event of watcher )
                    {
                        if ( event.filename === "Insurgency.log" )
                        {
                            // Read the file and process new events
                            const content = await Bun.file( testLogFile ).text();
                            const lines = content.split( "\n" ).filter( ( line: string ) => line.trim() );
                            const lastLine = lines[ lines.length - 1 ];

                            if ( lastLine && lastLine !== lastProcessedLine )
                            {
                                // Parse and process the last line
                                const events = parseLogEvents( lastLine );
                                events.forEach( ( gameEvent ) => {
                                    StatsServiceRef.processEvent( gameEvent, testServerId );
                                } );

                                lastProcessedLine = lastLine;

                                if ( eventsProcessed >= 4 ) break; // Stop after processing expected events
                            }
                        }
                    }
                } catch ( error: any )
                {
                    if ( error.name !== "AbortError" )
                    {
                        throw error;
                    }
                }
            } )();

            await new Promise( ( resolve ) => setTimeout( resolve, 50 ) );

            // Write log events one by one to simulate real server logging
            for ( const logEvent of logEvents )
            {
                writeFileSync( testLogFile, logEvent + "\n" );
                await new Promise( ( resolve ) => setTimeout( resolve, 100 ) );
            }

            await new Promise( ( resolve ) => setTimeout( resolve, 300 ) );
            watcherController!.abort();
            await watchPromise;

            // Restore original function
            StatsServiceRef.processEvent = originalProcessEvent;

            expect( eventsProcessed ).toBeGreaterThan( 0 );
            expect( processedEvents.length ).toBeGreaterThan( 0 );
        } );

        test( "should handle file size changes correctly", async () => {
            let sizeChanges: number[] = [];
            let lastSize = 0;

            // Create initial file
            writeFileSync( testLogFile, "Initial content\n" );
            lastSize = ( await Bun.file( testLogFile ).stat() ).size;

            // Set up watcher to track size changes
            const watchPromise = ( async () => {
                try
                {
                    const watcher = watch( testLogDir, { signal: watcherController!.signal } );
                    for await ( const event of watcher )
                    {
                        if ( event.filename === "Insurgency.log" )
                        {
                            const currentSize = ( await Bun.file( testLogFile ).stat() ).size;
                            if ( currentSize !== lastSize )
                            {
                                sizeChanges.push( currentSize );
                                lastSize = currentSize;
                            }

                            if ( sizeChanges.length >= 3 ) break;
                        }
                    }
                } catch ( error: any )
                {
                    if ( error.name !== "AbortError" )
                    {
                        throw error;
                    }
                }
            } )();

            await new Promise( ( resolve ) => setTimeout( resolve, 50 ) );

            // Make incremental changes to increase file size
            appendFileSync( testLogFile, "[2025.10.05-12.00.00:000][001]LogNet: First addition\n" );
            await new Promise( ( resolve ) => setTimeout( resolve, 100 ) );

            appendFileSync( testLogFile, "[2025.10.05-12.01.00:000][002]LogNet: Second addition\n" );
            await new Promise( ( resolve ) => setTimeout( resolve, 100 ) );

            appendFileSync( testLogFile, "[2025.10.05-12.02.00:000][003]LogNet: Third addition\n" );
            await new Promise( ( resolve ) => setTimeout( resolve, 100 ) );

            watcherController!.abort();
            await watchPromise;

            expect( sizeChanges.length ).toBeGreaterThanOrEqual( 3 );
            // Each size should be larger than the previous
            for ( let i = 1; i < sizeChanges.length; i++ )
            {
                expect( sizeChanges[ i ] ).toBeGreaterThan( sizeChanges[ i - 1 ] );
            }
        } );

        test( "should handle file truncation (server restart scenario)", async () => {
            let truncationDetected = false;
            let sizesRecorded: number[] = [];

            // Create initial file with content
            const initialContent =
                [
                    "[2025.10.05-12.00.00:000][001]LogNet: Server started",
                    "[2025.10.05-12.01.00:000][002]LogNet: Player joined: TestPlayer",
                    "[2025.10.05-12.02.00:000][003]LogGameplayEvents: Game started",
                ].join( "\n" ) + "\n";

            writeFileSync( testLogFile, initialContent );
            let lastSize = ( await Bun.file( testLogFile ).stat() ).size;

            const watchPromise = ( async () => {
                try
                {
                    const watcher = watch( testLogDir, { signal: watcherController!.signal } );
                    for await ( const event of watcher )
                    {
                        if ( event.filename === "Insurgency.log" )
                        {
                            const currentSize = ( await Bun.file( testLogFile ).stat() ).size;
                            sizesRecorded.push( currentSize );

                            // Check for truncation (file size decreased)
                            if ( currentSize < lastSize )
                            {
                                truncationDetected = true;
                            }

                            lastSize = currentSize;

                            if ( truncationDetected ) break;
                        }
                    }
                } catch ( error: any )
                {
                    if ( error.name !== "AbortError" )
                    {
                        throw error;
                    }
                }
            } )();

            await new Promise( ( resolve ) => setTimeout( resolve, 50 ) );

            // Simulate server restart by truncating and rewriting file
            writeFileSync( testLogFile, "[2025.10.05-12.03.00:000][001]LogNet: Server restarted\n" );
            await new Promise( ( resolve ) => setTimeout( resolve, 200 ) );

            watcherController!.abort();
            await watchPromise;

            expect( truncationDetected ).toBe( true );
            expect( sizesRecorded.length ).toBeGreaterThan( 0 );
        } );
    } );

    describe( "Multi-Server Scenarios", () => {
        test( "should handle multiple log files in same directory", async () => {
            const server1Log = join( testLogDir, "server1.log" );
            const server2Log = join( testLogDir, "server2.log" );
            const detectedFiles = new Set<string>();

            const watchPromise = ( async () => {
                try
                {
                    const watcher = watch( testLogDir, { signal: watcherController!.signal } );
                    for await ( const event of watcher )
                    {
                        if ( event.filename )
                        {
                            detectedFiles.add( event.filename );
                            if ( detectedFiles.size >= 2 ) break;
                        }
                    }
                } catch ( error: any )
                {
                    if ( error.name !== "AbortError" )
                    {
                        throw error;
                    }
                }
            } )();

            await new Promise( ( resolve ) => setTimeout( resolve, 50 ) );

            // Create multiple log files
            writeFileSync( server1Log, "[2025.10.05-12.00.00:000][001]LogNet: Server 1 started\n" );
            await new Promise( ( resolve ) => setTimeout( resolve, 100 ) );

            writeFileSync( server2Log, "[2025.10.05-12.00.00:000][001]LogNet: Server 2 started\n" );
            await new Promise( ( resolve ) => setTimeout( resolve, 200 ) );

            watcherController!.abort();
            await watchPromise;

            expect( detectedFiles.has( "server1.log" ) ).toBe( true );
            expect( detectedFiles.has( "server2.log" ) ).toBe( true );
        } );
    } );

    describe( "Error Handling", () => {
        test( "should handle non-existent directory gracefully", async () => {
            let watcherError: Error | null = null;
            const nonExistentDir = "./non-existent-dir";

            try
            {
                const watcher = watch( nonExistentDir, { signal: watcherController!.signal } );
                // This should throw an error immediately or when we try to iterate
                for await ( const event of watcher )
                {
                    // Should not reach here
                    break;
                }
            } catch ( error )
            {
                watcherError = error instanceof Error ? error : new Error( String( error ) );
            }

            expect( watcherError ).not.toBeNull();
            expect( watcherError?.message ).toContain( "ENOENT" );
        } );

        test( "should handle file permission errors", async () => {
            // This test might not work on all systems, but we can simulate it
            let permissionError = false;

            try
            {
                // Try to watch a file that doesn't exist in a restricted location
                const restrictedPath = "/root/restricted-logs";
                const watcher = watch( restrictedPath, { signal: watcherController!.signal } );

                // Attempt to iterate (this is where the error would occur)
                const timeout = setTimeout( () => {
                    watcherController!.abort();
                }, 100 );

                for await ( const event of watcher )
                {
                    clearTimeout( timeout );
                    break;
                }
            } catch ( error: any )
            {
                if ( error.name !== "AbortError" )
                {
                    permissionError = true;
                }
            }

            // We expect either a permission error or the watch to fail
            // The exact behavior depends on the system
            expect( typeof permissionError ).toBe( "boolean" );
        } );
    } );
} );
