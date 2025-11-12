# Startup Catch-up Implementation

## Overview

The startup catch-up feature ensures that when the tracker starts after an Insurgency server is already running a match, it processes historical events from the log file to maintain complete data integrity.

## How It Works

### Detection Logic

When the watcher processes a server log file for the first time (`offset == 0` and no saved log time), it performs these checks:

1. **File Modification Check**

   - Reads the log file's last modification time
   - Uses adaptive thresholds based on SAW (Sandstorm Admin Wrapper) detection:
     - **With SAW**: 1 minute threshold (SAW polls RCON every 4 seconds, keeping log fresh)
     - **Without SAW**: 6 hours threshold (catches servers that are idle or low activity)

2. **SAW Detection**

   - Scans the last 100 lines of the log file for `LogRcon:` entries
   - If recent RCON logs are found (within 30 seconds), SAW is detected as active
   - This enables more aggressive catch-up since log files stay fresh

3. **Recent Map Event Check**
   - Searches backwards through the log file for the most recent map event
   - Prioritizes `MapTravel` events (runtime map changes) over `MapLoad` (server start)
   - Considers events within the last 30 minutes as "recent"
   - Returns map name, scenario, timestamp, and line number

### Catch-up Process

If both conditions are met (file recently modified AND recent map event found):

1. **Match Creation**

   - Checks if an active match already exists for the server
   - If not, creates a new match with:
     - Map name and scenario from the found map event
     - Start time from the map event timestamp
     - Player team extracted from scenario (Security/Insurgents)

2. **Historical Event Processing**

   - Marks the current file size as the catch-up end point
   - Processes all events from the map event line to the current position
   - Uses the same `parser.ParseAndProcess` logic to ensure consistency
   - Preserves original timestamps from the log entries
   - Processes events in exact chronological order

3. **Watcher Startup**
   - Saves the catch-up end offset to the database
   - Watcher starts monitoring from the saved offset forward
   - No duplicate processing - catch-up phase is complete before watching begins

## Edge Cases Handled

### Empty Server

- File not recently modified → Skip catch-up
- Watcher starts normally from current position

### Mid-Match Startup

- File recently modified + recent map event found
- Full catch-up processes all kills, deaths, objectives, etc.
- Complete match data captured despite late start

### About to Map Travel

- If tracker starts right before map change
- Catch-up captures old match data
- New MapTravel event creates new match naturally

### SAW vs Non-SAW

- **With SAW**: Aggressive 1-minute threshold catches active servers quickly
- **Without SAW**: Conservative 6-hour threshold prevents false positives

### Multiple Players in Objectives

- Fixed in parser: All players get credit (not just first)
- `tryProcessObjectiveDestroyed`: All players get `objectives_destroyed` incremented
- `tryProcessObjectiveCaptured`: All players get `objectives_captured` incremented

## Code Structure

### Files Modified

1. **internal/watcher/watcher.go**

   - `checkStartupCatchup()`: Main catch-up detection and orchestration
   - `hasRecentRconLogs()`: SAW detection via RCON log scanning
   - `processHistoricalEvents()`: Sequential processing of historical events
   - `parseTimestampFromLog()`: Utility for timestamp parsing

2. **internal/parser/parser.go**
   - `FindLastMapEvent()`: Exported function to find recent map events
   - Used internally for orphaned event match creation
   - Now also used by watcher for catch-up detection

### Key Functions

```go
// Check if catch-up is needed, return offset to start from
func (w *Watcher) checkStartupCatchup(filePath, serverID string) (int, bool)

// Detect SAW by scanning for recent RCON logs
func (w *Watcher) hasRecentRconLogs(filePath string, threshold time.Duration) bool

// Process events from startLine to endOffset
func (w *Watcher) processHistoricalEvents(filePath, serverID string, startLine int, endOffset int64) int
```

## Testing Scenarios

To validate the implementation:

1. **Start tracker after server is running**

   - Verify match is created with correct map/scenario
   - Confirm all players, kills, deaths are captured
   - Check objectives are properly attributed

2. **Start tracker on empty server**

   - Verify no catch-up is triggered
   - Confirm normal operation from start

3. **Start tracker with SAW active**

   - Verify 1-minute threshold is used
   - Confirm SAW detection via log output

4. **Start tracker without SAW**
   - Verify 6-hour threshold is used
   - Confirm detection works for idle servers

## Logging

The implementation includes detailed logging for debugging:

```
[Catchup] SAW detected for <serverID>, using 1-minute threshold
[Catchup] Starting catch-up for <serverID>: file modified 45.3s ago, map 'Ministry' loaded 127.8s ago
[Catchup] Created match for <serverID>: Ministry (Scenario_Ministry_Checkpoint_Security) at 2025-01-04 15:23:12
[Catchup] Completed for <serverID>: processed 1247 lines from 89 to 524288
```

## Performance Considerations

- Sequential processing prevents race conditions
- Single-pass through log file for historical events
- No duplicate processing (catch-up completes before watcher starts)
- Efficient reverse search for map events (scans from end of file)

## Future Enhancements

Potential improvements:

1. **Configurable Thresholds**: Allow users to set custom thresholds in config
2. **Progress Callbacks**: Report catch-up progress for long processing
3. **Partial Match Recovery**: Handle cases where map event is very old
4. **Multi-file Catch-up**: Process across log rotations if needed
