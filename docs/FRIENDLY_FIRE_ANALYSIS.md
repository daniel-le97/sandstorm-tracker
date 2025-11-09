# Friendly Fire Analysis System

## Overview

The friendly fire incident tracking system captures detailed context about each teamkill to enable analysis and classification of whether kills were accidental or intentional.

## Database Schema

### Table: `friendly_fire_incidents`

Stores individual friendly fire kills with contextual information:

**Core Fields:**

- `match` - Match where FF occurred (relation to matches)
- `killer` - Player who committed FF (relation to players)
- `victim` - Player who was killed (relation to players)
- `weapon` - Weapon used (e.g., "M67_Grenade", "M4A1", "Technical_Gun")
- `timestamp` - Exact time of the incident

**Context Fields for Analysis:**

- `killer_team` / `victim_team` - Team IDs (should match for true FF)
- `time_since_match_start_seconds` - How long into the match (spawn kills vs mid-game)
- `time_since_last_ff_seconds` - Time between this and killer's previous FF
- `killer_total_kills_in_match` - Total kills by killer at time of FF
- `killer_ff_count_in_match` - Number of FF kills by this player so far in match
- `is_explosive_weapon` - Boolean (grenades, RPGs, airstrikes have higher accident rate)
- `is_vehicle_weapon` - Boolean (vehicle weapons often accident prone)
- `map` / `game_mode` - Environmental context

**Classification Fields:**

- `accident_classification` - Enum: likely_accident, possibly_intentional, likely_intentional, unclassified
- `confidence_score` - Float 0.0-1.0 indicating confidence in classification

## Usage

### 1. Recording a Friendly Fire Incident

When your log parser detects a teamkill:

```go
import (
    "time"
    "sandstorm-tracker/internal/database"
)

// In your kill event handler
if isTeamKill {
    // Determine weapon type
    isExplosive := strings.Contains(weapon, "Grenade") ||
                   strings.Contains(weapon, "RPG") ||
                   strings.Contains(weapon, "Airstrike")

    isVehicle := strings.Contains(weapon, "Technical") ||
                 strings.Contains(weapon, "Helicopter")

    // Get current match context
    match, _ := database.GetActiveMatch(ctx, app, serverID)
    timeSinceStart := time.Since(*match.StartTime).Seconds()

    // Get killer's current stats
    killerStats, _ := getKillerStatsFromMatch(match.ID, killerID)

    // Get time since last FF for this killer
    recentFF, _ := database.GetRecentFriendlyFireIncidents(ctx, app, match.ID, &killerID, 1)
    var timeSinceLastFF *float64
    if len(recentFF) > 0 {
        diff := time.Since(recentFF[0].Timestamp).Seconds()
        timeSinceLastFF = &diff
    }

    // Record the incident
    incident := &database.FriendlyFireIncident{
        MatchID:                   match.ID,
        KillerID:                  killerID,
        VictimID:                  victimID,
        Weapon:                    weapon,
        Timestamp:                 time.Now(),
        KillerTeam:                &killerTeam,
        VictimTeam:                &victimTeam,
        TimeSinceMatchStartSecs:   &timeSinceStart,
        TimeSinceLastFFSecs:       timeSinceLastFF,
        KillerTotalKillsInMatch:   &killerStats.TotalKills,
        KillerFFCountInMatch:      &killerStats.FFCount,
        IsExplosiveWeapon:         isExplosive,
        IsVehicleWeapon:           isVehicle,
        Map:                       match.Map,
        GameMode:                  &match.Mode,
    }

    err := database.RecordFriendlyFireIncident(ctx, app, incident)
    if err != nil {
        log.Printf("Failed to record FF incident: %v", err)
    }
}
```

### 2. Analyzing Player Patterns

Get aggregate stats about a player's friendly fire history:

```go
stats, err := database.GetFriendlyFireStats(ctx, app, playerID)
if err != nil {
    log.Printf("Error getting FF stats: %v", err)
}

fmt.Printf("FF Rate: %.2f%%\n", stats["ff_percentage"])
fmt.Printf("Most common weapon: %s\n", stats["most_common_ff_weapon"])
fmt.Printf("Explosive FF count: %d\n", stats["explosive_ff_count"])
```

### 3. Classifying Behavior

Automatically classify if a player's FF pattern is accidental or intentional:

