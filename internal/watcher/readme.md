## Watcher & Chronos Package Responsibilities

### 1. Monitor Log Directory

- Watch one or more directories for new or modified log files using `fsnotify`.
- Detect when new log files are created (e.g., after server restart/log rotation).
- Ensure every single log line is processed exactly once and in strict sequential order, even after restarts, log rotation, or crashes.

### 2. Track File Offsets

- Maintain the last read position (offset) for each log file.
- Persist offsets after every processed line to guarantee no duplication or loss.
- Resume reading from the correct offset after restarts or file changes.

### 3. Tail Log Files

- Read new lines appended to each log file in real time, in the order they are written.
- Handle partial lines and ensure no events are missed or duplicated.

### 4. Handle Log Rotation

- Detect when a log file is rotated, truncated, or replaced.
- Finish reading any remaining lines in the old file before switching to the new one.
- Start tailing the new log file automatically.

### 5. Parse and Process Events

- Parse each log line into a structured event.
- Pass events to the appropriate handler for database updates or further processing.
- Guarantee event processing order matches log order.

### 6. Error Handling and Recovery

- Log errors and attempt to recover from file access issues.
- Handle cases where files are temporarily unavailable, deleted, or truncated.
- On crash or restart, replay unprocessed lines from the last saved offset.

### 7. Graceful Shutdown

- Stop watching files and clean up resources on shutdown.
- Ensure all processed offsets are persisted before exit.

---

