import { afterAll, afterEach, beforeAll, beforeEach, describe, expect, test } from "bun:test";
import { mkdirSync, rmSync, unlinkSync } from "fs";
import { join } from "path";
import { AppConfig } from "../src/config";
import { ConfigLoader, ConfigurationError } from "../src/config-loader";

describe( "Configuration Loader", () => {
    const testDir = "./test-configs";
    const validJsonConfig = join( testDir, "valid.json" );
    const validTomlConfig = join( testDir, "valid.toml" );
    const invalidJsonConfig = join( testDir, "invalid.json" );
    const invalidTomlConfig = join( testDir, "invalid.toml" );
    const testLogDir = join( testDir, "logs" );

    beforeAll( () => {
        // Create test directory structure
        try
        {
            mkdirSync( testDir, { recursive: true } );
            mkdirSync( testLogDir, { recursive: true } );
        } catch ( e )
        {
            // Directory might already exist
        }

        // Create valid JSON config
        const validJsonContent = {
            servers: [
                {
                    id: "test-server-1",
                    name: "Test Server 1",
                    logPath: testLogDir,
                    serverId: "60844f66-b93b-4fe1-afc4-a0a91b493865",
                    enabled: true,
                    description: "Test server for config loader tests",
                },
            ],
            database: {
                path: "test_config_loader.db",
                enableWAL: true,
                cacheSize: 1000,
            },
            logging: {
                level: "info",
                enableServerLogs: true,
            },
            performance: {
                debounceDelay: 100,
                maxConcurrentWatchers: 5,
            },
        };

        // Create valid TOML config (escape backslashes for Windows paths)
        const validTomlContent = `[[servers]]
id = "test-server-1"
name = "Test Server 1"
logPath = "${ testLogDir.replace( /\\/g, '\\\\' ) }"
serverId = "60844f66-b93b-4fe1-afc4-a0a91b493865"
enabled = true
description = "Test server for config loader tests"

[database]
path = "test_config_loader.db"
enableWAL = true
cacheSize = 1000

[logging]
level = "info"
enableServerLogs = true

[performance]
debounceDelay = 100
maxConcurrentWatchers = 5`;

        // Create invalid JSON config (malformed JSON)
        const invalidJsonContent = `{
            "servers": [
                {
                    "id": "test-server-1",
                    "name": "Test Server 1"
                    // Missing comma and other required fields
                }
            }`;

        // Create invalid TOML config (malformed TOML)
        const invalidTomlContent = `[servers
id = "test-server-1"
missing_bracket = true`;

        // Write test config files
        Bun.write( validJsonConfig, JSON.stringify( validJsonContent, null, 2 ) );
        Bun.write( validTomlConfig, validTomlContent );
        Bun.write( invalidJsonConfig, invalidJsonContent );
        Bun.write( invalidTomlConfig, invalidTomlContent );
    } );

    beforeEach( () => {
        // Reset config loader state before each test
        ConfigLoader.reset();

        // Clear environment variables
        delete process.env.SANDSTORM_CONFIG_PATH;
        delete process.env.SANDSTORM_DB_PATH;
        delete process.env.SANDSTORM_LOG_LEVEL;
        delete process.env.SANDSTORM_MAX_CONCURRENT_SERVERS;
    } );

    afterEach( () => {
        // Clean up any test database files
        try
        {
            unlinkSync( "test_config_loader.db" );
        } catch ( e )
        {
            // File doesn't exist, that's fine
        }
    } );

    describe( "JSON Configuration Loading", () => {
        test( "should load valid JSON configuration", async () => {
            const config = await ConfigLoader.loadConfig( validJsonConfig );
            expect( config ).toBeInstanceOf( AppConfig );
            expect( config.servers ).toHaveLength( 1 );
            expect( config.servers[ 0 ].id ).toBe( "test-server-1" );
            expect( config.servers[ 0 ].name ).toBe( "Test Server 1" );
            expect( config.servers[ 0 ].enabled ).toBe( true );
            expect( config.database.path ).toBe( "test_config_loader.db" );
            expect( config.logging.level ).toBe( "info" );
        } );

        test( "should throw error for invalid JSON syntax", async () => {
            await expect( ConfigLoader.loadConfig( invalidJsonConfig ) ).rejects.toThrow( ConfigurationError );
        } );

        test( "should detect JSON format correctly", () => {
            expect( ConfigLoader.isSupportedFormat( "config.json" ) ).toBe( true );
            expect( ConfigLoader.isSupportedFormat( "CONFIG.JSON" ) ).toBe( true );
        } );
    } );

    describe( "TOML Configuration Loading", () => {
        test( "should load valid TOML configuration", async () => {
            const config = await ConfigLoader.loadConfig( validTomlConfig );

            expect( config ).toBeInstanceOf( AppConfig );
            expect( config.servers ).toHaveLength( 1 );
            expect( config.servers[ 0 ].id ).toBe( "test-server-1" );
            expect( config.servers[ 0 ].name ).toBe( "Test Server 1" );
            expect( config.servers[ 0 ].enabled ).toBe( true );
            expect( config.database.path ).toBe( "test_config_loader.db" );
            expect( config.logging.level ).toBe( "info" );
        } );

        test( "should throw error for invalid TOML syntax", async () => {
            await expect( ConfigLoader.loadConfig( invalidTomlConfig ) ).rejects.toThrow( ConfigurationError );
        } );

        test( "should detect TOML format correctly", () => {
            expect( ConfigLoader.isSupportedFormat( "config.toml" ) ).toBe( true );
            expect( ConfigLoader.isSupportedFormat( "CONFIG.TOML" ) ).toBe( true );
        } );
    } );

    describe( "Environment Variable Overrides", () => {
        test( "should apply database path override", async () => {
            process.env.SANDSTORM_DB_PATH = "override_database.db";

            const config = await ConfigLoader.loadConfig( validJsonConfig );
            expect( config.database.path ).toBe( "override_database.db" );
        } );

        test( "should apply log level override", async () => {
            process.env.SANDSTORM_LOG_LEVEL = "debug";

            const config = await ConfigLoader.loadConfig( validJsonConfig );
            expect( config.logging.level ).toBe( "debug" );
        } );

        test( "should apply max concurrent servers override", async () => {
            process.env.SANDSTORM_MAX_CONCURRENT_SERVERS = "15";

            const config = await ConfigLoader.loadConfig( validJsonConfig );
            expect( config.performance.maxConcurrentWatchers ).toBe( 15 );
        } );

        test( "should ignore invalid log level override", async () => {
            process.env.SANDSTORM_LOG_LEVEL = "invalid_level";

            const config = await ConfigLoader.loadConfig( validJsonConfig );
            expect( config.logging.level ).toBe( "info" ); // Should keep original
        } );

        test( "should use environment config path", async () => {
            process.env.SANDSTORM_CONFIG_PATH = validTomlConfig;

            const config = await ConfigLoader.loadConfig();
            expect( config.servers[ 0 ].id ).toBe( "test-server-1" );
        } );

        test( "should require servers when no config file and no environment servers", async () => {
            // Ensure no config path and no server env vars are present
            delete process.env.SANDSTORM_CONFIG_PATH;
            delete process.env.SANDSTORM_SERVER_1_ID;
            delete process.env.SANDSTORM_SERVER_1_NAME;
            delete process.env.SANDSTORM_SERVER_1_LOG_PATH;
            delete process.env.SANDSTORM_SERVER_1_SERVER_ID;
            delete process.env.SANDSTORM_SERVER_1_ENABLED;

            // Do not remove any repo-level config files; tests must not delete
            // user configuration present in the repository.

            await expect( ConfigLoader.loadConfig() ).rejects.toThrow( "No configuration file found" );
        } );

        test( "should load servers from environment variables when provided", async () => {
            // Provide a valid set of environment variables for one server
            process.env.SANDSTORM_SERVER_1_ID = "env-server-1";
            process.env.SANDSTORM_SERVER_1_NAME = "Env Server 1";
            process.env.SANDSTORM_SERVER_1_LOG_PATH = testLogDir;
            process.env.SANDSTORM_SERVER_1_SERVER_ID = "60844f66-b93b-4fe1-afc4-a0a91b493866";
            process.env.SANDSTORM_SERVER_1_ENABLED = "true";

            // Ensure no explicit config path is used
            delete process.env.SANDSTORM_CONFIG_PATH;

            // Do not remove any repo-level config files; tests must not delete
            // user configuration present in the repository.

            const config = await ConfigLoader.loadConfig();
            expect( config ).toBeInstanceOf( AppConfig );
            expect( config.servers ).toHaveLength( 1 );
            expect( config.servers[ 0 ].id ).toBe( "env-server-1" );

            // Clean up env vars for subsequent tests
            delete process.env.SANDSTORM_SERVER_1_ID;
            delete process.env.SANDSTORM_SERVER_1_NAME;
            delete process.env.SANDSTORM_SERVER_1_LOG_PATH;
            delete process.env.SANDSTORM_SERVER_1_SERVER_ID;
            delete process.env.SANDSTORM_SERVER_1_ENABLED;
        } );
    } );

    describe( "Configuration Validation", () => {
        test( "should validate valid JSON config file", async () => {
            const validation = await ConfigLoader.validateConfigFile( validJsonConfig );
            expect( validation.valid ).toBe( true );
            expect( validation.errors ).toHaveLength( 0 );
        } );

        test( "should validate valid TOML config file", async () => {
            const validation = await ConfigLoader.validateConfigFile( validTomlConfig );
            expect( validation.valid ).toBe( true );
            expect( validation.errors ).toHaveLength( 0 );
        } );

        test( "should detect invalid JSON config", async () => {
            const validation = await ConfigLoader.validateConfigFile( invalidJsonConfig );
            expect( validation.valid ).toBe( false );
            expect( validation.errors.length ).toBeGreaterThan( 0 );
        } );

        test( "should detect invalid TOML config", async () => {
            const validation = await ConfigLoader.validateConfigFile( invalidTomlConfig );
            expect( validation.valid ).toBe( false );
            expect( validation.errors.length ).toBeGreaterThan( 0 );
        } );

        test( "should handle non-existent config file", async () => {
            const validation = await ConfigLoader.validateConfigFile( "non-existent.json" );
            expect( validation.valid ).toBe( false );
            expect( validation.errors.length ).toBeGreaterThan( 0 );
        } );
    } );

    describe( "Sample Configuration Generation", () => {
        const sampleJsonPath = join( testDir, "sample.json" );
        const sampleTomlPath = join( testDir, "sample.toml" );

        afterEach( () => {
            try
            {
                unlinkSync( sampleJsonPath );
                unlinkSync( sampleTomlPath );
            } catch ( e )
            {
                // Files might not exist
            }
        } );

        test( "should create sample JSON configuration", async () => {
            await ConfigLoader.createSampleConfig( sampleJsonPath, "json" );

            const file = Bun.file( sampleJsonPath );
            expect( await file.exists() ).toBe( true );

            const content = await file.text();
            expect( () => JSON.parse( content ) ).not.toThrow();

            // Validate the generated config structure is valid JSON and parseable
            const validation = await ConfigLoader.validateConfigFile( sampleJsonPath );
            // The config structure should be valid (validateConfigFile doesn't check paths)
            expect( validation.valid ).toBe( true );
        } );

        test( "should create sample TOML configuration", async () => {
            await ConfigLoader.createSampleConfig( sampleTomlPath, "toml" );

            const file = Bun.file( sampleTomlPath );
            expect( await file.exists() ).toBe( true );

            const content = await file.text();
            expect( content ).toContain( "[[servers]]" );
            expect( content ).toContain( "[database]" );
            expect( content ).toContain( "[logging]" );
            expect( content ).toContain( "[performance]" );
        } );
    } );

    describe( "Configuration Saving", () => {
        const saveJsonPath = join( testDir, "saved.json" );
        const saveTomlPath = join( testDir, "saved.toml" );

        afterEach( () => {
            try
            {
                unlinkSync( saveJsonPath );
                unlinkSync( saveTomlPath );
            } catch ( e )
            {
                // Files might not exist
            }
        } );

        test( "should save configuration as JSON", async () => {
            // Load a config first
            await ConfigLoader.loadConfig( validJsonConfig );

            // Save it
            await ConfigLoader.saveConfig( saveJsonPath, "json" );

            const file = Bun.file( saveJsonPath );
            expect( await file.exists() ).toBe( true );

            const content = await file.text();
            const parsed = JSON.parse( content );
            expect( parsed.servers ).toHaveLength( 1 );
            expect( parsed.servers[ 0 ].id ).toBe( "test-server-1" );
        } );

        test( "should save configuration as TOML", async () => {
            // Load a config first
            await ConfigLoader.loadConfig( validJsonConfig );

            // Save it
            await ConfigLoader.saveConfig( saveTomlPath, "toml" );

            const file = Bun.file( saveTomlPath );
            expect( await file.exists() ).toBe( true );

            const content = await file.text();
            expect( content ).toContain( "[[servers]]" );
            expect( content ).toContain( 'id = "test-server-1"' );
        } );

        test( "should throw error when saving without loaded config", async () => {
            await expect( ConfigLoader.saveConfig( saveJsonPath ) ).rejects.toThrow( "No configuration loaded" );
        } );
    } );

    describe( "Format Detection and Support", () => {
        test( "should list supported formats", () => {
            const formats = ConfigLoader.getSupportedFormats();
            expect( formats ).toContain( "json" );
            expect( formats ).toContain( "toml" );
        } );

        test( "should detect supported formats correctly", () => {
            expect( ConfigLoader.isSupportedFormat( "config.json" ) ).toBe( true );
            expect( ConfigLoader.isSupportedFormat( "config.toml" ) ).toBe( true );
            expect( ConfigLoader.isSupportedFormat( "config.yaml" ) ).toBe( false );
            expect( ConfigLoader.isSupportedFormat( "config.xml" ) ).toBe( false );
            expect( ConfigLoader.isSupportedFormat( "config.txt" ) ).toBe( false );
        } );
    } );

    describe( "Configuration Access Methods", () => {
        test( "should get current configuration", async () => {
            expect( ConfigLoader.getConfig() ).toBeNull();

            await ConfigLoader.loadConfig( validJsonConfig );
            const config = ConfigLoader.getConfig();
            expect( config ).not.toBeNull();
            expect( config?.servers ).toHaveLength( 1 );
        } );

        test( "should get enabled servers only", async () => {
            await ConfigLoader.loadConfig( validJsonConfig );
            const enabledServers = ConfigLoader.getEnabledServers();
            expect( enabledServers ).toHaveLength( 1 );
            expect( enabledServers[ 0 ].enabled ).toBe( true );
        } );

        test( "should get server by ID", async () => {
            await ConfigLoader.loadConfig( validJsonConfig );
            const server = ConfigLoader.getServerById( "test-server-1" );
            expect( server ).not.toBeUndefined();
            expect( server?.name ).toBe( "Test Server 1" );

            const nonExistentServer = ConfigLoader.getServerById( "non-existent" );
            expect( nonExistentServer ).toBeUndefined();
        } );

        test( "should get server by UUID", async () => {
            await ConfigLoader.loadConfig( validJsonConfig );
            const server = ConfigLoader.getServerByUuid( "60844f66-b93b-4fe1-afc4-a0a91b493865" );
            expect( server ).not.toBeUndefined();
            expect( server?.id ).toBe( "test-server-1" );

            const nonExistentServer = ConfigLoader.getServerByUuid( "00000000-0000-0000-0000-000000000000" );
            expect( nonExistentServer ).toBeUndefined();
        } );

        test( "should throw error accessing methods without loaded config", () => {
            expect( () => ConfigLoader.getEnabledServers() ).toThrow( "Configuration not loaded" );
            expect( () => ConfigLoader.getServerById( "test" ) ).toThrow( "Configuration not loaded" );
            expect( () => ConfigLoader.getServerByUuid( "test" ) ).toThrow( "Configuration not loaded" );
        } );
    } );

    describe( "Error Handling", () => {
        test( "should handle non-existent config file", async () => {
            await expect( ConfigLoader.loadConfig( "non-existent-file.json" ) ).rejects.toThrow( ConfigurationError );
        } );

        test( "should handle config validation errors", async () => {
            const invalidConfigPath = join( testDir, "validation-error.json" );
            const invalidConfig = {
                servers: [], // Empty servers array should fail validation
                database: { path: "test.db", enableWAL: true, cacheSize: 1000 },
                logging: { level: "info", enableServerLogs: true },
                performance: { debounceDelay: 100, maxConcurrentWatchers: 5 },
            };

            await Bun.write( invalidConfigPath, JSON.stringify( invalidConfig ) );

            await expect( ConfigLoader.loadConfig( invalidConfigPath ) ).rejects.toThrow( ConfigurationError );

            // Clean up
            unlinkSync( invalidConfigPath );
        } );

        test( "should handle path validation errors gracefully", async () => {
            const configWithBadPath = join( testDir, "bad-path.json" );
            const badPathConfig = {
                servers: [
                    {
                        id: "test-server",
                        name: "Test Server",
                        logPath: "/this/path/does/not/exist/and/never/will",
                        serverId: "60844f66-b93b-4fe1-afc4-a0a91b493865",
                        enabled: true,
                        description: "Test server with bad path",
                    },
                ],
                database: { path: "test.db", enableWAL: true, cacheSize: 1000 },
                logging: { level: "info", enableServerLogs: true },
                performance: { debounceDelay: 100, maxConcurrentWatchers: 5 },
            };

            await Bun.write( configWithBadPath, JSON.stringify( badPathConfig ) );

            await expect( ConfigLoader.loadConfig( configWithBadPath ) ).rejects.toThrow( ConfigurationError );

            // Clean up
            unlinkSync( configWithBadPath );
        } );
    } );

    describe( "Singleton Behavior", () => {
        test( "should return cached configuration on subsequent calls", async () => {
            const config1 = await ConfigLoader.loadConfig( validJsonConfig );
            const config2 = await ConfigLoader.loadConfig( validTomlConfig ); // Different file, should return cached

            expect( config1 ).toBe( config2 ); // Same instance
        } );

        test( "should reset configuration properly", async () => {
            await ConfigLoader.loadConfig( validJsonConfig );
            expect( ConfigLoader.getConfig() ).not.toBeNull();

            ConfigLoader.reset();
            expect( ConfigLoader.getConfig() ).toBeNull();
        } );
    } );

    // Clean up test files after all tests
    afterAll( () => {
        try
        {
            rmSync( testDir, { recursive: true, force: true } );
        } catch ( e )
        {
            // Directory cleanup failed, not critical
        }
    } );
} );
