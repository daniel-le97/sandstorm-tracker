import { watch } from "fs/promises";
import CommandHandler from "./command-handler";
import type { ServerConfig } from "./config";
import { ConfigLoader } from "./config-loader";
import { upsertServer } from "./database";
import type { ChatEvent, GameEvent } from "./events";
import { parseLogEvents } from "./events";
import StatsService from "./stats-service";
import { info, debug, warn, error } from "./lib/logger";
import { getStatements } from "./database";
import { sep } from "path";

// Server watcher state with health monitoring
interface ServerWatcher {
    config: ServerConfig;
    serverId: number; // Database ID
    lastProcessedTime: number;
    lastProcessedLine: string;
    fileSize: number;
    logFilePath: string;
    isHealthy: boolean;
    lastHealthCheck: number;
    errorCount: number;
    lastError: Error | null;
    logFileId?: number; // tracking row id in log_files
    logOpenTime?: string; // parsed log open time (ISO)
    linesProcessed: number; // number of lines processed for this log file
    fileSizeBytes: number; // current file size snapshot
}

const serverWatchers = new Map<string, ServerWatcher>();
const debounceDelay = 100; // 100ms debounce
const HEALTH_CHECK_INTERVAL = 30000; // 30 seconds
const MAX_ERROR_COUNT = 5;
let healthCheckTimer: ReturnType<typeof setInterval> | null = null;
let shuttingDown = false;
const activeDirWatchers = new Map<string, any>();

/**
 * Retry function with exponential backoff
 */
export async function retryWithBackoff<T> ( fn: () => Promise<T>, maxRetries: number = 3, baseDelay: number = 1000 ): Promise<T> {
    let lastError: Error;

    for ( let attempt = 1; attempt <= maxRetries; attempt++ )
    {
        try
        {
            return await fn();
        } catch ( err )
        {
            lastError = err instanceof Error ? err : new Error( String( err ) );

            if ( attempt === maxRetries )
            {
                throw lastError;
            }

            const delay = baseDelay * Math.pow( 2, attempt - 1 );
            warn( `⏳ Retry attempt ${ attempt }/${ maxRetries } failed, waiting ${ delay }ms...` );
            await new Promise( ( resolve ) => setTimeout( resolve, delay ) );
        }
    }

    throw lastError!;
}

/**
 * Schedule periodic health checks for all server watchers
 */
export function scheduleHealthChecks (): void {
    if ( healthCheckTimer ) return;
    // healthCheckTimer = setInterval( () => {
    //     performHealthChecks().catch( ( err ) => {
    //         error( "Health check failed:", err );
    //     } );
    // }, HEALTH_CHECK_INTERVAL );

    info( `Health checks scheduled every ${ HEALTH_CHECK_INTERVAL / 1000 }s` );
}

/**
 * Perform health checks on all server watchers
 */
export async function performHealthChecks (): Promise<void> {
    const unhealthyServers: string[] = [];

    for ( const [ serverId, watcher ] of serverWatchers )
    {
        try
        {
            // Check if log file is still accessible
            const file = Bun.file( watcher.logFilePath );
            await file.exists();

            // Check if we've had recent activity or too many errors
            const timeSinceLastActivity = Date.now() - watcher.lastProcessedTime;
            const wasHealthy = watcher.isHealthy;

            watcher.isHealthy = watcher.errorCount < MAX_ERROR_COUNT && timeSinceLastActivity < 300000; // 5 minutes
            watcher.lastHealthCheck = Date.now();

            if ( !watcher.isHealthy && wasHealthy )
            {
                warn( `Server "${ watcher.config.name }" marked as unhealthy (errors: ${ watcher.errorCount })` );
                unhealthyServers.push( watcher.config.name );
            } else if ( watcher.isHealthy && !wasHealthy )
            {
                info( `Server "${ watcher.config.name }" recovered and marked as healthy` );
                watcher.errorCount = 0;
                watcher.lastError = null;
            }
        } catch ( err )
        {
            watcher.isHealthy = false;
            watcher.errorCount++;
            watcher.lastError = err instanceof Error ? err : new Error( String( err ) );
            watcher.lastHealthCheck = Date.now();

            if ( !unhealthyServers.includes( watcher.config.name ) )
            {
                unhealthyServers.push( watcher.config.name );
            }
        }
    }

    if ( unhealthyServers.length > 0 )
    {
        warn( `Unhealthy servers detected: ${ unhealthyServers.join( ", " ) }` );
    }
}

