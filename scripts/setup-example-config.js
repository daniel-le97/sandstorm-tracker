#!/usr/bin/env node

/**
 * Cross-platform setup script for example configuration
 * Creates platform-appropriate example configurations
 */

import { existsSync, mkdirSync, writeFileSync } from "fs";
import { homedir, platform } from "os";
import { join } from "path";

const currentPlatform = platform();
const isWindows = currentPlatform === "win32";
const isMac = currentPlatform === "darwin";
const isLinux = currentPlatform === "linux";

console.log(`🔧 Setting up example configuration for ${currentPlatform}...`);

// Create config directory if it doesn't exist
const configDir = join(process.cwd(), "config");
if (!existsSync(configDir)) {
    mkdirSync(configDir, { recursive: true });
    console.log("✅ Created config directory");
}

// Platform-specific paths
const getExamplePaths = () => {
    if (isWindows) {
        return {
            server1:
                "C:\\\\Program Files\\\\Steam\\\\steamapps\\\\common\\\\sandstorm-server\\\\Insurgency\\\\Saved\\\\Logs",
            server2: "C:\\\\GameServers\\\\Sandstorm1\\\\Insurgency\\\\Saved\\\\Logs",
            server3: "C:\\\\GameServers\\\\Sandstorm2\\\\Insurgency\\\\Saved\\\\Logs",
            database: join(homedir(), "AppData", "Local", "SandstormTracker", "sandstorm_stats.db").replace(
                /\\/g,
                "\\\\"
            ),
        };
    } else if (isMac) {
        return {
            server1: "/Applications/Steam/steamapps/common/sandstorm-server/Insurgency/Saved/Logs",
            server2: "/opt/sandstorm-server1/Insurgency/Saved/Logs",
            server3: "/opt/sandstorm-server2/Insurgency/Saved/Logs",
            database: join(homedir(), "Library/Application Support/SandstormTracker/sandstorm_stats.db"),
        };
    } else {
        return {
            server1: "/home/steam/.local/share/Steam/steamapps/common/sandstorm-server/Insurgency/Saved/Logs",
            server2: "/opt/sandstorm-server1/Insurgency/Saved/Logs",
            server3: "/opt/sandstorm-server2/Insurgency/Saved/Logs",
            database: join(homedir(), ".local/share/SandstormTracker/sandstorm_stats.db"),
        };
    }
};

const paths = getExamplePaths();

// Create platform-specific .env file
const envContent = `# Sandstorm Multi-Server Tracker Configuration
# Platform: ${currentPlatform}
# Generated: ${new Date().toISOString()}

# Database Configuration
SANDSTORM_DB_PATH=${paths.database}
SANDSTORM_LOG_LEVEL=info

# Server 1 - Main Server
SANDSTORM_SERVER_1_ID=main-server
SANDSTORM_SERVER_1_NAME=Main Sandstorm Server
SANDSTORM_SERVER_1_LOG_PATH=${paths.server1}
SANDSTORM_SERVER_1_SERVER_ID=60844f66-b93b-4fe1-afc4-a0a91b493865
SANDSTORM_SERVER_1_ENABLED=true
SANDSTORM_SERVER_1_DESCRIPTION=Primary game server

# Server 2 - Secondary Server
SANDSTORM_SERVER_2_ID=secondary-server
SANDSTORM_SERVER_2_NAME=Secondary Sandstorm Server
SANDSTORM_SERVER_2_LOG_PATH=${paths.server2}
SANDSTORM_SERVER_2_SERVER_ID=a1b2c3d4-e5f6-7890-abcd-ef1234567890
SANDSTORM_SERVER_2_ENABLED=true
SANDSTORM_SERVER_2_DESCRIPTION=Secondary game server

# Server 3 - Event Server (Disabled by default)
SANDSTORM_SERVER_3_ID=event-server
SANDSTORM_SERVER_3_NAME=Event Sandstorm Server
SANDSTORM_SERVER_3_LOG_PATH=${paths.server3}
SANDSTORM_SERVER_3_SERVER_ID=12345678-9abc-def0-1234-56789abcdef0
SANDSTORM_SERVER_3_ENABLED=false
SANDSTORM_SERVER_3_DESCRIPTION=Special events and tournaments

# Performance Settings (Optional)
# SANDSTORM_MAX_CONCURRENT_SERVERS=6
# SANDSTORM_DEBOUNCE_DELAY_MS=500
`;

// Write the .env file
const envPath = join(process.cwd(), ".env");
writeFileSync(envPath, envContent);
console.log(`✅ Created platform-specific .env file for ${currentPlatform}`);

// Create JSON configuration as alternative
const jsonConfig = {
    database: {
        path: paths.database,
        enableWAL: true,
        cacheSize: 1000,
    },
    logging: {
        level: "info",
        enableServerLogs: true,
    },
    performance: {
        debounceDelay: 500,
        maxConcurrentWatchers: 6,
    },
    servers: [
        {
            id: "main-server",
            name: "Main Sandstorm Server",
            logPath: paths.server1,
            serverId: "60844f66-b93b-4fe1-afc4-a0a91b493865",
            enabled: true,
            description: "Primary game server",
        },
        {
            id: "secondary-server",
            name: "Secondary Sandstorm Server",
            logPath: paths.server2,
            serverId: "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
            enabled: true,
            description: "Secondary game server",
        },
        {
            id: "event-server",
            name: "Event Sandstorm Server",
            logPath: paths.server3,
            serverId: "12345678-9abc-def0-1234-56789abcdef0",
            enabled: false,
            description: "Special events and tournaments",
        },
    ],
};

const configPath = join(configDir, "local.json");
writeFileSync(configPath, JSON.stringify(jsonConfig, null, 2));
console.log("✅ Created JSON configuration file");

// Show instructions
console.log("\\n📋 Setup Complete!");
console.log("\\n🚀 To get started:");
console.log("1. Edit the paths in .env to match your actual server installations");
console.log("2. Run: bun run dev");
console.log("\\n🔧 Alternative configuration:");
console.log(`Set SANDSTORM_CONFIG_PATH=${configPath} to use JSON config`);
console.log("\\n📖 See README.md for detailed setup instructions");

if (isWindows) {
    console.log("\\n💡 Windows Tips:");
    console.log("- Use PowerShell or Command Prompt for best compatibility");
    console.log("- Ensure Bun is in your PATH environment variable");
    console.log("- Use double backslashes (\\\\\\\\) in paths within JSON files");
} else if (isMac) {
    console.log("\\n🍎 macOS Tips:");
    console.log("- Use Terminal.app or iTerm2");
    console.log("- Ensure Bun is accessible via Homebrew or direct installation");
    console.log("- Check file permissions if you encounter access errors");
} else {
    console.log("\\n🐧 Linux Tips:");
    console.log("- Ensure your user has read access to server log directories");
    console.log("- Consider running servers under a dedicated user account");
    console.log("- Use systemd for production server management");
}
