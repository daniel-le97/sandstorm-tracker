// Debug script for chat commands issue

// Set test environment
process.env.TEST_DB_PATH = "debug_chat.db";

// Import modules
const dbFunctions = await import("./src/database.ts");
const StatsService = (await import("./src/stats-service")).default;

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

console.log("Processing Bob join...");
StatsService.processEvent(
    {
        type: "player_join",
        timestamp: "2025.10.05-12.00.30:000",
        data: { playerName: "Bob" },
        rawLine: "test",
    },
    testServerDbId
);

// Check players exist
const aliceResult = dbFunctions.getStatements().getPlayerByName.get("Alice", testServerDbId);
console.log("Alice in DB:", aliceResult);

const bobResult = dbFunctions.getStatements().getPlayerByName.get("Bob", testServerDbId);
console.log("Bob in DB:", bobResult);

console.log("Processing kill event...");
try {
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
    console.log("Kill event processed successfully!");
} catch (error) {
    console.error("Kill event failed:", error);
}

// Check if Alice steam ID was updated
const aliceAfterKill = dbFunctions.getStatements().getPlayerByName.get("Alice", testServerDbId);
console.log("Alice after kill:", aliceAfterKill);

// Try to get Alice by Steam ID
const aliceBySteam = dbFunctions.getStatements().getPlayer.get("76561198000000001", testServerDbId);
console.log("Alice by Steam ID:", aliceBySteam);

// Try stats query
try {
    const stats = StatsService.getPlayerStats("76561198000000001", testServerDbId);
    console.log("Alice stats:", stats);
} catch (error) {
    console.error("Stats query failed:", error);
}
