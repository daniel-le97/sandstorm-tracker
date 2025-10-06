import { EventEmitter } from 'events';
import { Socket } from 'net';
import type {
    RconConfig,
    RconResponse,
    QueuedCommand,
    ConnectionStats,
    RconPacket
} from './types.js';
import { RconClientState, RconPacketType } from './types.js';
import { RconProtocol } from './protocol.js';

/**
 * Insurgency: Sandstorm Source RCON Client
 * Full-featured client with reconnection, command queueing, and event handling
 */
export class RconClient extends EventEmitter {
    private config: Required<RconConfig>;
    private socket: Socket | null = null;
    private state: RconClientState = RconClientState.DISCONNECTED;
    private packetId: number = 1;
    /** Map keyed by primary command ID */
    private commandQueue = new Map<number, QueuedCommand>();
    private buffer = Buffer.alloc( 0 );
    private reconnectTimer: NodeJS.Timeout | null = null;
    private reconnectAttempts: number = 0;
    private stats: ConnectionStats;

    constructor ( config: RconConfig ) {
        super();

        this.config = {
            timeout: 5000,
            reconnectDelay: 3000,
            maxReconnectAttempts: 5,
            encoding: 'ascii',
            tcpTimeout: 30000,
            keepAlive: true,
            keepAliveInitialDelay: 0,
            ...config
        };

        this.stats = {
            connected: false,
            authenticated: false,
            state: RconClientState.DISCONNECTED,
            commandsSent: 0,
            responsesReceived: 0,
            reconnectAttempts: 0,
            uptime: 0
        };

        this.validateConfig();
    }

    /**
     * Connect to the RCON server
     */
    async connect (): Promise<void> {
        if ( this.state !== RconClientState.DISCONNECTED )
        {
            throw new Error( `Cannot connect: client is ${ this.state }` );
        }

        return new Promise( ( resolve, reject ) => {
            this.setState( RconClientState.CONNECTING );
            this.emit( 'debug', `Connecting to ${ this.config.host }:${ this.config.port }` );

            this.socket = new Socket();
            this.setupSocketEvents();

            const connectTimeout = setTimeout( () => {
                this.cleanup();
                reject( new Error( `Connection timeout after ${ this.config.timeout }ms` ) );
            }, this.config.timeout );

            this.socket.once( 'connect', () => {
                clearTimeout( connectTimeout );
                this.setState( RconClientState.CONNECTED );
                this.stats.connected = true;
                this.stats.connectTime = new Date();
                this.emit( 'connect' );
                this.authenticate().then( resolve ).catch( reject );
            } );

            this.socket.once( 'error', ( error ) => {
                clearTimeout( connectTimeout );
                reject( error );
            } );

            this.socket.connect( this.config.port, this.config.host );
        } );
    }

    /**
     * Disconnect from the server
     */
    async disconnect (): Promise<void> {
        if ( this.state === RconClientState.DISCONNECTED || this.state === RconClientState.DESTROYED )
        {
            return;
        }

        this.emit( 'debug', 'Disconnecting...' );
        this.cleanup();
        this.setState( RconClientState.DISCONNECTED );
        this.stats.connected = false;
        this.stats.authenticated = false;
        this.emit( 'disconnect' );
    }

    /**
     * Execute a command on the server
     */
    async execute ( command: string ): Promise<RconResponse> {
        if ( this.state !== RconClientState.AUTHENTICATED )
        {
            throw new Error( `Cannot execute command: client is ${ this.state }` );
        }

        if ( !command.trim() )
        {
            throw new Error( 'Command cannot be empty' );
        }

        return new Promise( ( resolve, reject ) => {
            const id = this.getNextPacketId();
            // Terminator packet uses a different ID to act as sentinel for multi-packet response.
            const terminatorId = this.getNextPacketId();
            const packet = RconProtocol.createCommandPacket( id, command );
            const terminatorPacket = RconProtocol.createTerminatorPacket( terminatorId );

            const timeout = setTimeout( () => {
                this.commandQueue.delete( id );
                reject( new Error( `Command timeout: ${ command }` ) );
            }, this.config.timeout );

            const queuedCommand: QueuedCommand = {
                id,
                terminatorId,
                command,
                resolve,
                reject,
                timestamp: Date.now(),
                timeout,
                parts: []
            };

            this.commandQueue.set( id, queuedCommand );
            // Send main command followed by sentinel terminator request
            this.sendPacket( packet );
            this.sendPacket( terminatorPacket );
            this.stats.commandsSent++;
            this.stats.lastCommand = new Date();

            this.emit( 'debug', `Executing command: ${ command } (ID: ${ id })` );
        } );
    }

