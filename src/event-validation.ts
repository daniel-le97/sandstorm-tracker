import * as v from 'valibot';

// Player kill / team kill / suicide
export const PlayerKillEventSchema = v.object( {
    type: v.union( [ v.literal( 'player_kill' ), v.literal( 'team_kill' ), v.literal( 'suicide' ) ] ),
    timestamp: v.string(),
    data: v.object( {
        killer: v.string(),
        killerSteamId: v.string(),
        killerTeam: v.union( [ v.number(), v.string() ] ),
        victim: v.string(),
        victimSteamId: v.string(),
        victimTeam: v.union( [ v.number(), v.string() ] ),
        weapon: v.string(),
    } ),
    rawLine: v.string(),
} );

export type PlayerKillEvent = v.InferOutput<typeof PlayerKillEventSchema>;

// Player join
export const PlayerJoinEventSchema = v.object( {
    type: v.literal( 'player_join' ),
    timestamp: v.string(),
    data: v.object( {
        playerName: v.string(),
        steamId: v.optional( v.string() ),
    } ),
    rawLine: v.string(),
} );
export type PlayerJoinEvent = v.InferOutput<typeof PlayerJoinEventSchema>;

// Player leave / disconnect
export const PlayerLeaveEventSchema = v.object( {
    type: v.union( [ v.literal( 'player_leave' ), v.literal( 'player_disconnect' ) ] ),
    timestamp: v.string(),
    data: v.object( {
        playerName: v.optional( v.string() ),
        steamId: v.optional( v.string() ),
    } ),
    rawLine: v.string(),
} );
export type PlayerLeaveEvent = v.InferOutput<typeof PlayerLeaveEventSchema>;

// Chat command
export const ChatCommandEventSchema = v.object( {
    type: v.literal( 'chat_command' ),
    timestamp: v.string(),
    data: v.object( {
        playerName: v.string(),
        steamId: v.string(),
        command: v.string(),
        args: v.optional( v.array( v.string() ) ),
    } ),
    rawLine: v.string(),
} );
export type ChatCommandEvent = v.InferOutput<typeof ChatCommandEventSchema>;

// Game over
export const GameOverEventSchema = v.object( {
    type: v.literal( 'game_over' ),
    timestamp: v.string(),
    data: v.object( {} ),
    rawLine: v.string(),
} );
export type GameOverEvent = v.InferOutput<typeof GameOverEventSchema>;

// Round over
export const RoundOverEventSchema = v.object( {
    type: v.literal( 'round_over' ),
    timestamp: v.string(),
    data: v.object( {
        roundNumber: v.number(),
        winningTeam: v.number(),
        winReason: v.string(),
    } ),
    rawLine: v.string(),
} );
export type RoundOverEvent = v.InferOutput<typeof RoundOverEventSchema>;

// Map load
export const MapLoadEventSchema = v.object( {
    type: v.literal( 'map_load' ),
    timestamp: v.string(),
    data: v.object( {
        mapName: v.string(),
        scenario: v.string(),
        team: v.optional( v.string() ),
        maxPlayers: v.optional( v.number() ),
        lighting: v.optional( v.string() ),
    } ),
    rawLine: v.string(),
} );
export type MapLoadEvent = v.InferOutput<typeof MapLoadEventSchema>;

// Difficulty set
export const DifficultyEventSchema = v.object( {
    type: v.literal( 'difficulty_set' ),
    timestamp: v.string(),
    data: v.object( {
        difficulty: v.number(),
    } ),
    rawLine: v.string(),
} );
export type DifficultyEvent = v.InferOutput<typeof DifficultyEventSchema>;

// Fall damage
export const FallDamageEventSchema = v.object( {
    type: v.literal( 'fall_damage' ),
    timestamp: v.string(),
    data: v.object( {
        damage: v.number(),
    } ),
    rawLine: v.string(),
} );
export type FallDamageEvent = v.InferOutput<typeof FallDamageEventSchema>;

// Map vote start
export const MapVoteStartEventSchema = v.object( {
    type: v.literal( 'map_vote_start' ),
    timestamp: v.string(),
    data: v.object( {} ),
    rawLine: v.string(),
} );
export type MapVoteStartEvent = v.InferOutput<typeof MapVoteStartEventSchema>;

// Union of all event types
export const EventSchema = v.union( [
    PlayerKillEventSchema,
    PlayerJoinEventSchema,
    PlayerLeaveEventSchema,
    ChatCommandEventSchema,
    GameOverEventSchema,
    RoundOverEventSchema,
    MapLoadEventSchema,
    DifficultyEventSchema,
    FallDamageEventSchema,
    MapVoteStartEventSchema,
] );
export type GameEvent = v.InferOutput<typeof EventSchema>;

/**
 * Validate a parsed event object against the Valibot schema.
 * Throws a ValiError if validation fails.
 */
export function validateEvent ( event: unknown ): GameEvent {
    return v.parse( EventSchema, event );
}

export default {
    PlayerKillEventSchema,
    PlayerJoinEventSchema,
    PlayerLeaveEventSchema,
    ChatCommandEventSchema,
    GameOverEventSchema,
    RoundOverEventSchema,
    MapLoadEventSchema,
    DifficultyEventSchema,
    FallDamageEventSchema,
    MapVoteStartEventSchema,
    EventSchema,
    validateEvent,
};
