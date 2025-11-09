# Chat Commands Feature

## Overview

The chat command system allows players to query their stats and server information directly via in-game chat commands. The system detects commands (starting with `!`) in the LogChat output, processes them, and responds via RCON `say` commands.

## Available Commands

### `!kdr` or `!kills`

Shows player's kills, deaths, and K/D ratio across all matches.

- **Example response:** `ArmoredBear - Kills: 42, Deaths: 15, K/D: 2.80`

### `!stats` or `!rank`

Shows player's overall stats including score, playtime, score/minute, and rank.

- **Example response:** `ArmoredBear - Score: 1250, Time: 45m, Score/min: 27.78, Rank: #3`

### `!top` or `!leaderboard`

Shows top 3 players by score per minute.

- **Example response:** `Top Players: 1. ProGamer (45.2), 2. Elite (40.8), 3. ArmoredBear (27.8)`

### `!guns` or `!weapons`

Shows player's personal top 3 weapons by kills across all their matches.

- **Example response:** `ArmoredBear's Top 3 Weapons by Kills:`
  - `#1: M4A1 - 42 kills`
  - `#2: AKM - 38 kills`
  - `#3: M16A2 - 25 kills`

## Implementation Details

### Log Pattern

Chat commands are detected using this log pattern:

```
[timestamp][frame]LogChat: Display: PlayerName(SteamID) Global Chat: !command
```

The regex pattern handles optional whitespace in the frame counter field to accommodate both `[427]` and `[ 34]` formats.

### Architecture

```
Log File → Parser → ChatCommandHandler → Database Helpers → RCON Response
                                          ↓
                                    Stats Queries
```

**Components:**

1. **Parser (`internal/parser/parser.go`)**: Detects chat command pattern
2. **ChatCommandHandler (`internal/parser/chat_commands.go`)**: Routes commands to handlers
3. **Database Helpers (`internal/database/helpers.go`)**: Aggregates stats from PocketBase
4. **RCON Integration (`internal/app/app.go`)**: Sends responses back to game server

### Database Queries

#### K/D Stats

```go
GetPlayerTotalKD(ctx, playerID) -> (kills, deaths, kd)
```

Sums kills and deaths across all `match_player_stats` records.

#### Player Stats

```go
GetPlayerStats(ctx, playerID) -> PlayerStats{score, duration, scorePerMin}
```

Aggregates score and playtime, calculates score per minute.

#### Player Rank

```go
GetPlayerRank(ctx, playerID) -> (rank, totalPlayers)
```

Calculates player's position based on score/minute among all players.

#### Top Players

```go
GetTopPlayersByScorePerMin(ctx, limit) -> []TopPlayer
```

Returns leaderboard of players with highest score per minute.

#### Top Weapons

```go
GetTopWeapons(ctx, playerID, limit) -> []TopWeapon
```

Returns a specific player's most-used weapons by kill count. Aggregates weapon kills from `match_weapon_stats` filtered by the player ID.

### Response Format

All responses are sent via RCON using the `say <message>` command, which broadcasts to all players in the server.

## Testing

### Unit Tests

- **Pattern Matching** (`TestChatCommandParsing`): Verifies regex matches all command formats
- **Handler Integration** (`TestChatCommandHandler`): Tests command routing with mock RCON

### Test Coverage

- All 4 commands tested with real log line formats
- Edge cases: regular chat (no command), non-chat lines
- Whitespace handling in log frame counter

## Configuration

No additional configuration required. The system automatically:

- Detects chat commands in log files
- Queries player stats from existing database
- Responds via the configured RCON connection

## Dependencies

- **Parser**: Requires `ChatCommand` regex pattern in `newLogPatterns()`
- **Database**: Uses existing PocketBase collections (players, match_player_stats, match_weapon_stats)
- **RCON**: Requires RCON pool initialized in app (via `SetChatHandler`)

## Future Enhancements

Possible additions:

- `!server` - Server info (map, mode, player count)
- `!time` - Match time remaining
- `!score` - Current match score
- `!help` - List available commands
- Player-specific targeting: `!stats PlayerName`
- Cooldown system to prevent spam
- Private responses (via admin broadcast or player-specific RCON)

## Known Limitations

1. **Public Responses**: All responses are broadcast to entire server (RCON `say` limitation)
2. **No Persistence**: Command history not stored
3. **No Validation**: Assumes player exists in database (commands used during active match)
4. **Case Sensitivity**: Commands are case-insensitive but player names preserve original case

## Example Usage

**Player types in game:**

```
!kdr
```

**Server log shows:**

```
[2025.10.21-20.11.28:617][ 34]LogChat: Display: ArmoredBear(76561198995742987) Global Chat: !kdr
```

**Tracker processes command:**

1. Parser detects `!kdr` command for player 76561198995742987
2. Handler queries database for ArmoredBear's stats
3. Finds: 42 kills, 15 deaths
4. Calculates K/D: 2.80
5. Sends RCON: `say ArmoredBear - Kills: 42, Deaths: 15, K/D: 2.80`

**Player sees in game:**

```
[Server] ArmoredBear - Kills: 42, Deaths: 15, K/D: 2.80
```
