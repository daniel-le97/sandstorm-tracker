/**
 * Maps blueprint weapon names to user-friendly names
 * @param weaponBlueprintName - Full blueprint weapon name from logs
 * @returns Clean weapon name
 */
function getCleanWeaponName(weaponBlueprintName: string): string {
    // Common weapon mappings
    const weaponMap: Record<string, string> = {
        // Assault Rifles
        BP_Firearm_M16A4: "M16A4",
        BP_Firearm_AK74: "AK-74",
        BP_Firearm_M4A1: "M4A1",
        BP_Firearm_AKM: "AKM",
        BP_Firearm_SCARH: "SCAR-H",
        BP_Firearm_G36K: "G36K",
        BP_Firearm_AK12: "AK-12",
        BP_Firearm_M16A2: "M16A2",
        BP_Firearm_AKS74U: "AKS-74U",
        BP_Firearm_Alpha_AK: "Alpha AK",
        BP_Firearm_VHS: "VHS-2",

        // Sniper Rifles
        BP_Firearm_M24: "M24 SWS",
        BP_Firearm_Mosin: "Mosin Nagant",
        BP_Firearm_M110: "M110 SASS",
        BP_Firearm_SVD: "SVD",
        BP_Firearm_L96A1: "L96A1",

        // Shotguns
        BP_Firearm_M590A1: "M590A1",
        BP_Firearm_TOZ: "TOZ-194",

        // SMGs
        BP_Firearm_Sterling: "Sterling L2A3",
        BP_Firearm_UMP45: "UMP-45",
        BP_Firearm_MP5A2: "MP5A2",
        BP_Firearm_MP7: "MP7",
        BP_Firearm_Uzi: "Uzi",

        // LMGs
        BP_Firearm_M249: "M249 SAW",
        BP_Firearm_PKM: "PKM",
        BP_Firearm_M240B: "M240B",
        BP_Firearm_MG3: "MG3",

        // Pistols
        BP_Firearm_M9: "M9 Beretta",
        BP_Firearm_Makarov: "Makarov",
        BP_Firearm_M45: "M45A1",
        BP_Firearm_Welrod: "Welrod Mk II",
        BP_Firearm_PF940: "PF940C",

        // Launchers
        BP_Firearm_M203: "M203 Grenade Launcher",
        BP_Firearm_GP25: "GP-25 Grenade Launcher",
        BP_Firearm_RPG7: "RPG-7",
        BP_Firearm_AT4: "AT4",
        BP_Firearm_M3MAAWS: "M3 MAAWS",

        // Explosives and Special
        BP_Projectile_Molotov: "Molotov Cocktail",
        BP_Projectile_Grenade_Frag: "Frag Grenade",
        BP_Projectile_Grenade_Smoke: "Smoke Grenade",
        BP_Projectile_IED: "IED",
        BP_Character_Player: "Fall Damage",
        BP_Projectile_Rocket: "Rocket",
        BP_Projectile_40mm: "40mm Grenade",

        // Commander Fire Support
        BP_Projectile_Mortar_HE: "Mortar Strike",
        BP_Projectile_Mortar_Smoke: "Mortar Smoke",
        BP_Projectile_Artillery_HE: "Artillery Strike",
        BP_Projectile_Artillery_Smoke: "Artillery Smoke",
        BP_Vehicle_Helicopter_Gunship: "Attack Helicopter",
        BP_Projectile_Hellfire: "Hellfire Missile",
        BP_Projectile_Airstrike: "Air Strike",
        BP_Projectile_Strafe: "Strafing Run",
        BP_Projectile_Rocket_155mm: "155mm Artillery",
        BP_Projectile_Rocket_120mm: "120mm Mortar",

        // Melee
        BP_Melee_Knife: "Combat Knife",
        BP_Melee_Kukri: "Kukri",
        BP_Melee_Machete: "Machete",
    };

    // First try to find exact match
    for (const [blueprint, cleanName] of Object.entries(weaponMap)) {
        if (weaponBlueprintName.includes(blueprint)) {
            return cleanName;
        }
    }

    // If no match found, try to extract a reasonable name from the blueprint
    // Remove BP_Firearm_, BP_Projectile_, BP_Melee_ prefixes and suffixes
    let cleanName = weaponBlueprintName
        .replace(/^BP_Firearm_/, "")
        .replace(/^BP_Projectile_/, "")
        .replace(/^BP_Melee_/, "")
        .replace(/^BP_Character_/, "")
        .replace(/_C_\d+$/, "") // Remove _C_numbers at the end
        .replace(/_\d+$/, "") // Remove _numbers at the end
        .replace(/_/g, " ") // Replace underscores with spaces
        .trim();

    // If we still have a reasonable name, return it capitalized
    if (cleanName && cleanName.length > 0) {
        return cleanName
            .split(" ")
            .map((word) => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
            .join(" ");
    }

    // Last resort: return the original name
    return weaponBlueprintName;
}

// Event types based on the log patterns
export interface GameEvent {
    type: string;
    timestamp: string;
    data: Record<string, any>;
    rawLine: string;
}

export interface PlayerKillEvent extends GameEvent {
    type: "player_kill" | "team_kill" | "suicide";
    data: {
        killer: string;
        killerSteamId: string;
        killerTeam: number;
        victim: string;
        victimSteamId: string;
        victimTeam: number;
        weapon: string;
    };
}

export interface PlayerJoinEvent extends GameEvent {
    type: "player_join";
    data: {
        playerName: string;
    };
}

export interface PlayerLeaveEvent extends GameEvent {
    type: "player_leave" | "player_disconnect";
    data: {
        playerName?: string;
        steamId?: string;
    };
}

export interface ChatEvent extends GameEvent {
    type: "chat_command";
    data: {
        playerName: string;
        steamId: string;
        command: string;
        args?: string[];
    };
}

export interface GameOverEvent extends GameEvent {
    type: "game_over";
}

export interface RoundOverEvent extends GameEvent {
    type: "round_over";
    data: {
        roundNumber: number;
        winningTeam: number;
        winReason: string;
    };
}

export interface MapLoadEvent extends GameEvent {
    type: "map_load";
    data: {
        mapName: string;
        scenario: string;
        team?: string;
        maxPlayers?: number;
        lighting?: string;
    };
}

export interface DifficultyEvent extends GameEvent {
    type: "difficulty_set";
    data: {
        difficulty: number;
    };
}

/**
 * Parses a log line and returns a GameEvent if it matches any known patterns
 * @param logLine - Raw log line from the game server
 * @returns GameEvent object or null if no pattern matches
 */
export function parseLogEvent(logLine: string): GameEvent | null {
    const trimmedLine = logLine.trim();
    if (!trimmedLine) return null;

    // console.log(trimmedLine);
    // Extract timestamp from the log line
    const timestampMatch = trimmedLine.match(/^\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{3})\]/);
    const timestamp = timestampMatch?.[1] ?? "";

    // Game Over
    if (trimmedLine.includes("LogGameplayEvents: Display: Game over")) {
        return {
            type: "game_over",
            timestamp,
            data: {},
            rawLine: trimmedLine,
        };
    }

    // Player Join
    const joinMatch = trimmedLine.match(/LogNet: Join succeeded: (.+)$/);
    if (joinMatch?.[1]) {
        return {
            type: "player_join",
            timestamp,
            data: {
                playerName: joinMatch[1],
            },
            rawLine: trimmedLine,
        };
    }

    // Player Leave (RCON message)
    const leaveMatch = trimmedLine.match(/LogRcon: .+ << say See you later, (.+)!$/);
    if (leaveMatch?.[1]) {
        return {
            type: "player_leave",
            timestamp,
            data: {
                playerName: leaveMatch[1],
            },
            rawLine: trimmedLine,
        };
    }

    // Player Disconnect (EOS Anti-Cheat)
    const disconnectMatch = trimmedLine.match(/LogEOSAntiCheat: Display: ServerUnregisterClient: UserId \((\d+)\)/);
    if (disconnectMatch?.[1]) {
        return {
            type: "player_disconnect",
            timestamp,
            data: {
                steamId: disconnectMatch[1],
            },
            rawLine: trimmedLine,
        };
    }

    // Player Kills (including team kills and suicides)
    const killMatch = trimmedLine.match(
        /LogGameplayEvents: Display: (.+?)\[(\d+|INVALID), team (\d+)\] killed (.+?)\[(\d+|INVALID), team (\d+)\] with (.+)$/
    );
    if (
        killMatch &&
        killMatch[1] &&
        killMatch[2] &&
        killMatch[3] &&
        killMatch[4] &&
        killMatch[5] &&
        killMatch[6] &&
        killMatch[7]
    ) {
        const [, killer, killerSteamId, killerTeam, victim, victimSteamId, victimTeam, weapon] = killMatch;

        let eventType: "player_kill" | "team_kill" | "suicide" = "player_kill";

        // Check if it's a suicide (killer and victim are the same)
        if (killer === victim && killerSteamId === victimSteamId) {
            eventType = "suicide";
        }
        // Check if it's a team kill (same team but different players)
        else if (killerTeam === victimTeam && killer !== victim) {
            eventType = "team_kill";
        }

        return {
            type: eventType,
            timestamp,
            data: {
                killer,
                killerSteamId,
                killerTeam: parseInt(killerTeam),
                victim,
                victimSteamId,
                victimTeam: parseInt(victimTeam),
                weapon: getCleanWeaponName(weapon),
            },
            rawLine: trimmedLine,
        };
    }

    // Round Over
    const roundOverMatch = trimmedLine.match(
        /LogGameplayEvents: Display: Round (\d+) Over: Team (\d+) won \(win reason: (.+)\)$/
    );
    if (roundOverMatch?.[1] && roundOverMatch?.[2] && roundOverMatch?.[3]) {
        return {
            type: "round_over",
            timestamp,
            data: {
                roundNumber: parseInt(roundOverMatch[1]),
                winningTeam: parseInt(roundOverMatch[2]),
                winReason: roundOverMatch[3],
            },
            rawLine: trimmedLine,
        };
    }

    // Map Load
    const mapLoadMatch = trimmedLine.match(
        /LogLoad: LoadMap: \/Game\/Maps\/(.+?)\/(.+?)\?.*?Scenario=Scenario_(.+?)_(.+?)_(.+?)\?.*?MaxPlayers=(\d+)\?.*?Lighting=(.+?)(?:\?|$)/
    );
    if (
        mapLoadMatch &&
        mapLoadMatch[1] &&
        mapLoadMatch[3] &&
        mapLoadMatch[4] &&
        mapLoadMatch[5] &&
        mapLoadMatch[6] &&
        mapLoadMatch[7]
    ) {
        const [, mapFolder, , scenario, mode, team, maxPlayers, lighting] = mapLoadMatch;
        return {
            type: "map_load",
            timestamp,
            data: {
                mapName: mapFolder,
                scenario: `${scenario}_${mode}`,
                team,
                maxPlayers: parseInt(maxPlayers),
                lighting,
            },
            rawLine: trimmedLine,
        };
    }

    // Difficulty Set
    const difficultyMatch = trimmedLine.match(/LogAI: Warning: AI difficulty set to ([\d.]+)$/);
    if (difficultyMatch?.[1]) {
        return {
            type: "difficulty_set",
            timestamp,
            data: {
                difficulty: parseFloat(difficultyMatch[1]),
            },
            rawLine: trimmedLine,
        };
    }

    // Chat Commands
    const chatMatch = trimmedLine.match(/LogChat: Display: (.+?)\((\d+)\) Global Chat: (!.+)$/);
    if (chatMatch?.[1] && chatMatch?.[2] && chatMatch?.[3]) {
        const [, playerName, steamId, message] = chatMatch;
        const commandParts = message.split(" ");
        const command = commandParts[0];
        const args = commandParts.slice(1);

        return {
            type: "chat_command",
            timestamp,
            data: {
                playerName,
                steamId,
                command,
                args: args.length > 0 ? args : undefined,
            },
            rawLine: trimmedLine,
        };
    }

    // Fall damage (can lead to suicide)
    if (trimmedLine.includes("LogSoldier: Applying") && trimmedLine.includes("fall damage")) {
        const fallDamageMatch = trimmedLine.match(/LogSoldier: Applying ([\d.]+) fall damage/);
        if (fallDamageMatch?.[1]) {
            return {
                type: "fall_damage",
                timestamp,
                data: {
                    damage: parseFloat(fallDamageMatch[1]),
                },
                rawLine: trimmedLine,
            };
        }
    }

    // Map Vote (start of voting)
    if (
        trimmedLine.includes("LogMapVoteManager: Display: Existing Vote Options:") ||
        trimmedLine.includes("LogMapVoteManager: Display: New Vote Options:")
    ) {
        return {
            type: "map_vote_start",
            timestamp,
            data: {},
            rawLine: trimmedLine,
        };
    }

    return null;
}

