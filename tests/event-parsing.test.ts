import { test, expect, describe } from 'bun:test';
import { parseLogEvent, parseLogEvents, type GameEvent, type PlayerKillEvent, type PlayerJoinEvent, type PlayerLeaveEvent, type ChatEvent, type RoundOverEvent, type MapLoadEvent, type DifficultyEvent } from '../src/events';

describe( 'Event Parsing', () => {
    test( 'Game Over Event', () => {
        const logLine = "[2025.10.04-14.31.05:706][800]LogGameplayEvents: Display: Game over";
        const event = parseLogEvent( logLine );

        expect( event?.type ).toBe( 'game_over' );
        expect( event?.timestamp ).toBe( '2025.10.04-14.31.05:706' );
    } );

    test( 'Player Join Event', () => {
        const logLine = "[2025.10.04-14.31.05:706][800]LogNet: Join succeeded: ArmoredBear";
        const event = parseLogEvent( logLine ) as PlayerJoinEvent;

        expect( event?.type ).toBe( 'player_join' );
        expect( event?.data.playerName ).toBe( 'ArmoredBear' );
    } );

    test( 'Player Leave Event (RCON)', () => {
        const logLine = "[2025.10.04-15.35.33:666][195]LogRcon: Console command received << say See you later, ArmoredBear!";
        const event = parseLogEvent( logLine ) as PlayerLeaveEvent;

        expect( event?.type ).toBe( 'player_leave' );
        expect( event?.data.playerName ).toBe( 'ArmoredBear' );
    } );

    test( 'Player Disconnect Event (EOS)', () => {
        const logLine = "[2025.10.04-15.56.27:095][530]LogEOSAntiCheat: Display: ServerUnregisterClient: UserId (76561198995742987)";
        const event = parseLogEvent( logLine );

        expect( event?.type ).toBe( 'player_disconnect' );
        expect( event?.data.steamId ).toBe( '76561198995742987' );
    } );

    test( 'Team Kill Event', () => {
        const logLine = "[2025.10.04-14.31.05:706][800]LogGameplayEvents: Display: ArmoredBear[76561198995742987, team 0] killed Teammate[76561198000000001, team 0] with BP_Firearm_M16A4";
        const event = parseLogEvent( logLine ) as PlayerKillEvent;

        expect( event?.type ).toBe( 'team_kill' );
        expect( event?.data.killer ).toBe( 'ArmoredBear' );
        expect( event?.data.victim ).toBe( 'Teammate' );
        expect( event?.data.killerTeam ).toBe( 0 );
        expect( event?.data.victimTeam ).toBe( 0 );
        expect( event?.data.weapon ).toBe( 'M16A4' );
    } );

    test( 'Player Kill Event (vs AI)', () => {
        const logLine = "[2025.10.04-14.31.05:706][800]LogGameplayEvents: Display: ArmoredBear[76561198995742987, team 0] killed AIBot[INVALID, team 1] with BP_Firearm_M16A4";
        const event = parseLogEvent( logLine ) as PlayerKillEvent;

        expect( event?.type ).toBe( 'player_kill' );
        expect( event?.data.killer ).toBe( 'ArmoredBear' );
        expect( event?.data.victim ).toBe( 'AIBot' );
        expect( event?.data.killerTeam ).toBe( 0 );
        expect( event?.data.victimTeam ).toBe( 1 );
        expect( event?.data.weapon ).toBe( 'M16A4' );
    } );

    test( 'Suicide Event (fall damage)', () => {
        const logLine = "[2025.10.04-14.31.05:706][800]LogGameplayEvents: Display: ArmoredBear[76561198995742987, team 0] killed ArmoredBear[76561198995742987, team 0] with BP_Character_Player_C_2147481419";
        const event = parseLogEvent( logLine );

        expect( event?.type ).toBe( 'suicide' );
        expect( event?.data.killer ).toBe( 'ArmoredBear' );
        expect( event?.data.victim ).toBe( 'ArmoredBear' );
        expect( event?.data.killerSteamId ).toBe( '76561198995742987' );
        expect( event?.data.victimSteamId ).toBe( '76561198995742987' );
    } );

    test( 'Suicide Event (molotov)', () => {
        const logLine = "[2025.10.04-14.31.05:706][800]LogGameplayEvents: Display: Player[76561198000000001, team 0] killed Player[76561198000000001, team 0] with BP_Projectile_Molotov_C_2147481420";
        const event = parseLogEvent( logLine );

        expect( event?.type ).toBe( 'suicide' );
        expect( event?.data.killer ).toBe( 'Player' );
        expect( event?.data.victim ).toBe( 'Player' );
        expect( event?.data.killerSteamId ).toBe( '76561198000000001' );
        expect( event?.data.victimSteamId ).toBe( '76561198000000001' );
    } );

    test( 'Round Over Event', () => {
        const logLine = "[2025.10.04-14.31.05:706][800]LogGameplayEvents: Display: Round 2 Over: Team 1 won (win reason: Elimination)";
        const event = parseLogEvent( logLine ) as RoundOverEvent;

        expect( event?.type ).toBe( 'round_over' );
        expect( event?.data.roundNumber ).toBe( 2 );
        expect( event?.data.winningTeam ).toBe( 1 );
        expect( event?.data.winReason ).toBe( 'Elimination' );
    } );

    test( 'Map Load Event', () => {
        const logLine = "[2025.10.04-14.31.05:706][800]LogLoad: LoadMap: /Game/Maps/Canyon/Canyon_Crossing_Checkpoint_Security?Scenario=Scenario_Canyon_Checkpoint_Security?MaxPlayers=8?Lighting=Day";
        const event = parseLogEvent( logLine ) as MapLoadEvent;

        expect( event?.type ).toBe( 'map_load' );
        expect( event?.data.mapName ).toBe( 'Canyon' );
        expect( event?.data.scenario ).toBe( 'Canyon_Checkpoint' );
        expect( event?.data.team ).toBe( 'Security' );
        expect( event?.data.maxPlayers ).toBe( 8 );
        expect( event?.data.lighting ).toBe( 'Day' );
    } );

    test( 'Difficulty Set Event', () => {
        const logLine = "[2025.10.04-14.31.05:706][800]LogAI: Warning: AI difficulty set to 0.5";
        const event = parseLogEvent( logLine ) as DifficultyEvent;

        expect( event?.type ).toBe( 'difficulty_set' );
        expect( event?.data.difficulty ).toBe( 0.5 );
    } );
} );

