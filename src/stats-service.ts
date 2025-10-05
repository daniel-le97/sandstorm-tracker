import { getStatements, initializeDatabase } from './database';
import type { GameEvent, PlayerKillEvent, ChatEvent, PlayerJoinEvent, PlayerLeaveEvent, RoundOverEvent, MapLoadEvent } from './events';

// Initialize database on import
initializeDatabase();

// Track active sessions and current game state
const activeSessions = new Map<string, { playerId: number, sessionId: number, joinTime: Date; }>();
let currentMap: string | null = null;
let currentRound = 1;
let currentMapId: number | null = null;

// Crash detection variables
let lastFileActivity: Date = new Date();
let crashCheckInterval: Timer | null = null;

// Start crash detection monitoring
function startCrashDetection (): void {
    if ( crashCheckInterval ) return; // Already running

    crashCheckInterval = setInterval( () => {
        const now = new Date();
        const timeSinceLastActivity = now.getTime() - lastFileActivity.getTime();
        const tenSecondsInMs = 10 * 1000;

        if ( timeSinceLastActivity > tenSecondsInMs && activeSessions.size > 0 )
        {
            console.log( '🚨 Server crash detected! Ending all active sessions...' );
            StatsService.handleServerCrash();
        }
    }, 5000 ); // Check every 5 seconds
}

// Update last activity timestamp
function updateLastActivity (): void {
    lastFileActivity = new Date();
}

export class StatsService {

    /**
     * Initialize the stats service and start crash detection
     */
    static initialize (): void {
        startCrashDetection();
    }

    /**
     * Update file activity timestamp - call this whenever log file changes
     */
    static updateActivity (): void {
        updateLastActivity();
    }

    /**
     * Handle server crash by ending all active sessions
     */
    static handleServerCrash (): void {
        const crashTime = new Date();

        for ( const [ playerName, session ] of activeSessions )
        {
            const durationMs = crashTime.getTime() - session.joinTime.getTime();
            const durationMinutes = Math.floor( durationMs / ( 1000 * 60 ) );

            // End the session with crash time
            getStatements().endSession.run(
                crashTime.toISOString(),
                durationMinutes,
                session.sessionId
            );

            // Update player's total playtime
            getStatements().updatePlayerPlaytime.run(
                durationMinutes,
                session.playerId
            );

            console.log( `📊 Crash: Ended session for ${ playerName } (${ durationMinutes } minutes)` );
        }

        activeSessions.clear();
    }

    /**
     * Process a game event and update database statistics
     */
    static processEvent ( event: GameEvent ): void {
        try
        {
            switch ( event.type )
            {
                case 'player_join':
                    this.handlePlayerJoin( event as PlayerJoinEvent );
                    break;
                case 'player_leave':
                case 'player_disconnect':
                    this.handlePlayerLeave( event as PlayerLeaveEvent );
                    break;
                case 'player_kill':
                case 'team_kill':
                case 'suicide':
                    this.handleKill( event as PlayerKillEvent );
                    break;
                case 'chat_command':
                    this.handleChatCommand( event as ChatEvent );
                    break;
                case 'round_over':
                    this.handleRoundOver( event as RoundOverEvent );
                    break;
                case 'map_load':
                    this.handleMapLoad( event as MapLoadEvent );
                    break;
            }
        } catch ( error )
        {
            console.error( `Error processing ${ event.type } event:`, error );
        }
    }

    /**
     * Handle player join events
     */
    private static handlePlayerJoin ( event: PlayerJoinEvent ): void {
        const timestamp = new Date();

        // Create or update player record
        const result = getStatements().upsertPlayer.get(
            'unknown', // We don't have steam_id from join event
            event.data.playerName
        ) as any;

        const playerId = result?.id;
        if ( !playerId ) return;

        // Start a new session
        const sessionResult = getStatements().startSession.get(
            playerId,
            timestamp.toISOString(),
            currentMap
        ) as any;

        const sessionId = sessionResult?.id;
        if ( sessionId )
        {
            activeSessions.set( event.data.playerName, {
                playerId,
                sessionId,
                joinTime: timestamp
            } );
        }

        console.log( `📊 Started tracking session for ${ event.data.playerName }` );
    }

    /**
     * Handle player leave/disconnect events
     */
    private static handlePlayerLeave ( event: PlayerLeaveEvent ): void {
        const playerName = event.data.playerName || 'Unknown';
        const timestamp = new Date();

        const session = activeSessions.get( playerName );
        if ( session )
        {
            // Calculate session duration
            const durationMs = timestamp.getTime() - session.joinTime.getTime();
            const durationMinutes = Math.floor( durationMs / ( 1000 * 60 ) );

            // End the session with calculated duration
            getStatements().endSession.run(
                timestamp.toISOString(),
                durationMinutes,
                session.sessionId
            );

            // Update player's total playtime
            getStatements().updatePlayerPlaytime.run(
                durationMinutes,
                session.playerId
            );

            activeSessions.delete( playerName );
            console.log( `📊 Ended session for ${ playerName } (${ durationMinutes } minutes)` );
        }
    }

