export class ServerConfig {
    public readonly id: string;
    public readonly name: string;
    public readonly logPath: string;
    public readonly serverId: string;
    public readonly enabled: boolean;
    public readonly description?: string;

    constructor(data: {
        id: string;
        name: string;
        logPath: string;
        serverId: string;
        enabled: boolean;
        description?: string;
    }) {
        // Sanitize and validate inputs
        this.id = this.sanitizeString(data.id, "Server ID");
        this.name = this.sanitizeString(data.name, "Server name");
        this.logPath = this.sanitizePath(data.logPath);
        this.serverId = this.validateUUID(data.serverId);
        this.enabled = Boolean(data.enabled);
        this.description = data.description ? this.sanitizeString(data.description, "Description", true) : undefined;
    }

    private sanitizeString(value: string, fieldName: string, optional = false): string {
        if (!value && !optional) {
            throw new Error(`${fieldName} is required`);
        }
        if (!value && optional) {
            return "";
        }
        const sanitized = value.trim();
        if (!sanitized && !optional) {
            throw new Error(`${fieldName} cannot be empty`);
        }
        if (sanitized.length > 200) {
            throw new Error(`${fieldName} cannot exceed 200 characters`);
        }
        // Remove potentially dangerous characters
        return sanitized.replace(/[<>\"'&]/g, "");
    }

    private sanitizePath(path: string): string {
        if (!path) {
            throw new Error("Log path is required");
        }
        const sanitized = path.trim();
        if (!sanitized) {
            throw new Error("Log path cannot be empty");
        }
        // Basic path validation - allow common path characters
        if (!/^[a-zA-Z0-9\\\/:._\- ]+$/.test(sanitized)) {
            throw new Error("Log path contains invalid characters");
        }
        return sanitized;
    }

    private validateUUID(uuid: string): string {
        if (!uuid) {
            throw new Error("Server UUID is required");
        }
        const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
        if (!uuidRegex.test(uuid)) {
            throw new Error("Server UUID must be a valid UUID format");
        }
        return uuid.toLowerCase();
    }

    // Convert to plain object for serialization
    toJSON() {
        return {
            id: this.id,
            name: this.name,
            logPath: this.logPath,
            serverId: this.serverId,
            enabled: this.enabled,
            description: this.description,
        };
    }

    // Create from plain object
    static fromJSON(data: any): ServerConfig {
        return new ServerConfig(data);
    }
}

export class DatabaseConfig {
    public readonly path: string;
    public readonly enableWAL: boolean;
    public readonly cacheSize: number;

    constructor(data: { path: string; enableWAL: boolean; cacheSize: number }) {
        this.path = this.sanitizePath(data.path);
        this.enableWAL = Boolean(data.enableWAL);
        this.cacheSize = this.validateCacheSize(data.cacheSize);
    }

    private sanitizePath(path: string): string {
        if (!path) {
            throw new Error("Database path is required");
        }
        const sanitized = path.trim();
        if (!sanitized) {
            throw new Error("Database path cannot be empty");
        }
        // Basic validation for database filename
        if (!/^[a-zA-Z0-9\\\/:._\- ]+\.db$/.test(sanitized)) {
            throw new Error("Database path must end with .db and contain only valid characters");
        }
        return sanitized;
    }

    private validateCacheSize(size: number): number {
        const cacheSize = Number(size);
        if (isNaN(cacheSize) || cacheSize < 100 || cacheSize > 10000) {
            throw new Error("Cache size must be between 100 and 10000");
        }
        return cacheSize;
    }

    toJSON() {
        return {
            path: this.path,
            enableWAL: this.enableWAL,
            cacheSize: this.cacheSize,
        };
    }

    static fromJSON(data: any): DatabaseConfig {
        return new DatabaseConfig(data);
    }
}

export class LoggingConfig {
    public readonly level: "debug" | "info" | "warn" | "error";
    public readonly enableServerLogs: boolean;

    constructor(data: { level: "debug" | "info" | "warn" | "error"; enableServerLogs: boolean }) {
        this.level = this.validateLogLevel(data.level);
        this.enableServerLogs = Boolean(data.enableServerLogs);
    }

    private validateLogLevel(level: string): "debug" | "info" | "warn" | "error" {
        const validLevels: Array<"debug" | "info" | "warn" | "error"> = ["debug", "info", "warn", "error"];
        if (!validLevels.includes(level as any)) {
            throw new Error(`Invalid log level: ${level}. Must be one of: ${validLevels.join(", ")}`);
        }
        return level as "debug" | "info" | "warn" | "error";
    }

    toJSON() {
        return {
            level: this.level,
            enableServerLogs: this.enableServerLogs,
        };
    }

    static fromJSON(data: any): LoggingConfig {
        return new LoggingConfig(data);
    }
}

export class PerformanceConfig {
    public readonly debounceDelay: number;
    public readonly maxConcurrentWatchers: number;

    constructor(data: { debounceDelay: number; maxConcurrentWatchers: number }) {
        this.debounceDelay = this.validateDebounceDelay(data.debounceDelay);
        this.maxConcurrentWatchers = this.validateMaxWatchers(data.maxConcurrentWatchers);
    }

    private validateDebounceDelay(delay: number): number {
        const debounceDelay = Number(delay);
        if (isNaN(debounceDelay) || debounceDelay < 50 || debounceDelay > 5000) {
            throw new Error("Debounce delay must be between 50 and 5000 milliseconds");
        }
        return debounceDelay;
    }

    private validateMaxWatchers(watchers: number): number {
        const maxWatchers = Number(watchers);
        if (isNaN(maxWatchers) || maxWatchers < 1 || maxWatchers > 50) {
            throw new Error("Max concurrent watchers must be between 1 and 50");
        }
        return maxWatchers;
    }

    toJSON() {
        return {
            debounceDelay: this.debounceDelay,
            maxConcurrentWatchers: this.maxConcurrentWatchers,
        };
    }

    static fromJSON(data: any): PerformanceConfig {
        return new PerformanceConfig(data);
    }
}

export class AppConfig {
    public readonly servers: ServerConfig[];
    public readonly database: DatabaseConfig;
    public readonly logging: LoggingConfig;
    public readonly performance: PerformanceConfig;

    constructor(data: { servers: any[]; database: any; logging: any; performance: any }) {
        this.servers = this.validateServers(data.servers);
        this.database = new DatabaseConfig(data.database);
        this.logging = new LoggingConfig(data.logging);
        this.performance = new PerformanceConfig(data.performance);
    }

    private validateServers(servers: any[]): ServerConfig[] {
        if (!Array.isArray(servers)) {
            throw new Error("Servers must be an array");
        }
        if (servers.length === 0) {
            throw new Error("At least one server must be configured");
        }
        if (servers.length > 10) {
            throw new Error("Maximum of 10 servers supported for performance reasons");
        }

        const serverConfigs = servers.map((server, index) => {
            try {
                return new ServerConfig(server);
            } catch (error) {
                throw new Error(
                    `Server ${index + 1} validation error: ${error instanceof Error ? error.message : error}`
                );
            }
        });

        // Check for duplicate server IDs
        const serverIds = serverConfigs.map((s) => s.id);
        const duplicateIds = serverIds.filter((id, index) => serverIds.indexOf(id) !== index);
        if (duplicateIds.length > 0) {
            throw new Error(`Duplicate server IDs found: ${duplicateIds.join(", ")}`);
        }

        // Check for duplicate server UUIDs
        const serverUuids = serverConfigs.map((s) => s.serverId);
        const duplicateUuids = serverUuids.filter((uuid, index) => serverUuids.indexOf(uuid) !== index);
        if (duplicateUuids.length > 0) {
            throw new Error(`Duplicate server UUIDs found: ${duplicateUuids.join(", ")}`);
        }

        return serverConfigs;
    }

    toJSON() {
        return {
            servers: this.servers.map((s) => s.toJSON()),
            database: this.database.toJSON(),
            logging: this.logging.toJSON(),
            performance: this.performance.toJSON(),
        };
    }

    static fromJSON(data: any): AppConfig {
        return new AppConfig(data);
    }
}

// Default configuration - created using classes for validation
export const defaultConfig: AppConfig = new AppConfig({
    servers: [
        {
            id: "server-1",
            name: "Main Server",
            logPath: "C:\\Users\\danie\\code\\sandstorm-admin-wrapper\\sandstorm-server\\Insurgency\\Saved\\Logs",
            serverId: "60844f66-b93b-4fe1-afc4-a0a91b493865",
            enabled: true,
            description: "Primary Sandstorm server",
        },
    ],
    database: {
        path: "sandstorm_stats.db",
        enableWAL: true,
        cacheSize: 1000,
    },
    logging: {
        level: "info",
        enableServerLogs: true,
    },
    performance: {
        debounceDelay: 100,
        maxConcurrentWatchers: 10,
    },
});

// Configuration validation functions (legacy support - validation is now in constructors)
export function validateServerConfig(server: any): string[] {
    try {
        new ServerConfig(server);
        return [];
    } catch (error) {
        return [error instanceof Error ? error.message : String(error)];
    }
}

export function validateConfig(config: any): string[] {
    try {
        new AppConfig(config);
        return [];
    } catch (error) {
        return [error instanceof Error ? error.message : String(error)];
    }
}

// Helper function to create config from JSON with proper validation
export function createConfigFromJSON(jsonData: any): AppConfig {
    return new AppConfig(jsonData);
}

// Helper function to create server config with validation
export function createServerConfig(data: any): ServerConfig {
    return new ServerConfig(data);
}
