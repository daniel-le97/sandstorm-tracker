import { describe, expect, it } from 'bun:test';
import { RconProtocol } from '../src/lib/rcon/protocol';
import { RconClient } from '../src/lib/rcon/client';
import { RconPacketType, RconClientState, type RconPacket } from '../src/lib/rcon/types';

describe('Rcon client behavior', () => {
        it( 'round-trips UTF-8 body with emoji and CJK', () => {
            const body = 'Hello, 世界 👋 — café — ñáç';
            const pkt: RconPacket = { id: 123, type: 0, body };
            const buf = RconProtocol.encode( pkt );
            const decoded = RconProtocol.decode( buf )[ 0 ];
            expect( decoded.body ).toBe( body );
            expect( Buffer.byteLength( body, 'utf8' ) ).toBeGreaterThan( body.length ); // shows multi-byte present
        } );
    it('aggregates multi-packet responses and resolves execute()', async () => {
        const client = new RconClient({ host: '127.0.0.1', port: 27015, password: 'pw' });
        // Mock socket to capture writes
        const writes: Buffer[] = [];
        (client as any).socket = {
            write: (b: Buffer) => writes.push(b),
            destroyed: false,
            removeAllListeners: () => { /* noop */ },
            destroy: function () { this.destroyed = true; }
        } as any;
        // Put client into authenticated state so execute() is allowed
        (client as any).state = RconClientState.AUTHENTICATED;

        const execPromise = client.execute('status');

        // Find queued command
    const queued = Array.from((client as any).commandQueue.values())[0] as any;
    expect(queued).toBeTruthy();
    const id: number = queued.id;
    const terminatorId: number = queued.terminatorId;

        // Simulate server sending a primary response then a terminator packet
        const resp1 = RconProtocol.encode({ id, type: RconPacketType.SERVERDATA_RESPONSE_VALUE, body: 'PART1-' });
        const term = RconProtocol.encode({ id: terminatorId, type: RconPacketType.SERVERDATA_RESPONSE_VALUE, body: '' });

        // Feed both in one chunk
        (client as any).handleData(Buffer.concat([resp1, term]));

        const resp = await execPromise;
        expect(resp.body).toBe('PART1-');
        // commandQueue should be cleared
        expect((client as any).commandQueue.size).toBe(0);
        // stats updated
        expect(client.getStats().responsesReceived).toBeGreaterThanOrEqual(1);
        client.destroy();
    });

    it('buffers partial packets across multiple data events', async () => {
        const client = new RconClient({ host: '127.0.0.1', port: 27015, password: 'pw' });
        (client as any).socket = {
            write: (_b: Buffer) => { /* noop */ },
            destroyed: false,
            removeAllListeners: () => { /* noop */ },
            destroy: function () { this.destroyed = true; }
        } as any;
        (client as any).state = RconClientState.AUTHENTICATED;

        const execPromise = client.execute('echo hi');
    const queued = Array.from((client as any).commandQueue.values())[0] as any;
    const id: number = queued.id;
    const terminatorId: number = queued.terminatorId;

        const resp = RconProtocol.encode({ id, type: RconPacketType.SERVERDATA_RESPONSE_VALUE, body: 'chunked-response' });
        const term = RconProtocol.encode({ id: terminatorId, type: RconPacketType.SERVERDATA_RESPONSE_VALUE, body: '' });

        const combined = Buffer.concat([resp, term]);

        // Send first half
        const mid = Math.floor(combined.length / 2);
        const first = combined.subarray(0, mid);
        const second = combined.subarray(mid);

        (client as any).handleData(first);
        // buffer should hold the partial data
        expect((client as any).buffer.length).toBe(first.length);

        // Send remainder
        (client as any).handleData(second);

        const response = await execPromise;
        expect(response.body).toBe('chunked-response');
        client.destroy();
    });

    it('rejects authentication when server responds with id -1 and emits authFailed', async () => {
        const client = new RconClient({ host: '127.0.0.1', port: 27015, password: 'secret' });
        (client as any).socket = {
            write: (_b: Buffer) => { /* noop */ },
            destroyed: false,
            removeAllListeners: () => { /* noop */ },
            destroy: function () { this.destroyed = true; }
        } as any;

        let emitted = false;
        client.on('authFailed', () => { emitted = true; });

        // Call the internal authenticate method (returns a Promise)
        const authPromise = (client as any).authenticate();

        // Simulate server sending failed auth response (id === -1)
        const failed = RconProtocol.encode({ id: -1, type: RconPacketType.SERVERDATA_AUTH_RESPONSE, body: '' });
        // Deliver via event emission which authenticate listens for
        client.emit('authResponse', { id: -1, type: RconPacketType.SERVERDATA_AUTH_RESPONSE, body: '' });

        try {
            await authPromise;
            // Should not reach here
            throw new Error('authenticate() should have rejected');
        } catch (err: any) {
            expect(err).toBeInstanceOf(Error);
            expect(emitted).toBe(true);
        }
        client.destroy();
    });

    it('throws when encoding a packet that exceeds the max packet size', () => {
        // MAX_PACKET_SIZE = 4096, protocol adds 14 bytes of overhead
        const overhead = 14; // protocol calculation uses body + 14
        const tooLarge = 4096 - overhead + 1;
        const body = 'a'.repeat(tooLarge);
        const pkt = { id: 1, type: RconPacketType.SERVERDATA_EXECCOMMAND, body };
        expect(() => RconProtocol.encode(pkt as any)).toThrow();
    });
});
