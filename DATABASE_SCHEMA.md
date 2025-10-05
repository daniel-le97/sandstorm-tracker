# Sandstorm Tracker Database Schema Relationship Model

## Overview

This is a multi-server game statistics tracking system for Insurgency: Sandstorm with comprehensive player, match, and performance tracking capabilities.

## Entity Relationship Diagram (Text-based)

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│     SERVERS     │────│     PLAYERS     │────│ PLAYER_SESSIONS │
│  (Multi-server) │    │  (Per-server)   │    │   (Join/Leave)  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│     MATCHES     │    │      KILLS      │    │   WEAPON_STATS  │
│ (Game sessions) │    │ (Kill tracking) │    │ (Per-weapon)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│MATCH_PARTICIPANTS│    │      MAPS       │    │ CHAT_COMMANDS   │
│  (Who played)   │    │   (Map info)    │    │ (!stats, !kdr)  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │
         │                       │
         ▼                       ▼
┌─────────────────┐    ┌─────────────────┐
│   MATCH_MAPS    │    │   MAP_ROUNDS    │
│ (Maps played)   │    │ (Round results) │
└─────────────────┘    └─────────────────┘
         │                       │
         │                       │
         ▼                       ▼
┌─────────────────┐    ┌─────────────────┐
│ SCHEMA_VERSION  │    │PLAYER_ROUND_STATS│
│  (Migration)    │    │ (Per-round perf) │
└─────────────────┘    └─────────────────┘
```

## Detailed Table Relationships

### Core Tables

#### 🖥️ **SERVERS** (Central Hub)

```sql
id (PK), server_id (UNIQUE), server_name, config_id (UNIQUE),
log_path, enabled, description, created_at, updated_at
```

**Relationships:**

- **One-to-Many** → `players`, `matches`, `kills`, `maps`, `player_sessions`, `weapon_stats`, etc.
- **Purpose:** Central configuration for multi-server tracking

#### 👤 **PLAYERS** (Per-Server Players)

```sql
id (PK), steam_id, player_name, server_id (FK → servers.id),
first_seen, last_seen, total_playtime_minutes, created_at, updated_at
UNIQUE(steam_id, server_id) -- Same player on different servers
```

**Relationships:**

- **Many-to-One** → `servers`
- **One-to-Many** → `player_sessions`, `kills` (as killer/victim), `weapon_stats`, `chat_commands`
- **Purpose:** Track unique players per server with aggregate stats

### Session & Activity Tracking

#### ⏱️ **PLAYER_SESSIONS** (Join/Leave Tracking)

```sql
id (PK), player_id (FK → players.id), server_id (FK → servers.id),
join_time, leave_time, duration_minutes, map_name, created_at
```

**Relationships:**

- **Many-to-One** → `players`, `servers`
- **Purpose:** Track individual play sessions for playtime calculation

#### 🎮 **MATCHES** (Game Sessions)

```sql
id (PK), server_id (FK → servers.id), match_name, start_time, end_time,
duration_minutes, status (active/completed/aborted), total_players,
max_players, created_at, updated_at
```

**Relationships:**

- **Many-to-One** → `servers`
- **One-to-Many** → `match_participants`, `match_maps`
- **Purpose:** Track complete game sessions/matches per server

#### 👥 **MATCH_PARTICIPANTS** (Who Played in Match)

```sql
id (PK), match_id (FK → matches.id), player_id (FK → players.id),
server_id (FK → servers.id), join_time, leave_time, duration_minutes,
final_kills, final_deaths, final_team_kills, final_suicides, final_score
UNIQUE(match_id, player_id) -- One entry per player per match
```

**Relationships:**

- **Many-to-One** → `matches`, `players`, `servers`
- **Purpose:** Link players to matches with final stats

### Combat & Performance Tracking

#### ⚔️ **KILLS** (Kill Event Tracking)

```sql
id (PK), killer_id (FK → players.id), victim_id (FK → players.id),
server_id (FK → servers.id), weapon, kill_type (player_kill/team_kill/suicide),
killer_team, victim_team, map_name, round_number, timestamp, created_at
```

**Relationships:**

- **Many-to-One** → `players` (killer), `players` (victim), `servers`
- **Purpose:** Record every kill event with context

#### 🔫 **WEAPON_STATS** (Per-Player Weapon Performance)

```sql
id (PK), player_id (FK → players.id), server_id (FK → servers.id),
weapon_name, kills, deaths, team_kills, suicides, created_at, updated_at
UNIQUE(player_id, weapon_name, server_id) -- One record per player/weapon/server
```

**Relationships:**

- **Many-to-One** → `players`, `servers`
- **Purpose:** Aggregate weapon statistics per player per server

#### 📊 **PLAYER_ROUND_STATS** (Per-Round Performance)

```sql
id (PK), player_id (FK → players.id), round_id (FK → map_rounds.id),
server_id (FK → servers.id), kills, deaths, team_kills, suicides, score
UNIQUE(player_id, round_id) -- One record per player per round
```

**Relationships:**

- **Many-to-One** → `players`, `map_rounds`, `servers`
- **Purpose:** Track detailed per-round player performance

### Map & Round Tracking

#### 🗺️ **MAPS** (Map Information)

```sql
id (PK), map_name, scenario, server_id (FK → servers.id),
created_at, updated_at
UNIQUE(map_name, server_id) -- Same map can exist on multiple servers
```

**Relationships:**

- **Many-to-One** → `servers`
- **One-to-Many** → `map_rounds`, `match_maps`
- **Purpose:** Track available maps per server

#### 🏁 **MAP_ROUNDS** (Round Results)

```sql
id (PK), map_id (FK → maps.id), server_id (FK → servers.id),
round_number, winning_team, win_reason, end_time, duration_seconds
```

**Relationships:**

- **Many-to-One** → `maps`, `servers`
- **One-to-Many** → `player_round_stats`
- **Purpose:** Track individual round outcomes

#### 🎯 **MATCH_MAPS** (Maps Played in Match)

```sql
id (PK), match_id (FK → matches.id), map_id (FK → maps.id),
server_id (FK → servers.id), sequence_order, start_time, end_time
UNIQUE(match_id, map_id, sequence_order) -- Ordered map rotation
```

**Relationships:**

- **Many-to-One** → `matches`, `maps`, `servers`
- **Purpose:** Track which maps were played in each match

### Communication & Commands

#### 💬 **CHAT_COMMANDS** (Command Usage Tracking)

```sql
id (PK), player_id (FK → players.id), server_id (FK → servers.id),
command (!stats, !kdr, !top, !guns), arguments, timestamp, created_at
```

**Relationships:**

- **Many-to-One** → `players`, `servers`
- **Purpose:** Track usage of chat commands for analytics

### System Management

#### 🔄 **SCHEMA_VERSION** (Database Migration Tracking)

```sql
id (PK), version, migration_date
```

**Relationships:** None (system table)
**Purpose:** Track database schema versions for migrations

## Key Design Principles

### 🎯 **Multi-Server Architecture**

- Every data table includes `server_id` foreign key
- Players are unique per server (same Steam ID can exist on multiple servers)
- Statistics are isolated per server but can be aggregated

### 🔗 **Referential Integrity**

- All foreign keys use `ON DELETE CASCADE` for clean data removal
- Unique constraints prevent duplicate data
- Indexes optimize query performance

### 📈 **Performance Optimization**

```sql
-- Key indexes for fast queries:
idx_players_steam_server      -- Fast player lookups
idx_kills_timestamp          -- Chronological kill queries
idx_matches_status           -- Active match queries
idx_weapon_stats_player_id   -- Player weapon stats
```

### 🎮 **Game Event Flow**

```
Player Joins Server → PLAYER_SESSIONS (join_time)
                   → PLAYERS (first_seen, last_seen updated)

