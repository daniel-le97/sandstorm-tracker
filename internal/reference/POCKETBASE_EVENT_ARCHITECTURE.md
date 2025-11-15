# Simplifying Your Event-Driven Architecture with PocketBase Hooks

## Current Architecture (Complex)

You currently have:

1. **Custom Event Bus** in `/internal/events/` with pub/sub pattern
2. **Parser** that detects events and processes them directly
3. **Manual coordination** between parser, watcher, jobs, etc.
4. **Separate event publishing** logic scattered throughout

## Simplified Architecture (Using PocketBase)

Instead, leverage PocketBase's built-in hook system as your event bus!

### Key Insight: PocketBase Records = Events

**Think of database writes as event publishing:**

- Parser writes a record to `match_events` collection → Event published!
- PocketBase triggers `OnRecordCreate("match_events")` → Event consumed!

## How It Works

### 1. Parser Becomes Simple Record Creator

```go
// OLD WAY: Parser does everything
func (p *LogParser) ProcessKill(line string) {
    // Parse kill
    kill := parseKill(line)

    // Store in database
    p.storeKill(kill)

    // Trigger score update
    p.scoreDebouncer.TriggerScoreUpdate(serverID)

    // Update match stats
    p.updateMatchStats(serverID)

    // Maybe send webhooks?
    // Maybe update real-time dashboard?
    // More and more responsibilities...
}

// NEW WAY: Parser just creates records
func (p *LogParser) ProcessKill(line string) {
    kill := parseKill(line)

    // Just create a record - that's it!
    record := core.NewRecord(p.pbApp.Collection("match_events"))
    record.Set("type", "player_kill")
    record.Set("server_id", serverID)
    record.Set("data", kill)
    record.Set("timestamp", time.Now())

    p.pbApp.Save(record)  // This triggers OnRecordCreate!
}
```

### 2. Event Handlers Subscribe via PocketBase Hooks

```go
// In your app.Bootstrap()
func (app *App) Bootstrap() error {
    // Subscribe to kill events
    app.OnRecordCreate("match_events").BindFunc(func(e *core.RecordEvent) error {
        eventType := e.Record.GetString("type")

        switch eventType {
        case "player_kill":
            return app.handlePlayerKill(e)
        case "player_join":
            return app.handlePlayerJoin(e)
        case "match_start":
            return app.handleMatchStart(e)
        case "round_end":
            return app.handleRoundEnd(e)
        }

        return e.Next()
    })

    // Subscribe to match updates
    app.OnRecordUpdate("matches").BindFunc(func(e *core.RecordEvent) error {
        return app.handleMatchUpdate(e)
    })

    return nil
}

// Handler functions are clean and focused
func (app *App) handlePlayerKill(e *core.RecordEvent) error {
    serverID := e.Record.GetString("server_id")

    // Trigger score update
    if app.scoreDebouncer != nil {
        app.scoreDebouncer.TriggerScoreUpdate(serverID)
    }

    // Update real-time dashboard
    app.notifyDashboard(e.Record)

    return e.Next()
}

func (app *App) handleRoundEnd(e *core.RecordEvent) error {
    serverID := e.Record.GetString("server_id")

    // Immediate score calculation on round end
    if app.scoreDebouncer != nil {
        app.scoreDebouncer.ExecuteImmediately(serverID)
    }

    return e.Next()
}
```

### 3. Database Collections as Event Streams

Create collections that represent events:

**`match_events` collection:**

```javascript
{
    "type": "string",           // "player_kill", "player_join", "round_start", etc.
    "server_id": "string",      // Which server
    "match_id": "relation",     // Link to match
    "timestamp": "datetime",
    "data": "json"              // Event-specific data
}
```

This gives you:

- ✅ **Event history** - All events are persisted automatically
- ✅ **Event replay** - Query past events anytime
- ✅ **Real-time updates** - PocketBase subscriptions work out of the box
- ✅ **Event filtering** - Use PocketBase's query API
- ✅ **Event auditing** - Built-in timestamps and tracking

## Complete Simplified Flow

```
┌─────────────────────────────────────────────────────────────┐
│ 1. LOG FILE CHANGES                                         │
│    Watcher detects new log lines                            │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│ 2. PARSER (Simplified)                                      │
│    - Parse log line                                         │
│    - Create record in "match_events"                        │
│    - Done! No other responsibilities                        │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼  (PocketBase Save triggers hooks)
┌─────────────────────────────────────────────────────────────┐
│ 3. POCKETBASE HOOKS (Your Event Bus)                        │
│    app.OnRecordCreate("match_events")                       │
│    ├─ handlePlayerKill() → Trigger score update            │
│    ├─ handlePlayerJoin() → Update player count             │
│    ├─ handleRoundEnd()   → Calculate final scores          │
│    └─ handleMatchStart() → Initialize match tracking       │
└─────────────────────────────────────────────────────────────┘
                      │
                      ▼ (Multiple subscribers can react)
┌─────────────────────────────────────────────────────────────┐
│ 4. SUBSCRIBERS (Decoupled)                                  │
│    - Score updater                                          │
│    - Real-time dashboard                                    │
│    - Discord webhooks                                       │
│    - Statistics aggregator                                  │
│    - ML model updater                                       │
└─────────────────────────────────────────────────────────────┘
```