/**
 * Parses multiple log lines and returns an array of events
 * @param logContent - Multiple log lines (e.g., from tail command)
 * @returns Array of parsed GameEvent objects
 */
export function parseLogEvents(logContent: string): GameEvent[] {
    const lines = logContent.split("\n");
    const events: GameEvent[] = [];

    for (const line of lines) {
        const event = parseLogEvent(line);
        if (event) {
            events.push(event);
        }
    }

    return events;
}

/**
 * Helper function to filter events by type
 * @param events - Array of GameEvent objects
 * @param eventType - Type of event to filter for
 * @returns Filtered array of events
 */
export function filterEventsByType<T extends GameEvent>(events: GameEvent[], eventType: string): T[] {
    return events.filter((event) => event.type === eventType) as T[];
}

/**
 * Helper function to get events within a time range
 * @param events - Array of GameEvent objects
 * @param startTime - Start timestamp (YYYY.MM.DD-HH.MM.SS:mmm format)
 * @param endTime - End timestamp (YYYY.MM.DD-HH.MM.SS:mmm format)
 * @returns Filtered array of events within the time range
 */
export function filterEventsByTimeRange(events: GameEvent[], startTime: string, endTime: string): GameEvent[] {
    return events.filter((event) => {
        return event.timestamp >= startTime && event.timestamp <= endTime;
    });
}

/**
 * Parse log events from a file with automatic server ID detection
 * @param logContent - Raw log file content
 * @param logFilePath - Full path to the log file (used to extract server ID)
 * @returns Object containing parsed events and detected server ID
 *
 * @example
 * const result = parseLogEventsFromFile(content, "/path/to/logs/server1.log");
 * console.log(result.serverId); // "server1"
 * console.log(result.events); // Array of GameEvent objects
 */
export function parseLogEventsFromFile(
    logContent: string,
    logFilePath: string
): {
    events: GameEvent[];
    serverId: string | null;
} {
    // Import PathUtils here to avoid circular dependencies
    const { PathUtils } = require("./cross-platform-utils");

    const serverId = PathUtils.extractServerIdFromLogPath(logFilePath);
    const events = parseLogEvents(logContent);

    return {
        events,
        serverId,
    };
}