// Initialize multi-server application with comprehensive error handling
export async function initializeApplication (): Promise<void> {
    info( "Initializing Sandstorm Multi-Server Tracker..." );

    // Load and validate configuration
    await ConfigLoader.loadConfig();
    const enabledServers = ConfigLoader.getEnabledServers();

    if ( enabledServers.length === 0 )
    {
        throw new Error( "No enabled servers found in configuration" );
    }

    info( `Found ${ enabledServers.length } enabled server(s)` );

    // Initialize stats service with error handling
    StatsService.initialize();
    debug( "✓ Statistics service initialized" );

    // Set up each server watcher with individual error handling
    const setupPromises = enabledServers.map( async ( serverConfig ) => {
        try
        {
            await setupServerWatcher( serverConfig );
            return { server: serverConfig.name, success: true };
        } catch ( err )
        {
            error( `Failed to setup watcher for server "${ serverConfig.name }":`, err );
            return { server: serverConfig.name, success: false, error: err };
        }
    } );

    const results = await Promise.allSettled( setupPromises );
    const successful = results.filter( ( r ) => ( r as any ).status === "fulfilled" && ( r as any ).value.success ).length;
    const failed = results.length - successful;

    if ( successful === 0 )
    {
        throw new Error( "All server watchers failed to initialize" );
    }

    info( `Multi-server tracker initialized successfully!` );
    info( `${ successful } server(s) active, ${ failed } failed` );
    debug( "Watching for file changes on active servers..." );

    // Schedule periodic health checks
    scheduleHealthChecks();
}

async function setupServerWatcher ( serverConfig: ServerConfig ): Promise<void> {
    // Register server in database
    const dbServerId = upsertServer(
        serverConfig.serverId,
        serverConfig.name,
        serverConfig.id,
        serverConfig.logPath,
        serverConfig.description
    );

    const logFilePath = `${ serverConfig.logPath }${ serverConfig.logPath.endsWith( "\\" ) || serverConfig.logPath.endsWith( "/" ) ? "" : "\\"
        }${ serverConfig.serverId }.log`;

    // Initialize file size
    let fileSize = 0;
    try
    {
        const stats = await Bun.file( logFilePath ).stat();
        fileSize = stats.size;
    } catch ( err )
    {
        debug( `Log file doesn't exist yet for ${ serverConfig.name }: ${ logFilePath }` );
    }

    // Create watcher state
    const watcher: ServerWatcher = {
        config: serverConfig,
        serverId: dbServerId,
        lastProcessedTime: 0,
        lastProcessedLine: "",
        fileSize: fileSize,
        logFilePath: logFilePath,
        isHealthy: true,
        lastHealthCheck: Date.now(),
        errorCount: 0,
        lastError: null,
        linesProcessed: 0,
        fileSizeBytes: fileSize,
    };

    serverWatchers.set( serverConfig.id, watcher );

    debug( `Server watcher configured: ${ serverConfig.name }` );
}

// Start watching all configured servers
export async function startWatching (): Promise<void> {
    const watchPromises: Promise<void>[] = [];

    // Group servers by log path to avoid duplicate watchers
    const pathToServers = new Map<string, ServerWatcher[]>();

    for ( const watcher of serverWatchers.values() )
    {
        const logDir = watcher.config.logPath;
        if ( !pathToServers.has( logDir ) )
        {
            pathToServers.set( logDir, [] );
        }
        pathToServers.get( logDir )!.push( watcher );
    }

    // Create one file system watcher per unique log directory
    for ( const [ logPath, watchers ] of pathToServers )
    {
        // Start watching for each server in this directory
        for ( const watcher of watchers )
        {
            if ( shuttingDown ) break;
            const p = watchLogDirectory( watcher.config.id );
            watchPromises.push( p );
        }
    }

    // Wait for all watchers (they run indefinitely)
    await Promise.all( watchPromises );
}

// Graceful shutdown: stop health checks and active watchers
export async function stopApp (): Promise<void> {
    if ( shuttingDown ) return;
    shuttingDown = true;

    info( 'Shutting down application...' );

    if ( healthCheckTimer )
    {
        clearInterval( healthCheckTimer );
        healthCheckTimer = null;
    }

    // signal active directory watchers to stop by clearing the map
    activeDirWatchers.clear();

    // give watchers some time to unwind
    await new Promise( ( r ) => setTimeout( r, 500 ) );

    info( 'Shutdown complete' );
}

