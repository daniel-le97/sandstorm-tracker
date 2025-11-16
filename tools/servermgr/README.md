# Insurgency: Sandstorm Server Manager

A standalone command-line tool for managing Insurgency: Sandstorm dedicated server instances using Sandstorm Admin Wrapper (SAW) configurations.

## Features

- **Start/Stop Servers**: Launch and terminate Insurgency server instances
- **Multi-Server Support**: Manage multiple server instances from SAW configuration
- **Process Management**: Track server processes with PID files
- **Server Updates**: Update SteamCMD and game server files
- **Status Monitoring**: Check running servers and detect stale processes
- **Configuration Management**: Apply server-specific config files

## Installation

### Build from Source

```bash
cd tools/servermgr
go build -o servermgr.exe
```

### Set Environment Variable (Optional)

Set the `SAW_PATH` environment variable to avoid passing `--saw-path` flag:

```powershell
$env:SAW_PATH = "C:\path\to\sandstorm-admin-wrapper"
```

Or set it permanently:

```powershell
[System.Environment]::SetEnvironmentVariable('SAW_PATH', 'C:\path\to\sandstorm-admin-wrapper', 'User')
```

## Usage

### Start a Server

Start a specific server by ID:

```bash
servermgr start server-1
```

Start with console logs (blocks until stopped):

```bash
servermgr start server-1 --logs
```

Start all configured servers:

```bash
servermgr start --all
```

### Stop a Server

Stop a specific server:

```bash
servermgr stop server-1
```

Stop all running servers:

```bash
servermgr stop --all
```

### Check Server Status

View running servers and detect stale PID files:

```bash
servermgr status
```

### List Available Servers

Display all servers from SAW configuration:

```bash
servermgr list
```

### Update Game Files

Update Insurgency: Sandstorm server to latest version:

```bash
servermgr update-game
```

Validate all server files (slower but thorough):

```bash
servermgr update-game --validate
```

Force update even if servers are running (not recommended):

```bash
servermgr update-game --force
```

### Update SteamCMD

Update SteamCMD to the latest version:

```bash
servermgr update-steamcmd
```

## Configuration

### SAW Path

Provide the SAW installation path using one of these methods:

1. **Environment Variable**: Set `SAW_PATH`
2. **Command Flag**: Use `--saw-path` flag on any command

Example:

```bash
servermgr start server-1 --saw-path "C:\SAW"
```

### Server Executable Path

By default, the tool looks for the server executable at:

```
{SAW_PATH}/sandstorm-server/Insurgency/Binaries/Win64/InsurgencyServer-Win64-Shipping.exe
```

Override this by setting the `INSURGENCY_SERVER_PATH` environment variable:

```powershell
$env:INSURGENCY_SERVER_PATH = "C:\custom\path\to\InsurgencyServer-Win64-Shipping.exe"
```

### Server Configurations

Server configurations are read from SAW's `server-configs.json`:

```
{SAW_PATH}/admin-interface/config/server-configs.json
```

Per-server config files are stored in:

```
{SAW_PATH}/server-config/{server-id}/
```

Supported config files:

- `Game.ini`
- `Engine.ini`
- `Admins.txt`
- `Bans.txt`
- `MapCycle.txt`
- `Motd.txt`

## Process Management

### PID Files

Server PIDs are tracked in `./data/{server-id}.pid` files.

### Detached Processes

Servers started without `--logs` flag run as detached background processes using PowerShell `Start-Process`. They continue running after the tool exits.

### Console Attached

Servers started with `--logs` flag run in the foreground and stop when you press Ctrl+C or close the terminal.

## Examples

### Complete Workflow

```bash
# List available servers
servermgr list

# Start a specific server in background
servermgr start my-coop-server

# Check if it's running
servermgr status

# Start another server with console logs
servermgr start my-pvp-server --logs

# (In another terminal) Stop the background server
servermgr stop my-coop-server

# Stop all servers and clean up stale PIDs
servermgr stop --all

# Update game server files
servermgr update-game --validate
```

### Update Workflow

```bash
# Stop all servers before updating
servermgr stop --all

# Update SteamCMD
servermgr update-steamcmd

# Update game server with validation
servermgr update-game --validate

# Start servers again
servermgr start --all
```

## Troubleshooting

### "SAW path not provided" Error

Set the `SAW_PATH` environment variable or use `--saw-path` flag.

### "Server executable not found" Error

Verify the Insurgency server is installed at:

```
{SAW_PATH}/sandstorm-server/Insurgency/Binaries/Win64/InsurgencyServer-Win64-Shipping.exe
```

Or set `INSURGENCY_SERVER_PATH` to the correct location.

### Stale PID Files

If servers crash or are killed externally, PID files may become stale. Use:

```bash
servermgr status
```

To detect stale PIDs, then run:

```bash
servermgr stop --all
```

To clean them up.

### Server Won't Start

1. Check if server is already running: `servermgr status`
2. Verify SAW configuration: `servermgr list`
3. Check server executable exists
4. Review logs in server log files

## Requirements

- Windows (uses PowerShell for process management)
- Sandstorm Admin Wrapper (SAW) installation
- Insurgency: Sandstorm Dedicated Server
- SteamCMD

## Technical Details

### Server Launch Arguments

Servers are started with these arguments:

- Map/scenario travel string with lighting and game mode
- `-Hostname`, `-MaxPlayers`, `-Port`, `-QueryPort`
- `-LogCmds=LogGameplayEvents Log` (enables gameplay event logging)
- `-LOCALLOGTIMES` (use local time in logs)
- `-AdminList=Admins`, `-MapCycle=MapCycle`
- Optional: `-Password`, `-Mutators`, `-CmdServerCheats`
- Custom args from SAW config

### Configuration Files

Server-specific configs are copied from:

```
{SAW_PATH}/server-config/{server-id}/
```

To:

```
{SAW_PATH}/sandstorm-server/Insurgency/Saved/Config/WindowsServer/
```

Before each server start.

## License

Part of the sandstorm-tracker project.
