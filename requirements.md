#todos:

## internal/watcher package
1. refactor watcher package to take in an App interface
2. change watcher catchup functionlity to:
   - send a2s server query to check if server is online and what map
   - find last map event that matches current map and then processevents for the current one
   - if none are found dont do catchup and proceed normally
3. we will need to update watcher tests for this

## internal/handlers and /assets packages
1. refactor our ui to use pocketbases js client library or find a way to update the ui from our backend
2. if we do pocketbase js client then we only need to do subcribes and unsubscribes as the intial data fetch would already be rendered into the html

## internal/app internal/watcher internal/jobs
1. ensure an rcon listplayers command is only ran once every 10 seconds per server unless there is a game over event





