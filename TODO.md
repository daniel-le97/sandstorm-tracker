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

<!-- # kill events requirements - COMPLETED ✅
## Data
1. player killed team mate
   - [2025.10.04-15.12.17:473][441]LogGameplayEvents: Display: ArmoredBear[76561198995742987, team 0] killed Rabbit[76561198995742956, team 0] with BP_Firearm_M16A4_C_2147481419
   - this does not increase the player who killed a teammates stats, but should increase their match_player_stats (friendly_fire_kills) field
   - this should still increase the player who died, match_player_stats (deaths) field
   - ✅ VERIFIED: handlePlayerKill() increments friendly_fire_kills and records FriendlyFireIncident

2. regular player kills
   - [2025.10.04-14.31.05:706][800]LogGameplayEvents: Display: ArmoredBear[76561198995742987, team 0] killed Marksman[INVALID, team 1] with BP_Firearm_M16A4_C_2147481419
   - ✅ VERIFIED: handlePlayerKill() increments kills for first killer in non-teamkill scenario

3. player died to enviroment/suicide
   - [2025.10.04-15.12.17:473][441]LogGameplayEvents: Display: ArmoredBear[76561198995742987, team 0] killed ArmoredBear[76561198995742987, team 0] with BP_Character_Player_C_2147481498
   - [2025.10.04-15.22.58:646][535]LogGameplayEvents: Display: ArmoredBear[76561198995742987, team 0] killed ArmoredBear[76561198995742987, team 0] with BP_Projectile_Molotov_C_2147480055
   - ✅ VERIFIED: handlePlayerKill() detects suicide and only increments victim deaths (no kills)

4. multi-player kills
   - [2025.10.04-21.29.31:291][459]LogGameplayEvents: Display: -=312th=- Rabbit[76561198262186571, team 0] + *OSS*0rigin[76561198007416544, team 0] killed Rifleman[INVALID, team 1] with BP_Projectile_GAU8_C_2147477120
   - for the first player in the killer section, they should get a kill in their match_player_stats (kills) field, it should also create or update a match_weapon_stats record for them
   - ✅ VERIFIED: handlePlayerKill() at index 0 increments kills + creates/updates weapon stats
   - for any other player listed, we are just going to increase their match_player_stats (assists) field
   - ✅ VERIFIED: handlePlayerKill() at index > 0 increments assists + updates weapon stats

## Testing
✅ TestKillEventFlow - Basic kill event flow
✅ TestFriendlyFireKillEvent - Team kills recorded as friendly_fire_kills
✅ TestMultiPlayerKillEvent - Multi-player kills with assists
✅ TestSuicideKillEvent - Suicides only increment deaths
✅ All tests PASSING (integration/parser_handler_test.go) -->

<!-- ## Game state
COMPLETED:
```
[ROUND_END] [2025.11.10-21.12.05:370][627]LogGameplayEvents: Display: Round 2 Over: Team 0 won (win reason: Objective)
[GAME_OVER] [2025.11.10-21.12.25:385][831]LogSession: Display: AINSGameSession::HandleMatchHasEnded
[MAP_VOTE] [2025.11.10-21.12.45:415][ 38]LogMapVoteManager: Display: New Vote Options:
[MAP_TRAVEL] [2025.11.10-21.12.58:822][846]LogGameMode: ProcessServerTravel: Oilfield?Scenario=Scenario_Refinery_Push_Insurgents?Game=?
[DISCONNECT] [2025.11.10-21.12.58:872][848]LogEOSAntiCheat: Display: ServerUnregisterClient: UserId (76561198995742987), Result: (EOS_Success)
[REGISTER] [2025.11.10-21.13.01:102][918]LogEOSAntiCheat: Display: ServerRegisterClient: Client: (76561198995742987) Result: (EOS_Success)
[DISCONNECT] [2025.11.10-21.13.19:601][ 23]LogEOSAntiCheat: Display: ServerUnregisterClient: UserId (76561198995742987), Result: (EOS_Success)
```
* note after a map_travel event is will disconnect each player and then rejoin them
* the second disconnect was actually the player leaving
- on GAME_OVER lets make sure the current match is ended gracefully, we will also want to make sure a players match_player_stats status is set to finished and is_currently_connected should be false
- on map_travel and map_load events we are creating a match

## game over map loading verification -->

<!-- ## player info
- players IP addresses should be tracked incase we need to do an IP ban for future use.
- an an IP field onto the players collection.
- [2025.11.08-17.35.54:333][663]LogNet: Server accepting post-challenge connection from: 127.0.0.1:52405
- the above is what a log for this looks like, we need to extract their ip
- this will be kept in memory using app.store()
- on the next player_login event, if there is a key for app.Store().Get("serverextid:lastIP") then we will use it to create the player. after it is used we will remove the key
- if we hit another connection event before it is used, we should discard both, because we cant be sure which client will actually connect first, and we will hopefully get it the next time they connect -->


<!-- ## insurgency server crashing recovery
COMPLETED:
- our servers collection has a file_creation_time, the fist line of a log file will be similar to "Log file open, 11/10/25 20:58:31"
- lets create another parser event for when we hit this
- the handler should update file_creation_time
- after this the event the next one that would be hit is a mapload event
- in either of these handlers we need to ensure there is not already an active match for the server, if there is we need to end it -->

<!-- ## weapon types
COMPLETED:
adding a weapon type to match_weapon_stats when creating a record
- BP_Projectile_F1_C_2147479037 -> Projectile
- BP_Firearm_M16A4_C_2147478730 -> Firearm -->

<!-- ## chat command parser should be refactored to also use our event architechure
COMPLETED:
- business logic should be moved to /handlers -->


## clean this repo for unused/out of date files

