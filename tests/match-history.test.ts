import { beforeEach, describe, expect, test } from "bun:test";
import { getStatements, initializeDatabase, type MatchDetails, type MatchHistory, type MatchParticipant } from "../src/database";
import type { MapLoadEvent, PlayerJoinEvent, PlayerLeaveEvent } from "../src/events";
import { StatsService } from "../src/stats-service";
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
        StatsService.clearAllState();

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
        StatsService.processEvent( mapLoadEvent, serverId );

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
        StatsService.processEvent( mapLoadEvent, serverId );

        // Get the active match
        const activeMatch = StatsService.getActiveMatch( serverId );
        expect( activeMatch ).toBeDefined();

        // Simulate player join
        const playerJoinEvent = {
            type: "player_join" as const,
            timestamp: "2024-01-15T10:31:00Z",
            data: {
                playerName: "TestPlayer1",
            },
        } as PlayerJoinEvent;

        StatsService.processEvent( playerJoinEvent, serverId );

        // Verify player was added to match
        const participants = statements.getMatchParticipants.all( activeMatch!.matchId, serverId ) as { player_id: string }[];
        expect( participants ).toHaveLength( 1 );
        expect( participants[ 0 ]?.player_id ).toBeDefined();
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
            data: {
                mapName: "Tell",
                scenario: "Checkpoint",
            },
        } as MapLoadEvent;
        StatsService.processEvent( mapLoadEvent, serverId );

        // Add multiple players with unique steam IDs
        const players = [ "Alice", "Bob", "Charlie" ];
        players.forEach( ( playerName, index ) => {
            const playerJoinEvent = {
                type: "player_join" as const,
                timestamp: `2024-01-15T10:3${ index + 1 }:00Z`,
                data: {
                    playerName,
                    steamId: `steam_${ playerName.toLowerCase() }_${ index }`
                },
                rawLine: `[2024.01.15-10.3${ index + 1 }.00:000] LogOnlineSession: Player "${ playerName }" joined (EOS: steam_${ playerName.toLowerCase() }_${ index })`
            };
            StatsService.processEvent( playerJoinEvent, serverId );
        } );

        const activeMatch = StatsService.getActiveMatch( serverId );
        const participants = statements.getMatchParticipants.all( activeMatch!.matchId, serverId );

        expect( participants ).toHaveLength( 3 );
        expect( activeMatch!.participants.size ).toBe( 3 );

        // Verify match player count was updated
        const match = statements.getMatchDetails.get( activeMatch!.matchId, serverId ) as any;
        expect( match.total_players ).toBe( 3 );
        expect( match.max_players ).toBe( 3 );
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
        StatsService.processEvent( mapLoadEvent, serverId );

        const playerJoinEvent = {
            type: "player_join" as const,
            timestamp: "2024-01-15T10:31:00Z",
            data: { playerName: "TestPlayer" },
        } as PlayerJoinEvent;
        StatsService.processEvent( playerJoinEvent, serverId );

        // Player leaves
        const playerLeaveEvent = {
            type: "player_leave" as const,
            timestamp: "2024-01-15T10:45:00Z",
            data: { playerName: "TestPlayer" },
        } as PlayerLeaveEvent;
        StatsService.processEvent( playerLeaveEvent, serverId );

        // Verify player participation was ended
        const activeMatch = StatsService.getActiveMatch( serverId );
        const participants = statements.getMatchParticipants.all( activeMatch!.matchId, serverId ) as MatchParticipant[];

        expect( participants ).toHaveLength( 1 );
        expect( participants[ 0 ].leave_time ).toBeDefined();
        expect( participants[ 0 ].duration_minutes ).toBeGreaterThan( 0 );
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
            rawLine: "[2024.01.15-10.30.00:000] LogGameMode: LoadMap: Tell?Scenario=Scenario_Tell_Checkpoint_Security"
        };
        StatsService.processEvent( mapLoadEvent, serverId );

        const activeMatch = StatsService.getActiveMatch( serverId );
        expect( activeMatch ).toBeDefined();

        // End match with a later timestamp to ensure duration > 0 (15 minutes later)
        StatsService.endMatch( serverId, "completed", "2024-01-15T10:45:00Z" );

        // Verify match was ended
        const endedActiveMatch = StatsService.getActiveMatch( serverId );
        expect( endedActiveMatch ).toBeUndefined();

        const match = statements.getMatchDetails.get( activeMatch!.matchId, serverId ) as any;
        expect( match.status ).toBe( "completed" );
        expect( match.end_time ).toBeDefined();
        expect( match.duration_minutes ).toBeGreaterThan( 0 );
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
            StatsService.processEvent( mapLoadEvent, serverId );
            // End match 1 hour later
            StatsService.endMatch( serverId, "completed", `2024-01-15T${ endHour.toString().padStart( 2, '0' ) }:30:00Z` );
        }

        // Get match history
        const history = StatsService.getMatchHistory( serverId, 10 ) as MatchHistory[];
        expect( history ).toHaveLength( 3 );

        // Should be ordered by start time descending (most recent first)
        expect( new Date( history[ 0 ].start_time ).getTime() ).toBeGreaterThan( new Date( history[ 1 ].start_time ).getTime() );
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
        StatsService.processEvent( mapLoadEvent, serverId );

        const activeMatch = StatsService.getActiveMatch( serverId );

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
            StatsService.processEvent( joinEvent, serverId );
        } );

        // Get match details
        const details = StatsService.getMatchDetails( activeMatch!.matchId, serverId ) as MatchDetails;

        expect( details.match ).toBeDefined();
        expect( details.participants ).toHaveLength( 2 );
        expect( details.maps ).toHaveLength( 1 );
        expect( details.maps[ 0 ].map_name ).toBe( "Tell" );
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
            rawLine: "[2024.01.15-10.30.00:000] LogGameMode: LoadMap: Tell?Scenario=Scenario_Tell_Checkpoint_Security"
        };
        StatsService.processEvent( mapLoadEvent, serverId );

        const playerJoinEvent = {
            type: "player_join" as const,
            timestamp: "2024-01-15T10:31:00Z",
            data: {
                playerName: "TestPlayer",
                steamId: "steam_testplayer_crash"
            },
            rawLine: "[2024.01.15-10.31.00:000] LogOnlineSession: Player \"TestPlayer\" joined (EOS: steam_testplayer_crash)"
        };
        StatsService.processEvent( playerJoinEvent, serverId );

        const activeMatch = StatsService.getActiveMatch( serverId );
        expect( activeMatch ).toBeDefined();

        // Simulate server crash
        StatsService.handleServerCrash( serverId );

        // Verify match was aborted
        const crashedActiveMatch = StatsService.getActiveMatch( serverId );
        expect( crashedActiveMatch ).toBeUndefined();

        const match = statements.getMatchDetails.get( activeMatch!.matchId, serverId ) as any;
        expect( match.status ).toBe( "aborted" );
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
        StatsService.processEvent( firstMapEvent, serverId );

        const activeMatch = StatsService.getActiveMatch( serverId );
        const initialMatchId = activeMatch!.matchId;

        // Load second map (should add to same match)
        const secondMapEvent = {
            type: "map_load" as const,
            timestamp: "2024-01-15T11:00:00Z",
            data: { mapName: "Crossing", scenario: "Push" },
            rawLine: "[2024.01.15-11.00.00:000] LogGameMode: LoadMap: Crossing?Scenario=Scenario_Crossing_Push_Security"
        };
        StatsService.processEvent( secondMapEvent, serverId );

        // Should still be the same active match
        const sameMatch = StatsService.getActiveMatch( serverId );
        expect( sameMatch!.matchId ).toBe( initialMatchId );

        // Verify both maps are tracked for the match
        const details = StatsService.getMatchDetails( initialMatchId, serverId ) as MatchDetails;
        expect( details.maps ).toHaveLength( 2 );
        expect( details.maps[ 0 ].sequence_order ).toBe( 1 );
        expect( details.maps[ 1 ].sequence_order ).toBe( 2 );
    } );
} );
