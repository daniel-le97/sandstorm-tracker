/**
 * Available Insurgency: Sandstorm RCON commands
 * Based on server's 'help' command output - only includes commands that actually work
 */
export const SandstormCommands = {
    // Information Commands
    help: 'help',
    listPlayers: 'listplayers',
    listBans: 'listbans',
    maps: 'maps',
    scenarios: 'scenarios',
    listGamemodeProperties: 'listgamemodeproperties',

    // Player Management
    kick: ( player: string, reason?: string ) =>
        `kick ${ player }${ reason ? ` ${ reason }` : '' }`,
    ban: ( player: string, duration?: number, reason?: string ) =>
        `ban ${ player }${ duration ? ` ${ duration }` : '' }${ reason ? ` ${ reason }` : '' }`,
    permban: ( player: string, reason?: string ) =>
        `permban ${ player }${ reason ? ` ${ reason }` : '' }`,
    banid: ( netId: string, duration?: number, reason?: string ) =>
        `banid ${ netId }${ duration ? ` ${ duration }` : '' }${ reason ? ` ${ reason }` : '' }`,
    unban: ( netId: string ) => `unban ${ netId }`,

    // Server Control
    say: ( message: string ) => `say ${ message }`,
    restartRound: ( swapTeams: boolean = false ) => `restartround ${ swapTeams ? 1 : 0 }`,

    // Level/Map Control
    travel: ( travelUrl: string ) => `travel ${ travelUrl }`,
    travelScenario: ( scenario: string ) => `travelscenario ${ scenario }`,

    // Gamemode Properties
    getGamemodeProperty: ( property: string ) => `gamemodeproperty ${ property }`,
    setGamemodeProperty: ( property: string, value: string | number ) => `gamemodeproperty ${ property } ${ value }`,

    // Filtered queries
    helpFilter: ( filter: string ) => `help ${ filter }`,
    mapsFilter: ( filter: string ) => `maps ${ filter }`,
    scenariosFilter: ( filter: string ) => `scenarios ${ filter }`,
    gamemodePropertiesFilter: ( filter: string ) => `listgamemodeproperties ${ filter }`
} as const;
