# Insurgency: Sandstorm Stats Tracker

A comprehensive real-time statistics tracking system for Insurgency: Sandstorm game servers.

## Features

### 📊 Real-time Event Processing

- Monitors game log files for real-time events
- Parses player joins, kills, deaths, and chat commands
- Intelligent weapon name mapping (e.g., `BP_Firearm_M16A4_C_123` → `M16A4`)

### 🗄️ SQLite Database Storage

- Persistent storage of player statistics
- Tracks kills, deaths, K/D ratios, weapon usage, playtime
- Optimized with indexes and prepared statements
- Session tracking for accurate playtime calculation

### 🤖 Chat Commands

Players can use these commands in-game:

- `!stats` - Show personal statistics (kills, deaths, K/D, playtime)
- `!kdr` - Show K/D ratio specifically
- `!top` - Show top 3 players by kills
- `!guns` - Show personal weapon statistics

### ⚡ Performance Optimized

- Debounced file watching (prevents spam during heavy logging)
- Prepared SQL statements for fast database operations
- Only processes events defined in `events.md`
- Size-based change detection to avoid reprocessing

## Database Schema

### Tables

- **players** - Basic player information and Steam IDs
- **player_sessions** - Track when players join/leave for playtime
- **kills** - Individual kill/death records with weapon info
- **maps** - Map information and rotation
- **map_rounds** - Individual game rounds and outcomes
- **player_round_stats** - Per-player, per-round statistics
- **weapon_stats** - Aggregated weapon usage statistics
- **chat_commands** - Command usage tracking

## Quick Start

1. **Install dependencies:**

   ```bash
   bun install
   ```

2. **Configure log path:**
   Edit `index.ts` and set your server's log file path:

   ```typescript
   const path = "C:\\Path\\To\\Your\\Server\\Logs";
   const serverId = "your-server-id";
   ```

3. **Run the tracker:**

   ```bash
   bun index.ts
   ```

4. **Database will be created automatically** at `sandstorm_stats.db`

## Testing

Run the comprehensive test suite:

```bash
# Test event parsing (25 test cases)
bun tests/tests.ts

# Test database functionality
bun test-database.ts

# Test chat commands
bun test-commands.ts

# Test complete integration
bun test-integration.ts
```

## File Structure

```
├── index.ts              # Main file watcher and event processor
├── events.ts             # Event parsing and weapon name mapping
├── database.ts           # SQLite schema and prepared statements
├── stats-service.ts      # Business logic for processing events
├── command-handler.ts    # Chat command responses
├── events.md             # Documentation of tracked events
├── tests/
│   └── tests.ts          # Comprehensive test suite
└── sandstorm_stats.db    # SQLite database (auto-created)
```

## Event Types Tracked

Based on `events.md`, the system tracks:

- Player joins/leaves
- Player kills (with weapon details)
- Team kills and suicides
- Round starts/ends
- Map changes
- Chat commands
- Admin actions
- Server events

## Weapon Name Mapping

The system includes mappings for 70+ weapons from blueprint paths to user-friendly names:

- `BP_Firearm_M16A4_C_*` → `M16A4`
- `BP_Firearm_AK74_C_*` → `AK-74`
- `BP_Explosive_F1_C_*` → `F1 Grenade`
- And many more...

## K/D Ratio Calculation

- When deaths = 0: Shows number of kills (e.g., 5 kills / 0 deaths = 5.0 K/D)
- When deaths > 0: Shows calculated ratio (e.g., 10 kills / 5 deaths = 2.0 K/D)
- Rounds to 2 decimal places for precision

## Performance Notes

- File watching uses debouncing (200ms) to handle rapid log updates
- Database operations use prepared statements for optimal performance
- Only processes the last line of log files to minimize CPU usage
- Size-based change detection prevents reprocessing the same content

## Chat Command Examples

```
Player: !stats
Bot: 📊 PlayerName: 25 kills, 8 deaths, K/D: 3.13, Score/min: 0, Playtime: 2h

Player: !kdr
Bot: 💀 PlayerName: 25 kills, 8 deaths, K/D ratio: 3.13

Player: !top
Bot: 🏆 Top 3 Players:
1. Alice: 45 kills (K/D: 4.5)
2. Bob: 32 kills (K/D: 2.67)
3. Charlie: 28 kills (K/D: 3.11)

Player: !guns
Bot: 🔫 PlayerName's Top Weapons:
1. M16A4: 15 kills
2. AK-74: 8 kills
3. M9 Pistol: 2 kills
```

## Requirements

- Bun runtime (latest version)
- Read access to Insurgency: Sandstorm server log files
- Windows PowerShell (for the current setup)

---

🎮 **Ready to track your server's statistics!** Simply run `bun index.ts` and watch the real-time stats come alive.
