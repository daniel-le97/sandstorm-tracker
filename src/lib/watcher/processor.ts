import { info, debug, warn, error } from "../../lib/console/logger";
import TrackerService from "../../trackerService";
import { parseLogEvents } from "../../events";
import { getStatements } from "../../database";
import type { GameEvent } from "../../events";
import type { ServerWatcher } from "./watcher";

const DEFAULT_DEBOUNCE = 100;
const MAX_ERROR_COUNT = 5;

export async function processServerLogChange ( watcher: ServerWatcher, debounceDelay = DEFAULT_DEBOUNCE ): Promise<void> {
    const now = Date.now();

    // Debounce rapid file changes
    if ( now - watcher.lastProcessedTime < debounceDelay )
    {
        return;
    }

    // Skip if watcher is marked as unhealthy
    if ( !watcher.isHealthy )
    {
        return;
    }

    try
    {
        // Check if file exists and is accessible
        const file = Bun.file( watcher.logFilePath );
        const exists = await file.exists();
        if ( !exists )
        {
            warn( `⚠️ Log file disappeared: ${ watcher.logFilePath }` );
            return;
        }

        // Check if file actually changed size (new content)
        const currentStats = await file.stat();
        if ( currentStats.size <= watcher.fileSize )
        {
            return; // No new content
        }

        // Handle case where file was truncated (server restart)
        if ( currentStats.size < watcher.fileSize )
        {
            debug( `Log file truncated for ${ watcher.config.name }, resetting position` );
            watcher.fileSize = 0;
            watcher.lastProcessedLine = "";
        }

        // Read just the last line to check if it's new
        const logContent = await file.text();
        const lines = logContent.split( "\n" );
        const nonEmpty = lines.filter( ( line: string ) => line.trim() );
        const lastLine = nonEmpty.pop();

        // Detect log open lines (supporting rollover)
        const openLineMatches = nonEmpty.filter( l => /Log file open,\s+\d{2}\/\d{2}\/\d{2}\s+\d{2}:\d{2}:\d{2}/.test( l ) );
        if ( openLineMatches.length )
        {
            const lastOpen = openLineMatches.pop()!;
            const m = lastOpen.match( /Log file open,\s+(\d{2})\/(\d{2})\/(\d{2})\s+(\d{2}):(\d{2}):(\d{2})/ );
            if ( m )
            {
                const [ , month, day, year2, hh, mm, ss ] = m;
                const fullYear = parseInt( year2 ) + 2000;
                const iso = new Date( `${ fullYear }-${ month }-${ day }T${ hh }:${ mm }:${ ss }Z` ).toISOString();
                if ( !watcher.logOpenTime || watcher.logOpenTime !== iso )
                {
                    watcher.logOpenTime = iso;
                    watcher.linesProcessed = 0;
                    try
                    {
                        const stmts = getStatements();
                        const result = stmts.upsertLogFile.get( watcher.serverId, watcher.logFilePath, iso, 0, currentStats.size ) as any;
                        watcher.logFileId = result?.id;
                        debug( `Detected log open${ watcher.logFileId ? ' (id ' + watcher.logFileId + ')' : '' } for ${ watcher.config.name } (${ iso })` );
                    } catch ( e )
                    {
                        warn( `Failed to record log file open time for ${ watcher.config.name }: ${ e instanceof Error ? e.message : e }` );
                    }
                }
            }
        }

        if ( !lastLine || lastLine === watcher.lastProcessedLine )
        {
            watcher.fileSize = currentStats.size; // Update size even if no new line
            return; // Skip if no new content or same line as before
        }

        // Process the log file and parse events
        const gameEvents = await processLogFile( watcher, lastLine );

        // Increment lines processed counter and persist periodically
        watcher.linesProcessed += 1;
        if ( watcher.logFileId && watcher.linesProcessed % 50 === 0 ) // batch every 50 lines
        {
            try
            {
                getStatements().updateLogFileLines.run( watcher.linesProcessed, currentStats.size, watcher.logFileId );
            } catch ( e )
            {
                debug( `Failed to update lines_processed (${ watcher.linesProcessed }) for log file id ${ watcher.logFileId }` );
            }
        }

        // Update tracking variables successfully
        watcher.lastProcessedTime = now;
        watcher.fileSize = currentStats.size;
        watcher.lastProcessedLine = lastLine;

        // Reset error count on successful processing
        if ( watcher.errorCount > 0 )
        {
            info( `✓ Server ${ watcher.config.name } recovered from errors` );
            watcher.errorCount = 0;
            watcher.lastError = null;
            watcher.isHealthy = true;
        }

        // Final update if this specific line triggered a state change and we haven't persisted recently
        if ( watcher.logFileId && watcher.linesProcessed < 50 )
        {
            try
            {
                getStatements().updateLogFileLines.run( watcher.linesProcessed, currentStats.size, watcher.logFileId );
            } catch { }
        }
    } catch ( err )
    {
        const errorMsg = err instanceof Error ? err.message : String( err );
        error( `❌ Error processing log change for ${ watcher.config.name }:`, errorMsg );

        // Update error state
        watcher.errorCount++;
        watcher.lastError = err instanceof Error ? err : new Error( String( err ) );
        watcher.isHealthy = watcher.errorCount < MAX_ERROR_COUNT;

        if ( !watcher.isHealthy )
        {
            error( `Server "${ watcher.config.name }" marked as unhealthy due to repeated errors` );
        }
    }
}

