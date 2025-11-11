1. ✅ this app is ideally meant to run 24/7 and before insurgency servers are started

- ✅ i want to also be able to start this tracker after the insurgency servers are started, ie this app needs to go down for maintenance
  - ✅ we kind of already have this implemented with findLastMapLoadEvent method, although this only looks for MapLoad events, we should probably look for MapTravel first incase the server is not on the default map
  - ✅ this should create the match and then go every each line of the log file until the end of file to then add the necessary events to the db
  - ✅ handle log rotation detection properly - if tracker is offline during server restart, log file is rotated and offset tracking becomes invalid

**Implementation:**

1. **Map Event Detection:**

   - Created `findLastMapEvent()` that searches for MapTravel events first (runtime map changes), then falls back to MapLoad (server start)
   - Returns the line number where the map event was found
   - Modified `getOrCreateMatchForEvent()` to create match with proper map info
   - Watcher naturally processes all events sequentially, no duplicate processing

2. **Log Rotation Detection:**
   - Added `log_file_creation_time` field to servers table (stores RFC3339 timestamp)
   - Parser extracts timestamp from first line: "Log file open, 11/10/25 20:58:31"
   - Watcher compares saved creation time with current log file creation time
   - If times differ → log was rotated → reset offset to 0
   - Fallback: If file size < saved offset → rotation detected (old method still works)
   - Prevents reading from wrong position when log rotates while tracker is offline

**Scenarios Handled:**

- ✅ Tracker starts after server (finds map event, continues from there)
- ✅ Server restarts while tracker offline (rotation detected via creation time)
- ✅ Server on non-default map (MapTravel detection)
- ✅ Multiple servers with different log rotation schedules
