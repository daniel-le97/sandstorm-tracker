import * as v from 'valibot';

export const MatchParticipantSchema = v.object( {
    id: v.number(),
    match_id: v.number(),
    player_id: v.number(),
    server_id: v.number(),
    join_time: v.string(),
    leave_time: v.nullable( v.string() ),
    duration_minutes: v.nullable( v.number() ),
    final_kills: v.number(),
    final_deaths: v.number(),
    final_team_kills: v.number(),
    final_suicides: v.number(),
    final_score: v.number(),
    created_at: v.string(),
    player_name: v.string(),
    steam_id: v.string(),
} );
export type MatchParticipant = v.InferOutput<typeof MatchParticipantSchema>;

export const MatchHistorySchema = v.object( {
    id: v.number(),
    server_id: v.number(),
    match_name: v.nullable( v.string() ),
    start_time: v.string(),
    end_time: v.nullable( v.string() ),
    duration_minutes: v.nullable( v.number() ),
    status: v.string(),
    total_players: v.number(),
    max_players: v.number(),
    created_at: v.string(),
    updated_at: v.string(),
    participant_count: v.number(),
    maps_played: v.nullable( v.string() ),
} );
export type MatchHistory = v.InferOutput<typeof MatchHistorySchema>;

export const MatchSchema = v.object( {
    id: v.number(),
    server_id: v.number(),
    match_name: v.nullable( v.string() ),
    start_time: v.string(),
    end_time: v.nullable( v.string() ),
    duration_minutes: v.nullable( v.number() ),
    status: v.union( [
        v.literal( 'active' ),
        v.literal( 'completed' ),
        v.literal( 'aborted' ),
    ] ),
    total_players: v.number(),
    max_players: v.number(),
    created_at: v.string(),
    updated_at: v.string(),
} );
export type Match = v.InferOutput<typeof MatchSchema>;

export const MatchMapSchema = v.object( {
    id: v.number(),
    match_id: v.number(),
    map_id: v.number(),
    server_id: v.number(),
    sequence_order: v.number(),
    start_time: v.nullable( v.string() ),
    end_time: v.nullable( v.string() ),
    created_at: v.string(),
    map_name: v.string(),
    scenario: v.nullable( v.string() ),
} );
export type MatchMap = v.InferOutput<typeof MatchMapSchema>;

export const MatchDetailsSchema = v.object( {
    match: MatchSchema,
    participants: v.array( MatchParticipantSchema ),
    maps: v.array( MatchMapSchema ),
} );
export type MatchDetails = v.InferOutput<typeof MatchDetailsSchema>;
