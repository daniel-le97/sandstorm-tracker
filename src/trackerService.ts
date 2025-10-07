import { getStatements } from "./database";
import type {
    ChatEvent,
    GameEvent,
    MapLoadEvent,
    PlayerJoinEvent,
    PlayerKillEvent,
    PlayerLeaveEvent,
    RoundOverEvent,
} from "./events";

// Track active sessions and current game state per server
const activeSessions = new Map<string, Map<string, { playerId: number; sessionId: number; joinTime: Date; }>>();
const currentMaps = new Map<string, string | null>();
const currentRounds = new Map<string, number>();
const currentMapIds = new Map<string, number | null>();

// Track active matches per server
const activeMatches = new Map<string, { matchId: number; startTime: Date; participants: Set<number>; }>();
const matchMapSequence = new Map<string, number>(); // Track map sequence in match

// Helper function to get server-specific session map
function getServerSessions ( serverId: number ): Map<string, { playerId: number; sessionId: number; joinTime: Date; }> {
    const serverKey = serverId.toString();
    if ( !activeSessions.has( serverKey ) )
    {
        activeSessions.set( serverKey, new Map() );
    }
    return activeSessions.get( serverKey )!;
}

// Crash detection variables per server
const lastFileActivity = new Map<string, Date>();
let crashCheckInterval: Timer | null = null;

// Start crash detection monitoring
function startCrashDetection (): void {
    if ( crashCheckInterval ) return; // Already running

    crashCheckInterval = setInterval( () => {
        const now = new Date();
        const tenSecondsInMs = 10 * 1000;

        // Check each server for crashes
        for ( const [ serverKey, lastActivity ] of lastFileActivity )
        {
            const timeSinceLastActivity = now.getTime() - lastActivity.getTime();
            const serverSessions = activeSessions.get( serverKey );

            if ( timeSinceLastActivity > tenSecondsInMs && serverSessions && serverSessions.size > 0 )
            {
                const serverId = parseInt( serverKey );
                console.log( `Server ${ serverId } crash detected! Ending all active sessions...` );
                TrackerService.handleServerCrash( serverId );
            }
        }
    }, 5000 ); // Check every 5 seconds
}

// Update last activity timestamp for a specific server
function updateLastActivity ( serverId: number ): void {
    lastFileActivity.set( serverId.toString(), new Date() );
}

export class TrackerService {
    /**
     * Initialize the stats service and start crash detection
     */
    static initialize (): void {
        startCrashDetection();
    }

    /**
     * Update file activity timestamp - call this whenever log file changes
     */
    static updateActivity ( serverId: number ): void {
        updateLastActivity( serverId );
    }

    /**
     * Handle server crash by ending all active sessions for a specific server
     */
    static handleServerCrash ( serverId: number ): void {
        const crashTime = new Date();
        const serverSessions = getServerSessions( serverId );

        for ( const [ playerName, session ] of serverSessions )
        {
            const durationMs = crashTime.getTime() - session.joinTime.getTime();
            const durationMinutes = Math.floor( durationMs / ( 1000 * 60 ) );

            // End the session with crash time
            getStatements().endSession.run( crashTime.toISOString(), durationMinutes, session.sessionId );

            // Update player's total playtime
            getStatements().updatePlayerPlaytime.run( durationMinutes, session.playerId );

            console.log( `Crash: Ended session for ${ playerName } on server ${ serverId } (${ durationMinutes } minutes)` );
        }

        serverSessions.clear();

        // End active match if exists
        const serverKey = serverId.toString();
        const activeMatch = activeMatches.get( serverKey );
        if ( activeMatch )
        {
            this.endMatch( serverId, "aborted" );
        }
    }

