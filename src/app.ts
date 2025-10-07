import CommandHandler from "./command-handler";
import type { ServerConfig } from "./config";
import { ConfigLoader } from "./config-loader";
import type { ChatEvent, GameEvent } from "./events";
import { parseLogEvents } from "./events";
import TrackerService from "./trackerService";
import { info, debug, warn, error } from "./lib/console/logger";
import { watchersService, type ServerWatcher } from "./lib/watcher/watcher";
import { processServerLogChange, processLogFile } from "./lib/watcher/processor";
import { getStatements } from "./database";

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
// Local debounce/health constants used by processServerLogChange
const debounceDelay = 100; // ms
const MAX_ERROR_COUNT = 5;

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
    TrackerService.initialize();
    debug( "✓ Statistics service initialized" );

    // Set up each server watcher with individual error handling via the centralized service
    const setupPromises = enabledServers.map( async ( serverConfig ) => {
        try
        {
            await watchersService.addServerWatcher( serverConfig );
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

    // Schedule periodic health checks in the service
    watchersService.scheduleHealthChecks();
}

// setupServerWatcher is now handled by watchersService.addServerWatcher()

// Start watching all configured servers
export async function startWatching (): Promise<void> {
    await watchersService.startWatching( processServerLogChange );
}

// Graceful shutdown: stop health checks and active watchers
export async function stopApp (): Promise<void> {
    await watchersService.stopAll();
}

// Directory watching moved into watchersService

// processServerLogChange/processLogFile are implemented in ./lib/watcher/processor