    /**
     * Get connection statistics
     */
    getStats (): ConnectionStats {
        return {
            ...this.stats,
            uptime: this.stats.connectTime ? Date.now() - this.stats.connectTime.getTime() : 0
        };
    }

    /**
     * Get current client state
     */
    getState (): RconClientState {
        return this.state;
    }

    /**
     * Check if client is connected and authenticated
     */
    isReady (): boolean {
        return this.state === RconClientState.AUTHENTICATED;
    }

    /**
     * Destroy the client and clean up resources
     */
    destroy (): void {
        this.emit( 'debug', 'Destroying client...' );
        this.cleanup();
        this.setState( RconClientState.DESTROYED );
        this.removeAllListeners();
    }

    /**
     * Enable automatic reconnection
     */
    enableAutoReconnect (): void {
        this.on( 'disconnect', ( error ) => {
            if ( this.state !== RconClientState.DESTROYED && error )
            {
                this.startReconnection();
            }
        } );
    }

    private validateConfig (): void {
        if ( !this.config.host )
        {
            throw new Error( 'Host is required' );
        }
        if ( !this.config.port || this.config.port < 1 || this.config.port > 65535 )
        {
            throw new Error( 'Valid port is required' );
        }
        if ( !this.config.password )
        {
            throw new Error( 'Password is required' );
        }
    }

    private setState ( newState: RconClientState ): void {
        const oldState = this.state;
        this.state = newState;
        this.stats.state = newState;
        this.emit( 'debug', `State changed: ${ oldState } -> ${ newState }` );
    }

    private setupSocketEvents (): void {
        if ( !this.socket ) return;

        this.socket.setTimeout( this.config.tcpTimeout );

        if ( this.config.keepAlive )
        {
            this.socket.setKeepAlive( true, this.config.keepAliveInitialDelay );
        }

        this.socket.on( 'data', this.handleData.bind( this ) );
        this.socket.on( 'close', this.handleClose.bind( this ) );
        this.socket.on( 'error', this.handleError.bind( this ) );
        this.socket.on( 'timeout', this.handleTimeout.bind( this ) );
    }

    private async authenticate (): Promise<void> {
        this.setState( RconClientState.AUTHENTICATING );

        return new Promise( ( resolve, reject ) => {
            const id = this.getNextPacketId();
            const packet = RconProtocol.createAuthPacket( id, this.config.password );

            const timeout = setTimeout( () => {
                reject( new Error( 'Authentication timeout' ) );
            }, this.config.timeout );

            const onResponse = ( response: RconPacket ) => {
                clearTimeout( timeout );
                this.removeListener( 'authResponse', onResponse );

                if ( RconProtocol.isAuthSuccess( packet, response ) )
                {
                    this.setState( RconClientState.AUTHENTICATED );
                    this.stats.authenticated = true;
                    this.reconnectAttempts = 0;
                    this.emit( 'authenticated' );
                    resolve();
                } else
                {
                    const error = new Error( 'Authentication failed' );
                    this.emit( 'authFailed', error );
                    reject( error );
                }
            };

            this.on( 'authResponse', onResponse );
            this.sendPacket( packet );
            this.emit( 'debug', 'Authenticating...' );
        } );
    }

    private handleData ( data: Buffer ): void {
        this.buffer = Buffer.concat( [ this.buffer, data ] );
        try
        {
            const packets = RconProtocol.decode( this.buffer );
            let consumed = 0;
            for ( const packet of packets )
            {
                // Size consumed: size field (4) + declared size
                const declaredSize = Buffer.byteLength( packet.body, 'utf8' ) + 14; // 4 id +4 type + body + 2 nulls + 4 size field
                consumed += declaredSize;
                this.handlePacket( packet );
            }
            if ( consumed > 0 )
            {
                this.buffer = this.buffer.subarray( consumed );
            }
        } catch ( error )
        {
            this.emit( 'error', error as Error );
        }
    }

