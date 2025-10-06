import { Buffer } from 'buffer';
import type { RconPacket } from './types.js';
import { RconPacketType } from './types.js';

/**
 * RCON Protocol Implementation
 * Handles packet encoding/decoding for Source RCON protocol
 */
export class RconProtocol {
    private static readonly PACKET_PADDING = 10; // 4 bytes size + 4 bytes id + 4 bytes type + 2 null terminators
    private static readonly MAX_PACKET_SIZE = 4096;
    private static readonly MIN_PACKET_SIZE = 14;

    /**
     * Encode an RCON packet to buffer
     */
    static encode ( packet: RconPacket ): Buffer {
        // Use utf8 to support broader character set (was ascii)
        const bodyBuffer = Buffer.from( packet.body, 'utf8' );
        // Total size = 4 bytes size + 4 bytes id + 4 bytes type + body + 2 null terminators
        const totalSize = 4 + 4 + 4 + bodyBuffer.length + 2;

        if ( totalSize > this.MAX_PACKET_SIZE )
        {
            throw new Error( `Packet too large: ${ totalSize } bytes (max: ${ this.MAX_PACKET_SIZE })` );
        }

        const buffer = Buffer.allocUnsafe( totalSize );
        let offset = 0;

        // Write size (excluding size field itself)
        buffer.writeInt32LE( totalSize - 4, offset );
        offset += 4;

        // Write packet ID
        buffer.writeInt32LE( packet.id, offset );
        offset += 4;

        // Write packet type
        buffer.writeInt32LE( packet.type, offset );
        offset += 4;

        // Write body
        bodyBuffer.copy( buffer, offset );
        offset += bodyBuffer.length;

        // Write null terminators
        buffer.writeUInt8( 0, offset );
        buffer.writeUInt8( 0, offset + 1 );

        return buffer;
    }

    /**
     * Decode RCON packets from buffer
     */
    static decode ( buffer: Buffer ): RconPacket[] {
        const packets: RconPacket[] = [];
        let offset = 0;

        while ( offset < buffer.length )
        {
            if ( buffer.length - offset < 4 )
            {
                // Not enough data for size field
                break;
            }

            const size = buffer.readInt32LE( offset );
            const totalSize = size + 4; // Include size field

            if ( totalSize < this.MIN_PACKET_SIZE )
            {
                throw new Error( `Invalid packet size: ${ totalSize } (min: ${ this.MIN_PACKET_SIZE })` );
            }

            if ( totalSize > this.MAX_PACKET_SIZE )
            {
                throw new Error( `Packet too large: ${ totalSize } bytes (max: ${ this.MAX_PACKET_SIZE })` );
            }

            if ( buffer.length - offset < totalSize )
            {
                // Not enough data for complete packet
                break;
            }

            offset += 4; // Skip size field

            // Read packet ID
            const id = buffer.readInt32LE( offset );
            offset += 4;

            // Read packet type
            const type = buffer.readInt32LE( offset ) as RconPacketType;
            offset += 4;

            // Read body (size - 10 bytes for id, type, and terminators)
            const bodySize = size - 10;
            const body = buffer.subarray( offset, offset + bodySize ).toString( 'utf8' );
            offset += bodySize;

            // Skip null terminators
            offset += 2;

            packets.push( { id, type, body } );
        }

        return packets;
    }

    /**
     * Create an authentication packet
     */
    static createAuthPacket ( id: number, password: string ): RconPacket {
        return {
            id,
            type: RconPacketType.SERVERDATA_AUTH,
            body: password
        };
    }

    /**
     * Create a command execution packet
     */
    static createCommandPacket ( id: number, command: string ): RconPacket {
        return {
            id,
            type: RconPacketType.SERVERDATA_EXECCOMMAND,
            body: command
        };
    }

    /**
     * Create an empty terminator packet used to delimit multi-packet responses.
     * Some Source engine servers send multiple SERVERDATA_RESPONSE_VALUE packets.
     * A trailing empty response (body="") from a distinct ID can be used as sentinel.
     */
    static createTerminatorPacket ( id: number ): RconPacket {
        return {
            id,
            type: RconPacketType.SERVERDATA_RESPONSE_VALUE,
            body: ''
        };
    }

    /**
     * Validate packet structure
     */
    static validatePacket ( packet: RconPacket ): boolean {
        if ( typeof packet.id !== 'number' || packet.id < 0 || packet.id > 0x7FFFFFFF )
        {
            return false;
        }

        if ( !Object.values( RconPacketType ).includes( packet.type ) )
        {
            return false;
        }

        if ( typeof packet.body !== 'string' )
        {
            return false;
        }

        return true;
    }

    /**
     * Check if packet is an authentication response
     */
    static isAuthResponse ( packet: RconPacket ): boolean {
        return packet.type === RconPacketType.SERVERDATA_AUTH_RESPONSE;
    }

    /**
     * Check if authentication was successful
     */
    static isAuthSuccess ( authPacket: RconPacket, responsePacket: RconPacket ): boolean {
        // Per protocol, failed auth returns ID = -1 in response.
        if ( responsePacket.id === -1 ) return false;
        return this.isAuthResponse( responsePacket ) && authPacket.id === responsePacket.id;
    }

    /**
     * Check if packet is a command response
     */
    static isCommandResponse ( packet: RconPacket ): boolean {
        return packet.type === RconPacketType.SERVERDATA_RESPONSE_VALUE;
    }
}
