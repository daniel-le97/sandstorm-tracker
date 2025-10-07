import { EventEmitter } from "events";
import type {
    GameEvent,
    PlayerKillEvent,
    PlayerJoinEvent,
    PlayerLeaveEvent,
    ChatEvent,
    GameOverEvent,
    RoundOverEvent,
    MapLoadEvent,
    DifficultyEvent,
} from "../../events";

/**
 * Map of event names to their payload types.
 * Add new mappings here when you add more typed events.
 */
export interface LogEventMap {
    player_kill: PlayerKillEvent;
    team_kill: PlayerKillEvent;
    suicide: PlayerKillEvent;
    player_join: PlayerJoinEvent;
    player_leave: PlayerLeaveEvent;
    player_disconnect: PlayerLeaveEvent;
    chat_command: ChatEvent;
    game_over: GameOverEvent;
    round_over: RoundOverEvent;
    map_load: MapLoadEvent;
    difficulty_set: DifficultyEvent;

    // Fallbacks for less-structured events parsed as GameEvent
    fall_damage: GameEvent;
    map_vote_start: GameEvent;
}

/**
 * A strongly-typed wrapper around node's EventEmitter.
 * Usage:
 *   import { eventBus } from './lib/emitter/emitter';
 *   // Example: eventBus.on('player_kill', (evt) => { /* evt is PlayerKillEvent */

export class TypedEventEmitter<T extends Record<string, any>> extends EventEmitter {
    // Provide typed overloads that remain compatible with EventEmitter's API.

    public override on<E extends keyof T & ( string | symbol )> ( event: E, listener: ( payload: T[ E ] ) => void ): this {
        return super.on( event as string, listener as ( ...args: any[] ) => void );
    }

    public override once<E extends keyof T & ( string | symbol )> ( event: E, listener: ( payload: T[ E ] ) => void ): this {
        return super.once( event as string, listener as ( ...args: any[] ) => void );
    }

    public override off<E extends keyof T & ( string | symbol )> ( event: E, listener: ( payload: T[ E ] ) => void ): this {
        return super.off( event as string, listener as ( ...args: any[] ) => void );
    }

    public override emit<E extends keyof T & ( string | symbol )> ( event: E, payload: T[ E ] ): boolean {
        return super.emit( event as string, payload as any );
    }

    public override addListener<E extends keyof T & ( string | symbol )> ( event: E, listener: ( payload: T[ E ] ) => void ): this {
        return super.addListener( event as string, listener as ( ...args: any[] ) => void );
    }

    public override removeListener<E extends keyof T & ( string | symbol )> ( event: E, listener: ( payload: T[ E ] ) => void ): this {
        return super.removeListener( event as string, listener as ( ...args: any[] ) => void );
    }
}

// Export a project-wide event bus typed to the log events
export const gameEventEmitter = new TypedEventEmitter<LogEventMap>();

export default gameEventEmitter;

