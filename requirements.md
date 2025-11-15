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

## internal/handlers and /assets packages

1. refactor our ui to use pocketbases js client library or find a way to update the ui from our backend
2. if we do pocketbase js client then we only need to do subcribes and unsubscribes as the intial data fetch would already be rendered into the html

## internal/app internal/watcher internal/jobs

1. ensure an rcon listplayers command is only ran once every 10 seconds per server unless there is a game over event

<!-- ## internal/logger
COMPLETED:
- this project uses pocketbase as a framework, and pocketbase provides its own logger which uses go's slog. this logger addionally logs everything to a sqlite db
- i want to also be able to log everything to a single log file
- if possible i want to also to be able to just use one of the loggers when needed for specific events
- it may be easiest to use a custom handler wrapper -->


// IGNORE
commit before changes:
ef470cc