describe( 'Chat Commands', () => {
    test( 'Chat Command: !stats', () => {
        const logLine = "[2025.10.04-14.31.05:706][800]LogChat: Display: ArmoredBear(76561198995742987) Global Chat: !stats";
        const event = parseLogEvent( logLine ) as ChatEvent;

        expect( event?.type ).toBe( 'chat_command' );
        expect( event?.data.playerName ).toBe( 'ArmoredBear' );
        expect( event?.data.steamId ).toBe( '76561198995742987' );
        expect( event?.data.command ).toBe( '!stats' );
        expect( event?.data.args ).toBeUndefined();
    } );

    test( 'Chat Command: !stats with player name', () => {
        const logLine = "[2025.10.04-14.31.05:706][800]LogChat: Display: ArmoredBear(76561198995742987) Global Chat: !stats PlayerName";
        const event = parseLogEvent( logLine ) as ChatEvent;

        expect( event?.type ).toBe( 'chat_command' );
        expect( event?.data.playerName ).toBe( 'ArmoredBear' );
        expect( event?.data.steamId ).toBe( '76561198995742987' );
        expect( event?.data.command ).toBe( '!stats' );
        expect( event?.data.args ).toEqual( [ 'PlayerName' ] );
    } );

    test( 'Chat Command: !stats with partial name', () => {
        const logLine = "[2025.10.04-14.31.05:706][800]LogChat: Display: ArmoredBear(76561198995742987) Global Chat: !stats Armored";
        const event = parseLogEvent( logLine ) as ChatEvent;

        expect( event?.type ).toBe( 'chat_command' );
        expect( event?.data.playerName ).toBe( 'ArmoredBear' );
        expect( event?.data.steamId ).toBe( '76561198995742987' );
        expect( event?.data.command ).toBe( '!stats' );
        expect( event?.data.args ).toEqual( [ 'Armored' ] );
    } );

    test( 'Chat Command: !kdr', () => {
        const logLine = "[2025.10.04-14.31.05:706][800]LogChat: Display: ArmoredBear(76561198995742987) Global Chat: !kdr";
        const event = parseLogEvent( logLine ) as ChatEvent;

        expect( event?.type ).toBe( 'chat_command' );
        expect( event?.data.command ).toBe( '!kdr' );
    } );

    test( 'Chat Command: !top', () => {
        const logLine = "[2025.10.04-14.31.05:706][800]LogChat: Display: ArmoredBear(76561198995742987) Global Chat: !top";
        const event = parseLogEvent( logLine ) as ChatEvent;

        expect( event?.type ).toBe( 'chat_command' );
        expect( event?.data.command ).toBe( '!top' );
    } );

    test( 'Chat Command: !guns', () => {
        const logLine = "[2025.10.04-14.31.05:706][800]LogChat: Display: ArmoredBear(76561198995742987) Global Chat: !guns";
        const event = parseLogEvent( logLine ) as ChatEvent;

        expect( event?.type ).toBe( 'chat_command' );
        expect( event?.data.command ).toBe( '!guns' );
    } );
} );

