// Clean, minimal implementation of the watcher service. Previous corruption removed.
import { watch } from "fs/promises";
import { basename } from "path";
import { debug, info, warn, error } from "../../lib/console/logger";
import { upsertServer } from "../../database";
import type { ServerConfig } from "../../config";
import { processServerLogChange } from "./processor";

export interface ServerWatcher {
    config: ServerConfig;
    serverId: number;          // Database id
    logFilePath: string;       // Absolute (or joined) path to <uuid>.log
    fileSize: number;          // Last seen file size
    fileSizeBytes: number;     // Mirror for legacy naming
    lastProcessedTime: number; // Timestamp of last successful process
    lastProcessedLine: string; // Cache to avoid duplicate line processing
    isHealthy: boolean;        // False after repeated errors
    errorCount: number;        // Incremented on processor errors
    lastError: Error | null;   // Last error encountered
    linesProcessed: number;    // Lines processed in current open session
    logFileId?: number;        // Set when a new log session is recorded
    logOpenTime?: string;      // ISO time of log open line
    lastHealthCheck?: number;  // For future health logic
}

type LogChangeHandler = ( watcher: ServerWatcher ) => Promise<void>;

class WatchersService {
    private watchers = new Map<string, ServerWatcher>(); // key: server config id
    private dirTasks = new Map<string, Promise<void>>(); // key: directory path
    private shuttingDown = false;
    private healthInterval: ReturnType<typeof setInterval> | null = null;

    addServerWatcher ( cfg: ServerConfig ): void {
        if ( this.watchers.has( cfg.id ) ) return;
        const dbId = upsertServer( cfg.serverId, cfg.name, cfg.id, cfg.logPath, cfg.description );
        const sep = cfg.logPath.endsWith( "\\" ) || cfg.logPath.endsWith( "/" ) ? "" : "\\";
        const logFilePath = `${ cfg.logPath }${ sep }${ cfg.serverId }.log`;
        const watcher: ServerWatcher = {
            config: cfg,
            serverId: dbId,
            logFilePath,
            fileSize: 0,
            fileSizeBytes: 0,
            lastProcessedTime: 0,
            lastProcessedLine: "",
            isHealthy: true,
            errorCount: 0,
            lastError: null,
            linesProcessed: 0,
            lastHealthCheck: Date.now(),
        };
        this.watchers.set( cfg.id, watcher );
        debug( `[watcher] registered server ${ cfg.name }` );
    }

    async startWatching ( handler: LogChangeHandler = processServerLogChange ): Promise<void> {
        if ( this.shuttingDown ) return;
        // Group watchers by directory
        const byDir = new Map<string, ServerWatcher[]>();
        for ( const w of this.watchers.values() )
        {
            const dir = w.config.logPath;
            if ( !byDir.has( dir ) ) byDir.set( dir, [] );
            byDir.get( dir )!.push( w );
        }
        for ( const [ dir, list ] of byDir )
        {
            if ( this.dirTasks.has( dir ) ) continue;
            this.dirTasks.set( dir, this.watchDirectory( dir, list, handler ) );
        }
        // Opportunistic initial processing for any logs that already exist (non-blocking, slight delay
        // so that tests creating file immediately after startWatching still get picked up)
        setTimeout( () => {
            for ( const w of this.watchers.values() )
            {
                ( async () => {
                    try
                    {
                        const f = Bun.file( w.logFilePath );
                        if ( await f.exists() )
                        {
                            await handler( w ); // will no-op if no new content
                        }
                    } catch { /* ignore */ }
                } )();
            }
        }, 100 );
    }

    scheduleHealthChecks (): void {
        if ( this.healthInterval ) return;
        this.healthInterval = setInterval( () => {
            const now = Date.now();
            for ( const w of this.watchers.values() )
            {
                // If no activity for 5 minutes after errors started, mark unhealthy
                if ( w.errorCount > 0 && now - w.lastProcessedTime > 5 * 60 * 1000 )
                {
                    if ( w.isHealthy )
                    {
                        warn( `[watcher] ${ w.config.name } marked unhealthy due to inactivity after errors` );
                    }
                    w.isHealthy = false;
                }
                w.lastHealthCheck = now;
            }
        }, 60_000 );
    }

    async stopAll (): Promise<void> {
        if ( this.shuttingDown ) return;
        this.shuttingDown = true;
        info( "Shutting down file watchers..." );
        if ( this.healthInterval )
        {
            clearInterval( this.healthInterval );
            this.healthInterval = null;
        }
        // Wait for directory loops to settle
        await Promise.allSettled( this.dirTasks.values() );
        this.dirTasks.clear();
        info( "Watchers stopped" );
    }

    getWatchers (): ReadonlyMap<string, ServerWatcher> {
        return this.watchers;
    }

    private async watchDirectory ( dir: string, servers: ServerWatcher[], handler: LogChangeHandler ): Promise<void> {
        debug( `[watcher] starting directory watcher: ${ dir }` );
        const filenames = new Set( servers.map( s => `${ s.config.serverId }.log` ) );
        let retry = 0;
        const maxRetries = 5;
        while ( !this.shuttingDown && retry <= maxRetries )
        {
            try
            {
                const iterator = watch( dir, { recursive: false } );
                debug( `[watcher] watching ${ dir }` );
                for await ( const evt of iterator )
                {
                    if ( this.shuttingDown ) break;
                    // Bun/Node fs.watch (promises) events: { eventType, filename }
                    if ( !evt.filename ) continue;
                    // Normalize to basename for cross-platform consistency
                    const name = basename( evt.filename.toString() );
                    if ( !filenames.has( name ) ) continue; // ignore unrelated files
                    const sw = servers.find( s => `${ s.config.serverId }.log` === name );
                    if ( !sw || !sw.isHealthy ) continue;
                    try
                    {
                        await handler( sw );
                    } catch ( e )
                    {
                        sw.errorCount++;
                        sw.lastError = e instanceof Error ? e : new Error( String( e ) );
                        if ( sw.errorCount > 5 && sw.isHealthy )
                        {
                            sw.isHealthy = false;
                            warn( `[watcher] server ${ sw.config.name } marked unhealthy after repeated errors` );
                        }
                    }
                }
                break; // normal exit (iterator ended)
            } catch ( e )
            {
                if ( this.shuttingDown ) break;
                retry++;
                const msg = e instanceof Error ? e.message : String( e );
                error( `[watcher] error watching ${ dir } (attempt ${ retry }): ${ msg }` );
                if ( retry > maxRetries )
                {
                    error( `[watcher] giving up on ${ dir }` );
                    break;
                }
                const delay = Math.min( 30_000, 500 * 2 ** ( retry - 1 ) );
                await new Promise( r => setTimeout( r, delay ) );
            }
        }
        debug( `[watcher] directory loop ended: ${ dir }` );
    }
}

export const watchersService = new WatchersService();
export default watchersService;

