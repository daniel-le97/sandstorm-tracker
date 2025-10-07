import { beforeEach, describe, expect, test } from "bun:test";
import { getStatements, initializeDatabase, type MatchDetails, type MatchHistory, type MatchParticipant } from "../src/database";
import { MatchDetailsSchema, MatchHistorySchema, MatchParticipantSchema } from "../src/validation";
import * as v from "valibot";
import type { MapLoadEvent, PlayerJoinEvent, PlayerLeaveEvent } from "../src/events";
import { TrackerService } from "../src/trackerService";
import { unlinkSync } from "fs";

// Use test database
const testDbPath = "tests/databases/test_match_history.db";
process.env.TEST_DB_PATH = testDbPath;

// Helper function to create mock events
function createMapLoadEvent ( mapName: string, scenario: string, timestamp: string ): MapLoadEvent {
    return {
        type: "map_load",
        timestamp,
        data: { mapName, scenario },
        rawLine: `[${ timestamp }] LogGameState: Match State: Loading scenario BP_Scenario_${ scenario }_${ mapName }`,
    };
}

function createPlayerJoinEvent ( playerName: string, timestamp: string ): PlayerJoinEvent {
    return {
        type: "player_join",
        timestamp,
        data: { playerName },
        rawLine: `[${ timestamp }] LogNet: Player '${ playerName }' joined the server`,
    };
}

function createPlayerLeaveEvent ( playerName: string, timestamp: string ): PlayerLeaveEvent {
    return {
        type: "player_leave",
        timestamp,
        data: { playerName },
        rawLine: `[${ timestamp }] LogNet: Player '${ playerName }' left the server`,
    };
}

