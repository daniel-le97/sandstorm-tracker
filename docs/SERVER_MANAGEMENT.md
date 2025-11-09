# Server Management

The sandstorm-tracker now includes built-in server management capabilities to start, stop, and monitor Insurgency: Sandstorm servers directly from the tracker application.

## Features

- **Start/Stop servers** from command line or web API
- **Monitor server status** and view running servers
- **Load configurations** from Sandstorm Admin Wrapper (SAW)
- **Automatic cleanup** when tracker shuts down

## Command Line Usage

### List Available Servers

View all servers configured in your SAW installation:

```bash
go run main.go server list --saw-path="C:/path/to/sandstorm-admin-wrapper"
```

### Start a Server

Start a specific server by ID:

```bash
# Start with logs in console (foreground mode)
go run main.go server start SERVER-ID-HERE --saw-path="C:/path/to/saw" --logs

# Start without logs (background mode)
go run main.go server start SERVER-ID-HERE --saw-path="C:/path/to/saw"
```

If you have `sawPath` configured in `sandstorm-tracker.yml`, you can omit the `--saw-path` flag:

```yaml
sawPath: "C:/Users/danie/code/forks/sandstorm-admin-wrapper"
```

Then simply:

```bash
go run main.go server start SERVER-ID-HERE --logs
```

### Stop a Server

Stop a running server:

```bash
go run main.go server stop SERVER-ID-HERE
```

### Check Server Status

View status of all managed servers:

```bash
go run main.go server status
```

View status of a specific server:

```bash
go run main.go server status SERVER-ID-HERE
```

## Web API Usage

### Start a Server

```bash
POST /api/server/start
Content-Type: application/json

{
  "server_id": "1d6407b7-f51b-4b1d-ad9e-faabbfbb7dde",
  "saw_path": "C:/path/to/sandstorm-admin-wrapper",
  "show_logs": false
}
```

### Stop a Server

```bash
POST /api/server/stop
Content-Type: application/json

{
  "server_id": "1d6407b7-f51b-4b1d-ad9e-faabbfbb7dde"
}
```

### Get Server Status

```bash
GET /api/server/status
```

Returns:

```json
{
  "servers": {
    "server-id-1": true,
    "server-id-2": false
  }
}
```

### List Available Servers

```bash
GET /api/server/list?saw_path=C:/path/to/sandstorm-admin-wrapper
```

Returns:

```json
{
  "servers": [
    {
      "id": "1d6407b7-f51b-4b1d-ad9e-faabbfbb7dde",
      "name": "My Server",
      "map": "Ministry",
      "mode": "Checkpoint_Security",
      "port": "27102",
      "query_port": "27131",
      "max_players": "20"
    }
  ]
}
```

## Configuration

Add SAW path to your `sandstorm-tracker.yml`:

```yaml
sawPath: "C:/Users/danie/code/forks/sandstorm-admin-wrapper"

servers:
  - name: "My Server"
    logPath: "C:/path/to/saw/sandstorm-server/Insurgency/Saved/Logs"
    enabled: true
```

## Environment Variables

- `INSURGENCY_SERVER_PATH`: Override the default server executable path
- `SAW_PATH`: Alternative to config file `sawPath` setting

## How It Works

1. **Configuration Loading**: The server manager loads server configurations from SAW's `admin-interface/config/server-configs.json`
2. **Process Management**: Each server runs as a child process managed by the tracker
3. **Automatic Cleanup**: When the tracker shuts down, all managed servers are automatically stopped
4. **Log Control**: Servers can output logs to console (`-stdout`) or to file (`-log=<id>.log`)

## Taskfile Integration

You can add server management tasks to your `Taskfile.yml`:

```yaml
tasks:
  server-start:
    desc: Start Insurgency server
    cmds:
      - go run main.go server start {{.CLI_ARGS}} --logs

  server-stop:
    desc: Stop Insurgency server
    cmds:
      - go run main.go server stop {{.CLI_ARGS}}

  server-list:
    desc: List available servers
    cmds:
      - go run main.go server list
```

Usage:

```bash
task server-start -- SERVER-ID
task server-stop -- SERVER-ID
task server-list
```

## Notes

- Servers started with `--logs` will display game logs directly in the console
- Servers started without `--logs` will write logs to a file but won't block the terminal
- All managed servers are automatically stopped when the tracker application terminates
- Server IDs are GUIDs from the SAW configuration (e.g., `1d6407b7-f51b-4b1d-ad9e-faabbfbb7dde`)