Kill Event        → KILLS (individual event)
                   → WEAPON_STATS (aggregated stats)
                   → PLAYER_ROUND_STATS (round performance)

Match Start       → MATCHES (new match)
                   → MATCH_PARTICIPANTS (players join)
                   → MATCH_MAPS (maps in rotation)

Player Leaves     → PLAYER_SESSIONS (leave_time, duration)
                   → PLAYERS (total_playtime_minutes updated)
```

### 📊 **Common Query Patterns**

#### Player Statistics

```sql
-- Get player stats for specific server
SELECT p.player_name, p.total_playtime_minutes,
       COUNT(k1.id) as kills, COUNT(k2.id) as deaths
FROM players p
LEFT JOIN kills k1 ON p.id = k1.killer_id
LEFT JOIN kills k2 ON p.id = k2.victim_id
WHERE p.server_id = ? AND p.steam_id = ?
```

#### Server Leaderboards

```sql
-- Top players by K/D ratio on server
SELECT p.player_name, COUNT(k1.id) as kills, COUNT(k2.id) as deaths,
       CASE WHEN COUNT(k2.id) = 0 THEN COUNT(k1.id)
            ELSE ROUND(CAST(COUNT(k1.id) AS FLOAT) / COUNT(k2.id), 2) END as kdr
FROM players p
LEFT JOIN kills k1 ON p.id = k1.killer_id AND k1.kill_type = 'player_kill'
LEFT JOIN kills k2 ON p.id = k2.victim_id
WHERE p.server_id = ?
GROUP BY p.id ORDER BY kdr DESC
```

#### Weapon Statistics

```sql
-- Player's favorite weapons
SELECT weapon_name, kills, deaths
FROM weapon_stats
WHERE player_id = ? AND server_id = ?
ORDER BY kills DESC
```

This schema provides comprehensive tracking for multi-server game statistics with excellent performance and data integrity!
