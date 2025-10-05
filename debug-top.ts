// Debug !top command issue

// Set test environment
process.env.TEST_DB_PATH = "debug_top.db";

// Import modules
const dbFunctions = await import("./src/database.ts");
const StatsService = (await import("./src/stats-service")).default;
const CommandHandler = (await import("./src/command-handler")).default;

console.log("Creating test server...");
const testServerDbId = dbFunctions.upsertServer(
    "debug-server",
    "Debug Server",
    "debug-config",
    "/test/logs/debug-server.log",
    "Debug server for testing"
);
console.log("Server created with ID:", testServerDbId);

console.log("Processing Alice join...");
StatsService.processEvent(
    {
        type: "player_join",
        timestamp: "2025.10.05-12.00.00:000",
        data: { playerName: "Alice" },
        rawLine: "test",
    },
    testServerDbId
);

console.log("Processing kill event...");
StatsService.processEvent(
    {
        type: "player_kill",
        timestamp: "2025.10.05-12.01.00:000",
        data: {
            killer: "Alice",
            killerSteamId: "76561198000000001",
            killerTeam: 0,
            victim: "Bob",
            victimSteamId: "76561198000000002",
            victimTeam: 1,
            weapon: "M16A4",
        },
        rawLine: "test",
    },
    testServerDbId
);

console.log("Testing !top command...");
try {
    const result = StatsService.getTopPlayers(testServerDbId, 3);
    console.log("getTopPlayers result:", result);

    const response = CommandHandler.handleCommand(
        {
            type: "chat_command",
            timestamp: "2025.10.05-12.02.00:000",
            data: {
                playerName: "Alice",
                steamId: "76561198000000001",
                command: "!top",
                args: undefined,
            },
            rawLine: "test",
        },
        testServerDbId
    );
    console.log("!top response:", response);
} catch (error) {
    console.error("!top command failed:", error);
}
