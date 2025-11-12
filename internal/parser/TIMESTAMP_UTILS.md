# Timestamp Utility Functions

## Overview

The timestamp utility functions allow you to replace timestamps in Insurgency: Sandstorm log lines with more recent timestamps. This is useful for testing parsers with old log data while ensuring the timestamps are recent enough to pass catch-up time checks (8-hour threshold).

## Functions

### `ReplaceTimestamp(line string, baseTime time.Time) string`

Replaces the timestamp in a single log line with a new timestamp.

**Example:**

```go
line := "[2025.10.04-21.27.51:780][866]LogNet: Join succeeded: Player"
newLine := ReplaceTimestamp(line, time.Now())
// Result: "[2025.11.11-14.30.00:123][866]LogNet: Join succeeded: Player"
```

### `ReplaceTimestampWithOffset(line string, baseTime, referenceTime time.Time) string`

Replaces the timestamp while preserving the relative timing from a reference timestamp. This maintains the time offsets between events.

**Example:**

```go
referenceTime := parseTime("2025.10.04-21.27.00:000")
line := "[2025.10.04-21.27.51:780][866]LogNet: Join succeeded: Player"
newLine := ReplaceTimestampWithOffset(line, time.Now(), referenceTime)
// Result: Current time + 51.78 seconds
```

### `UpdateLogTimestamps(lines []string, baseTime time.Time) ([]string, error)`

Updates all timestamps in a slice of log lines to be relative to a new base time. The first line's timestamp becomes the base time, and all subsequent lines maintain their original offset.

**Example:**

```go
lines := []string{
    "[2025.10.04-21.27.51:780][866]LogNet: Join succeeded: Player1",
    "[2025.10.04-21.27.56:890][490]LogNet: Join succeeded: Player2", // +5.11 seconds
    "[2025.10.04-21.28.22:312][561]LogChat: Display: Player1: !stats", // +30.532 seconds
}

baseTime := time.Now().Add(-30 * time.Minute) // 30 minutes ago
updatedLines, err := UpdateLogTimestamps(lines, baseTime)
// All lines now have timestamps starting from baseTime with same relative offsets
```

## Practical Use Case

When testing with old log files (like the test.txt with October timestamps), you can update them to recent timestamps:

```go
// Read old log file
logLines := readLogFile("test_data/test.txt")

// Update timestamps to be within the last hour
baseTime := time.Now().Add(-30 * time.Minute)
recentLines, err := UpdateLogTimestamps(logLines, baseTime)

// Now process with parser - timestamps will pass the 8-hour catch-up check
for _, line := range recentLines {
    parser.ParseLine(context.Background(), line, serverID)
}
```

## Timestamp Format

Insurgency: Sandstorm uses the format: `[YYYY.MM.DD-HH.MM.SS:mmm]`

Example: `[2025.11.11-14.30.00:123]`

- Date: 2025.11.11 (November 11, 2025)
- Time: 14.30.00 (14:30:00)
- Milliseconds: 123

## Testing

All functions are tested in `timestamp_utils_test.go`:

- `TestReplaceTimestamp` - Basic timestamp replacement
- `TestReplaceTimestampWithOffset` - Offset-preserving replacement
- `TestUpdateLogTimestamps` - Bulk update with relative timing
- `TestParseInsurgencyTimestamp` - Timestamp parsing
- `TestFormatInsurgencyTimestamp` - Timestamp formatting
- `TestParseAllEventsWithUpdatedTimestamps` - End-to-end event parsing with updated timestamps
- `TestLoginRequestEvent` - Verify login request parser works with updated timestamps

Run tests:

```bash
go test -v ./internal/parser -run "TestUpdate|TestParse|TestFormat|TestLogin"
```
