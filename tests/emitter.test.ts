import { beforeEach, describe, expect, it } from 'bun:test';
import eventBus from '../src/lib/emitter/emitter';

describe('Typed eventBus', () => {
  beforeEach(() => {
    // ensure a clean slate for each test
    eventBus.removeAllListeners();
  });

  it('calls listeners registered with on and returns true from emit', () => {
    const received: any[] = [];
    const payload = {
      type: 'player_join',
      timestamp: '2025.10.06-12.00.00:000',
      data: { playerName: 'Alice' },
      rawLine: 'dummy',
    };

    eventBus.on('player_join', (evt) => received.push(evt));

    const emitted = eventBus.emit('player_join', payload as any);
    expect(emitted).toBe(true);
    expect(received).toHaveLength(1);
    expect(received[0]).toEqual(payload);
  });

  it('calls once listeners only once', () => {
    let calls = 0;
    eventBus.once('player_join', () => calls++);

    eventBus.emit('player_join', { type: 'player_join' } as any);
    eventBus.emit('player_join', { type: 'player_join' } as any);

    expect(calls).toBe(1);
  });

  it('removeListener / off prevents future calls', () => {
    let calls = 0;
    const listener = () => calls++;

    eventBus.on('player_join', listener);
    eventBus.emit('player_join', { type: 'player_join' } as any);
    expect(calls).toBe(1);

    eventBus.off('player_join', listener);
    eventBus.emit('player_join', { type: 'player_join' } as any);
    expect(calls).toBe(1);
  });

  it('emit returns false when no listeners are registered', () => {
    // ensure no listeners
    eventBus.removeAllListeners();
    const result = eventBus.emit('player_join', { type: 'player_join' } as any);
    expect(result).toBe(false);
  });
});
