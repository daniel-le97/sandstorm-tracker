import { watch } from "fs/promises";
import CommandHandler from "./src/command-handler";
import type { ServerConfig } from "./src/config";
import { ConfigLoader } from "./src/config-loader";
import { upsertServer } from "./src/database";
import type { ChatEvent, GameEvent } from "./src/events";
import { parseLogEvents } from "./src/events";
import StatsService from "./src/stats-service";

// Server watcher state with health monitoring
interface ServerWatcher {
    config: ServerConfig;
    serverId: number; // Database ID
    lastProcessedTime: number;
    lastProcessedLine: string;
    fileSize: number;
    logFilePath: string;
    isHealthy: boolean;
    lastHealthCheck: number;
    errorCount: number;
    lastError: Error | null;
}

const serverWatchers = new Map<string, ServerWatcher>();
const debounceDelay = 100; // 100ms debounce
const HEALTH_CHECK_INTERVAL = 30000; // 30 seconds
const MAX_ERROR_COUNT = 5;

/**
 * Retry function with exponential backoff
 */
async function retryWithBackoff<T>(fn: () => Promise<T>, maxRetries: number = 3, baseDelay: number = 1000): Promise<T> {
    let lastError: Error;

    for (let attempt = 1; attempt <= maxRetries; attempt++) {
        try {
            return await fn();
        } catch (error) {
            lastError = error instanceof Error ? error : new Error(String(error));

            if (attempt === maxRetries) {
                throw lastError;
            }

            const delay = baseDelay * Math.pow(2, attempt - 1);
            console.warn(`⏳ Retry attempt ${attempt}/${maxRetries} failed, waiting ${delay}ms...`);
            await new Promise((resolve) => setTimeout(resolve, delay));
        }
    }

    throw lastError!;
}

/**
 * Schedule periodic health checks for all server watchers
 */
function scheduleHealthChecks(): void {
    setInterval(() => {
        performHealthChecks().catch((error) => {
            console.error("Health check failed:", error);
        });
    }, HEALTH_CHECK_INTERVAL);

    console.log(`Health checks scheduled every ${HEALTH_CHECK_INTERVAL / 1000}s`);
}

/**
 * Perform health checks on all server watchers
 */
async function performHealthChecks(): Promise<void> {
    const unhealthyServers: string[] = [];

    for (const [serverId, watcher] of serverWatchers) {
        try {
            // Check if log file is still accessible
            const file = Bun.file(watcher.logFilePath);
            await file.exists();

            // Check if we've had recent activity or too many errors
            const timeSinceLastActivity = Date.now() - watcher.lastProcessedTime;
            const wasHealthy = watcher.isHealthy;

            watcher.isHealthy = watcher.errorCount < MAX_ERROR_COUNT && timeSinceLastActivity < 300000; // 5 minutes
            watcher.lastHealthCheck = Date.now();

            if (!watcher.isHealthy && wasHealthy) {
                console.warn(`Server "${watcher.config.name}" marked as unhealthy (errors: ${watcher.errorCount})`);
                unhealthyServers.push(watcher.config.name);
            } else if (watcher.isHealthy && !wasHealthy) {
                console.log(`Server "${watcher.config.name}" recovered and marked as healthy`);
                watcher.errorCount = 0;
                watcher.lastError = null;
            }
        } catch (error) {
            watcher.isHealthy = false;
            watcher.errorCount++;
            watcher.lastError = error instanceof Error ? error : new Error(String(error));
            watcher.lastHealthCheck = Date.now();

            if (!unhealthyServers.includes(watcher.config.name)) {
                unhealthyServers.push(watcher.config.name);
            }
        }
    }

    if (unhealthyServers.length > 0) {
        console.warn(`Unhealthy servers detected: ${unhealthyServers.join(", ")}`);
    }
}

