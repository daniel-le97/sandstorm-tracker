# Server ID Extraction Examples

This document shows how the Sandstorm Tracker automatically detects server IDs from log file paths.

## How It Works

The system extracts the server ID from the log file name using the pattern `{server_id}.log`. This happens automatically during log processing, so you don't need to manually configure server IDs in most cases.

## File Naming Convention

Your Sandstorm server log files should follow this naming pattern:

```
{server_id}.log
```

Where `{server_id}` is a unique identifier for your server containing only:

- Letters (a-z, A-Z)
- Numbers (0-9)
- Hyphens (-)
- Underscores (\_)
- Length: 1-50 characters
- Must start with a letter or number

## Examples

### ✅ Valid Server ID Examples

```bash
# Simple server names
main-server.log          → Server ID: "main-server"
server1.log              → Server ID: "server1"
hardcore.log             → Server ID: "hardcore"

# Descriptive server names
casual-pvp.log           → Server ID: "casual-pvp"
competitive_server.log   → Server ID: "competitive_server"
training-ground.log      → Server ID: "training-ground"

# Mixed case examples
MainServer.log           → Server ID: "MainServer"
Hardcore-PvP.log         → Server ID: "Hardcore-PvP"
Event_Server_2024.log    → Server ID: "Event_Server_2024"
```

### ❌ Invalid Examples

```bash
# These will NOT work
.log                     → Empty server ID
-invalid.log            → Can't start with hyphen
_invalid.log            → Can't start with underscore
server with spaces.log   → Spaces not allowed
server@home.log         → Special characters not allowed
```

## Full Path Examples

### Windows Paths

```bash
C:\Program Files\Steam\steamapps\common\sandstorm-server\Insurgency\Saved\Logs\main-server.log
D:\GameServers\Hardcore\Insurgency\Saved\Logs\hardcore-pvp.log
E:\Servers\Training\Insurgency\Saved\Logs\training-ground.log
```

### Mac/Linux Paths

```bash
/Users/admin/servers/main/Insurgency/Saved/Logs/main-server.log
/home/gameserver/casual/Insurgency/Saved/Logs/casual-server.log
/opt/sandstorm/competitive/Insurgency/Saved/Logs/competitive-server.log
```

## Configuration Integration

When using this naming convention, your server configuration becomes simpler:

### Environment Variables (.env)

```bash
# You only need to specify the log directory path
# The server ID will be automatically detected from the filename

SERVER1_LOG_PATH=C:\Servers\Main\Insurgency\Saved\Logs
SERVER1_SERVER_ID=60844f66-b93b-4fe1-afc4-a0a91b493865  # Your actual Sandstorm server UUID
# The tracker will look for: main-server.log in the above directory

SERVER2_LOG_PATH=C:\Servers\Hardcore\Insurgency\Saved\Logs
SERVER2_SERVER_ID=a1b2c3d4-e5f6-7890-abcd-ef1234567890
# The tracker will look for: hardcore-server.log in the above directory
```

### JSON Configuration

```json
{
  "servers": [
    {
      "id": "main-server",
      "name": "Main Sandstorm Server",
      "logPath": "C:\\Servers\\Main\\Insurgency\\Saved\\Logs",
      "serverId": "60844f66-b93b-4fe1-afc4-a0a91b493865"
    },
    {
      "id": "hardcore-server",
      "name": "Hardcore PvP Server",
      "logPath": "C:\\Servers\\Hardcore\\Insurgency\\Saved\\Logs",
      "serverId": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
    }
  ]
}
```

## Testing Server ID Extraction

You can test server ID extraction using the built-in utilities:

```typescript
import { PathUtils } from "./src/cross-platform-utils";

// Test various log file paths
console.log(PathUtils.extractServerIdFromLogPath("main-server.log"));
// Output: "main-server"

console.log(PathUtils.extractServerIdFromLogPath("/path/to/logs/hardcore-pvp.log"));
// Output: "hardcore-pvp"

console.log(PathUtils.extractServerIdFromLogPath("C:\\Logs\\competitive_server.log"));
// Output: "competitive_server"

// Validate server ID format
console.log(PathUtils.isValidServerId("main-server"));
// Output: true

console.log(PathUtils.isValidServerId("invalid server"));
// Output: false
```

## Best Practices

1. **Use descriptive names**: Choose server IDs that clearly identify your server's purpose
2. **Keep it simple**: Shorter names are easier to manage and read in logs
3. **Be consistent**: Use a consistent naming pattern across all your servers
4. **Test first**: Validate your server ID format before deploying

### Recommended Naming Patterns

```bash
# By purpose
main-server.log
casual-server.log
hardcore-server.log
competitive-server.log
training-server.log

# By region/location
us-east-server.log
eu-west-server.log
asia-server.log

# By game mode
coop-server.log
pvp-server.log
checkpoint-server.log

# By number/sequence
server-01.log
server-02.log
server-03.log
```

This automatic detection makes server management much easier and reduces configuration complexity!
