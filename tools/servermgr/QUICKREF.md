# Server Manager Tool - Quick Reference

## Installation

```bash
cd tools/servermgr
go build -o servermgr.exe
```

## Environment Setup

```powershell
# Set SAW path (optional, to avoid --saw-path flag)
$env:SAW_PATH = "C:\path\to\sandstorm-admin-wrapper"

# Set custom server executable path (optional)
$env:INSURGENCY_SERVER_PATH = "C:\custom\path\to\InsurgencyServer-Win64-Shipping.exe"
```

## Common Commands

### Server Operations

| Command                           | Description                  |
| --------------------------------- | ---------------------------- |
| `servermgr start server-1`        | Start specific server        |
| `servermgr start server-1 --logs` | Start with console output    |
| `servermgr start --all`           | Start all configured servers |
| `servermgr stop server-1`         | Stop specific server         |
| `servermgr stop --all`            | Stop all servers             |
| `servermgr status`                | Show running servers         |
| `servermgr list`                  | List available servers       |

### Updates

| Command                            | Description                    |
| ---------------------------------- | ------------------------------ |
| `servermgr update-steamcmd`        | Update SteamCMD                |
| `servermgr update-game`            | Update game server             |
| `servermgr update-game --validate` | Update and validate files      |
| `servermgr update-game --force`    | Force update (not recommended) |

## Workflows

### Daily Operations

```bash
# Morning - Start servers
servermgr start --all

# Check status
servermgr status

# Evening - Stop servers
servermgr stop --all
```

### Update Workflow

```bash
# 1. Stop all servers
servermgr stop --all

# 2. Update SteamCMD
servermgr update-steamcmd

# 3. Update game with validation
servermgr update-game --validate

# 4. Restart servers
servermgr start --all
```

### Troubleshooting

```bash
# Check for stale processes
servermgr status

# Clean up stale PID files
servermgr stop --all

# Verify server configuration
servermgr list
```

## File Locations

### Configuration Files

- SAW configs: `{SAW_PATH}/admin-interface/config/server-configs.json`
- Server configs: `{SAW_PATH}/server-config/{server-id}/`
- PID files: `./data/{server-id}.pid`

### Server Executable

- Default: `{SAW_PATH}/sandstorm-server/Insurgency/Binaries/Win64/InsurgencyServer-Win64-Shipping.exe`
- Override: Set `INSURGENCY_SERVER_PATH` environment variable

### Log Files

- Server logs: Created in working directory as `{server-id}.log`
- SAW logs: `{SAW_PATH}/sandstorm-server/Insurgency/Saved/Logs/`

## Supported Config Files

Per-server configs in `{SAW_PATH}/server-config/{server-id}/`:

- `Game.ini` - Game settings
- `Engine.ini` - Engine settings
- `Admins.txt` - Admin list
- `Bans.txt` - Ban list
- `MapCycle.txt` - Map rotation
- `Motd.txt` - Message of the day

## Tips

1. **Background Servers**: Without `--logs`, servers run as detached processes
2. **Console Servers**: With `--logs`, server stops when you close terminal
3. **Multiple Servers**: Each server needs unique ports in SAW config
4. **PID Tracking**: Tool tracks PIDs in `./data/` directory
5. **Clean Shutdown**: Always use `stop` command instead of killing processes

## Examples

### Start Multiple Servers

```bash
# Start first server in background
servermgr start coop-server

# Start second server with console logs
servermgr start pvp-server --logs
```

### Monitor Servers

```bash
# Check running processes
servermgr status

# Output shows:
# - Running servers with PIDs
# - Which server ID each PID belongs to
# - Stale PID files that need cleanup
```

### SAW Path Options

```bash
# Option 1: Environment variable
$env:SAW_PATH = "C:\SAW"
servermgr start server-1

# Option 2: Command flag
servermgr start server-1 --saw-path "C:\SAW"

# Option 3: Global flag
servermgr --saw-path "C:\SAW" start server-1
```

## Troubleshooting Guide

### Error: "SAW path not provided"

**Solution**: Set `SAW_PATH` environment variable or use `--saw-path` flag

### Error: "Server executable not found"

**Solution**:

1. Verify server is installed at expected location
2. Or set `INSURGENCY_SERVER_PATH` to correct path

### Error: "Server is already running"

**Solution**:

1. Check status: `servermgr status`
2. If stale PID, run: `servermgr stop --all`

### Server Won't Start

**Check**:

1. Is SAW config valid? `servermgr list`
2. Are ports already in use?
3. Is server executable accessible?
4. Check server log files for errors

### Stale PID Files

**Clean up**:

```bash
servermgr status  # Identify stale PIDs
servermgr stop --all  # Clean them up
```

## Integration with Tracker

The standalone `servermgr` tool is independent from the tracker app. They can run simultaneously:

- **Tracker**: Monitors server logs and stores stats in PocketBase
- **Server Manager**: Manages server processes (start/stop/update)

Both tools can read the same SAW configuration but operate independently.
