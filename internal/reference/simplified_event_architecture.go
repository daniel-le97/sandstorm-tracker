package reference

import (
	"encoding/json"
	"fmt"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// ============================================================================
// SIMPLIFIED EVENT-DRIVEN ARCHITECTURE USING POCKETBASE HOOKS
// ============================================================================

// This example shows how to replace a custom event bus with PocketBase hooks
//
// KEY ARCHITECTURE DECISIONS:
// 1. Single "events" collection with type field and JSON data field
// 2. Each event type gets its own dedicated hook handler
// 3. serverID is stored in the JSON data field (not as separate column)
// 4. Handlers extract data from JSON and process independently
//
// BENEFITS:
// - Clean separation: Parser creates records, handlers process them
// - Easy to add new event handlers without touching parser
// - Built-in persistence and event history
// - Real-time subscriptions work automatically
// - Each handler is focused and testable

// ============================================================================
// 1. EVENT TYPES (Just constants, no custom structs needed)
// ============================================================================

const (
	EventPlayerKill        = "player_kill"
	EventPlayerJoin        = "player_join"
	EventPlayerLeave       = "player_leave"
	EventMatchStart        = "match_start"
	EventMatchEnd          = "match_end"
	EventRoundStart        = "round_start"
	EventRoundEnd          = "round_end"
	EventObjectiveCaptured = "objective_captured"
)

// ============================================================================
// 2. SIMPLIFIED PARSER (Just creates records)
// ============================================================================

type SimplifiedParser struct {
	pbApp core.App
}

func NewSimplifiedParser(app core.App) *SimplifiedParser {
	return &SimplifiedParser{pbApp: app}
}

// CreateEvent is the only method needed - just create a record!
func (p *SimplifiedParser) CreateEvent(eventType, serverID string, data map[string]interface{}) error {
	collection, err := p.pbApp.FindCollectionByNameOrId("events")
	if err != nil {
		return fmt.Errorf("collection not found: %w", err)
	}

	record := core.NewRecord(collection)
	record.Set("type", eventType)

	// Add server_id to the data JSON (handlers extract it from there)
	data["server_id"] = serverID

	// Store event-specific data as JSON
	dataJSON, _ := json.Marshal(data)
	record.Set("data", string(dataJSON))

	// This Save() triggers OnRecordCreate hooks automatically!
	if err := p.pbApp.Save(record); err != nil {
		return fmt.Errorf("failed to save event: %w", err)
	}

	return nil
}

// Example: Parser processing a kill event
func (p *SimplifiedParser) ProcessKillEvent(serverID, killerName, victimName, weapon string) error {
	return p.CreateEvent(EventPlayerKill, serverID, map[string]interface{}{
		"killer_name": killerName,
		"victim_name": victimName,
		"weapon":      weapon,
		// server_id is added automatically by CreateEvent
	})
}

// Example: Parser processing a player join
func (p *SimplifiedParser) ProcessPlayerJoin(serverID, playerName, steamID string) error {
	return p.CreateEvent(EventPlayerJoin, serverID, map[string]interface{}{
		"player_name": playerName,
		"steam_id":    steamID,
		// server_id is added automatically by CreateEvent
	})
}

// ============================================================================
// 3. APP WITH EVENT HANDLERS (Hooks are your event subscribers)
// ============================================================================

type SimplifiedApp struct {
	*pocketbase.PocketBase
	Parser         *SimplifiedParser
	ScoreDebouncer ScoreDebouncer
}

// ScoreDebouncer interface (from your actual code)
type ScoreDebouncer interface {
	TriggerScoreUpdate(serverID string)
	ExecuteImmediately(serverID string)
}

func NewSimplifiedApp() *SimplifiedApp {
	pb := pocketbase.New()

	app := &SimplifiedApp{
		PocketBase: pb,
	}

	app.Parser = NewSimplifiedParser(pb)

	return app
}

// Bootstrap sets up all event handlers via PocketBase hooks
// Each event type gets its own dedicated handler
func (app *SimplifiedApp) Bootstrap() error {
	// Player Kill Events
	app.OnRecordCreate("events").BindFunc(app.handlePlayerKill)

	// Player Join Events
	app.OnRecordCreate("events").BindFunc(app.handlePlayerJoin)

	// Player Leave Events
	app.OnRecordCreate("events").BindFunc(app.handlePlayerLeave)

	// Round End Events
	app.OnRecordCreate("events").BindFunc(app.handleRoundEnd)

	// Match Start Events
	app.OnRecordCreate("events").BindFunc(app.handleMatchStart)

	// Match End Events
	app.OnRecordCreate("events").BindFunc(app.handleMatchEnd)

	// You can also have hooks for other collections
	app.OnRecordUpdate("matches").BindFunc(app.handleMatchUpdate)

	return nil
}

// ============================================================================
// 4. EVENT HANDLERS (Clean, focused functions)
// ============================================================================

// Each handler extracts data and serverID from the event record's JSON data field

func (app *SimplifiedApp) handlePlayerKill(e *core.RecordEvent) error {
	if e.Record.GetString("type") != EventPlayerKill {
		return e.Next()
	}
	// Parse event data from JSON field
	var data map[string]interface{}
	json.Unmarshal([]byte(e.Record.GetString("data")), &data)

	// Extract serverID from data (it's stored in the JSON)
	serverID, _ := data["server_id"].(string)

	app.Logger().Info("Player kill event",
		"killer", data["killer_name"],
		"victim", data["victim_name"],
		"weapon", data["weapon"],
		"server", serverID,
	)

	// Trigger score update (debounced)
	if app.ScoreDebouncer != nil {
		app.ScoreDebouncer.TriggerScoreUpdate(serverID)
	}

	// Could add more handlers:
	// - Update kill feed
	// - Send Discord notification
	// - Update player stats
	// - Trigger ML model update

	return e.Next()
}

func (app *SimplifiedApp) handlePlayerJoin(e *core.RecordEvent) error {
	if e.Record.GetString("type") != EventPlayerJoin {
		return e.Next()
	}
	var data map[string]interface{}
	json.Unmarshal([]byte(e.Record.GetString("data")), &data)

	serverID, _ := data["server_id"].(string)

	app.Logger().Info("Player joined",
		"player", data["player_name"],
		"steam_id", data["steam_id"],
		"server", serverID,
	)

	// Update active player count
	// Send welcome message
	// Log to analytics

	return e.Next()
}

func (app *SimplifiedApp) handlePlayerLeave(e *core.RecordEvent) error {
	if e.Record.GetString("type") != EventPlayerLeave {
		return e.Next()
	}
	var data map[string]interface{}
	json.Unmarshal([]byte(e.Record.GetString("data")), &data)

	serverID, _ := data["server_id"].(string)

	app.Logger().Info("Player left", "server", serverID)

	// Update active player count
	// Clean up temporary data

	return e.Next()
}

func (app *SimplifiedApp) handleRoundEnd(e *core.RecordEvent) error {
	if e.Record.GetString("type") != EventRoundEnd {
		return e.Next()
	}
	var data map[string]interface{}
	json.Unmarshal([]byte(e.Record.GetString("data")), &data)

	serverID, _ := data["server_id"].(string)

	app.Logger().Info("Round ended", "server", serverID)

	// Immediate score calculation on round end
	if app.ScoreDebouncer != nil {
		app.ScoreDebouncer.ExecuteImmediately(serverID)
	}

	// Update round statistics
	// Save round snapshot

	return e.Next()
}

func (app *SimplifiedApp) handleMatchStart(e *core.RecordEvent) error {
	if e.Record.GetString("type") != EventMatchStart {
		return e.Next()
	}
	var data map[string]interface{}
	json.Unmarshal([]byte(e.Record.GetString("data")), &data)

	serverID, _ := data["server_id"].(string)

	app.Logger().Info("Match started",
		"server", serverID,
		"map", data["map"],
		"scenario", data["scenario"],
	)

	// Initialize match tracking
	// Reset player stats
	// Start match timer

	return e.Next()
}

func (app *SimplifiedApp) handleMatchEnd(e *core.RecordEvent) error {
	if e.Record.GetString("type") != EventMatchEnd {
		return e.Next()
	}
	var data map[string]interface{}
	json.Unmarshal([]byte(e.Record.GetString("data")), &data)

	serverID, _ := data["server_id"].(string)

	app.Logger().Info("Match ended", "server", serverID)

	// Finalize scores
	// Save match results
	// Send match summary

	return e.Next()
}

func (app *SimplifiedApp) handleMatchUpdate(e *core.RecordEvent) error {
	
	matchID := e.Record.Id

	app.Logger().Info("Match updated", "match_id", matchID)

	// Handle match state changes
	// Update real-time dashboard
	// Notify subscribers

	return e.Next()
}

// ============================================================================
// 5. QUERYING EVENTS (Event history for free!)
// ============================================================================

// GetServerEvents retrieves all events for a specific server
// Note: serverID is extracted from the JSON data field
func (app *SimplifiedApp) GetServerEvents(serverID string) ([]*core.Record, error) {
	records, err := app.FindRecordsByFilter(
		"events",
		"server = {:serverID}", // Using the relation field from the schema
		"-created",             // Order by created timestamp descending
		100,                    // Limit
		0,                      // Offset
		dbx.Params{"serverID": serverID},
	)
	if err != nil {
		return nil, err
	}

	return records, nil
}

// GetKillEvents retrieves all kill events for a server
func (app *SimplifiedApp) GetKillEvents(serverID string) ([]*core.Record, error) {
	records, err := app.FindRecordsByFilter(
		"events",
		"server = {:serverID} && type = {:type}",
		"-created",
		1000,
		0,
		dbx.Params{
			"serverID": serverID,
			"type":     EventPlayerKill,
		},
	)
	if err != nil {
		return nil, err
	}

	return records, nil
}

// GetRecentEvents retrieves recent events across all servers
func (app *SimplifiedApp) GetRecentEvents(limit int) ([]*core.Record, error) {
	records, err := app.FindRecordsByFilter(
		"events",
		"",
		"-created",
		limit,
		0,
	)
	if err != nil {
		return nil, err
	}

	return records, nil
}

// GetEventsByType retrieves events of a specific type
func (app *SimplifiedApp) GetEventsByType(eventType string, limit int) ([]*core.Record, error) {
	records, err := app.FindRecordsByFilter(
		"events",
		"type = {:type}",
		"-created",
		limit,
		0,
		dbx.Params{"type": eventType},
	)
	if err != nil {
		return nil, err
	}

	return records, nil
}

// ============================================================================
// 6. REAL-TIME SUBSCRIPTIONS (Frontend integration)
// ============================================================================

/*
Frontend JavaScript can subscribe to events directly:

```javascript
// Subscribe to all events
pb.collection('events').subscribe('*', (e) => {
    if (e.record.type === 'player_kill') {
        const data = JSON.parse(e.record.data);
        updateKillFeed(data);
    }

    if (e.record.type === 'round_end') {
        showRoundEndScreen(e.record);
    }
}, { filter: 'type = "player_kill" || type = "round_end"' });

// Subscribe to specific server events
pb.collection('events').subscribe('*', (e) => {
    const data = JSON.parse(e.record.data);
    updateServerDashboard(e.record, data);
}, { filter: 'server = "server-123"' });
```
*/

// ============================================================================
// 7. COMPARISON: OLD VS NEW
// ============================================================================

/*
OLD ARCHITECTURE (Complex):
┌─────────────────┐
│   Parser        │ (Many responsibilities)
│ ├─ Parse        │
│ ├─ Store        │
│ ├─ Publish      │
│ └─ Coordinate   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Custom Event   │
│     Bus         │
└────────┬────────┘
         │
         ▼
┌─────────────────────────────────┐
│  Manual Subscribers             │
│  (Must register/unregister)     │
└─────────────────────────────────┘

NEW ARCHITECTURE (Simple):
┌─────────────────┐
│   Parser        │ (One job: create records)
│ └─ CreateEvent()│
└────────┬────────┘
         │
         ▼ (PocketBase.Save)
┌─────────────────────────────────┐
│  PocketBase Hooks               │
│  (Automatic event routing)      │
│  ├─ OnRecordCreate              │
│  ├─ OnRecordUpdate              │
│  └─ OnRecordDelete              │
└────────┬────────────────────────┘
         │
         ▼
┌─────────────────────────────────┐
│  Event Handlers                 │
│  (Automatically triggered)      │
└─────────────────────────────────┘
*/

// ============================================================================
// 8. BENEFITS
// ============================================================================

/*
✅ SIMPLER CODE
   - Parser has one job: create records
   - No custom event bus implementation
   - No manual subscriber management

✅ BUILT-IN PERSISTENCE
   - All events stored automatically
   - Event history queryable anytime
   - No separate event store needed

✅ REAL-TIME OUT OF THE BOX
   - Frontend subscribes to collection
   - PocketBase handles WebSockets
   - No custom real-time logic needed

✅ EASIER TESTING
   - Test hooks independently
   - Query test events from database
   - No mocking event bus

✅ EASIER DEBUGGING
   - View events in PocketBase admin
   - Query event timeline
   - Trace event flow through logs

✅ SCALABLE
   - Add new event types: just add a handler
   - Add new subscribers: just add a hook
   - No architectural changes needed
*/
