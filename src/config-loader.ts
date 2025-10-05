import { access, constants } from "fs/promises";
import { AppConfig, ServerConfig, createConfigFromJSON, defaultConfig } from "./config";
import { EnvUtils, PathUtils } from "./cross-platform-utils";

export class ConfigurationError extends Error {
    constructor ( message: string, public details?: string[] ) {
        super( message );
        this.name = "ConfigurationError";
    }
}

export type ConfigFormat = "json" | "toml" | "auto";

export class ConfigLoader {
    private static config: AppConfig | null = null;
    private static readonly SUPPORTED_EXTENSIONS = [ ".json", ".toml" ];

    /**
     * Create a sample configuration file
     */
    static async createSampleConfig ( filePath: string, format: "json" | "toml" = "json" ): Promise<void> {
        const sampleConfig = {
            servers: [
                {
                    id: "server-1",
                    name: "Main Server",
                    logPath:
                        process.platform === "win32"
                            ? "C:\\SteamCMD\\steamapps\\common\\sandstorm_server\\Insurgency\\Saved\\Logs"
                            : "/opt/sandstorm/Insurgency/Saved/Logs",
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
        };

        let content: string;
        if ( format === "toml" )
        {
            content = this.convertToTOML( sampleConfig );
        } else
        {
            content = JSON.stringify( sampleConfig, null, 2 );
        }

        await Bun.write( filePath, content );
        console.log( `Created sample configuration file: ${ filePath }` );
    }

    /**
     * Convert config object to TOML format (simple implementation)
     */
    private static convertToTOML ( obj: any ): string {
        let toml = "";

        // Add servers array
        if ( obj.servers && Array.isArray( obj.servers ) )
        {
            obj.servers.forEach( ( server: any, index: number ) => {
                toml += `[[servers]]\n`;
                toml += `id = "${ server.id }"\n`;
                toml += `name = "${ server.name }"\n`;
                toml += `logPath = "${ server.logPath }"\n`;
                toml += `serverId = "${ server.serverId }"\n`;
                toml += `enabled = ${ server.enabled }\n`;
                if ( server.description )
                {
                    toml += `description = "${ server.description }"\n`;
                }
                toml += `\n`;
            } );
        }

        // Add database section
        if ( obj.database )
        {
            toml += `[database]\n`;
            toml += `path = "${ obj.database.path }"\n`;
            toml += `enableWAL = ${ obj.database.enableWAL }\n`;
            toml += `cacheSize = ${ obj.database.cacheSize }\n\n`;
        }

        // Add logging section
        if ( obj.logging )
        {
            toml += `[logging]\n`;
            toml += `level = "${ obj.logging.level }"\n`;
            toml += `enableServerLogs = ${ obj.logging.enableServerLogs }\n\n`;
        }

        // Add performance section
        if ( obj.performance )
        {
            toml += `[performance]\n`;
            toml += `debounceDelay = ${ obj.performance.debounceDelay }\n`;
            toml += `maxConcurrentWatchers = ${ obj.performance.maxConcurrentWatchers }\n\n`;
        }

        return toml;
    }

    /**
     * Load configuration with support for JSON and TOML formats
     */
    static async loadConfig ( configPath?: string ): Promise<AppConfig> {
        if ( this.config )
        {
            return this.config;
        }

        try
        {
            // Use provided path or environment variable
            let finalConfigPath = configPath || process.env.SANDSTORM_CONFIG_PATH;
            let config: AppConfig;

            // If no explicit path, check for both TOML and JSON in default location
            if ( !finalConfigPath )
            {
                const tomlPath = "sandstorm-config.toml";
                const jsonPath = "sandstorm-config.json";
                const tomlExists = await Bun.file( tomlPath ).exists();
                const jsonExists = await Bun.file( jsonPath ).exists();
                if ( tomlExists && jsonExists )
                {
                    console.warn( `Both TOML (${ tomlPath }) and JSON (${ jsonPath }) config files found. Using TOML.` );
                    finalConfigPath = tomlPath;
                } else if ( tomlExists )
                {
                    finalConfigPath = tomlPath;
                } else if ( jsonExists )
                {
                    finalConfigPath = jsonPath;
                }
            }

            if ( finalConfigPath )
            {
                config = await this.loadFromFile( finalConfigPath );
            } else
            {
                // Start with default config
                config = defaultConfig;

                // Load servers from environment variables if available
                const envServers = await this.loadServersFromEnvironment();
                if ( envServers.length > 0 )
                {
                    // Create new config with environment servers
                    config = new AppConfig( {
                        servers: envServers,
                        database: config.database.toJSON(),
                        logging: config.logging.toJSON(),
                        performance: config.performance.toJSON(),
                    } );
                }
            }

            // Override database settings from environment
            config = this.applyEnvironmentOverrides( config );

            // Configuration is already validated by classes

            // Validate server paths exist and are accessible
            await this.validateServerPaths( config.servers );

            // Validate database path is writable
            await this.validateDatabasePath( config.database.path );

            this.config = config;
            console.log( `Loaded and validated configuration with ${ config.servers.length } server(s)` );
            return config;
        } catch ( error )
        {
            if ( error instanceof ConfigurationError )
            {
                console.error( "Configuration Error:", error.message );
                if ( error.details )
                {
                    error.details.forEach( ( detail ) => console.error( "  -", detail ) );
                }
            } else
            {
                console.error( "Unexpected error loading configuration:", error );
            }
            throw error;
        }
    }

    /**
     * Load configuration from file with auto-detection of format (JSON or TOML)
     */
    private static async loadFromFile ( configPath: string ): Promise<AppConfig> {
        try
        {
            // Check if file exists and is readable
            await access( configPath, constants.R_OK );

            const format = this.detectConfigFormat( configPath );
            const file = Bun.file( configPath );
            const configText = await file.text();

            let configData: any;

            if ( format === "toml" )
            {
                configData = Bun.TOML.parse( configText );
                console.log( `Loaded TOML configuration from ${ configPath }` );
            } else
            {
                configData = JSON.parse( configText );
                console.log( `Loaded JSON configuration from ${ configPath }` );
            }

            // Create config using class constructor for validation
            return createConfigFromJSON( configData );
        } catch ( error )
        {
            if ( error instanceof SyntaxError )
            {
                throw new ConfigurationError(
                    `Invalid ${ this.detectConfigFormat( configPath ).toUpperCase() } in config file: ${ configPath }`,
                    [ error.message ]
                );
            }
            throw new ConfigurationError( `Failed to load config file: ${ configPath }`, [ String( error ) ] );
        }
    }

    /**
     * Detect configuration format from file extension
     */
    private static detectConfigFormat ( configPath: string ): "json" | "toml" {
        const extension = configPath.toLowerCase().split( "." ).pop();
        return extension === "toml" ? "toml" : "json";
    }

    /**
     * Legacy method for backward compatibility
     */
    private static async loadFromJsonFile ( configPath: string ): Promise<AppConfig> {
        return this.loadFromFile( configPath );
    }

    /**
     * Apply environment variable overrides to configuration
     */
    private static applyEnvironmentOverrides ( config: AppConfig ): AppConfig {
        const configData = config.toJSON();

        if ( process.env.SANDSTORM_DB_PATH )
        {
            configData.database.path = process.env.SANDSTORM_DB_PATH;
        }

        if ( process.env.SANDSTORM_LOG_LEVEL )
        {
            const level = process.env.SANDSTORM_LOG_LEVEL.toLowerCase();
            if ( [ "debug", "info", "warn", "error" ].includes( level ) )
            {
                configData.logging.level = level as "debug" | "info" | "warn" | "error";
            } else
            {
                console.warn( `Invalid log level: ${ level }. Using default: info` );
            }
        }

        if ( process.env.SANDSTORM_MAX_CONCURRENT_SERVERS )
        {
            const maxConcurrent = parseInt( process.env.SANDSTORM_MAX_CONCURRENT_SERVERS );
            if ( !isNaN( maxConcurrent ) && maxConcurrent > 0 )
            {
                configData.performance.maxConcurrentWatchers = maxConcurrent;
            }
        }

        // Return new config instance with overrides
        return new AppConfig( configData );
    }

    /**
     * Validate that all server log paths exist and are accessible with cross-platform support
     */
    private static async validateServerPaths ( servers: ServerConfig[] ): Promise<void> {
        const errors: string[] = [];
        const warnings: string[] = [];

        for ( const server of servers.filter( ( s ) => s.enabled ) )
        {
            try
            {
                // Normalize the path for the current platform
                const normalizedPath = PathUtils.normalizePath( server.logPath );

                // Validate path format
                if ( !PathUtils.isValidSandstormLogPath( normalizedPath ) )
                {
                    warnings.push( `Server "${ server.name }": Path format may be incorrect - ${ normalizedPath }` );
                }

                // Check if path is absolute (recommended)
                if ( !PathUtils.isAbsolutePath( normalizedPath ) )
                {
                    warnings.push(
                        `Server "${ server.name }": Relative path detected, absolute paths recommended - ${ normalizedPath }`
                    );
                }

                // Test accessibility
                await access( normalizedPath, constants.R_OK );
                console.log( `✓ Server "${ server.name }" log path validated: ${ normalizedPath }` );
            } catch ( error )
            {
                const suggestion = this.getSuggestedPaths( server.logPath );
                errors.push(
                    `Server "${ server.name }" (${ server.id }): Log path not accessible - ${ server.logPath }${ suggestion }`
                );
            }
        }

        // Show warnings but don't fail validation
        if ( warnings.length > 0 )
        {
            console.warn( "\nConfiguration Warnings:" );
            warnings.forEach( ( warning ) => console.warn( `  - ${ warning }` ) );
        }

        if ( errors.length > 0 )
        {
            throw new ConfigurationError( "Server path validation failed", errors );
        }
    }

    /**
     * Get suggested path corrections for common issues
     */
    private static getSuggestedPaths ( originalPath: string ): string {
        const platformInfo = PathUtils.getPlatformInfo();
        const commonPaths = PathUtils.getCommonServerPaths();

        let suggestions = "\n    Common paths for your platform:";

        if ( platformInfo.isWindows )
        {
            suggestions += "\n    - " + commonPaths.windows.slice( 0, 2 ).join( "\n    - " );
        } else if ( platformInfo.isMac )
        {
            suggestions += "\n    - " + commonPaths.mac.slice( 0, 2 ).join( "\n    - " );
        } else
        {
            suggestions += "\n    - " + commonPaths.linux.slice( 0, 2 ).join( "\n    - " );
        }

        return suggestions;
    }

    /**
     * Validate database directory exists or can be created
     */
    private static async validateDatabasePath ( dbPath: string ): Promise<void> {
        try
        {
            const dbDir = dbPath.substring( 0, dbPath.lastIndexOf( "/" ) || dbPath.lastIndexOf( "\\" ) );
            if ( dbDir )
            {
                // Check if directory exists or try to access parent
                try
                {
                    await access( dbDir, constants.W_OK );
                } catch
                {
                    // Directory might not exist, check parent or current dir
                    await access( ".", constants.W_OK );
                }
            }
            console.log( `✓ Database path validated: ${ dbPath }` );
        } catch ( error )
        {
            throw new ConfigurationError( `Database path not writable: ${ dbPath }`, [ String( error ) ] );
        }
    }

    /**
     * Load server configurations from environment variables with validation
     * Format: SANDSTORM_SERVER_<INDEX>_<PROPERTY>=value
     */
    private static async loadServersFromEnvironment (): Promise<ServerConfig[]> {
        const servers: ServerConfig[] = [];
        const serverIndexes = new Set<number>();
        const errors: string[] = [];

        // Find all server environment variables
        if ( typeof process !== "undefined" && process.env )
        {
            Object.keys( process.env ).forEach( ( key ) => {
                const match = key.match( /^SANDSTORM_SERVER_(\d+)_/ );
                if ( match && match[ 1 ] )
                {
                    const index = parseInt( match[ 1 ] );
                    if ( index > 0 && index <= 50 )
                    {
                        // Reasonable limit
                        serverIndexes.add( index );
                    } else
                    {
                        errors.push( `Invalid server index: ${ index }. Must be between 1 and 50.` );
                    }
                }
            } );

            // Build server configs with validation
            for ( const index of Array.from( serverIndexes ).sort() )
            {
                const prefix = `SANDSTORM_SERVER_${ index }_`;

                const id = EnvUtils.getEnvVar( `${ prefix }ID` );
                const name = EnvUtils.getEnvVar( `${ prefix }NAME` );
                const rawLogPath = EnvUtils.getEnvVar( `${ prefix }LOG_PATH` );
                const logPath = rawLogPath ? PathUtils.normalizePath( rawLogPath ) : "";
                const serverId = EnvUtils.getEnvVar( `${ prefix }SERVER_ID` );
                const enabled = EnvUtils.getEnvBoolean( `${ prefix }ENABLED`, true );
                const description = EnvUtils.getEnvVar( `${ prefix }DESCRIPTION` );

                // Validate required fields
                const missingFields: string[] = [];
                if ( !id ) missingFields.push( "ID" );
                if ( !name ) missingFields.push( "NAME" );
                if ( !logPath ) missingFields.push( "LOG_PATH" );
                if ( !serverId ) missingFields.push( "SERVER_ID" );

                if ( missingFields.length > 0 )
                {
                    errors.push( `Server ${ index }: Missing required fields: ${ missingFields.join( ", " ) }` );
                    continue;
                }

                // Validate server ID format (should be UUID)
                if ( !this.isValidUuid( serverId! ) )
                {
                    errors.push( `Server ${ index }: SERVER_ID must be a valid UUID, got: ${ serverId }` );
                    continue;
                }

                // Check for duplicate IDs
                if ( servers.some( ( s ) => s.id === id ) )
                {
                    errors.push( `Server ${ index }: Duplicate server ID: ${ id }` );
                    continue;
                }

                if ( servers.some( ( s ) => s.serverId === serverId ) )
                {
                    errors.push( `Server ${ index }: Duplicate server UUID: ${ serverId }` );
                    continue;
                }

                servers.push(
                    new ServerConfig( {
                        id: id!,
                        name: name!,
                        logPath: logPath!,
                        serverId: serverId!,
                        enabled,
                        description,
                    } )
                );

                console.log( `Server ${ index } (${ name }): ${ enabled ? "Enabled" : "Disabled" }` );
            }
        }

        if ( errors.length > 0 )
        {
            throw new ConfigurationError( "Environment server configuration errors", errors );
        }

        return servers;
    }

    /**
     * Validate UUID format
     */
    private static isValidUuid ( uuid: string ): boolean {
        const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
        return uuidRegex.test( uuid );
    }

    /**
     * Get the current configuration
     */
    static getConfig (): AppConfig | null {
        return this.config;
    }

    /**
     * Reset configuration (useful for testing)
     */
    static reset (): void {
        this.config = null;
    }

    /**
     * Get enabled servers only
     */
    static getEnabledServers (): ServerConfig[] {
        if ( !this.config )
        {
            throw new Error( "Configuration not loaded. Call loadConfig() first." );
        }
        return this.config.servers.filter( ( server ) => server.enabled );
    }

    /**
     * Get server configuration by ID
     */
    static getServerById ( serverId: string ): ServerConfig | undefined {
        if ( !this.config )
        {
            throw new Error( "Configuration not loaded. Call loadConfig() first." );
        }
        return this.config.servers.find( ( server ) => server.id === serverId );
    }

    /**
     * Get server configuration by server UUID
     */
    static getServerByUuid ( serverUuid: string ): ServerConfig | undefined {
        if ( !this.config )
        {
            throw new Error( "Configuration not loaded. Call loadConfig() first." );
        }
        return this.config.servers.find( ( server ) => server.serverId === serverUuid );
    }

    /**
     * Save current configuration to file
     */
    static async saveConfig ( filePath: string, format: "json" | "toml" = "json" ): Promise<void> {
        if ( !this.config )
        {
            throw new Error( "No configuration loaded. Call loadConfig() first." );
        }

        const configData = this.config.toJSON();
        let content: string;

        if ( format === "toml" )
        {
            content = this.convertToTOML( configData );
        } else
        {
            content = JSON.stringify( configData, null, 2 );
        }

        await Bun.write( filePath, content );
        console.log( `Saved configuration to: ${ filePath }` );
    }

    /**
     * Validate a configuration file without loading it
     */
    static async validateConfigFile ( filePath: string ): Promise<{ valid: boolean; errors: string[]; }> {
        try
        {
            await access( filePath, constants.R_OK );

            const format = this.detectConfigFormat( filePath );
            const file = Bun.file( filePath );
            const configText = await file.text();

            let configData: any;

            if ( format === "toml" )
            {
                configData = Bun.TOML.parse( configText );
            } else
            {
                configData = JSON.parse( configText );
            }

            // Try to create config to validate
            createConfigFromJSON( configData );

            return { valid: true, errors: [] };
        } catch ( error )
        {
            return {
                valid: false,
                errors: [ error instanceof Error ? error.message : String( error ) ],
            };
        }
    }

    /**
     * List supported configuration file formats
     */
    static getSupportedFormats (): string[] {
        return [ "json", "toml" ];
    }

    /**
     * Check if a file extension is supported
     */
    static isSupportedFormat ( filePath: string ): boolean {
        const extension = `.${ filePath.toLowerCase().split( "." ).pop() }`;
        return this.SUPPORTED_EXTENSIONS.includes( extension );
    }
}
