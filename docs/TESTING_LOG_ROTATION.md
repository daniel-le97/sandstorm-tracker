# Testing Log Rotation Handling

## Scenario: Tracker Starts After Server Restart

This guide helps you test that the tracker correctly handles log rotation when Insurgency servers have restarted while the tracker was offline.

### What Happens During Server Restart

1. **Insurgency server restarts**

   - Old `Insurgency.log` → renamed to `Insurgency.log.1` or deleted
   - New `Insurgency.log` created
   - First line: `Log file open, MM/DD/YY HH:MM:SS` (with current time)
   - File starts at 0 bytes

2. **Tracker database state (before restart)**

   - `servers.offset` = 50,000 bytes (where it left off)
   - `servers.log_file_creation_time` = "2025-11-10T20:00:00Z" (old timestamp)

3. **Tracker restarts and detects rotation**
   - Reads saved offset (50,000) from database
   - Opens `Insurgency.log`
   - Extracts timestamp from first line → "2025-11-11T13:30:00Z"
   - **Compares**: New time ≠ saved time → **ROTATION DETECTED**
   - Resets offset to 0
   - Processes log from beginning
   - Updates `log_file_creation_time` in database

### Manual Testing Steps

#### Option 1: Simulate with Real Server

1. **Start tracker** and let it process some logs

   ```bash
   go run main.go serve
   ```

2. **Check database** before rotation:

   ```sql
   SELECT name, offset, log_file_creation_time FROM servers;
   ```

   Example output:

   ```
   name          | offset | log_file_creation_time
   Main Server   | 45231  | 2025-11-10T20:00:00Z
   ```

3. **Stop tracker**:

   ```bash
   Ctrl+C
   ```

4. **Restart Insurgency server** (this rotates the log)

   - Or manually simulate: Delete `Insurgency.log` and create new one with different timestamp

5. **Start tracker again**:

   ```bash
   go run main.go serve
   ```

6. **Check logs** for rotation detection message:

   ```
   Log rotation detected for main-server (creation time changed: 2025-11-10 20:00:00 → 2025-11-11 13:30:00)
   ```

7. **Verify database** was updated:
   ```sql
   SELECT name, offset, log_file_creation_time FROM servers;
   ```
   Should show new timestamp and reset offset.

#### Option 2: Manual Simulation (Without Server)

1. **Create test log directory**:

   ```bash
   mkdir test-logs
   ```

2. **Create initial log file** `test-logs/Insurgency.log`:

   ```
   Log file open, 11/10/25 20:00:00
   [2025.11.10-20.00.05:123][1]LogWorld: Bringing World /Game/Maps/Ministry.Ministry up for play
   [2025.11.10-20.00.10:456][5]LogGameMode: Match State Changed from EnteringMap to WaitingToStart
   (add ~100 lines of events)
   ```

3. **Configure tracker** to watch `test-logs`:

   ```yaml
   servers:
     - name: "Test Server"
       logPath: "./test-logs"
       enabled: true
   ```

4. **Start tracker** and verify it processes the log:

   ```bash
   go run main.go serve
   ```

   Check database for initial offset.

5. **Stop tracker** (Ctrl+C)

6. **Simulate log rotation**:

   ```bash
   # Delete old log
   rm test-logs/Insurgency.log

   # Create new log with different timestamp
   echo "Log file open, 11/11/25 13:30:00" > test-logs/Insurgency.log
   echo "[2025.11.11-13.30.05:123][1]LogWorld: Bringing World /Game/Maps/Hideout.Hideout up for play" >> test-logs/Insurgency.log
   ```

7. **Start tracker again** and watch for rotation detection:

   ```bash
   go run main.go serve
   ```

8. **Verify in logs**:
   ```
   Log rotation detected for test-server (creation time changed: ...)
   ```

### What to Verify

✅ **Rotation Detection Message**: Should see log message indicating rotation was detected

✅ **Offset Reset**: Database `servers.offset` should be reset to 0, then increase as file is processed

✅ **Timestamp Updated**: `servers.log_file_creation_time` should be updated to new time

✅ **Match Creation**: Should create new match for the current map (not resume old match)

✅ **No Duplicate Events**: Events should not be duplicated (processed only once)

### Fallback Detection Test

If timestamp extraction fails, the tracker uses file size as fallback:

1. **Setup**: Tracker has saved offset = 50,000 bytes
2. **Rotation**: New log file is only 5,000 bytes
3. **Detection**: `current_size (5,000) < saved_offset (50,000)` → rotation detected
4. **Result**: Offset reset to 0

Test this by creating a smaller log file than the saved offset.

### Automated Tests

Run the unit tests:

```bash
# Test timestamp extraction
go test ./internal/watcher/ -v -run TestExtractLogFileCreationTime

# View documentation tests
go test ./internal/watcher/ -v -run TestLogRotation
```

### Troubleshooting

**Problem**: Rotation not detected

- Check first line of log has correct format: `Log file open, MM/DD/YY HH:MM:SS`
- Verify `log_file_creation_time` is saved in database
- Check logs for error messages about timestamp parsing

**Problem**: Events processed twice

- Check if offset was properly reset to 0
- Verify file was processed from beginning after rotation

**Problem**: Old events still showing

- Database may have stale match data from before rotation
- Matches have `end_time` field - verify old match was ended
- Check `match.created` timestamp vs log file creation time
