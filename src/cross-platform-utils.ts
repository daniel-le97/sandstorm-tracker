import { platform } from "os";
import { join, normalize, resolve, sep } from "path";

/**
 * Cross-platform path utilities for handling server log paths and configuration
 */
export class PathUtils {
    /**
     * Normalize path separators for the current platform
     */
    static normalizePath(inputPath: string): string {
        // Handle empty or null paths
        if (!inputPath || inputPath.trim() === "") {
            return "";
        }

        // Normalize the path using Node.js path module
        let normalizedPath = normalize(inputPath.trim());

        // Ensure consistent path separators
        if (platform() === "win32") {
            // On Windows, convert forward slashes to backslashes
            normalizedPath = normalizedPath.replace(/\//g, "\\");
        } else {
            // On Unix-like systems, convert backslashes to forward slashes
            normalizedPath = normalizedPath.replace(/\\/g, "/");
        }

        return normalizedPath;
    }

    /**
     * Join path components using the appropriate separator for the current platform
     */
    static joinPaths(...paths: string[]): string {
        if (paths.length === 0) return "";

        // Filter out empty paths and normalize each component
        const cleanPaths = paths.filter((p) => p && p.trim() !== "").map((p) => this.normalizePath(p));

        return join(...cleanPaths);
    }

    /**
     * Resolve a path to an absolute path
     */
    static resolvePath(inputPath: string): string {
        return resolve(this.normalizePath(inputPath));
    }

    /**
     * Get the platform-appropriate path separator
     */
    static getPathSeparator(): string {
        return sep;
    }

    /**
     * Check if a path is absolute
     */
    static isAbsolutePath(inputPath: string): boolean {
        const normalizedPath = this.normalizePath(inputPath);

        if (platform() === "win32") {
            // Windows: Check for drive letter (C:\) or UNC path (\\server\)
            return /^[A-Za-z]:\\/.test(normalizedPath) || /^\\\\/.test(normalizedPath);
        } else {
            // Unix-like: Check if starts with /
            return normalizedPath.startsWith("/");
        }
    }

    /**
     * Convert a Windows-style path to Unix-style or vice versa
     */
    static convertPathStyle(
        inputPath: string,
        targetPlatform: "win32" | "posix" = platform() as "win32" | "posix"
    ): string {
        const normalized = this.normalizePath(inputPath);

        if (targetPlatform === "win32") {
            return normalized.replace(/\//g, "\\");
        } else {
            return normalized.replace(/\\/g, "/");
        }
    }

    /**
     * Generate example paths for different platforms
     */
    static getExampleLogPaths(): { windows: string; mac: string; linux: string } {
        return {
            windows: "C:\\Program Files\\Steam\\steamapps\\common\\sandstorm-server\\Insurgency\\Saved\\Logs",
            mac: "/Users/username/Library/Application Support/Steam/steamapps/common/sandstorm-server/Insurgency/Saved/Logs",
            linux: "/home/username/.local/share/Steam/steamapps/common/sandstorm-server/Insurgency/Saved/Logs",
        };
    }

    /**
     * Get common Sandstorm server installation paths by platform
     */
    static getCommonServerPaths(): { windows: string[]; mac: string[]; linux: string[] } {
        return {
            windows: [
                "C:\\Program Files (x86)\\Steam\\steamapps\\common\\sandstorm-server",
                "C:\\Program Files\\Steam\\steamapps\\common\\sandstorm-server",
                "D:\\Steam\\steamapps\\common\\sandstorm-server",
                "C:\\SandstormServer",
                "C:\\GameServers\\Sandstorm",
            ],
            mac: [
                "/Users/Shared/Steam/steamapps/common/sandstorm-server",
                "/Applications/Steam/steamapps/common/sandstorm-server",
                "/opt/sandstorm-server",
                "~/Library/Application Support/Steam/steamapps/common/sandstorm-server",
            ],
            linux: [
                "/home/steam/.local/share/Steam/steamapps/common/sandstorm-server",
                "/opt/sandstorm-server",
                "/usr/local/games/sandstorm-server",
                "/srv/sandstorm-server",
                "~/steamcmd/sandstorm-server",
            ],
        };
    }

    /**
     * Validate that a path looks like a valid Sandstorm log directory
     */
    static isValidSandstormLogPath(inputPath: string): boolean {
        const normalized = this.normalizePath(inputPath);

        // Should end with something like "Insurgency/Saved/Logs" or "Insurgency\\Saved\\Logs"
        const pathPattern = /[Ii]nsurgency[\\\/][Ss]aved[\\\/][Ll]ogs\s*$/;
        return pathPattern.test(normalized);
    }

    /**
     * Get platform-specific environment variable names for paths
     */
    static getPathEnvironmentVariables(): string[] {
        if (platform() === "win32") {
            return ["PATH", "PATHEXT"];
        } else {
            return ["PATH", "LD_LIBRARY_PATH"];
        }
    }

    /**
     * Extract server ID from a log file path
     * Expects log files to be named as "{server_id}.log"
     *
     * @param logFilePath - Full path to the log file or just the filename
     * @returns Server ID string or null if not found
     *
     * @example
     * extractServerIdFromLogPath("/path/to/logs/server1.log") // Returns "server1"
     * extractServerIdFromLogPath("C:\\Logs\\my-server.log") // Returns "my-server"
     * extractServerIdFromLogPath("invalid.txt") // Returns null
     */
    static extractServerIdFromLogPath(logFilePath: string): string | null {
        if (!logFilePath || typeof logFilePath !== "string") {
            return null;
        }

        // Normalize the path and extract just the filename
        const normalizedPath = this.normalizePath(logFilePath.trim());
        const filename = normalizedPath.split(/[\\\/]/).pop(); // Get last segment (filename)

        if (!filename) {
            return null;
        }

        // Check if filename ends with .log
        if (!filename.toLowerCase().endsWith(".log")) {
            return null;
        }

        // Extract server ID by removing the .log extension
        const serverId = filename.slice(0, -4); // Remove last 4 characters (.log)

        // Validate server ID (should not be empty and should contain valid characters)
        if (!serverId || serverId.trim() === "") {
            return null;
        }

        // Optional: Add validation for server ID format if needed
        // For now, just return the extracted ID
        return serverId.trim();
    }

    /**
     * Validate if a string is a valid server ID format
     * Server IDs should be alphanumeric with hyphens/underscores allowed
     */
    static isValidServerId(serverId: string): boolean {
        if (!serverId || typeof serverId !== "string") {
            return false;
        }

        const trimmed = serverId.trim();

        // Server ID should be 1-50 characters, alphanumeric plus hyphens and underscores
        const serverIdPattern = /^[a-zA-Z0-9][a-zA-Z0-9_-]{0,49}$/;
        return serverIdPattern.test(trimmed);
    }

    /**
     * Get the current platform information
     */
    static getPlatformInfo(): {
        platform: string;
        isWindows: boolean;
        isMac: boolean;
        isLinux: boolean;
        pathSeparator: string;
        homeDirectory: string;
    } {
        const currentPlatform = platform();
        const homeDir = process.env.HOME || process.env.USERPROFILE || process.env.HOMEPATH || "";

        return {
            platform: currentPlatform,
            isWindows: currentPlatform === "win32",
            isMac: currentPlatform === "darwin",
            isLinux: currentPlatform === "linux",
            pathSeparator: sep,
            homeDirectory: homeDir,
        };
    }

    /**
     * Create a cross-platform file URL from a path
     */
    static pathToFileURL(inputPath: string): string {
        const resolved = this.resolvePath(inputPath);

        if (platform() === "win32") {
            // Windows: Convert backslashes to forward slashes and add file:// protocol
            return "file:///" + resolved.replace(/\\/g, "/").replace(/^[A-Za-z]:/, (match) => match.toLowerCase());
        } else {
            // Unix-like: Just add file:// protocol
            return "file://" + resolved;
        }
    }

    /**
     * Get default database paths by platform
     */
    static getDefaultDatabasePaths(): { windows: string; mac: string; linux: string } {
        const platformInfo = this.getPlatformInfo();

        return {
            windows: this.joinPaths(
                platformInfo.homeDirectory,
                "AppData",
                "Local",
                "SandstormTracker",
                "sandstorm_stats.db"
            ),
            mac: this.joinPaths(
                platformInfo.homeDirectory,
                "Library",
                "Application Support",
                "SandstormTracker",
                "sandstorm_stats.db"
            ),
            linux: this.joinPaths(
                platformInfo.homeDirectory,
                ".local",
                "share",
                "SandstormTracker",
                "sandstorm_stats.db"
            ),
        };
    }
}

/**
 * Environment variable utilities for cross-platform configuration
 */
export class EnvUtils {
    /**
     * Get an environment variable with cross-platform fallbacks
     */
    static getEnvVar(name: string, fallbacks: string[] = [], defaultValue: string = ""): string {
        // Try the primary name first
        let value = process.env[name];
        if (value !== undefined && value !== "") {
            return value.trim();
        }

        // Try fallback names
        for (const fallback of fallbacks) {
            value = process.env[fallback];
            if (value !== undefined && value !== "") {
                return value.trim();
            }
        }

        return defaultValue;
    }

    /**
     * Set an environment variable if it doesn't already exist
     */
    static setEnvDefault(name: string, value: string): void {
        if (process.env[name] === undefined || process.env[name] === "") {
            process.env[name] = value;
        }
    }

    /**
     * Get boolean value from environment variable
     */
    static getEnvBoolean(name: string, defaultValue: boolean = false): boolean {
        const value = this.getEnvVar(name).toLowerCase();

        if (value === "" && defaultValue !== undefined) {
            return defaultValue;
        }

        return value === "true" || value === "1" || value === "yes" || value === "on";
    }

    /**
     * Get numeric value from environment variable
     */
    static getEnvNumber(name: string, defaultValue: number = 0): number {
        const value = this.getEnvVar(name);

        if (value === "") {
            return defaultValue;
        }

        const parsed = parseInt(value, 10);
        return isNaN(parsed) ? defaultValue : parsed;
    }

    /**
     * Parse environment variables for server configuration with cross-platform paths
     */
    static parseServerEnvironmentVars(): Record<string, any> {
        const servers: Record<string, any> = {};
        const serverPattern = /^SANDSTORM_SERVER_(\d+)_(.+)$/;

        for (const [key, value] of Object.entries(process.env)) {
            const match = key.match(serverPattern);
            if (match && value !== undefined) {
                const [, serverIndex, property] = match;
                const serverKey = `server_${serverIndex}`;

                if (!servers[serverKey]) {
                    servers[serverKey] = {};
                }

                // Normalize path values
                if (property === "LOG_PATH" && value) {
                    servers[serverKey][property.toLowerCase()] = PathUtils.normalizePath(value);
                } else if (property) {
                    servers[serverKey][property.toLowerCase()] = value;
                }
            }
        }

        return servers;
    }

    /**
     * Get home directory with cross-platform support
     */
    static getHomeDirectory(): string {
        return PathUtils.getPlatformInfo().homeDirectory;
    }

    /**
     * Get temp directory with cross-platform support
     */
    static getTempDirectory(): string {
        return this.getEnvVar("TMPDIR", ["TEMP", "TMP"], "/tmp");
    }
}

export default { PathUtils, EnvUtils };
