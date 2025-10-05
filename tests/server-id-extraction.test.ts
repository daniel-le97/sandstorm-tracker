import { describe, expect, test } from "bun:test";
import { PathUtils } from "../src/cross-platform-utils";
import { parseLogEventsFromFile } from "../src/events";

describe("Server ID Extraction", () => {
    test("PathUtils.extractServerIdFromLogPath extracts server ID correctly", () => {
        // Test various log file paths
        expect(PathUtils.extractServerIdFromLogPath("server1.log")).toBe("server1");
        expect(PathUtils.extractServerIdFromLogPath("/path/to/logs/main-server.log")).toBe("main-server");
        expect(PathUtils.extractServerIdFromLogPath("C:\\Logs\\test_server.log")).toBe("test_server");
        expect(PathUtils.extractServerIdFromLogPath("/var/log/sandstorm/my-game-server.log")).toBe("my-game-server");

        // Test invalid cases
        expect(PathUtils.extractServerIdFromLogPath("notALogFile.txt")).toBe(null);
        expect(PathUtils.extractServerIdFromLogPath(".log")).toBe(null);
        expect(PathUtils.extractServerIdFromLogPath("")).toBe(null);
        expect(PathUtils.extractServerIdFromLogPath("/path/to/directory/")).toBe(null);

        // Test edge cases
        expect(PathUtils.extractServerIdFromLogPath("server.with.dots.log")).toBe("server.with.dots");
        expect(PathUtils.extractServerIdFromLogPath("   spaced-server.log   ")).toBe("spaced-server");
    });

    test("PathUtils.isValidServerId validates server IDs correctly", () => {
        // Valid server IDs
        expect(PathUtils.isValidServerId("server1")).toBe(true);
        expect(PathUtils.isValidServerId("main-server")).toBe(true);
        expect(PathUtils.isValidServerId("test_server")).toBe(true);
        expect(PathUtils.isValidServerId("Server123")).toBe(true);
        expect(PathUtils.isValidServerId("a")).toBe(true); // Single character

        // Invalid server IDs
        expect(PathUtils.isValidServerId("")).toBe(false);
        expect(PathUtils.isValidServerId("   ")).toBe(false);
        expect(PathUtils.isValidServerId("-invalid")).toBe(false); // Can't start with hyphen
        expect(PathUtils.isValidServerId("_invalid")).toBe(false); // Can't start with underscore
        expect(PathUtils.isValidServerId("server with spaces")).toBe(false);
        expect(PathUtils.isValidServerId("server@invalid")).toBe(false);
        expect(PathUtils.isValidServerId("a".repeat(51))).toBe(false); // Too long (51 chars)
    });

    test("parseLogEventsFromFile extracts server ID from file path", () => {
        const sampleLogContent = `[2024.10.05-12.30.45:123][456]LogGameplayEvents: Display: Game over
[2024.10.05-12.30.46:124][457]LogNet: Join succeeded: TestPlayer`;

        const result = parseLogEventsFromFile(sampleLogContent, "/path/to/logs/test-server.log");

        expect(result.serverId).toBe("test-server");
        expect(result.events).toHaveLength(2);
        expect(result.events[0].type).toBe("game_over");
        expect(result.events[1].type).toBe("player_join");
    });

    test("parseLogEventsFromFile handles invalid file paths gracefully", () => {
        const sampleLogContent = `[2024.10.05-12.30.45:123][456]LogGameplayEvents: Display: Game over`;

        const result = parseLogEventsFromFile(sampleLogContent, "invalid.txt");

        expect(result.serverId).toBe(null);
        expect(result.events).toHaveLength(1);
        expect(result.events[0].type).toBe("game_over");
    });
});

describe("Cross-Platform Path Handling", () => {
    test("Server ID extraction works on Windows paths", () => {
        const windowsPath =
            "C:\\Program Files\\Steam\\steamapps\\common\\sandstorm-server\\Insurgency\\Saved\\Logs\\main-server.log";
        expect(PathUtils.extractServerIdFromLogPath(windowsPath)).toBe("main-server");
    });

    test("Server ID extraction works on Unix paths", () => {
        const unixPath = "/home/user/sandstorm-server/Insurgency/Saved/Logs/production-server.log";
        expect(PathUtils.extractServerIdFromLogPath(unixPath)).toBe("production-server");
    });

    test("Server ID extraction handles mixed path separators", () => {
        const mixedPath = "C:/Games/Server\\Logs/event_server.log";
        expect(PathUtils.extractServerIdFromLogPath(mixedPath)).toBe("event_server");
    });
});