// Initialize multi-server application with comprehensive error handling
async function initializeApplication(): Promise<void> {
    console.log("Initializing Sandstorm Multi-Server Tracker...");

    try {
        // Load and validate configuration
        const config = await ConfigLoader.loadConfig();
        const enabledServers = ConfigLoader.getEnabledServers();

        if (enabledServers.length === 0) {
            console.error("No enabled servers found in configuration!");
            console.error("Please check your configuration and ensure at least one server is enabled.");
            process.exit(1);
        }

        console.log(`Found ${enabledServers.length} enabled server(s)`);

        // Initialize stats service with error handling
        try {
            StatsService.initialize();
            console.log("✓ Statistics service initialized");
        } catch (error) {
            console.error("Failed to initialize statistics service:", error);
            throw error;
        }

        // Set up each server watcher with individual error handling
        const setupPromises = enabledServers.map(async (serverConfig) => {
            try {
                await setupServerWatcher(serverConfig);
                return { server: serverConfig.name, success: true };
            } catch (error) {
                console.error(`Failed to setup watcher for server "${serverConfig.name}":`, error);
                return { server: serverConfig.name, success: false, error };
            }
        });

        const results = await Promise.allSettled(setupPromises);
        const successful = results.filter((r) => r.status === "fulfilled" && r.value.success).length;
        const failed = results.length - successful;

        if (successful === 0) {
            console.error("All server watchers failed to initialize!");
            process.exit(1);
        }

        console.log(`Multi-server tracker initialized successfully!`);
        console.log(`${successful} server(s) active, ${failed} failed`);
        console.log("Watching for file changes on active servers...");

        // Schedule periodic health checks
        scheduleHealthChecks();
    } catch (error) {
        console.error("Critical error during initialization:", error);
        console.error("Application cannot continue. Please check your configuration and try again.");
        process.exit(1);
    }
}

async function setupServerWatcher(serverConfig: ServerConfig): Promise<void> {
    try {
        console.log(`Setting up watcher for server: ${serverConfig.name} (${serverConfig.id})`);

        // Register server in database
        const dbServerId = upsertServer(
            serverConfig.serverId,
            serverConfig.name,
            serverConfig.id,
            serverConfig.logPath,
            serverConfig.description
        );

        const logFilePath = `${serverConfig.logPath}${
            serverConfig.logPath.endsWith("\\") || serverConfig.logPath.endsWith("/") ? "" : "\\"
        }${serverConfig.serverId}.log`;

        // Initialize file size
        let fileSize = 0;
        try {
            const stats = await Bun.file(logFilePath).stat();
            fileSize = stats.size;
        } catch (error) {
            console.log(`Log file doesn't exist yet for ${serverConfig.name}: ${logFilePath}`);
        }

        // Create watcher state
        const watcher: ServerWatcher = {
            config: serverConfig,
            serverId: dbServerId,
            lastProcessedTime: 0,
            lastProcessedLine: "",
            fileSize: fileSize,
            logFilePath: logFilePath,
            isHealthy: true,
            lastHealthCheck: Date.now(),
            errorCount: 0,
            lastError: null,
        };

        serverWatchers.set(serverConfig.id, watcher);

        console.log(`Server watcher configured: ${serverConfig.name}`);
    } catch (error) {
        console.error(`Failed to setup watcher for ${serverConfig.name}:`, error);
    }
}

// Start watching all configured servers
async function startWatching(): Promise<void> {
    const watchPromises: Promise<void>[] = [];

    // Group servers by log path to avoid duplicate watchers
    const pathToServers = new Map<string, ServerWatcher[]>();

    for (const watcher of serverWatchers.values()) {
        const logDir = watcher.config.logPath;
        if (!pathToServers.has(logDir)) {
            pathToServers.set(logDir, []);
        }
        pathToServers.get(logDir)!.push(watcher);
    }

    // Create one file system watcher per unique log directory
    for (const [logPath, watchers] of pathToServers) {
        // Start watching for each server in this directory
        for (const watcher of watchers) {
            watchPromises.push(watchLogDirectory(watcher.config.id));
        }
    }

    // Wait for all watchers (they run indefinitely)
    await Promise.all(watchPromises);
}

