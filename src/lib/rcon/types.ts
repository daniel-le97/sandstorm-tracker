/**
 * RCON Protocol Types and Interfaces
 */

export enum RconPacketType {
    SERVERDATA_AUTH = 3,
    SERVERDATA_AUTH_RESPONSE = 2,
    SERVERDATA_EXECCOMMAND = 2,
    SERVERDATA_RESPONSE_VALUE = 0
}

export interface RconPacket {
    id: number;
    type: RconPacketType;
    body: string;
}

export interface RconResponse {
    id: number;
    body: string;
    timestamp: Date;
}

export interface RconConfig {
    host: string;
    port: number;
    password: string;
    timeout?: number;
    reconnectDelay?: number;
    maxReconnectAttempts?: number;
    encoding?: BufferEncoding;
    tcpTimeout?: number;
    keepAlive?: boolean;
    keepAliveInitialDelay?: number;
}

export interface RconClientEvents {
    connect: () => void;
    disconnect: ( error?: Error ) => void;
    authenticated: () => void;
    authFailed: ( error: Error ) => void;
    response: ( response: RconResponse ) => void;
    command: ( command: string, response: RconResponse ) => void;
    error: ( error: Error ) => void;
    reconnecting: ( attempt: number, maxAttempts: number ) => void;
    reconnected: () => void;
    reconnectFailed: () => void;
    debug: ( message: string ) => void;
}

export interface QueuedCommand {
    /** Primary command packet ID */
    id: number;
    /** Terminator packet ID used to detect end of multi-packet response */
    terminatorId: number;
    command: string;
    resolve: ( response: RconResponse ) => void;
    reject: ( error: Error ) => void;
    timestamp: number;
    timeout?: NodeJS.Timeout;
    /** Collected response body parts for multi-packet responses */
    parts: string[];
}

export enum RconClientState {
    DISCONNECTED = 'DISCONNECTED',
    CONNECTING = 'CONNECTING',
    CONNECTED = 'CONNECTED',
    AUTHENTICATING = 'AUTHENTICATING',
    AUTHENTICATED = 'AUTHENTICATED',
    RECONNECTING = 'RECONNECTING',
    DESTROYED = 'DESTROYED'
}

export interface ConnectionStats {
    connected: boolean;
    authenticated: boolean;
    state: RconClientState;
    connectTime?: Date;
    lastCommand?: Date;
    commandsSent: number;
    responsesReceived: number;
    reconnectAttempts: number;
    uptime: number;
}
