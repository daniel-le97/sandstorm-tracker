#!/usr/bin/env bun
import { mkdir } from "fs/promises";
import { ConfigLoader } from "./src/config-loader";

async function testConfigLoader() {
    console.log("Testing Configuration Loader with JSON and TOML support\n");

    try {
        // Create mock log directories for testing
        const mockDirs = ["/opt/sandstorm/Insurgency/Saved/Logs", "/opt/sandstorm2/Insurgency/Saved/Logs"];

        for (const dir of mockDirs) {
            try {
                await mkdir(dir, { recursive: true });
                console.log(`Created mock directory: ${dir}`);
            } catch (error) {
                // Skip if can't create (e.g., permissions)
                console.log(`⚠️  Could not create mock directory: ${dir}`);
            }
        }

        // Test JSON config loading
        console.log("\nTesting JSON configuration loading...");
        const jsonConfig = await ConfigLoader.loadConfig("./test-config.json");
        console.log(`✅ JSON Config loaded: ${jsonConfig.servers.length} servers configured`);
        console.log(`   Database: ${jsonConfig.database.path}`);
        console.log(`   Log level: ${jsonConfig.logging.level}\n`);

        // Reset for next test
        ConfigLoader.reset();

        // Test TOML config loading
        console.log("Testing TOML configuration loading...");
        const tomlConfig = await ConfigLoader.loadConfig("./test-config.toml");
        console.log(`✅ TOML Config loaded: ${tomlConfig.servers.length} servers configured`);
        console.log(`   Database: ${tomlConfig.database.path}`);
        console.log(`   Log level: ${tomlConfig.logging.level}\n`);

        // Test config validation
        console.log("Testing configuration validation...");
        const jsonValidation = await ConfigLoader.validateConfigFile("./test-config.json");
        const tomlValidation = await ConfigLoader.validateConfigFile("./test-config.toml");

        console.log(`   JSON config valid: ${jsonValidation.valid}`);
        console.log(`   TOML config valid: ${tomlValidation.valid}\n`);

        // Test creating sample configs
        console.log("Creating sample configuration files...");
        await ConfigLoader.createSampleConfig("./sample-config.json", "json");
        await ConfigLoader.createSampleConfig("./sample-config.toml", "toml");
        console.log("✅ Sample configuration files created\n");

        // Test format detection
        console.log("Testing format detection...");
        console.log(`   JSON supported: ${ConfigLoader.isSupportedFormat("config.json")}`);
        console.log(`   TOML supported: ${ConfigLoader.isSupportedFormat("config.toml")}`);
        console.log(`   TXT supported: ${ConfigLoader.isSupportedFormat("config.txt")}`);
        console.log(`   Supported formats: ${ConfigLoader.getSupportedFormats().join(", ")}\n`);

        // Test environment override
        console.log("Testing environment variable overrides...");
        process.env.SANDSTORM_DB_PATH = "override_stats.db";
        process.env.SANDSTORM_LOG_LEVEL = "debug";

        ConfigLoader.reset();
        const overrideConfig = await ConfigLoader.loadConfig("./test-config.json");
        console.log(`   Database path override: ${overrideConfig.database.path}`);
        console.log(`   Log level override: ${overrideConfig.logging.level}\n`);

        console.log("All configuration loader tests passed!");
    } catch (error) {
        console.error("❌ Configuration loader test failed:", error);
        process.exit(1);
    }
}

// Run the test
testConfigLoader();