async function watchLogDirectory(serverId: string): Promise<void> {
    const watcher = serverWatchers.get(serverId);
    if (!watcher) {
        console.error(`No watcher found for server ID: ${serverId}`);
        return;
    }

    const logPath = watcher.config.logPath;
    const watchers = Array.from(serverWatchers.values()).filter((w) => w.config.logPath === logPath);

    console.log(`� Starting file watcher for directory: ${logPath}`);

    let retryCount = 0;
    const maxRetries = 5;

    while (retryCount < maxRetries) {
        try {
            const dirWatcher = watch(logPath);
            console.log(`✓ Successfully watching: ${logPath}`);

            for await (const event of dirWatcher) {
                try {
                    // Skip backup files and other irrelevant files
                    if (!event.filename || event.filename.includes("backup") || event.filename.includes("tmp")) {
                        continue;
                    }

                    // Find which server this file belongs to
                    for (const serverWatcher of watchers) {
                        if (!serverWatcher.isHealthy) {
                            continue; // Skip unhealthy watchers
                        }

                        const expectedFilename = "Insurgency.log"; // Standard log filename

                        if (event.filename === expectedFilename) {
                            await processServerLogChange(serverWatcher);
                            break;
                        }
                    }
                } catch (processingError) {
                    const errorMsg =
                        processingError instanceof Error ? processingError.message : String(processingError);
                    console.error(`Error processing file event for ${event.filename}:`, errorMsg);

                    // Update watcher error state
                    for (const serverWatcher of watchers) {
                        if (serverWatcher.config.id === serverId) {
                            serverWatcher.errorCount++;
                            serverWatcher.lastError =
                                processingError instanceof Error ? processingError : new Error(String(processingError));
                            break;
                        }
                    }
                }
            }

            // If we get here, the watcher ended normally - shouldn't happen
            console.warn(`File watcher ended unexpectedly for ${logPath}`);
            break;
        } catch (error) {
            retryCount++;
            const errorMsg = error instanceof Error ? error.message : String(error);
            console.error(`File watcher error (attempt ${retryCount}/${maxRetries}) for ${logPath}:`, errorMsg);

            // Update watcher error state
            if (watcher) {
                watcher.errorCount++;
                watcher.lastError = error instanceof Error ? error : new Error(String(error));
                watcher.isHealthy = watcher.errorCount < MAX_ERROR_COUNT;
            }

            if (retryCount >= maxRetries) {
                console.error(`File watcher permanently failed for ${logPath} after ${maxRetries} attempts`);
                if (watcher) {
                    watcher.isHealthy = false;
                }
                return;
            }

            // Wait before retrying with exponential backoff
            const delay = Math.pow(2, retryCount) * 2000; // Start at 4s, max ~1 minute
            console.log(`⏳ Retrying file watcher in ${delay}ms...`);
            await new Promise((resolve) => setTimeout(resolve, delay));
        }
    }
}

async function processServerLogChange(watcher: ServerWatcher): Promise<void> {
    const now = Date.now();

    // Debounce rapid file changes
    if (now - watcher.lastProcessedTime < debounceDelay) {
        return;
    }

    // Skip if watcher is marked as unhealthy
    if (!watcher.isHealthy) {
        return;
    }

    try {
        // Check if file exists and is accessible
        const file = Bun.file(watcher.logFilePath);
        const exists = await file.exists();
        if (!exists) {
            console.warn(`⚠️ Log file disappeared: ${watcher.logFilePath}`);
            return;
        }

        // Check if file actually changed size (new content)
        const currentStats = await file.stat();
        if (currentStats.size <= watcher.fileSize) {
            return; // No new content
        }

        // Handle case where file was truncated (server restart)
        if (currentStats.size < watcher.fileSize) {
            console.log(`Log file truncated for ${watcher.config.name}, resetting position`);
            watcher.fileSize = 0;
            watcher.lastProcessedLine = "";
        }

        // Read just the last line to check if it's new
        const logContent = await file.text();
        const lines = logContent.split("\n");
        const lastLine = lines.filter((line: string) => line.trim()).pop();

        if (!lastLine || lastLine === watcher.lastProcessedLine) {
            watcher.fileSize = currentStats.size; // Update size even if no new line
            return; // Skip if no new content or same line as before
        }

        // Process the log file and parse events
        const gameEvents = await processLogFile(watcher, lastLine);

        // Update tracking variables successfully
        watcher.lastProcessedTime = now;
        watcher.fileSize = currentStats.size;
        watcher.lastProcessedLine = lastLine;

        // Reset error count on successful processing
        if (watcher.errorCount > 0) {
            console.log(`✓ Server ${watcher.config.name} recovered from errors`);
            watcher.errorCount = 0;
            watcher.lastError = null;
            watcher.isHealthy = true;
        }
    } catch (error) {
        const errorMsg = error instanceof Error ? error.message : String(error);
        console.error(`❌ Error processing log change for ${watcher.config.name}:`, errorMsg);

        // Update error state
        watcher.errorCount++;
        watcher.lastError = error instanceof Error ? error : new Error(String(error));
        watcher.isHealthy = watcher.errorCount < MAX_ERROR_COUNT;

        if (!watcher.isHealthy) {
            console.error(`Server "${watcher.config.name}" marked as unhealthy due to repeated errors`);
        }
    }
}