    private handlePacket ( packet: RconPacket ): void {
        this.emit( 'debug', `Received packet: ID=${ packet.id }, Type=${ packet.type }, Body="${ packet.body }"` );

        if ( RconProtocol.isAuthResponse( packet ) )
        {
            // Detect failed auth (ID -1) earlier and surface error
            if ( packet.id === -1 )
            {
                this.emit( 'authResponse', packet );
                return;
            }
            this.emit( 'authResponse', packet );
            return;
        }

        if ( RconProtocol.isCommandResponse( packet ) )
        {
            // Match either primary id or terminator id to locate queued command
            let queuedCommand = this.commandQueue.get( packet.id );
            if ( !queuedCommand )
            {
                // Search by terminator id
                for ( const qc of this.commandQueue.values() )
                {
                    if ( qc.terminatorId === packet.id )
                    {
                        queuedCommand = qc;
                        break;
                    }
                }
            }

            if ( queuedCommand )
            {
                if ( packet.id === queuedCommand.id )
                {
                    // Primary response part
                    if ( packet.body ) queuedCommand.parts.push( packet.body );
                } else if ( packet.id === queuedCommand.terminatorId )
                {
                    // Terminator: finalize
                    clearTimeout( queuedCommand.timeout );
                    this.commandQueue.delete( queuedCommand.id );
                    const fullBody = queuedCommand.parts.join( '' );
                    const response: RconResponse = {
                        id: queuedCommand.id,
                        body: fullBody,
                        timestamp: new Date()
                    };
                    this.stats.responsesReceived++;
                    this.emit( 'response', response );
                    this.emit( 'command', queuedCommand.command, response );
                    queuedCommand.resolve( response );
                } else
                {
                    // Unexpected extra packet relating to the command
                    if ( packet.body ) queuedCommand.parts.push( packet.body );
                }
            }
        }
    }

    private handleClose (): void {
        this.emit( 'debug', 'Socket closed' );
        const wasConnected = this.stats.connected;
        this.stats.connected = false;
        this.stats.authenticated = false;

        if ( wasConnected && this.state !== RconClientState.DESTROYED )
        {
            this.setState( RconClientState.DISCONNECTED );
            this.emit( 'disconnect', new Error( 'Connection closed' ) );
        }
    }

    private handleError ( error: Error ): void {
        this.emit( 'debug', `Socket error: ${ error.message }` );
        this.emit( 'error', error );
    }

    private handleTimeout (): void {
        this.emit( 'debug', 'Socket timeout' );
        this.handleError( new Error( 'Socket timeout' ) );
    }

    private sendPacket ( packet: RconPacket ): void {
        if ( !this.socket || this.socket.destroyed )
        {
            throw new Error( 'Socket is not connected' );
        }

        try
        {
            const buffer = RconProtocol.encode( packet );
            this.socket.write( buffer );
            this.emit( 'debug', `Sent packet: ID=${ packet.id }, Type=${ packet.type }, Size=${ buffer.length }` );
        } catch ( error )
        {
            this.emit( 'error', error as Error );
        }
    }

    private getNextPacketId (): number {
        this.packetId = ( this.packetId % 0x7FFFFFFF ) + 1;
        return this.packetId;
    }

    private cleanup (): void {
        // Clear reconnection timer
        if ( this.reconnectTimer )
        {
            clearTimeout( this.reconnectTimer );
            this.reconnectTimer = null;
        }

        // Clear command queue
        this.commandQueue.forEach( cmd => {
            clearTimeout( cmd.timeout );
            cmd.reject( new Error( 'Connection closed' ) );
        } );
        this.commandQueue.clear();

        // Close socket
        if ( this.socket )
        {
            this.socket.removeAllListeners();
            if ( !this.socket.destroyed )
            {
                this.socket.destroy();
            }
            this.socket = null;
        }

        // Reset buffer
        this.buffer = Buffer.alloc( 0 );
    }

    private startReconnection (): void {
        if ( this.state === RconClientState.DESTROYED ||
            this.reconnectAttempts >= this.config.maxReconnectAttempts )
        {
            this.emit( 'reconnectFailed' );
            return;
        }

        this.setState( RconClientState.RECONNECTING );
        this.reconnectAttempts++;
        this.stats.reconnectAttempts++;

        this.emit( 'reconnecting', this.reconnectAttempts, this.config.maxReconnectAttempts );

        this.reconnectTimer = setTimeout( async () => {
            try
            {
                await this.connect();
                this.emit( 'reconnected' );
                this.emit( 'debug', `Reconnected after ${ this.reconnectAttempts } attempts` );
            } catch ( error )
            {
                this.emit( 'debug', `Reconnection attempt ${ this.reconnectAttempts } failed: ${ ( error as Error ).message }` );
                this.startReconnection();
            }
        }, this.config.reconnectDelay );
    }
}