export async function processLogFile ( watcher: ServerWatcher, logLine: string ): Promise<GameEvent[]> {
    try
    {
        // Parse just the provided log line
        const events = parseLogEvents( logLine );

        // Only process if we found actual events we care about
        if ( events.length === 0 )
        {
            return []; // Silent - no events we care about
        }

        // Update file activity timestamp
        TrackerService.updateActivity( watcher.serverId );

        events.forEach( ( event ) => {
            // Process event in database with server context
            TrackerService.processEvent( event, watcher.serverId );

            // Only log events that are specifically listed in events.md
            const serverPrefix = `[${ watcher.config.name }]`;

            switch ( event.type )
            {
                case "game_over":
                    info( `${ serverPrefix } Game over!` );
                    break;
                case "player_join":
                    info( `${ serverPrefix } Player joined: ${ event.data.playerName }` );
                    break;
                case "player_leave":
                case "player_disconnect":
                    info( `${ serverPrefix } Player left: ${ event.data.playerName || event.data.steamId }` );
                    break;
                case "team_kill":
                    info(
                        `${ serverPrefix } TEAM KILL: ${ event.data.killer } killed teammate ${ event.data.victim } with ${ event.data.weapon }`
                    );
                    break;
                case "player_kill":
                    info(
                        `${ serverPrefix } ⚔️ ${ event.data.killer } killed ${ event.data.victim } with ${ event.data.weapon }`
                    );
                    break;
                case "suicide":
                    info( `${ serverPrefix } ${ event.data.killer } committed suicide with ${ event.data.weapon }` );
                    break;
                case "difficulty_set":
                    info( `${ serverPrefix } ⚙️ AI difficulty set to: ${ event.data.difficulty }` );
                    break;
                case "round_over":
                    info(
                        `${ serverPrefix } Round ${ event.data.roundNumber } over: Team ${ event.data.winningTeam } won (${ event.data.winReason })`
                    );
                    break;
                case "map_load":
                    info( `${ serverPrefix } Loading map: ${ event.data.mapName } (${ event.data.scenario })` );
                    break;
                case "chat_command":
                    // Handle chat commands with server context
                    // Command handling is side-effectful and lives in the app's CLI layer; keep minimal here
                    debug( `${ serverPrefix } Chat command received: ${ event.data.command }` );
                    break;
                // Ignore any other events not listed in events.md
                default:
                    // Silent - don't log events we don't care about
                    break;
            }
        } );

        return events;
    } catch ( err )
    {
        error( `❌ Error processing log file for ${ watcher.config.name }:`, err );
        return [];
    }
}
