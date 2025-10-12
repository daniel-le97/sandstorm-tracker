## Chronos Package: Log Replay & Catch-Up

Chronos is responsible for catching up on all log data and events in strict sequential order, especially when the application starts after the game server has already been running.

- On startup, scan the log directory for all relevant log files (including the current active one).
- For each log file, check for a saved offset. If none, start from the beginning; if present, resume from the saved offset.
- Read and process every line from the offset to the end of each file, in order.
- For the current active log file, continue tailing new lines as they are written.
- Always update and persist the offset after processing each line, so no events are missed or duplicated, even across restarts.
- Integrate with the watcher to ensure all historical and new events are processed in strict order.
