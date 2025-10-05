import { parseLogEvent, parseLogEvents } from './events';
import type { GameEvent, PlayerKillEvent, ChatEvent } from './events';


const lines = await Bun.file('logfile.log').text();
const logLines = lines.split('\n').filter(line => line.trim());

// Example 1: Parsing each log log
export function processLogFile (): GameEvent[] {
    const events: GameEvent[] = [];
    logLines.forEach( line => {
        const parsedEvents = parseLogEvents( line );
        events.push( ...parsedEvents );
    } );
    return events;
}

const events = processLogFile();
events.forEach( event  => {
    console.log(event);
} )