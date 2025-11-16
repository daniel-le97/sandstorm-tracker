# Advanced PocketBase Patterns & Techniques

## 1. Custom Hooks with Priority & Transaction Management

### Pattern: Ordered Hook Execution

```go
// Multiple hooks can execute in defined order using Priority
app.OnRecordCreate("players").BindFunc(func(e *core.RecordCreateEvent) error {
	// Priority 1 - runs first
	return e.Next()
}, 1)

app.OnRecordCreate("players").BindFunc(func(e *core.RecordCreateEvent) error {
	// Priority 2 - runs second
	// Can rely on work from Priority 1
	return e.Next()
}, 2)

// All hooks run in a database transaction automatically
// Return error to rollback entire transaction
```

### Pattern: Atomic Multi-Collection Updates

```go
// Your updatePlayersFromRcon pattern - wrap multiple saves in one transaction
err := app.RunInTransaction(func(txApp core.App) error {
	for _, player := range players {
		// Find player
		playerRecord, err := findPlayerByName(txApp, player.Name)
		if err != nil {
			return err // Rollback all updates
		}

		// Update stats - if any save fails, entire batch rolls back
		if err := updatePlayerMatchScore(txApp, logger, matchID, playerRecord.Id, player.Score); err != nil {
			return fmt.Errorf("failed to update player: %w", err)
		}
	}
	return nil // All succeed together
})
```

---

## 2. Event-Driven Architecture

### Pattern: Hook-Based Event Bus

Instead of cron jobs, use hooks as an event bus (your current approach):

```go
// Game events trigger score updates via hooks
app.OnRecordCreate("events").BindFunc(func(e *core.RecordCreateEvent) error {
	eventType := e.Record.GetString("type")
	serverID := e.Record.GetString("server_id")

	// Route to handlers based on event type
	switch eventType {
	case "player_kill":
		scoreDebouncer.TriggerScoreUpdate(serverID)
	case "objective_captured":
		scoreDebouncer.TriggerScoreUpdateFixed(serverID, 2*time.Second)
	case "round_end":
		scoreDebouncer.ExecuteImmediately(serverID)
	}

	return e.Next()
}, 1) // Priority 1 - trigger immediately
```

### Pattern: Debounced Event Processing

Your ScoreDebouncer pattern prevents thundering herd:

```go
// Multiple events within 10s window = 1 query
scoreDebouncer := jobs.NewScoreDebouncer(
	app,
	cfg,
	10*time.Second,  // Debounce window
	30*time.Second,  // Max wait before forcing update
)

// Events trigger debounced update
scoreDebouncer.TriggerScoreUpdate(serverID)
scoreDebouncer.TriggerScoreUpdate(serverID) // Ignored, same window
scoreDebouncer.TriggerScoreUpdate(serverID) // Ignored, same window
// After 10s: Single RCON query executes
```

---

## 3. Custom Record Validation & Enrichment

### Pattern: Pre-Save Validation & Data Enrichment

```go
app.OnRecordBeforeSave("match_player_stats").BindFunc(func(e *core.RecordSaveEvent) error {
	record := e.Record

	// Validation
	score := record.GetInt("score")
	if score < 0 {
		return fmt.Errorf("score cannot be negative")
	}

	// Data enrichment
	createdTime := record.GetDateTime("created")
	playTime := int(time.Since(createdTime.Time()).Seconds())
	record.Set("calculated_playtime", playTime)

	// Cross-collection validation
	matchID := record.GetString("match")
	match, err := e.App.FindRecordById("matches", matchID)
	if err != nil {
		return fmt.Errorf("match not found")
	}

	// Prevent stats from being added to completed matches
	if match.GetDateTime("end_time").Time().After(time.Now()) {
		return fmt.Errorf("cannot add stats to inactive match")
	}

	return e.Next()
}, 1)
```

---

## 4. Complex Queries with Filters

### Pattern: Related Record Filtering

```go
// Get all player stats for matches in the last 7 days
records, err := app.FindRecordsByFilter(
	"match_player_stats",
	"match.created >= {:start} && match.created < {:end}",
	"-match.created", // Sort by match creation time
	1000,
	0,
	map[string]any{
		"start": time.Now().AddDate(0, 0, -7).Format(time.RFC3339),
		"end":   time.Now().Format(time.RFC3339),
	},
)
```

