# Multi-Server Configuration Examples

This directory contains example configurations for setting up the Sandstorm tracker to monitor multiple servers simultaneously.

## Configuration Method 1: Environment Variables

Set environment variables in your system or `.env` file:

```bash
# Server 1 - Hardcore Mode
SANDSTORM_SERVER_1_ID=hardcore-server
SANDSTORM_SERVER_1_NAME=Hardcore Mode Server
SANDSTORM_SERVER_1_LOG_PATH=C:\Servers\Hardcore\Insurgency\Saved\Logs
SANDSTORM_SERVER_1_SERVER_ID=60844f66-b93b-4fe1-afc4-a0a91b493865
SANDSTORM_SERVER_1_ENABLED=true
SANDSTORM_SERVER_1_DESCRIPTION=Hardcore difficulty server with friendly fire

# Server 2 - Casual Mode
SANDSTORM_SERVER_2_ID=casual-server
SANDSTORM_SERVER_2_NAME=Casual Mode Server
SANDSTORM_SERVER_2_LOG_PATH=C:\Servers\Casual\Insurgency\Saved\Logs
SANDSTORM_SERVER_2_SERVER_ID=a1b2c3d4-e5f6-7890-abcd-ef1234567890
SANDSTORM_SERVER_2_ENABLED=true
SANDSTORM_SERVER_2_DESCRIPTION=Casual server for new players

# Server 3 - Competitive
SANDSTORM_SERVER_3_ID=competitive-server
SANDSTORM_SERVER_3_NAME=Competitive Server
SANDSTORM_SERVER_3_LOG_PATH=C:\Servers\Competitive\Insurgency\Saved\Logs
SANDSTORM_SERVER_3_SERVER_ID=12345678-9abc-def0-1234-56789abcdef0
SANDSTORM_SERVER_3_ENABLED=true
SANDSTORM_SERVER_3_DESCRIPTION=Ranked competitive matches

# Server 4 - Co-op PvE
SANDSTORM_SERVER_4_ID=coop-server
SANDSTORM_SERVER_4_NAME=Co-op PvE Server
SANDSTORM_SERVER_4_LOG_PATH=C:\Servers\Coop\Insurgency\Saved\Logs
SANDSTORM_SERVER_4_SERVER_ID=fedcba98-7654-3210-fedc-ba9876543210
SANDSTORM_SERVER_4_ENABLED=true
SANDSTORM_SERVER_4_DESCRIPTION=Player vs AI cooperative gameplay

# Server 5 - Modded Server
SANDSTORM_SERVER_5_ID=modded-server
SANDSTORM_SERVER_5_NAME=Modded Server
SANDSTORM_SERVER_5_LOG_PATH=C:\Servers\Modded\Insurgency\Saved\Logs
SANDSTORM_SERVER_5_SERVER_ID=abcdef12-3456-7890-abcd-ef123456789a
SANDSTORM_SERVER_5_ENABLED=true
SANDSTORM_SERVER_5_DESCRIPTION=Server with custom modifications and maps

# Server 6 - Training Server (Disabled)
SANDSTORM_SERVER_6_ID=training-server
SANDSTORM_SERVER_6_NAME=Training Server
SANDSTORM_SERVER_6_LOG_PATH=C:\Servers\Training\Insurgency\Saved\Logs
SANDSTORM_SERVER_6_SERVER_ID=98765432-1fed-cba0-9876-543210fedcba
SANDSTORM_SERVER_6_ENABLED=false
SANDSTORM_SERVER_6_DESCRIPTION=Training server for clan practice sessions

# Database Configuration
SANDSTORM_DB_PATH=multi_server_sandstorm_stats.db
SANDSTORM_LOG_LEVEL=info
```

## Configuration Method 2: JSON File

Create a `config/servers.json` file:

See `six-servers.json` for a complete example.

## Running the Tracker

1. **Using Environment Variables:**

   ```bash
   # Set your environment variables (see above)
   bun run index.ts
   ```

2. **Using JSON Configuration:**
   ```bash
   # Create your config/servers.json file first
   SANDSTORM_CONFIG_PATH=./config/servers.json bun run index.ts
   ```

## Server Log Paths

Make sure your server log paths are correct for your Insurgency: Sandstorm server installations:

- **Windows:** `C:\Path\To\Your\Server\Insurgency\Saved\Logs`
- **Linux:** `/path/to/your/server/Insurgency/Saved/Logs`

## Server UUIDs

You can find your server UUIDs in the server configuration files or admin panel. Each server must have a unique UUID.

## Database

The tracker automatically creates and manages a SQLite database. With multi-server support:

- Each server's data is isolated
- Players can have separate statistics per server
- Chat commands show server-specific stats
- Database migration handles existing single-server data

## Performance

The tracker is optimized for multiple servers:

- Efficient file watching (one watcher per log directory)
- Debounced event processing
- Indexed database queries
- Configurable performance parameters

## Monitoring

The tracker provides detailed logging with server context:

```
[Hardcore Mode Server] 🎮 Player joined: PlayerName
[Casual Mode Server] ⚔️ Player1 killed Player2 with AK-74
[Competitive Server] 💬 Player used command: !stats
```