describe( 'Special Events', () => {
    test( 'Fall Damage Event', () => {
        const logLine = "[2025.10.04-14.31.05:706][800]LogSoldier: Applying 50.0 fall damage to player";
        const event = parseLogEvent( logLine );

        expect( event?.type ).toBe( 'fall_damage' );
        expect( event?.data.damage ).toBe( 50.0 );
    } );

    test( 'Map Vote Event (start)', () => {
        const logLine = "[2025.10.04-14.31.05:706][800]LogMapVoteManager: Display: Existing Vote Options: Canyon, Precinct, Crossing";
        const event = parseLogEvent( logLine );

        expect( event?.type ).toBe( 'map_vote_start' );
    } );
} );

describe( 'Edge Cases', () => {
    test( 'Invalid log line returns null', () => {
        const logLine = "This is not a valid log line";
        const event = parseLogEvent( logLine );

        expect( event ).toBeNull();
    } );

    test( 'Empty line returns null', () => {
        const logLine = "";
        const event = parseLogEvent( logLine );

        expect( event ).toBeNull();
    } );

    test( 'Whitespace line returns null', () => {
        const logLine = "   \t  \n  ";
        const event = parseLogEvent( logLine );

        expect( event ).toBeNull();
    } );
} );

describe( 'Multiple Events', () => {
    test( 'Multiple events parsing', () => {
        const logContent = `[2025.10.04-14.31.05:706][800]LogGameplayEvents: Display: Game over
[2025.10.04-14.31.05:706][800]LogNet: Join succeeded: ArmoredBear
[2025.10.04-14.31.05:706][800]LogChat: Display: ArmoredBear(76561198995742987) Global Chat: !stats`;

        const events = parseLogEvents( logContent );

        expect( events ).toHaveLength( 3 );
        expect( events[ 0 ]?.type ).toBe( 'game_over' );
        expect( events[ 1 ]?.type ).toBe( 'player_join' );
        expect( events[ 2 ]?.type ).toBe( 'chat_command' );
    } );

    test( 'Mixed valid/invalid lines parsing', () => {
        const logContent = `[2025.10.04-14.31.05:706][800]LogGameplayEvents: Display: Game over
This is invalid
[2025.10.04-14.31.05:706][800]LogNet: Join succeeded: ArmoredBear
Another invalid line`;

        const events = parseLogEvents( logContent );

        expect( events ).toHaveLength( 2 );
        expect( events[ 0 ]?.type ).toBe( 'game_over' );
        expect( events[ 1 ]?.type ).toBe( 'player_join' );
    } );
} );