### Pattern: Aggregation via Multiple Queries

```go
// Get top 10 players by kills (requires manual aggregation in Go)
// PocketBase doesn't have GROUP BY, so you aggregate in code

playerKills := make(map[string]int)

records, _ := app.FindRecordsByFilter(
	"match_player_stats",
	"",
	"-created",
	10000, // Get many records
	0,
	map[string]any{},
)

for _, record := range records {
	playerID := record.GetString("player")
	kills := record.GetInt("kills")
	playerKills[playerID] += kills
}

// Convert to sorted slice
type PlayerKill struct {
	PlayerID string
	Kills    int
}
var topKillers []PlayerKill
for pid, k := range playerKills {
	topKillers = append(topKillers, PlayerKill{pid, k})
}
sort.Slice(topKillers, func(i, j int) bool {
	return topKillers[i].Kills > topKillers[j].Kills
})
```

---

## 5. Cron Jobs & Scheduled Tasks

### Pattern: Multiple Cron Jobs with Different Schedules

```go
// Archive old data - daily at 2 AM
scheduler.MustAdd("archive_data", "0 2 * * *", func() {
	archiveOldData(app)
})

// Update rankings - every 6 hours
scheduler.MustAdd("update_rankings", "0 */6 * * *", func() {
	updatePlayerRankings(app)
})

// Health check - every 5 minutes
scheduler.MustAdd("health_check", "*/5 * * * *", func() {
	checkServerHealth(app)
})

// Reset daily stats - midnight UTC
scheduler.MustAdd("reset_daily_stats", "0 0 * * *", func() {
	resetDailyStats(app)
})
```

### Pattern: Cron with Context Timeout

```go
scheduler.MustAdd("long_operation", "0 1 * * *", func() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// This operation will be cancelled if it takes >30 min
	if err := longRunningOperation(ctx, app); err != nil {
		logger.Error("Long operation failed", "error", err)
	}
})
```

---

## 6. Middleware & Request Handling

### Pattern: Custom Authentication Middleware

```go
// Hook into before request
app.OnServe().BindFunc(func(e *core.ServeEvent) error {
	e.Router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Custom auth logic
			token := r.Header.Get("Authorization")

			// Validate token against your system
			if !isValidToken(token) {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	})
	return e.Next()
})
```

### Pattern: Custom Routes with Business Logic

```go
app.OnServe().BindFunc(func(e *core.ServeEvent) error {
	e.Router.AddRoute(echo.Route{
		Method:  http.MethodGet,
		Path:    "/api/top-players",
		Handler: getTopPlayers,
		Middlewares: []echo.MiddlewareFunc{
			// Add auth middleware
		},
	})
	return e.Next()
})

func getTopPlayers(c echo.Context) error {
	app := c.Get(core.ContextKeyApp).(core.App)

	// Complex logic with transactions
	var topPlayers []map[string]any

	err := app.RunInTransaction(func(txApp core.App) error {
		// Get stats
		records, err := txApp.FindRecordsByFilter(
			"match_player_stats",
			"",
			"-kills",
			10,
			0,
			map[string]any{},
		)
		if err != nil {
			return err
		}

		for _, record := range records {
			topPlayers = append(topPlayers, record.PublicExport())
		}
		return nil
	})

	if err != nil {
		return c.JSON(500, map[string]string{"error": err.Error()})
	}

	return c.JSON(200, topPlayers)
}
```

---

## 7. Record Relationships & Cascading Operations

### Pattern: Cascade Deletes via Hooks

```go
// When a player is deleted, clean up their stats
app.OnRecordBeforeDelete("players").BindFunc(func(e *core.RecordDeleteEvent) error {
	playerID := e.Record.Id

	// Find all stats for this player
	stats, err := e.App.FindRecordsByFilter(
		"match_player_stats",
		"player = {:id}",
		"",
		10000,
		0,
		map[string]any{"id": playerID},
	)
	if err != nil {
		return err // Don't delete player if we can't clean up
	}

	// Delete all their stats in transaction
	return e.App.RunInTransaction(func(txApp core.App) error {
		for _, stat := range stats {
			if err := txApp.Delete(stat); err != nil {
				return err
			}
		}
		return nil
	})
}, 1)
```