    /**
     * Process a game event and update database statistics
     */
    static processEvent ( event: GameEvent, serverId: number ): void {
        try
        {
            switch ( event.type )
            {
                case "player_join":
                    this.handlePlayerJoin( event as PlayerJoinEvent, serverId );
                    break;
                case "player_leave":
                case "player_disconnect":
                    this.handlePlayerLeave( event as PlayerLeaveEvent, serverId );
                    break;
                case "player_kill":
                case "team_kill":
                case "suicide":
                    this.handleKill( event as PlayerKillEvent, serverId );
                    break;

                case "round_over":
                    this.handleRoundOver( event as RoundOverEvent, serverId );
                    break;
                case "map_load":
                    this.handleMapLoad( event as MapLoadEvent, serverId );
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
    private static handlePlayerJoin ( event: PlayerJoinEvent, serverId: number ): void {
        const timestamp = new Date( this.parseTimestamp( event.timestamp ) );

        // Create or update player record
        // Generate unique steamId when not available to avoid conflicts
        const steamId = ( event.data as any ).steamId || `unknown_${ event.data.playerName }_${ Date.now() }`;
        const result = getStatements().upsertPlayer.get(
            steamId,
            event.data.playerName,
            serverId
        ) as any;

        const playerId = result?.id;
        if ( !playerId ) return;

        // Start a new session
        const currentMap = currentMaps.get( serverId.toString() ) || null;
        const sessionResult = getStatements().startSession.get(
            playerId,
            serverId,
            timestamp.toISOString(),
            currentMap
        ) as any;

        const sessionId = sessionResult?.id;
        if ( sessionId )
        {
            const serverSessions = getServerSessions( serverId );
            serverSessions.set( event.data.playerName, {
                playerId,
                sessionId,
                joinTime: timestamp,
            } );

            // Add player to active match if one exists
            const serverKey = serverId.toString();
            const activeMatch = activeMatches.get( serverKey );
            if ( activeMatch && !activeMatch.participants.has( playerId ) )
            {
                activeMatch.participants.add( playerId );

                // Add to match participants table
                getStatements().addMatchParticipant.run(
                    activeMatch.matchId,
                    playerId,
                    serverId,
                    timestamp.toISOString()
                );

                // Update match player counts
                getStatements().updateMatchPlayerCount.run(
                    activeMatch.participants.size,
                    activeMatch.participants.size,
                    activeMatch.matchId,
                    serverId
                );

                console.log( `Added ${ event.data.playerName } to match ${ activeMatch.matchId }` );
            }
        }

        console.log( `Started tracking session for ${ event.data.playerName } on server ${ serverId }` );
    }

    /**
     * Handle player leave/disconnect events
     */
    private static handlePlayerLeave ( event: PlayerLeaveEvent, serverId: number ): void {
        const playerName = event.data.playerName || "Unknown";
        const timestamp = new Date( this.parseTimestamp( event.timestamp ) );

        const serverSessions = getServerSessions( serverId );
        const session = serverSessions.get( playerName );
        if ( session )
        {
            // Calculate session duration
            const durationMs = timestamp.getTime() - session.joinTime.getTime();
            const durationMinutes = Math.floor( durationMs / ( 1000 * 60 ) );

            // End the session with calculated duration
            getStatements().endSession.run( timestamp.toISOString(), durationMinutes, session.sessionId );

            // Update player's total playtime
            getStatements().updatePlayerPlaytime.run( durationMinutes, session.playerId );

            // Handle match participant departure
            const serverKey = serverId.toString();
            const activeMatch = activeMatches.get( serverKey );
            if ( activeMatch && activeMatch.participants.has( session.playerId ) )
            {
                // Get final player stats for match
                const playerStats = getStatements().getPlayerStats.get( serverId, serverId, "unknown", serverId ) as any;

                // End match participation with final stats
                getStatements().endMatchParticipant.run(
                    timestamp.toISOString(),
                    durationMinutes,
                    playerStats?.total_kills || 0,
                    playerStats?.total_deaths || 0,
                    playerStats?.team_kills || 0,
                    playerStats?.suicides || 0,
                    0, // score - not tracked yet
                    activeMatch.matchId,
                    session.playerId
                );

                activeMatch.participants.delete( session.playerId );

                // Update match player count
                getStatements().updateMatchPlayerCount.run(
                    activeMatch.participants.size,
                    activeMatch.participants.size,
                    activeMatch.matchId,
                    serverId
                );
            }

            serverSessions.delete( playerName );
            console.log( `Ended session for ${ playerName } on server ${ serverId } (${ durationMinutes } minutes)` );
        }
    }

    /**
     * Handle kill events (player_kill, team_kill, suicide)
     */
    private static handleKill ( event: PlayerKillEvent, serverId: number ): void {
        const timestamp = new Date( this.parseTimestamp( event.timestamp ) );

        // Get or create killer (try by Steam ID first, then by name if not found)
        let killer = getStatements().getPlayer.get( event.data.killerSteamId, serverId ) as any;
        if ( !killer )
        {
            // Check if player exists with name but unknown steam_id
            const existingPlayer = getStatements().getPlayerByName.get( event.data.killer, serverId ) as any;
            if ( existingPlayer && existingPlayer.steam_id === "unknown" )
            {
                // Update existing player with real Steam ID
                getStatements().updatePlayerSteamId.run( event.data.killerSteamId, existingPlayer.id );
                killer = getStatements().getPlayer.get( event.data.killerSteamId, serverId ) as any;
            } else
            {
                // Create new player
                killer = getStatements().upsertPlayer.get( event.data.killerSteamId, event.data.killer, serverId ) as any;
            }
        }

        // Get or create victim (if different from killer)
        let victim = null;
        if ( event.data.killer !== event.data.victim )
        {
            victim = getStatements().getPlayer.get( event.data.victimSteamId, serverId ) as any;
            if ( !victim )
            {
                // Check if player exists with name but unknown steam_id
                const existingPlayer = getStatements().getPlayerByName.get( event.data.victim, serverId ) as any;
                if ( existingPlayer && existingPlayer.steam_id === "unknown" )
                {
                    // Update existing player with real Steam ID
                    getStatements().updatePlayerSteamId.run( event.data.victimSteamId, existingPlayer.id );
                    victim = getStatements().getPlayer.get( event.data.victimSteamId, serverId ) as any;
                } else
                {
                    // Create new player
                    victim = getStatements().upsertPlayer.get(
                        event.data.victimSteamId,
                        event.data.victim,
                        serverId
                    ) as any;
                }
            }
        }

        const killerId = killer?.id;
        const victimId = victim?.id || killerId; // For suicide, victim = killer

        if ( !killerId ) return;

        // Insert kill record
        const currentMap = currentMaps.get( serverId.toString() ) || null;
        const currentRound = currentRounds.get( serverId.toString() ) || 1;
        getStatements().insertKill.run(
            killerId,
            victimId,
            serverId,
            event.data.weapon,
            event.type,
            event.data.killerTeam,
            event.data.victimTeam,
            currentMap,
            currentRound,
            timestamp.toISOString()
        );

        // Update weapon stats for killer
        if ( event.type === "player_kill" )
        {
            getStatements().upsertWeaponStats.run( killerId, serverId, event.data.weapon, 1, 0, 0, 0 );
        } else if ( event.type === "team_kill" )
        {
            getStatements().upsertWeaponStats.run( killerId, serverId, event.data.weapon, 0, 0, 1, 0 );
        } else if ( event.type === "suicide" )
        {
            getStatements().upsertWeaponStats.run( killerId, serverId, event.data.weapon, 0, 0, 0, 1 );
        }

        // Update weapon stats for victim (deaths)
        if ( victimId && victimId !== killerId && ( event.type === "player_kill" || event.type === "team_kill" ) )
        {
            getStatements().upsertWeaponStats.run( victimId, serverId, event.data.weapon, 0, 1, 0, 0 );
        }

        console.log( `Recorded ${ event.type }: ${ event.data.killer } → ${ event.data.victim } (${ event.data.weapon })` );
    }



    /**
     * Handle round over events
     */
    private static handleRoundOver ( event: RoundOverEvent, serverId: number ): void {
        const serverKey = serverId.toString();
        const currentMapId = currentMapIds.get( serverKey );
        if ( !currentMapId ) return;

        const timestamp = new Date( this.parseTimestamp( event.timestamp ) );

        // Insert round record (skipped for now due to schema issues)
        // const roundResult = getStatements().insertRound?.get( ... ) as any;

        currentRounds.set( serverKey, event.data.roundNumber + 1 );

        console.log(
            `Recorded round ${ event.data.roundNumber } end on server ${ serverId }: Team ${ event.data.winningTeam } won (${ event.data.winReason })`
        );
    }

    /**
     * Handle map load events
     */
    private static handleMapLoad ( event: MapLoadEvent, serverId: number ): void {
        const serverKey = serverId.toString();
        const timestamp = new Date( this.parseTimestamp( event.timestamp ) );

        currentMaps.set( serverKey, event.data.mapName );
        currentRounds.set( serverKey, 1 );

        // Create or update map record
        const mapResult = getStatements().upsertMap.get( event.data.mapName, event.data.scenario, serverId ) as any;
        currentMapIds.set( serverKey, mapResult?.id );

        // Handle match logic
        const existingMatch = activeMatches.get( serverKey );

        if ( existingMatch )
        {
            // Map change within existing match - add to match maps
            const currentSequence = ( matchMapSequence.get( serverKey ) || 0 ) + 1;
            matchMapSequence.set( serverKey, currentSequence );

            if ( mapResult?.id )
            {
                getStatements().addMatchMap.run(
                    existingMatch.matchId,
                    mapResult.id,
                    serverId,
                    currentSequence,
                    timestamp.toISOString()
                );
            }
        } else
        {
            // Start a new match
            const matchName = `Match ${ timestamp.toISOString().split( "T" )[ 0 ] } ${ event.data.mapName }`;
            const matchResult = getStatements().startMatch.get(
                serverId,
                matchName,
                timestamp.toISOString(),
                0, // initial player count
                0 // initial max player count
            ) as any;

            if ( matchResult?.id )
            {
                activeMatches.set( serverKey, {
                    matchId: matchResult.id,
                    startTime: timestamp,
                    participants: new Set(),
                } );
                matchMapSequence.set( serverKey, 1 );

                // Add the first map to the match
                if ( mapResult?.id )
                {
                    getStatements().addMatchMap.run( matchResult.id, mapResult.id, serverId, 1, timestamp.toISOString() );
                }

                console.log( `Started new match ${ matchResult.id } on server ${ serverId }: ${ event.data.mapName }` );
            }
        }

        console.log( `Loaded map on server ${ serverId }: ${ event.data.mapName } (${ event.data.scenario })` );
    }

    /**
     * Get player statistics
     */
    static getPlayerStats ( steamId: string, serverId: number ) {
        return getStatements().getPlayerStats.get( serverId, serverId, steamId, serverId );
    }

    /**
     * Get top players by kills
     */
    static getTopPlayers ( serverId: number, limit = 10 ) {
        return getStatements().getTopPlayers.all( serverId, serverId, serverId, limit );
    }

    /**
     * Get player's weapon statistics
     */
    static getPlayerWeapons ( steamId: string, serverId: number, limit = 5 ) {
        // The prepared statement expects steamId, serverId, serverId, limit
        return getStatements().getPlayerWeapons.all( steamId, serverId, serverId, limit );
    }

    /**
     * Parse timestamp from log format to Date
     */
    private static parseTimestamp ( timestamp: string ): string {
        if ( !timestamp ) return new Date().toISOString();

        // Check if it's already an ISO string (for tests)
        if ( timestamp.match( /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d{3})?Z?$/ ) )
        {
            return new Date( timestamp ).toISOString();
        }

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
     * Handle server shutdown - end all active sessions for a specific server
     */
    static endAllSessions ( serverId: number ): void {
        const timestamp = new Date();
        const serverSessions = getServerSessions( serverId );

        for ( const [ playerName, session ] of serverSessions )
        {
            getStatements().endSession.run( timestamp.toISOString(), timestamp.toISOString(), session.playerId );
            console.log( `Ended session for ${ playerName } on server ${ serverId } (server shutdown)` );
        }

        serverSessions.clear();
    }

    /**
     * Clear all in-memory state for testing purposes
     */
    static clearAllState (): void {
        activeSessions.clear();
        currentMaps.clear();
        currentRounds.clear();
        currentMapIds.clear();
        activeMatches.clear();
        matchMapSequence.clear();
    }

    /**
     * Get current active sessions count for a specific server
     */
    static getActiveSessionCount ( serverId: number ): number {
        const serverSessions = getServerSessions( serverId );
        return serverSessions.size;
    }

    /**
     * Get active sessions for testing
     */
    static getActiveSessions ( serverId: number ): Map<string, { playerId: number; sessionId: number; joinTime: Date; }> {
        const serverSessions = getServerSessions( serverId );
        return new Map( serverSessions );
    }

    /**
     * Get all active sessions across all servers
     */
    static getAllActiveSessions (): Map<string, Map<string, { playerId: number; sessionId: number; joinTime: Date; }>> {
        return new Map( activeSessions );
    }

    /**
     * Get total active session count across all servers
     */
    static getTotalActiveSessionCount (): number {
        let total = 0;
        for ( const serverSessions of activeSessions.values() )
        {
            total += serverSessions.size;
        }
        return total;
    }

    /**
     * Get last file change time for testing
     */
    static getLastFileChange ( serverId: number ): Date {
        const serverKey = serverId.toString();
        return lastFileActivity.get( serverKey ) || new Date();
    }

    /**
     * Update last file change time for testing
     */
    static updateLastFileChange ( serverId: number ): void {
        updateLastActivity( serverId );
    }

    /**
     * Get player stats by player name for testing
     */
    static getPlayerStatsByName ( playerName: string, serverId: number ) {
        return getStatements().getPlayerByName.get( playerName, serverId );
    }

    /**
     * End an active match with the specified status
     */
    static endMatch ( serverId: number, status: "completed" | "aborted" = "completed", timestamp?: string ): void {
        const serverKey = serverId.toString();
        const activeMatch = activeMatches.get( serverKey );

        if ( !activeMatch ) return;

        const endTime = timestamp ? new Date( this.parseTimestamp( timestamp ) ) : new Date();
        const durationMinutes = Math.floor( ( endTime.getTime() - activeMatch.startTime.getTime() ) / ( 1000 * 60 ) );

        // Update match status
        if ( status === "completed" )
        {
            getStatements().endMatch.run( endTime.toISOString(), durationMinutes, activeMatch.matchId, serverId );
        } else
        {
            getStatements().abortMatch.run( endTime.toISOString(), activeMatch.matchId, serverId );
        }

        // End participation for all remaining participants
        for ( const playerId of activeMatch.participants )
        {
            const playerStats = getStatements().getPlayerStats.get( serverId, serverId, "unknown", serverId ) as any;

            getStatements().endMatchParticipant.run(
                endTime.toISOString(),
                durationMinutes,
                playerStats?.total_kills || 0,
                playerStats?.total_deaths || 0,
                playerStats?.team_kills || 0,
                playerStats?.suicides || 0,
                0, // score
                activeMatch.matchId,
                playerId
            );
        }

        // Clean up state
        activeMatches.delete( serverKey );
        matchMapSequence.delete( serverKey );

        console.log(
            `${ status === "completed" ? "Completed" : "Aborted" } match ${ activeMatch.matchId
            } on server ${ serverId } (${ durationMinutes } minutes)`
        );
    }

    /**
     * Get active match for a server
     */
    static getActiveMatch ( serverId: number ) {
        return activeMatches.get( serverId.toString() );
    }

    /**
     * Get match history for a server
     */
    static getMatchHistory ( serverId: number, limit = 10 ) {
        return getStatements().getMatchHistory.all( serverId, limit );
    }

    /**
     * Get match details including participants and maps
     */
    static getMatchDetails ( matchId: number, serverId: number ) {
        const match = getStatements().getMatchDetails.get( matchId, serverId );
        const participants = getStatements().getMatchParticipants.all( matchId, serverId );
        const maps = getStatements().getMatchMaps.all( matchId, serverId );

        return {
            match,
            participants,
            maps,
        };
    }

    /**
     * Get player's match history
     */
    static getPlayerMatchHistory ( steamId: string, serverId: number, limit = 5 ) {
        return getStatements().getPlayerMatchHistory.all( steamId, serverId, serverId, limit );
    }
}

export default TrackerService;
