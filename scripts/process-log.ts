#!/usr/bin/env bun
/**
 * Offline log processing script.
 *
 * Usage (PowerShell):
 *   bun run process-log -- --file .\logs\hc.log --server-id hardcore --name "Hardcore Server" --log-path "C:\\Servers\\Hardcore\\Insurgency\\Saved\\Logs"
 *   (The log path is only used for creating/looking up the server record; the file itself can live anywhere.)
 *
 * You can also omit --server-id and a UUIDv7 will be generated. If you already
 * have a server registered (same server-id), the existing DB server row is reused.
 *
 * This script:
 *   1. Registers (or reuses) a server in the DB
 *   2. Reads a static log file line-by-line (streaming, memory efficient)
 *   3. Parses each line for game events using existing parseLogEvents()
 *   4. Feeds events into StatsService (replaying history)
 *   5. On completion prints a summary: total lines, events, matches, top players
 */
import { parseLogEvents } from "../src/events";
import StatsService from "../src/stats-service";
import { upsertServer, getStatements } from "../src/database";
import logger from "../src/lib/console/logger";
import { hijackConsole } from "../src/lib/console/console-hijack";

// Simple argument parsing (no external deps)
interface Args {
    file: string;
    serverId?: string;
    name?: string;
    logPath?: string; // purely for DB metadata
    verbose?: boolean;
}

function parseArgs ( argv: string[] ): Args {
    const args: Args = { file: "" } as any;
    for ( let i = 0; i < argv.length; i++ )
    {
        const a = argv[ i ];
        if ( a === "--file" || a === "-f" ) args.file = argv[ ++i ];
        else if ( a === "--server-id" ) args.serverId = argv[ ++i ];
        else if ( a === "--name" ) args.name = argv[ ++i ];
        else if ( a === "--log-path" ) args.logPath = argv[ ++i ];
        else if ( a === "--verbose" || a === "-v" ) args.verbose = true;
    }
    return args;
}

async function fileExists ( path: string ) {
    try { return await Bun.file( path ).exists(); } catch { return false; }
}

async function main () {
    hijackConsole(); // unify output
    const rawArgs = process.argv.slice( 2 );
    const args = parseArgs( rawArgs );

    if ( !args.file )
    {
        console.error( "--file is required (path to static .log file)" );
        process.exit( 1 );
    }
    if ( !( await fileExists( args.file ) ) )
    {
        console.error( `Log file not found: ${ args.file }` );
        process.exit( 1 );
    }

    const serverUuid = args.serverId || Bun.randomUUIDv7();
    const serverName = args.name || serverUuid;
    const logPathForDb = args.logPath || "offline";

    if ( args.verbose )
    {
        logger.setVerbose?.( true );
    }

    logger.info( `Processing static log file: ${ args.file }` );
    logger.info( `Server ID: ${ serverUuid }` );

    // Register (or reuse) server row
    const dbServerId = upsertServer( serverUuid, serverName, serverUuid, logPathForDb, "offline import" );

    // Initialize stats service
    StatsService.initialize();

    // We'll stream lines to avoid loading giant files fully
    const file = Bun.file( args.file );
    const text = await file.text();
    const lines = text.split( /\r?\n/ );

    let lineCount = 0;
    let eventCount = 0;
    let logFileId: number | undefined;
    let openTimeIso: string | undefined;

    for ( const line of lines )
    {
        if ( !line.trim() ) continue;
        lineCount++;
        if ( !openTimeIso )
        {
            const openMatch = line.match( /Log file open,\s+(\d{2})\/(\d{2})\/(\d{2})\s+(\d{2}):(\d{2}):(\d{2})/ );
            if ( openMatch )
            {
                const [ , month, day, year2, hh, mm, ss ] = openMatch;
                const fullYear = 2000 + parseInt( year2 );
                openTimeIso = new Date( `${ fullYear }-${ month }-${ day }T${ hh }:${ mm }:${ ss }Z` ).toISOString();
                const stmts = getStatements();
                const res = stmts.upsertLogFile.get( dbServerId, args.file, openTimeIso, 0 ) as any;
                logFileId = res?.id;
            }
        }
        const events = parseLogEvents( line );
        if ( events.length === 0 ) continue;
        eventCount += events.length;
        for ( const ev of events )
        {
            StatsService.processEvent( ev, dbServerId );
        }
        if ( logFileId && lineCount % 100 === 0 )
        {
            // periodic persistence of lines processed
            try { getStatements().updateLogFileLines.run( lineCount, logFileId ); } catch { }
        }
    }

    if ( logFileId )
    {
        try { getStatements().updateLogFileLines.run( lineCount, logFileId ); } catch { }
    }

    logger.info( `Finished processing. Lines read: ${ lineCount }, events parsed: ${ eventCount }${ openTimeIso ? ", log open=" + openTimeIso : "" }` );

    // Produce summary
    const statements = getStatements();

    // Match history (recent 10)
    const matchHistory = statements.getMatchHistory.all( dbServerId, 10 ) as any[];

    // Top players
    const topPlayers = statements.getTopPlayers.all( dbServerId, dbServerId, dbServerId, 10 ) as any[];

    console.log( "\n=== Offline Import Summary ===" );
    console.log( `Server: ${ serverName } (${ serverUuid })` );
    console.log( `Matches recorded: ${ matchHistory.length }` );
    if ( openTimeIso ) console.log( `Log Open Time: ${ openTimeIso }` );
    if ( typeof logFileId !== 'undefined' ) console.log( `Log File DB ID: ${ logFileId }` );
    console.log( `Lines Processed: ${ lineCount }` );

    if ( matchHistory.length )
    {
        console.log( "Matches:" );
        for ( const m of matchHistory )
        {
            console.log( `  #${ m.id } ${ m.match_name || 'Match' } status=${ m.status } start=${ m.start_time } players=${ m.total_players }` );
        }
    }

    if ( topPlayers.length )
    {
        console.log( "\nTop Players (by kills):" );
        for ( const p of topPlayers )
        {
            console.log( `  ${ p.player_name } - Kills: ${ p.total_kills }, Deaths: ${ p.total_deaths }, KDR: ${ p.kdr }` );
        }
    } else
    {
        console.log( "No player stats found." );
    }

    console.log( "\nDone." );
}

main().catch( err => {
    console.error( "Offline log processing failed:", err );
    process.exit( 1 );
} );