---

## 8. Batch Operations & Bulk Updates

### Pattern: Efficient Batch Processing

```go
// Update many records efficiently with transaction batching
func bulkUpdatePlayerScores(app core.App, updates map[string]int32) error {
	return app.RunInTransaction(func(txApp core.App) error {
		for playerID, newScore := range updates {
			record, err := txApp.FindRecordById("players", playerID)
			if err != nil {
				continue // Skip if not found
			}

			record.Set("score", newScore)
			if err := txApp.Save(record); err != nil {
				return err // Rollback all on first error
			}
		}
		return nil
	})
}
```

---

## 9. Logging & Monitoring

### Pattern: Audit Trail via Hooks

```go
app.OnRecordAfterUpdate("match_player_stats").BindFunc(func(e *core.RecordUpdateEvent) error {
	// Log who/what changed
	logger.Info("Player stats updated",
		"playerID", e.Record.GetString("player"),
		"matchID", e.Record.GetString("match"),
		"oldScore", e.Record.Original().GetInt("score"),
		"newScore", e.Record.GetInt("score"),
		"timestamp", time.Now(),
	)

	return e.Next()
})
```

### Pattern: Performance Monitoring

```go
scheduler.MustAdd("monitor_db", "*/10 * * * *", func() {
	// Check database size
	stats, _ := app.Dao().GetStats()

	logger.Info("Database stats",
		"tables", stats.Total,
		"size", stats.Size,
	)
})
```

---

## 10. Plugin Architecture

### Pattern: Modular Feature Plugins

Your app already uses this! Example structure:

```go
// internal/plugins/rankings/rankings.go
type RankingsPlugin struct {
	app    core.App
	logger *slog.Logger
}

func (p *RankingsPlugin) Register() {
	// Calculate rankings hourly
	scheduler := p.app.Cron()
	scheduler.MustAdd("rankings", "0 * * * *", p.calculateRankings)

	// Add API routes
	p.app.OnServe().BindFunc(p.registerRoutes)

	// Add hooks for real-time updates
	p.app.OnRecordCreate("match_player_stats").BindFunc(p.onStatsCreate)
}

func (p *RankingsPlugin) calculateRankings() {
	// Your ranking logic here
}
```

---

## 11. Testing Patterns

### Pattern: Test with Transactions

```go
import "sandstorm-tracker/tests"

func TestPlayerScoreUpdate(t *testing.T) {
	// Create test app with test database
	testApp := tests.NewTestApp("./test_pb_data")
	defer testApp.Cleanup()

	err := testApp.RunInTransaction(func(txApp core.App) error {
		// Test in transaction context
		player, _ := txApp.FindRecordById("players", "test_id")
		player.Set("score", 100)
		return txApp.Save(player)
	})

	if err != nil {
		t.Fatalf("Failed: %v", err)
	}
}
```

---

## 12. Real-Time Subscriptions

### Pattern: WebSocket Subscriptions (Built into PocketBase)

```go
// Client-side (JavaScript)
pb.collection('match_player_stats').subscribe('*', function (e) {
	console.log(e.action); // create, update, delete
	console.log(e.record); // The record data
});

// Server-side hook to log subscription activity
app.OnRecordCreate("match_player_stats").BindFunc(func(e *core.RecordCreateEvent) error {
	// This triggers WebSocket broadcasts automatically
	return e.Next()
})
```

---

## Key Takeaways for Your Project

1. **Transactions First**: Wrap related updates in `RunInTransaction()` to prevent partial failures
2. **Hook Priority**: Use Priority to control hook execution order
3. **Event-Driven**: Let hooks drive your business logic instead of cron jobs where possible
4. **Debouncing**: Your ScoreDebouncer pattern prevents query storms - excellent design
5. **Batch Processing**: Always limit queries (10,000+) to avoid memory issues
6. **Contextual Timeouts**: Use context.WithTimeout for long operations
7. **Logging Everything**: Hook into record lifecycle for audit trails
8. **Modular Design**: Keep plugins separate (your architecture is already doing this)

Your current implementation already uses several of these patterns correctly!