async function watchLogDirectory ( serverId: string ): Promise<void> {
    const watcher = serverWatchers.get( serverId );
    if ( !watcher )
    {
        error( `No watcher found for server ID: ${ serverId }` );
        return;
    }

    const logPath = watcher.config.logPath;
    const watchers = Array.from( serverWatchers.values() ).filter( ( w ) => w.config.logPath === logPath );

    debug( `� Starting file watcher for directory: ${ logPath }` );

    let retryCount = 0;
    const maxRetries = 5;

    // register active watcher
    activeDirWatchers.set( serverId, true );

    while ( retryCount < maxRetries )
    {
        try
        {
            const dirWatcher = watch( logPath );
            debug( `✓ Successfully watching: ${ logPath }` );

            for await ( const event of dirWatcher )
            {
                console.log( event.filename );
                const file = Bun.file( `${ logPath }\\${ event.filename }` );
                const fileContent = ( await file.text() ).trim().split( '\n' );
                console.log( fileContent.pop() );
                try
                {
                    // Skip backup files and other irrelevant files
                    if ( !event.filename || event.filename.includes( "backup" ) || event.filename.includes( "tmp" ) )
                    {
                        continue;
                    }

                    // Find which server this file belongs to
                    for ( const serverWatcher of watchers )
                    {
                        if ( !serverWatcher.isHealthy )
                        {
                            // continue; // Skip unhealthy watchers
                        }

                        // Each server writes to a file named after its server UUID
                        // (configured as `serverId` in the server config). Build the
                        // expected filename per-server so multiple servers sharing a
                        // directory can be distinguished.
                        const expectedFilename = `${ serverWatcher.config.serverId }.log`;

                        if ( event.filename === expectedFilename )
                        {
                            await processServerLogChange( serverWatcher );
                            break;
                        }
                    }
                } catch ( processingError )
                {
                    const errorMsg =
                        processingError instanceof Error ? processingError.message : String( processingError );
                    error( `Error processing file event for ${ event.filename }:`, errorMsg );

                    // Update watcher error state
                    for ( const serverWatcher of watchers )
                    {
                        if ( serverWatcher.config.id === serverId )
                        {
                            serverWatcher.errorCount++;
                            serverWatcher.lastError =
                                processingError instanceof Error ? processingError : new Error( String( processingError ) );
                            break;
                        }
                    }
                }
            }

            // If we get here, the watcher ended normally - shouldn't happen
            warn( `File watcher ended unexpectedly for ${ logPath }` );
            break;
        } catch ( err )
        {
            retryCount++;
            const errorMsg = err instanceof Error ? err.message : String( err );
            error( `File watcher error (attempt ${ retryCount }/${ maxRetries }) for ${ logPath }:`, errorMsg );

            // Update watcher error state
            if ( watcher )
            {
                watcher.errorCount++;
                watcher.lastError = err instanceof Error ? err : new Error( String( err ) );
                watcher.isHealthy = watcher.errorCount < MAX_ERROR_COUNT;
            }

            if ( retryCount >= maxRetries )
            {
                error( `File watcher permanently failed for ${ logPath } after ${ maxRetries } attempts` );
                if ( watcher )
                {
                    watcher.isHealthy = false;
                }
                return;
            }

            // Wait before retrying with exponential backoff
            const delay = Math.pow( 2, retryCount ) * 2000; // Start at 4s, max ~1 minute
            debug( `⏳ Retrying file watcher in ${ delay }ms...` );
            await new Promise( ( resolve ) => setTimeout( resolve, delay ) );
        }
    }
}

async function processServerLogChange ( watcher: ServerWatcher ): Promise<void> {
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

// Function to process a log line from a specific server
async function processLogFile ( watcher: ServerWatcher, logLine: string ): Promise<GameEvent[]> {
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
        StatsService.updateActivity( watcher.serverId );

        events.forEach( ( event ) => {
            // Process event in database with server context
            StatsService.processEvent( event, watcher.serverId );

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
                    const response = CommandHandler.handleCommand( event as ChatEvent, watcher.serverId );
                    info( `${ serverPrefix } ${ event.data.playerName } used command: ${ event.data.command } ${ event.data.args?.join( " " ) || "" }` );
                    if ( response )
                    {
                        info( `${ serverPrefix } Response: ${ response }` );
                    }
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