// Function to process a log line from a specific server
async function processLogFile(watcher: ServerWatcher, logLine: string): Promise<GameEvent[]> {
    try {
        // Parse just the provided log line
        const events = parseLogEvents(logLine);

        // Only process if we found actual events we care about
        if (events.length === 0) {
            return []; // Silent - no events we care about
        }

        // Update file activity timestamp
        StatsService.updateActivity(watcher.serverId);

        events.forEach((event) => {
            // Process event in database with server context
            StatsService.processEvent(event, watcher.serverId);

            // Only log events that are specifically listed in events.md
            const serverPrefix = `[${watcher.config.name}]`;

            switch (event.type) {
                case "game_over":
                    console.log(`${serverPrefix} Game over!`);
                    break;
                case "player_join":
                    console.log(`${serverPrefix} Player joined: ${event.data.playerName}`);
                    break;
                case "player_leave":
                case "player_disconnect":
                    console.log(`${serverPrefix} Player left: ${event.data.playerName || event.data.steamId}`);
                    break;
                case "team_kill":
                    console.log(
                        `${serverPrefix} TEAM KILL: ${event.data.killer} killed teammate ${event.data.victim} with ${event.data.weapon}`
                    );
                    break;
                case "player_kill":
                    console.log(
                        `${serverPrefix} ⚔️ ${event.data.killer} killed ${event.data.victim} with ${event.data.weapon}`
                    );
                    break;
                case "suicide":
                    console.log(`${serverPrefix} ${event.data.killer} committed suicide with ${event.data.weapon}`);
                    break;
                case "difficulty_set":
                    console.log(`${serverPrefix} ⚙️ AI difficulty set to: ${event.data.difficulty}`);
                    break;
                case "round_over":
                    console.log(
                        `${serverPrefix} Round ${event.data.roundNumber} over: Team ${event.data.winningTeam} won (${event.data.winReason})`
                    );
                    break;
                case "map_load":
                    console.log(`${serverPrefix} Loading map: ${event.data.mapName} (${event.data.scenario})`);
                    break;
                case "chat_command":
                    // Handle chat commands with server context
                    const response = CommandHandler.handleCommand(event as ChatEvent, watcher.serverId);
                    console.log(
                        `${serverPrefix} ${event.data.playerName} used command: ${event.data.command} ${
                            event.data.args?.join(" ") || ""
                        }`
                    );
                    if (response) {
                        console.log(`${serverPrefix} Response: ${response}`);
                    }
                    break;
                // Ignore any other events not listed in events.md
                default:
                    // Silent - don't log events we don't care about
                    break;
            }
        });

        return events;
    } catch (error) {
        console.error(`❌ Error processing log file for ${watcher.config.name}:`, error);
        return [];
    }
}

// Start the multi-server application
async function main(): Promise<void> {
    try {
        await initializeApplication();
        await startWatching();
    } catch (error) {
        console.error("❌ Failed to start multi-server tracker:", error);
        process.exit(1);
    }
}

// Start the application
main().catch(console.error);