## Benefits of This Approach

### 1. **Drastically Simpler Code**

```go
// OLD: Parser has many responsibilities
type LogParser struct {
    pbApp          core.App
    scoreDebouncer ScoreDebouncer
    dashboardNotifier DashboardNotifier
    webhookSender WebhookSender
    statsAggregator StatsAggregator
    // ... more dependencies
}

// NEW: Parser has one job
type LogParser struct {
    pbApp core.App
}
```

### 2. **Easy to Add Features**

Want to add Discord notifications? Just add another hook:

```go
app.OnRecordCreate("match_events").BindFunc(func(e *core.RecordEvent) error {
    if e.Record.GetString("type") == "player_kill" {
        return sendDiscordNotification(e.Record)
    }
    return e.Next()
})
```

### 3. **Built-in Real-time Updates**

Your frontend can subscribe directly:

```javascript
pb.collection("match_events").subscribe("*", (e) => {
  if (e.record.type === "player_kill") {
    updateKillFeed(e.record);
  }
});
```

### 4. **Event History for Free**

Query past events anytime:

```go
events, _ := app.FindRecordsByFilter(
    "match_events",
    "match_id = {:matchId} && type = 'player_kill'",
    dbx.Params{"matchId": matchID},
)
```

### 5. **Testability**

Test hooks independently:

```go
func TestHandlePlayerKill(t *testing.T) {
    app := setupTestApp()

    // Create test event record
    record := createTestKillEvent()

    // Test handler directly
    err := app.handlePlayerKill(&core.RecordEvent{
        Record: record,
    })

    assert.NoError(t, err)
}
```

## Migration Strategy

### Phase 1: Keep Current System, Add Events

```go
func (p *LogParser) ProcessKill(line string) {
    kill := parseKill(line)

    // Old behavior (keep working)
    p.storeKill(kill)
    p.scoreDebouncer.TriggerScoreUpdate(serverID)

    // NEW: Also create event record
    p.createEventRecord("player_kill", serverID, kill)
}
```

### Phase 2: Move Logic to Hooks

```go
app.OnRecordCreate("match_events").BindFunc(func(e *core.RecordEvent) error {
    if e.Record.GetString("type") == "player_kill" {
        // Move score update logic here
        app.scoreDebouncer.TriggerScoreUpdate(
            e.Record.GetString("server_id"),
        )
    }
    return e.Next()
})
```

### Phase 3: Remove Old Code

```go
func (p *LogParser) ProcessKill(line string) {
    kill := parseKill(line)

    // Just create event - hooks handle the rest!
    p.createEventRecord("player_kill", serverID, kill)
}
```

## Recommendations

### 1. Use a Single `match_events` Collection

All parser events go here with a `type` field:

- `player_kill`
- `player_join`
- `player_leave`
- `round_start`
- `round_end`
- `match_start`
- `match_end`
- `objective_captured`

### 2. Keep Match/Player Tables Separate

Use `match_events` for the event stream, but keep your existing:

- `matches` - Current match state
- `players` - Player records
- `match_player_stats` - Aggregated stats

### 3. Use Hook Priorities

PocketBase hooks run in order, so you can chain:

```go
// Priority 1: Update state
app.OnRecordCreate("match_events").BindFunc(func(e *core.RecordEvent) error {
    return updateMatchState(e)
})

// Priority 2: Trigger calculations (after state is updated)
app.OnRecordCreate("match_events").BindFunc(func(e *core.RecordEvent) error {
    return triggerScoreUpdate(e)
})
```

### 4. Error Handling

Hooks should not block the pipeline:

```go
app.OnRecordCreate("match_events").BindFunc(func(e *core.RecordEvent) error {
    // Non-critical operations shouldn't fail the event
    if err := sendWebhook(e); err != nil {
        app.Logger().Error("webhook failed", "error", err)
        // Don't return error - allow other hooks to run
    }
    return e.Next()
})
```

## Your Custom Event Bus is Redundant

You don't need `/internal/events/` anymore because:

- ❌ EventBus.Publish() → ✅ PocketBase Save()
- ❌ EventBus.Subscribe() → ✅ app.OnRecordCreate()
- ❌ Custom event types → ✅ Collection records
- ❌ Manual threading → ✅ PocketBase handles it
- ❌ Custom serialization → ✅ JSON in database

## Summary

**Before:**

- Parser → EventBus.Publish() → Multiple subscribers → Database
- Complex event routing logic
- Manual coordination
- Hard to debug event flow

**After:**

- Parser → PocketBase Save() → Hooks triggered automatically
- Database IS the event bus
- PocketBase handles coordination
- Easy to trace events in database

This is the **"database as event stream"** pattern, and PocketBase is perfectly designed for it!
