import { watch } from "fs/promises";
import { parseLogEvents } from './src/events';
import type { GameEvent, ChatEvent } from './src/events';
import StatsService from './src/stats-service';
import CommandHandler from './src/command-handler';

const path = "C:\\Users\\danie\\code\\sandstorm-admin-wrapper\\sandstorm-server\\Insurgency\\Saved\\Logs";
const serverId = "60844f66-b93b-4fe1-afc4-a0a91b493865";
const logFilePath = `${ path }\\${ serverId }.log`;

// Initialize stats service with crash detection
StatsService.initialize();

console.log( "Watching for file changes..." );

// Function to read and parse the last line of the log file
async function processLogFile (): Promise<GameEvent[]> {
    try
    {
        // Read the file content
        const logContent = await Bun.file( logFilePath ).text();
        const lines = logContent.split( '\n' );

        // Get only the last non-empty line
        const lastLine = lines.filter( line => line.trim() ).pop();
        if ( !lastLine ) return [];

        // Parse just the last line
        const events = parseLogEvents( lastLine );

        // Only process if we found actual events we care about
        if ( events.length === 0 )
        {
            return []; // Silent - no events we care about
        }

        // Update file activity timestamp
        StatsService.updateActivity();

        events.forEach( event => {
            // Process event in database first
            StatsService.processEvent( event );
            // Only log events that are specifically listed in events.md
            switch ( event.type )
            {
                case 'game_over':
                    console.log( `🎯 Game over!` );
                    break;
                case 'player_join':
                    console.log( `🎮 Player joined: ${ event.data.playerName }` );
                    break;
                case 'player_leave':
                case 'player_disconnect':
                    console.log( `👋 Player left: ${ event.data.playerName || event.data.steamId }` );
                    break;
                case 'team_kill':
                    console.log( `� TEAM KILL: ${ event.data.killer } killed teammate ${ event.data.victim } with ${ event.data.weapon }` );
                    break;
                case 'player_kill':
                    console.log( `� ${ event.data.killer } killed ${ event.data.victim } with ${ event.data.weapon }` );
                    break;
                case 'suicide':
                    console.log( `💥 ${ event.data.killer } committed suicide with ${ event.data.weapon }` );
                    break;
                case 'difficulty_set':
                    console.log( `⚙️  AI difficulty set to: ${ event.data.difficulty }` );
                    break;
                case 'round_over':
                    console.log( `🏁 Round ${ event.data.roundNumber } over: Team ${ event.data.winningTeam } won (${ event.data.winReason })` );
                    break;
                case 'map_load':
                    console.log( `🗺️  Loading map: ${ event.data.mapName } (${ event.data.scenario })` );
                    break;
                case 'chat_command':
                    // Handle chat commands and show response
                    const response = CommandHandler.handleCommand( event as ChatEvent );
                    console.log( `💬 ${ event.data.playerName } used command: ${ event.data.command } ${ event.data.args?.join( ' ' ) || '' }` );
                    if ( response )
                    {
                        console.log( `📢 Response: ${ response }` );
                    }
                    break;
                // Ignore any other events not listed in events.md
                default:
                    // Silent - don't log events we don't care about
                    break;
            }
        } );

        return events;
    } catch ( error )
    {
        console.error( 'Error processing log file:', error );
        return [];
    }
}


// Debounce variables to prevent multiple rapid fires
let lastProcessedTime = 0;
let lastProcessedLine = '';
let fileSize = 0;
const debounceDelay = 100; // 100ms debounce

try
{
    // Get initial file size
    const stats = await Bun.file( logFilePath ).stat();
    fileSize = stats.size;
    // Silent initialization
} catch ( error )
{
    console.log( `⚠️  Log file doesn't exist yet: ${ logFilePath }` );
} const watcher = watch( path );
for await ( const event of watcher )
{
    if ( event.filename && event.filename.includes( 'backup' ) )
    {
        continue;
    }
    if ( event.filename && event.filename == `${ serverId }.log` )
    {
        const now = Date.now();

        // Debounce rapid file changes
        if ( now - lastProcessedTime < debounceDelay )
        {
            continue;
        }

        try
        {
            // Check if file actually changed size (new content)
            const currentStats = await Bun.file( logFilePath ).stat();
            if ( currentStats.size <= fileSize )
            {
                continue; // No new content
            }

            // Read just the last line to check if it's new
            const logContent = await Bun.file( logFilePath ).text();
            const lines = logContent.split( '\n' );
            const lastLine = lines.filter( line => line.trim() ).pop();

            if ( !lastLine || lastLine === lastProcessedLine )
            {
                fileSize = currentStats.size; // Update size even if no new line
                continue; // Skip if no new content or same line as before
            }

            // Process the log file and parse events
            const gameEvents = await processLogFile();

            // Update tracking variables
            lastProcessedTime = now;
            lastProcessedLine = lastLine;
            fileSize = currentStats.size;

        } catch ( error )
        {
            console.error( 'Error reading log file:', error );
        }
    }
}

