import { RconClient } from "./client";
import { SandstormCommands } from "./sandstorm-commands";
export { RconClient } from './client';
export { SandstormCommands } from './sandstorm-commands';

// Utility to check if running in different environments
export const isNode = typeof process !== 'undefined' && process.versions?.node;
export const isBun = typeof Bun !== 'undefined';

/**
 * Get environment variable with fallback
 * @param key - Environment variable name
 * @param defaultValue - Default value if not set
 * @returns Environment variable value or default
 */
export function getEnv ( key: string, defaultValue?: string ): string | undefined {
    return process.env[ key ] ?? defaultValue;
}

/**
 * Get required environment variable (throws if not set)
 * @param key - Environment variable name
 * @returns Environment variable value
 * @throws Error if environment variable is not set
 */
export function getRequiredEnv ( key: string ): string {
    const value = process.env[ key ];
    if ( value === undefined )
    {
        throw new Error( `Required environment variable ${ key } is not set` );
    }
    return value;
}

const CONFIG = {
    host: getEnv( 'RCON_HOST', '127.0.0.1' )!,
    port: parseInt( getEnv( 'RCON_PORT', '27015' )! ),
    password: getEnv( 'RCON_PASSWORD', 'default-password' )!,
    logPath: getEnv( 'LOG_PATH', './default.log' )!,
    timeout: 5000,
    reconnectDelay: 3000,
    maxReconnectAttempts: 2
};

// Lazy RCON execution - connect, execute, disconnect
export async function executeSayCommand ( message: string, description: string ): Promise<void> {
    try
    {
        const client = new RconClient( CONFIG );
        await client.connect();
        await client.execute( SandstormCommands.say( message ) );
        await client.disconnect();
        console.log( description );
    } catch ( error )
    {
        console.log( `❌ RCON failed (${ description }): ${ ( error as Error ).message }` );
    }
}

/**
 * Execute an arbitrary RCON command (multi-packet safe) using a short-lived client
 */
export async function executeCommandOnce ( command: string ): Promise<string | undefined> {
    const client = new RconClient( CONFIG );
    try
    {
        await client.connect();
        const res = await client.execute( command );
        return res.body;
    } catch ( e )
    {
        console.error( 'RCON command failed', ( e as Error ).message );
        return undefined;
    } finally
    {
        await client.disconnect().catch( () => { } );
    }
}