describe( "Match History", () => {
    beforeEach( () => {
        // Clean up any existing test database
        try
        {
            unlinkSync( testDbPath );
        } catch ( e )
        {
            // File doesn't exist, that's fine
        }

        // Initialize fresh database for each test
        initializeDatabase();

        // Clear all in-memory state from previous tests
        TrackerService.clearAllState();

        // Clear all database tables for clean state
        const db = require( "../src/database" ).default();
        db.exec( `
            DELETE FROM match_maps;
            DELETE FROM match_participants;
            DELETE FROM matches;
            DELETE FROM weapon_stats;
            DELETE FROM player_round_stats;
            DELETE FROM map_rounds;
            DELETE FROM maps;
            DELETE FROM kills;
            DELETE FROM player_sessions;
            DELETE FROM players;
            DELETE FROM servers;
        `);
    } );

    test( "should create matches table", () => {
        const database = require( "../src/database" ).default();
        const tables = database.prepare( "SELECT name FROM sqlite_master WHERE type='table' AND name='matches'" ).all();
        expect( tables ).toHaveLength( 1 );
        expect( tables[ 0 ].name ).toBe( "matches" );
    } );

    test( "should create match participants table", () => {
        const database = require( "../src/database" ).default();
        const tables = database
            .prepare( "SELECT name FROM sqlite_master WHERE type='table' AND name='match_participants'" )
            .all();
        expect( tables ).toHaveLength( 1 );
        expect( tables[ 0 ].name ).toBe( "match_participants" );
    } );

    test( "should create match maps table", () => {
        const database = require( "../src/database" ).default();
        const tables = database
            .prepare( "SELECT name FROM sqlite_master WHERE type='table' AND name='match_maps'" )
            .all();
        expect( tables ).toHaveLength( 1 );
        expect( tables[ 0 ].name ).toBe( "match_maps" );
    } );

    test( "should start a new match when map loads", () => {
        const statements = getStatements();

        // Create a server first
        const server = statements.upsertServer.get(
            "test-server",
            "Test Server",
            "test-config",
            "/test/path",
            "Test Description"
        ) as any;
        const serverId = server.id;

        // Simulate map load event
        const mapLoadEvent = createMapLoadEvent( "Tell", "Checkpoint", "2024-01-15T10:30:00Z" );
        TrackerService.processEvent( mapLoadEvent, serverId );

        // Verify match was created
        const matches = statements.getActiveMatches.all( serverId ) as any[];
        expect( matches ).toHaveLength( 1 );
        expect( matches[ 0 ].status ).toBe( "active" );
        expect( matches[ 0 ].server_id ).toBe( serverId );
    } );

    test( "should add players to matches when they join", () => {
        const statements = getStatements();

        // Create server and start match
        const server = statements.upsertServer.get(
            "test-server",
            "Test Server",
            "test-config",
            "/test/path",
            "Test Description"
        ) as any;
        const serverId = server.id;

        // Start a match via map load
        const mapLoadEvent = {
            type: "map_load" as const,
            timestamp: "2024-01-15T10:30:00Z",
            data: {
                mapName: "Tell",
                scenario: "Checkpoint",
            },
        } as MapLoadEvent;
        TrackerService.processEvent( mapLoadEvent, serverId );

        // Get the active match
        const activeMatch = TrackerService.getActiveMatch( serverId );
        expect( activeMatch ).toBeDefined();

        // Simulate player join
        const playerJoinEvent = {
            type: "player_join" as const,
            timestamp: "2024-01-15T10:31:00Z",
            data: {
                playerName: "TestPlayer1",
            },
        } as PlayerJoinEvent;

        TrackerService.processEvent( playerJoinEvent, serverId );

        // Verify player was added to match
        const participants = statements.getMatchParticipants.all( activeMatch!.matchId, serverId ) as { player_id: string; }[];
        const participantsRaw = statements.getMatchParticipants.all( activeMatch!.matchId, serverId );
        // Validate each participant with Valibot
        const validatedParticipants = participantsRaw.map( p => v.parse( MatchParticipantSchema, p ) );
        expect( validatedParticipants ).toHaveLength( 1 );
        expect( validatedParticipants[ 0 ]?.player_id ).toBeDefined();
    } );

    test( "should track multiple players in a match", () => {
        const statements = getStatements();
        // Create server and start match
        const server = statements.upsertServer.get(
            "test-server",
            "Test Server",
            "test-config",
            "/test/path",
            "Test Description"
        ) as any;
        const serverId = server.id;
        // Start match
        const mapLoadEvent = {
            type: "map_load" as const,
            timestamp: "2024-01-15T10:30:00Z",
            data: { mapName: "Tell", scenario: "Checkpoint" },
        } as MapLoadEvent;
        TrackerService.processEvent( mapLoadEvent, serverId );
        // Add multiple players
        const players = [ "Alice", "Bob", "Charlie" ];
        players.forEach( ( playerName, index ) => {
            TrackerService.processEvent( {
                type: "player_join",
                timestamp: `2024-01-15T10:3${ index + 1 }:00Z`,
                data: {
                    playerName,
                    steamId: `steam_${ playerName.toLowerCase() }_${ index }`
                },
                rawLine: `joined`
            }, serverId );
        } );
        const activeMatch = TrackerService.getActiveMatch( serverId );
        expect( activeMatch ).toBeDefined();
        expect( activeMatch!.participants.size ).toBe( 3 );
        // Validate participants
        const participantsRaw = statements.getMatchParticipants.all( activeMatch!.matchId, serverId );
        const validatedParticipants = participantsRaw.map( p => v.parse( MatchParticipantSchema, p ) );
        expect( validatedParticipants ).toHaveLength( 3 );
        validatedParticipants.forEach( ( p, i ) => {
            expect( p.player_name ).toBe( players[ i ] );
            expect( p.steam_id ).toBe( `steam_${ players[ i ].toLowerCase() }_${ i }` );
        } );
        // Validate match player count
        let matchRaw = statements.getMatchDetails.get( activeMatch!.matchId, serverId );
        if ( ( matchRaw as any ).maps_played === undefined ) ( matchRaw as any ).maps_played = null;
        const validatedMatch = v.parse( MatchHistorySchema, matchRaw );
        expect( validatedMatch.total_players ).toBe( 3 );
        expect( validatedMatch.max_players ).toBe( 3 );
    } );

    test( "should handle player leaving during match", () => {
        const statements = getStatements();

        // Setup server and match with players
        const server = statements.upsertServer.get(
            "test-server",
            "Test Server",
            "test-config",
            "/test/path",
            "Test Description"
        ) as any;
        const serverId = server.id;

        // Start match and add player
        const mapLoadEvent = {
            type: "map_load" as const,
            timestamp: "2024-01-15T10:30:00Z",
            data: { mapName: "Tell", scenario: "Checkpoint" },
        } as MapLoadEvent;
        TrackerService.processEvent( mapLoadEvent, serverId );

        const playerJoinEvent = {
            type: "player_join" as const,
            timestamp: "2024-01-15T10:31:00Z",
            data: { playerName: "TestPlayer" },
        } as PlayerJoinEvent;
        TrackerService.processEvent( playerJoinEvent, serverId );

        // Player leaves
        const playerLeaveEvent = {
            type: "player_leave" as const,
            timestamp: "2024-01-15T10:45:00Z",
            data: { playerName: "TestPlayer" },
        } as PlayerLeaveEvent;
        TrackerService.processEvent( playerLeaveEvent, serverId );

        // Verify player participation was ended
        const activeMatch = TrackerService.getActiveMatch( serverId );
        const participants = statements.getMatchParticipants.all( activeMatch!.matchId, serverId ) as MatchParticipant[];

        const participantsRaw3 = statements.getMatchParticipants.all( activeMatch!.matchId, serverId );
        const validatedParticipants3 = participantsRaw3.map( p => v.parse( MatchParticipantSchema, p ) );
        expect( validatedParticipants3 ).toHaveLength( 1 );
        expect( validatedParticipants3[ 0 ].leave_time ).toBeDefined();
        expect( validatedParticipants3[ 0 ].duration_minutes ).toBeGreaterThan( 0 );
        expect( activeMatch!.participants.size ).toBe( 0 );
    } );

    test( "should end match manually", () => {
        const statements = getStatements();
        // Setup server and match
        const server = statements.upsertServer.get(
            "test-server",
            "Test Server",
            "test-config",
            "/test/path",
            "Test Description"
        ) as any;
        const serverId = server.id;
        // Start match
        const mapLoadEvent = {
            type: "map_load" as const,
            timestamp: "2024-01-15T10:30:00Z",
            data: { mapName: "Tell", scenario: "Checkpoint" },
            rawLine: "LoadMap"
        };
        TrackerService.processEvent( mapLoadEvent, serverId );
        const activeMatch = TrackerService.getActiveMatch( serverId );
        expect( activeMatch ).toBeDefined();
        // End match with a later timestamp to ensure duration > 0 (15 minutes later)
        TrackerService.endMatch( serverId, "completed", "2024-01-15T10:45:00Z" );
        // Verify match was ended
        expect( TrackerService.getActiveMatch( serverId ) ).toBeUndefined();
        // Validate match
        let matchRaw = statements.getMatchDetails.get( activeMatch!.matchId, serverId );
        if ( ( matchRaw as any ).maps_played === undefined ) ( matchRaw as any ).maps_played = null;
        const validatedMatch = v.parse( MatchHistorySchema, matchRaw );
        expect( validatedMatch.status ).toBe( "completed" );
        expect( validatedMatch.end_time ).toBeDefined();
        expect( validatedMatch.duration_minutes ).toBeGreaterThan( 0 );
    } );

    test( "should get match history", () => {
        const statements = getStatements();

        // Create server
        const server = statements.upsertServer.get(
            "test-server",
            "Test Server",
            "test-config",
            "/test/path",
            "Test Description"
        ) as any;
        const serverId = server.id;

        // Create multiple matches
        for ( let i = 0; i < 3; i++ )
        {
            const startHour = 10 + i;
            const endHour = startHour + 1;
            const mapLoadEvent = {
                type: "map_load" as const,
                timestamp: `2024-01-15T${ startHour.toString().padStart( 2, '0' ) }:30:00Z`,
                data: { mapName: `Map${ i }`, scenario: "Checkpoint" },
                rawLine: `[2024.01.15-${ startHour.toString().padStart( 2, '0' ) }.30.00:000] LogGameMode: LoadMap: Map${ i }?Scenario=Scenario_Map${ i }_Checkpoint_Security`
            };
            TrackerService.processEvent( mapLoadEvent, serverId );
            // End match 1 hour later
            TrackerService.endMatch( serverId, "completed", `2024-01-15T${ endHour.toString().padStart( 2, '0' ) }:30:00Z` );
        }

        // Get match history
        const history = TrackerService.getMatchHistory( serverId, 10 ) as MatchHistory[];
        const historyRaw = TrackerService.getMatchHistory( serverId, 10 );
        const validatedHistory = historyRaw.map( h => v.parse( MatchHistorySchema, h ) );
        expect( validatedHistory ).toHaveLength( 3 );

        // Should be ordered by start time descending (most recent first)
        expect( new Date( validatedHistory[ 0 ].start_time ).getTime() ).toBeGreaterThan( new Date( validatedHistory[ 1 ].start_time ).getTime() );
    } );

    test( "should get detailed match information", () => {
        const statements = getStatements();

        // Setup server and match
        const server = statements.upsertServer.get(
            "test-server",
            "Test Server",
            "test-config",
            "/test/path",
            "Test Description"
        ) as any;
        const serverId = server.id;

        // Start match and add players
        const mapLoadEvent = {
            type: "map_load" as const,
            timestamp: "2024-01-15T10:30:00Z",
            data: { mapName: "Tell", scenario: "Checkpoint" },
            rawLine: "[2024.01.15-10.30.00:000] LogGameMode: LoadMap: Tell?Scenario=Scenario_Tell_Checkpoint_Security"
        };
        TrackerService.processEvent( mapLoadEvent, serverId );

        const activeMatch = TrackerService.getActiveMatch( serverId );

        // Add players
        [ "Alice", "Bob" ].forEach( ( name, index ) => {
            const joinEvent = {
                type: "player_join" as const,
                timestamp: `2024-01-15T10:3${ index + 1 }:00Z`,
                data: {
                    playerName: name,
                    steamId: `steam_${ name.toLowerCase() }_detail_${ index }`
                },
                rawLine: `[2024.01.15-10.3${ index + 1 }.00:000] LogOnlineSession: Player "${ name }" joined (EOS: steam_${ name.toLowerCase() }_detail_${ index })`
            };
            TrackerService.processEvent( joinEvent, serverId );
        } );

        // Get match details
        const details = TrackerService.getMatchDetails( activeMatch!.matchId, serverId ) as MatchDetails;
        const detailsRaw = TrackerService.getMatchDetails( activeMatch!.matchId, serverId );
        const validatedDetails = v.parse( MatchDetailsSchema, detailsRaw );

        expect( validatedDetails.match ).toBeDefined();
        expect( validatedDetails.participants ).toHaveLength( 2 );
        expect( validatedDetails.maps ).toHaveLength( 1 );
        expect( validatedDetails.maps[ 0 ].map_name ).toBe( "Tell" );
    } );

    test( "should handle server crash during match", () => {
        const statements = getStatements();
        // Setup server and match with players
        const server = statements.upsertServer.get(
            "test-server",
            "Test Server",
            "test-config",
            "/test/path",
            "Test Description"
        ) as any;
        const serverId = server.id;
        // Start match and add player
        const mapLoadEvent = {
            type: "map_load" as const,
            timestamp: "2024-01-15T10:30:00Z",
            data: { mapName: "Tell", scenario: "Checkpoint" },
            rawLine: "LoadMap"
        };
        TrackerService.processEvent( mapLoadEvent, serverId );
        TrackerService.processEvent( {
            type: "player_join",
            timestamp: "2024-01-15T10:31:00Z",
            data: {
                playerName: "TestPlayer",
                steamId: "steam_testplayer_crash"
            },
            rawLine: "joined"
        }, serverId );
        const activeMatch = TrackerService.getActiveMatch( serverId );
        expect( activeMatch ).toBeDefined();
        // Simulate server crash
        TrackerService.handleServerCrash( serverId );
        // Verify match was aborted
        expect( TrackerService.getActiveMatch( serverId ) ).toBeUndefined();
        // Validate match
        let matchRaw = statements.getMatchDetails.get( activeMatch!.matchId, serverId );
        if ( ( matchRaw as any ).maps_played === undefined ) ( matchRaw as any ).maps_played = null;
        const validatedMatch = v.parse( MatchHistorySchema, matchRaw );
        expect( validatedMatch.status ).toBe( "aborted" );
    } );

    test( "should track map changes within a match", () => {
        const statements = getStatements();

        // Setup server
        const server = statements.upsertServer.get(
            "test-server",
            "Test Server",
            "test-config",
            "/test/path",
            "Test Description"
        ) as any;
        const serverId = server.id;

        // Start match with first map
        const firstMapEvent = {
            type: "map_load" as const,
            timestamp: "2024-01-15T10:30:00Z",
            data: { mapName: "Tell", scenario: "Checkpoint" },
            rawLine: "[2024.01.15-10.30.00:000] LogGameMode: LoadMap: Tell?Scenario=Scenario_Tell_Checkpoint_Security"
        };
        TrackerService.processEvent( firstMapEvent, serverId );

        const activeMatch = TrackerService.getActiveMatch( serverId );
        const initialMatchId = activeMatch!.matchId;

        // Load second map (should add to same match)
        const secondMapEvent = {
            type: "map_load" as const,
            timestamp: "2024-01-15T11:00:00Z",
            data: { mapName: "Crossing", scenario: "Push" },
            rawLine: "[2024.01.15-11.00.00:000] LogGameMode: LoadMap: Crossing?Scenario=Scenario_Crossing_Push_Security"
        };
        TrackerService.processEvent( secondMapEvent, serverId );

        // Should still be the same active match
        const sameMatch = TrackerService.getActiveMatch( serverId );
        expect( sameMatch!.matchId ).toBe( initialMatchId );

        // Verify both maps are tracked for the match
        const details = TrackerService.getMatchDetails( initialMatchId, serverId ) as MatchDetails;
        const detailsRaw2 = TrackerService.getMatchDetails( initialMatchId, serverId );
        const validatedDetails2 = v.parse( MatchDetailsSchema, detailsRaw2 );
        expect( validatedDetails2.maps ).toHaveLength( 2 );
        expect( validatedDetails2.maps[ 0 ].sequence_order ).toBe( 1 );
        expect( validatedDetails2.maps[ 1 ].sequence_order ).toBe( 2 );
    } );
} );
