import type { ChatEvent } from "./events";
import TrackerService from "./trackerService";

// Type definitions for database results
interface PlayerStats {
    player_name: string;
    steam_id: string;
    total_playtime_minutes: number;
    total_kills: number;
    total_deaths: number;
    team_kills: number;
    suicides: number;
    kdr: number;
}

interface TopPlayer {
    player_name: string;
    total_kills: number;
    total_deaths: number;
    kdr: number;
}

interface WeaponStat {
    weapon_name: string;
    kills: number;
    deaths: number;
    team_kills: number;
    suicides: number;
}

export class CommandHandler {
    /**
     * Process a chat command and return response (if any)
     */
    static handleCommand ( event: ChatEvent, serverId?: number ): string | null {
        const { command, args, playerName, steamId } = event.data;
        const serverDbId = serverId || 1; // Default to server ID 1 for backward compatibility

        try
        {
            switch ( command.toLowerCase() )
            {
                case "!stats":
                    return this.handleStatsCommand( steamId, playerName, serverDbId, args );

                case "!kdr":
                    return this.handleKdrCommand( steamId, playerName, serverDbId );

                case "!top":
                    return this.handleTopCommand( serverDbId, args );

                case "!guns":
                case "!weapons":
                    return this.handleWeaponsCommand( steamId, playerName, serverDbId );

                default:
                    return null; // Unknown command
            }
        } catch ( error )
        {
            console.error( `Error handling command ${ command }:`, error );
            return `Error processing command ${ command }`;
        }
    }

    /**
     * Handle !stats command - show player stats (self or another player) for specific server
     */
    private static handleStatsCommand ( steamId: string, playerName: string, serverId: number, args?: string[] ): string {
        let targetSteamId = steamId;
        let targetName = playerName;

        // If a player name is provided, try to find that player
        if ( args && args.length > 0 )
        {
            const searchName = args.join( " " ).toLowerCase();
            // TODO: Implement player search by partial name
            // For now, just show requesting player's stats
            targetName = searchName;
        }

        const stats = TrackerService.getPlayerStats( targetSteamId, serverId ) as PlayerStats | null;

        if ( !stats )
        {
            return `Stats not found for ${ targetName } on this server`;
        }

        const playtimeHours = Math.round( ( ( stats.total_playtime_minutes || 0 ) / 60 ) * 10 ) / 10;
        const scorePerMin =
            stats.total_playtime_minutes > 0
                ? Math.round( ( ( stats.total_kills || 0 ) / ( stats.total_playtime_minutes || 1 ) ) * 60 * 100 ) / 100
                : 0;

        return `${ stats.player_name }: ${ stats.total_kills } kills, ${ stats.total_deaths } deaths, K/D: ${ stats.kdr }, Score/min: ${ scorePerMin }, Playtime: ${ playtimeHours }h`;
    }

    /**
     * Handle !kdr command - show kill/death ratio for specific server
     */
    private static handleKdrCommand ( steamId: string, playerName: string, serverId: number ): string {
        const stats = TrackerService.getPlayerStats( steamId, serverId ) as PlayerStats | null;

        if ( !stats )
        {
            return `Stats not found for ${ playerName } on this server`;
        }

        return `${ stats.player_name }: ${ stats.total_kills } kills, ${ stats.total_deaths } deaths, K/D ratio: ${ stats.kdr }`;
    }

    /**
     * Handle !top command - show top players for specific server
     */
    private static handleTopCommand ( serverId: number, args?: string[] ): string {
        const limit = args && args[ 0 ] ? Math.min( parseInt( args[ 0 ] ) || 3, 10 ) : 3;
        const topPlayers = TrackerService.getTopPlayers( serverId, limit ) as TopPlayer[];

        if ( !topPlayers || topPlayers.length === 0 )
        {
            return "No player statistics available yet on this server";
        }

        let response = `Top ${ limit } Players (This Server):\n`;
        topPlayers.forEach( ( player, index: number ) => {
            response += `${ index + 1 }. ${ player.player_name }: ${ player.total_kills } kills (K/D: ${ player.kdr })\n`;
        } );

        return response.trim();
    }

    /**
     * Handle !guns/!weapons command - show player's weapon stats for specific server
     */
    private static handleWeaponsCommand ( steamId: string, playerName: string, serverId: number ): string {
        const weapons = TrackerService.getPlayerWeapons( steamId, serverId, 3 ) as WeaponStat[];

        if ( !weapons || weapons.length === 0 )
        {
            return `No weapon stats found for ${ playerName }`;
        }

        let response = `${ playerName }'s Top Weapons:\n`;
        weapons.forEach( ( weapon, index: number ) => {
            response += `${ index + 1 }. ${ weapon.weapon_name }: ${ weapon.kills } kills`;
            if ( weapon.deaths > 0 )
            {
                response += ` (${ weapon.deaths } deaths)`;
            }
            response += "\n";
        } );

        return response.trim();
    }

    /**
     * Search for players by partial name match
     * Returns the best match or null if not found
     */
    private static findPlayerByName ( searchName: string ): { steamId: string; playerName: string; } | null {
        // TODO: Implement database query to search players by partial name
        // This would require adding a method to TrackerService
        return null;
    }
}

export default CommandHandler;