    /**
     * Handle kill events (player_kill, team_kill, suicide)
     */
    private static handleKill ( event: PlayerKillEvent ): void {
        const timestamp = new Date( this.parseTimestamp( event.timestamp ) );

        // Get or create killer
        const killer = getStatements().upsertPlayer.get(
            event.data.killerSteamId,
            event.data.killer
        ) as any;

        // Get or create victim (if different from killer)
        let victim = null;
        if ( event.data.killer !== event.data.victim )
        {
            victim = getStatements().upsertPlayer.get(
                event.data.victimSteamId,
                event.data.victim
            ) as any;
        }

        const killerId = killer?.id;
        const victimId = victim?.id || killerId; // For suicide, victim = killer

        if ( !killerId ) return;

        // Insert kill record
        getStatements().insertKill.run(
            killerId,
            victimId,
            event.data.weapon,
            event.type,
            event.data.killerTeam,
            event.data.victimTeam,
            currentMap,
            currentRound,
            timestamp.toISOString()
        );

        // Update weapon stats for killer
        if ( event.type === 'player_kill' )
        {
            getStatements().upsertWeaponStats.run( killerId, event.data.weapon, 1, 0, 0, 0 );
        } else if ( event.type === 'team_kill' )
        {
            getStatements().upsertWeaponStats.run( killerId, event.data.weapon, 0, 0, 1, 0 );
        } else if ( event.type === 'suicide' )
        {
            getStatements().upsertWeaponStats.run( killerId, event.data.weapon, 0, 0, 0, 1 );
        }

        // Update weapon stats for victim (deaths)
        if ( victimId && victimId !== killerId && ( event.type === 'player_kill' || event.type === 'team_kill' ) )
        {
            getStatements().upsertWeaponStats.run( victimId, event.data.weapon, 0, 1, 0, 0 );
        }

        console.log( `📊 Recorded ${ event.type }: ${ event.data.killer } → ${ event.data.victim } (${ event.data.weapon })` );
    }

    /**
     * Handle chat command events
     */
    private static handleChatCommand ( event: ChatEvent ): void {
        const timestamp = new Date( this.parseTimestamp( event.timestamp ) );

        // Get or create player
        const player = getStatements().upsertPlayer.get(
            event.data.steamId,
            event.data.playerName
        ) as any;

        if ( !player?.id ) return;

        // Record chat command
        getStatements().insertChatCommand.run(
            player.id,
            event.data.command,
            event.data.args?.join( ' ' ) || null,
            timestamp.toISOString()
        );

        console.log( `📊 Recorded command: ${ event.data.playerName } used ${ event.data.command }` );
    }

    /**
     * Handle round over events
     */
    private static handleRoundOver ( event: RoundOverEvent ): void {
        if ( !currentMapId ) return;

        const timestamp = new Date( this.parseTimestamp( event.timestamp ) );

        // Insert round record (skipped for now due to schema issues)
        // const roundResult = getStatements().insertRound?.get( ... ) as any;

        currentRound = event.data.roundNumber + 1;

        console.log( `📊 Recorded round ${ event.data.roundNumber } end: Team ${ event.data.winningTeam } won (${ event.data.winReason })` );
    }

    /**
     * Handle map load events
     */
    private static handleMapLoad ( event: MapLoadEvent ): void {
        currentMap = event.data.mapName;
        currentRound = 1;

        // Create or update map record
        const mapResult = getStatements().upsertMap.get(
            event.data.mapName,
            event.data.scenario
        ) as any;

        currentMapId = mapResult?.id;

        console.log( `📊 Loaded map: ${ event.data.mapName } (${ event.data.scenario })` );
    }

    /**
     * Get player statistics
     */
    static getPlayerStats ( steamId: string ) {
        return getStatements().getPlayerStats.get( steamId );
    }

    /**
     * Get top players by kills
     */
    static getTopPlayers ( limit = 10 ) {
        return getStatements().getTopPlayers.all( limit );
    }

    /**
     * Get player's weapon statistics
     */
    static getPlayerWeapons ( steamId: string, limit = 5 ) {
        // The prepared statement expects steamId, not player.id
        return getStatements().getPlayerWeapons.all( steamId, limit );
    }

    /**
     * Parse timestamp from log format to Date
     */
    private static parseTimestamp ( timestamp: string ): string {
        if ( !timestamp ) return new Date().toISOString();

        // Convert format [2025.10.04-15.23.38:790] to ISO string
        const match = timestamp.match( /(\d{4})\.(\d{2})\.(\d{2})-(\d{2})\.(\d{2})\.(\d{2}):(\d{3})/ );
        if ( match )
        {
            const [ , year, month, day, hour, minute, second, ms ] = match;
            return new Date( `${ year }-${ month }-${ day }T${ hour }:${ minute }:${ second }.${ ms }Z` ).toISOString();
        }

        return new Date().toISOString();
    }

    /**
     * Handle server shutdown - end all active sessions
     */
    static endAllSessions (): void {
        const timestamp = new Date();

        for ( const [ playerName, session ] of activeSessions )
        {
            getStatements().endSession.run(
                timestamp.toISOString(),
                timestamp.toISOString(),
                session.playerId
            );
            console.log( `📊 Ended session for ${ playerName } (server shutdown)` );
        }

        activeSessions.clear();
    }

    /**
     * Get current active sessions count
     */
    static getActiveSessionCount (): number {
        return activeSessions.size;
    }

    /**
     * Get active sessions for testing
     */
    static getActiveSessions (): Map<string, { playerId: number, sessionId: number, joinTime: Date; }> {
        return new Map( activeSessions );
    }

    /**
     * Get last file change time for testing
     */
    static getLastFileChange (): Date {
        return new Date( lastFileActivity );
    }

    /**
     * Update last file change time for testing
     */
    static updateLastFileChange (): void {
        updateLastActivity();
    }

    /**
     * Get player stats by player name for testing
     */
    static getPlayerStatsByName ( playerName: string ) {
        return getStatements().getPlayerByName.get( playerName );
    }
}

export default StatsService;