```go
classification, confidence, err := database.GetFriendlyFirePattern(ctx, app, playerID)
if err != nil {
    log.Printf("Error classifying FF pattern: %v", err)
}

fmt.Printf("Classification: %s (confidence: %.2f)\n", classification, confidence)

// Possible classifications:
// - "likely_accident" (confidence > 0.65) - Normal gameplay accidents
// - "possibly_intentional" (0.40 < confidence < 0.65) - Borderline cases
// - "likely_intentional" (confidence > 0.70) - Pattern suggests griefing
// - "unclassified" - Not enough data
```

## Classification Algorithm

The current heuristic-based classifier considers:

### Accident Indicators (higher score = more likely accident):

- **Low FF rate**: < 5% of total kills
- **Explosive/vehicle weapons**: > 70% of FF with explosives/vehicles
- **Time between incidents**: Average > 120 seconds
- **Mid-game kills**: Few FF in first minute (spawn area)

### Intentional Indicators (higher score = more likely intentional):

- **High FF rate**: > 20% of total kills
- **Spawn camping**: > 3 FF kills in first minute
- **Rapid succession**: Average < 30 seconds between FF
- **Direct fire weapons**: Low explosive/vehicle weapon percentage

## Advanced Analysis Queries

### Most Common FF Weapon by Map

```sql
SELECT map, weapon, COUNT(*) as count
FROM friendly_fire_incidents
WHERE is_explosive_weapon = 0
GROUP BY map, weapon
ORDER BY count DESC
```

### Players with Suspicious Patterns

```sql
SELECT
    p.name,
    COUNT(*) as total_ff,
    (SELECT SUM(kills) FROM match_player_stats WHERE player = p.id) as total_kills,
    AVG(time_since_last_ff_seconds) as avg_time_between
FROM friendly_fire_incidents ff
JOIN players p ON p.id = ff.killer
WHERE accident_classification = 'likely_intentional'
GROUP BY p.id
HAVING total_ff > 10
ORDER BY total_ff DESC
```

### Time-of-Match Distribution

```sql
SELECT
    CASE
        WHEN time_since_match_start_seconds < 60 THEN '0-1 min'
        WHEN time_since_match_start_seconds < 300 THEN '1-5 min'
        WHEN time_since_match_start_seconds < 600 THEN '5-10 min'
        ELSE '10+ min'
    END as time_bracket,
    COUNT(*) as incidents,
    AVG(CASE WHEN is_explosive_weapon = 1 THEN 1.0 ELSE 0.0 END) as explosive_rate
FROM friendly_fire_incidents
GROUP BY time_bracket
```

## Future Enhancements

### Machine Learning Integration

Once you have enough data (1000+ incidents), you can:

1. **Export training data**:

   ```sql
   SELECT
       killer_ff_count_in_match,
       time_since_match_start_seconds,
       time_since_last_ff_seconds,
       is_explosive_weapon,
       is_vehicle_weapon,
       (killer_ff_count_in_match * 1.0 / killer_total_kills_in_match) as ff_rate_in_match,
       accident_classification as label
   FROM friendly_fire_incidents
   WHERE accident_classification IS NOT NULL
   ```

2. **Train a classifier** (Python example):

   ```python
   from sklearn.ensemble import RandomForestClassifier
   from sklearn.model_selection import train_test_split

   # Load data, train model
   model = RandomForestClassifier()
   model.fit(X_train, y_train)

   # Feature importance analysis
   print(model.feature_importances_)
   ```

3. **Update Go code** to call ML model via API or ONNX runtime

### Admin Dashboard Features

- Real-time FF alerts for suspicious patterns
- Player FF history visualization
- Weapon-specific FF rates
- Map hotspots for FF incidents
- Appeal system for false positives

## Integration with Admin Actions

Example workflow for automated moderation:

```go
// After recording FF incident
classification, confidence, err := database.GetFriendlyFirePattern(ctx, app, killerID)

if classification == "likely_intentional" && confidence > 0.80 {
    // Get recent history
    recentIncidents, _ := database.GetRecentFriendlyFireIncidents(ctx, app, matchID, &killerID, 5)

    if len(recentIncidents) >= 3 {
        // 3+ likely intentional FF in this match
        // Take action: warn, kick, or ban
        sendAdminAlert(killerID, "Suspicious FF pattern detected")

        // Option: auto-kick after threshold
        if len(recentIncidents) >= 5 {
            kickPlayer(serverID, killerID, "Excessive team killing")
        }
    }
}
```

## Performance Considerations

- Indexes on `match`, `killer`, `timestamp`, and `accident_classification` ensure fast queries
- Consider archiving old incidents (> 6 months) to separate table if data grows large
- Classification can run asynchronously after incident recording to avoid blocking gameplay
