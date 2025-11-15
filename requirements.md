# todos:

<!-- ## internal/watcher package
COMPLETED:
1. refactor watcher package to take in an App interface of my main App struct in /internal/app - cancelled
2. change watcher catchup functionlity to:
   - send a2s server query to check if server is online and what map
   - find last map event that matches current map and then processevents for the current one
   - if none are found dont do catchup and proceed normally
3. we will need to update watcher tests for this -->

<!-- ## internal/watcher package - Sequential Event Processing
COMPLETED: Implemented per-server worker queues to ensure sequential event processing
- Each server has its own buffered channel (queue) and dedicated worker goroutine
- Events from fsnotify are enqueued to the appropriate server's queue
- Workers process events sequentially, maintaining the order Sandstorm writes to log files
- No more race conditions - events for each server are guaranteed to be processed in order
- Different servers can still process events concurrently (good for performance)
- Proper cleanup: queues are closed on shutdown, workers exit cleanly -->

<!-- ## internal/handlers and /assets packages
COMPLETED:
1. ✅ UI uses self-hosted PocketBase JS SDK (no external CDN dependencies)
2. ✅ Server-side rendering for initial page load (full HTML with data)
3. ✅ PocketBase client subscriptions for real-time updates via WebSockets
4. ✅ Server status page implemented with real-time match/player subscriptions
5. ✅ Static files served from embedded assets/static/ directory
6. ✅ Example subscriptions and documentation in docs/REALTIME_UPDATES.md
Next: Add data attributes to templates and implement direct DOM updates (currently reloads page) -->

<!-- ## internal/app internal/watcher internal/jobs
COMPLETED:
 - refactor our score cron job, if there has not been a parser event in over a minute, we can assume server is no longer active
 - if server is active we do not want to try to update scores if there are no players -->

<!-- 1. ensure an rcon listplayers command is only ran once every 10 seconds per server unless there is a game over event -->

<!-- ## internal/logger
COMPLETED:
- this project uses pocketbase as a framework, and pocketbase provides its own logger which uses go's slog. this logger addionally logs everything to a sqlite db
- i want to also be able to log everything to a single log file
- if possible i want to also to be able to just use one of the loggers when needed for specific events
- it may be easiest to use a custom handler wrapper -->


im thinking of refactoring my app more to use pocketbases built in features such as 
```
app.OnRecordCreate("events").BindFunc(func(e *core.RecordEvent) error {
        // e.App
        // e.Record

        return e.Next()
    })
```
i can setup an event bus that my app struct holds and have these handlers publish to the event bus
then other parts of my app can subscribe to these events, ie my parser will publish events when it parses them



// IGNORE
commit before changes:
ef470cc
