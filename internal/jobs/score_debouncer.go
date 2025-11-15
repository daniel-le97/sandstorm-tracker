package jobs

import (
	"log/slog"
	"sync"
	"time"

	"sandstorm-tracker/internal/config"
	"sandstorm-tracker/internal/util"
)

// ScoreDebouncer handles debounced score updates per server
// When game events occur (kills, objectives, etc), we trigger a score update
// The update is debounced so multiple events within the debounce window
// result in only one RCON query and database update
// If events keep happening rapidly, scores will still update after maxWait duration
type ScoreDebouncer struct {
	app            AppInterface
	cfg            *config.Config
	logger         *slog.Logger
	debounceWindow time.Duration
	maxWait        time.Duration // Maximum time to wait before forcing an update

	mu             sync.Mutex
	timers         map[string]*time.Timer // serverID -> debounce timer
	firstTriggerAt map[string]time.Time   // serverID -> time of first trigger in current window
}

// NewScoreDebouncer creates a new score debouncer
// maxWait should typically be 2-3x the debounceWindow to ensure updates during continuous activity
func NewScoreDebouncer(app AppInterface, cfg *config.Config, debounceWindow time.Duration, maxWait time.Duration) *ScoreDebouncer {
	return &ScoreDebouncer{
		app:            app,
		cfg:            cfg,
		logger:         app.Logger().With("component", "SCORE_DEBOUNCER"),
		debounceWindow: debounceWindow,
		maxWait:        maxWait,
		timers:         make(map[string]*time.Timer),
		firstTriggerAt: make(map[string]time.Time),
	}
}

// TriggerScoreUpdate signals that a score-affecting event occurred for a server
// The actual score update will be debounced and executed after the debounce window
// However, if events keep happening, update will be forced after maxWait duration
func (d *ScoreDebouncer) TriggerScoreUpdate(serverID string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()

	// Check if this is the first trigger in a new window
	firstTrigger, exists := d.firstTriggerAt[serverID]
	if !exists {
		// First trigger - record the time
		d.firstTriggerAt[serverID] = now
		firstTrigger = now
	}

	// Calculate time since first trigger
	timeSinceFirst := now.Sub(firstTrigger)

	// If we've exceeded maxWait, force update immediately
	if timeSinceFirst >= d.maxWait {
		d.logger.Debug("Max wait exceeded, forcing score update",
			"serverID", serverID, "timeSinceFirst", timeSinceFirst)

		// Stop existing timer if any
		if timer, exists := d.timers[serverID]; exists {
			timer.Stop()
		}

		// Execute immediately in goroutine
		go func() {
			d.executeScoreUpdate(serverID)

			// Clean up
			d.mu.Lock()
			delete(d.timers, serverID)
			delete(d.firstTriggerAt, serverID)
			d.mu.Unlock()
		}()

		return
	}

	// Normal debounce behavior - stop existing timer if any
	if timer, exists := d.timers[serverID]; exists {
		timer.Stop()
	}

	// Create a new timer that will execute the score update after the debounce window
	d.timers[serverID] = time.AfterFunc(d.debounceWindow, func() {
		d.executeScoreUpdate(serverID)

		// Clean up the timer and first trigger time
		d.mu.Lock()
		delete(d.timers, serverID)
		delete(d.firstTriggerAt, serverID)
		d.mu.Unlock()
	})

	d.logger.Debug("Score update triggered",
		"serverID", serverID, "debounce", d.debounceWindow, "timeSinceFirst", timeSinceFirst)
}

// executeScoreUpdate performs the actual RCON query and database update
func (d *ScoreDebouncer) executeScoreUpdate(serverID string) {
	// Find the server config for this serverID
	var serverCfg *config.ServerConfig
	for i, sc := range d.cfg.Servers {
		if !sc.Enabled {
			continue
		}

		pathServerID, err := util.GetServerIdFromPath(sc.LogPath)
		if err != nil {
			continue
		}

		if pathServerID == serverID {
			serverCfg = &d.cfg.Servers[i]
			break
		}
	}

	if serverCfg == nil {
		d.logger.Warn("Could not find server config for score update", "serverID", serverID)
		return
	}

	d.logger.Info("Executing score update", "server", serverCfg.Name, "serverID", serverID, "component", "SCORE_DEBOUNCER")

	// Find or create server record
	serverRecord, err := getOrCreateServerFromConfig(d.app, d.logger, *serverCfg)
	if err != nil {
		d.logger.Error("Failed to get server record", "component", "SCORE_DEBOUNCER", "server", serverCfg.Name, "error", err)
		return
	}
	d.logger.Debug("Got server record", "component", "SCORE_DEBOUNCER", "server", serverCfg.Name, "recordID", serverRecord.Id)

	// Find active match for this server
	activeMatch, err := getActiveMatchForServer(d.app, serverRecord.Id)
	if err != nil {
		d.logger.Error("Failed to get active match", "component", "SCORE_DEBOUNCER", "server", serverCfg.Name, "error", err)
		return
	}

	if activeMatch == nil {
		d.logger.Info("No active match found for server, skipping score update", "component", "SCORE_DEBOUNCER", "server", serverCfg.Name)
		return
	}
	d.logger.Debug("Found active match", "component", "SCORE_DEBOUNCER", "server", serverCfg.Name, "matchID", activeMatch.Id)

	// Query players via RCON
	d.logger.Debug("Querying players via RCON", "component", "SCORE_DEBOUNCER", "server", serverCfg.Name, "serverID", serverID)
	players, err := queryPlayersViaRcon(d.app, serverID)
	if err != nil {
		d.logger.Error("Failed to query players via RCON", "component", "SCORE_DEBOUNCER", "server", serverCfg.Name, "serverID", serverID, "error", err)
		return
	}

	// Update player scores in database
	if len(players) > 0 {
		d.logger.Info("Updating player scores", "component", "SCORE_DEBOUNCER", "count", len(players), "server", serverCfg.Name, "matchID", activeMatch.Id)
		updatePlayersFromRcon(d.app, d.logger, activeMatch.Id, players)
		d.logger.Info("Finished updating player scores", "component", "SCORE_DEBOUNCER", "server", serverCfg.Name)
	} else {
		d.logger.Info("No players found via RCON", "component", "SCORE_DEBOUNCER", "server", serverCfg.Name)
	}
}

// Stop cancels all pending score updates
func (d *ScoreDebouncer) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()

	for serverID, timer := range d.timers {
		timer.Stop()
		d.logger.Debug("Stopped pending score update", "serverID", serverID)
	}

	d.timers = make(map[string]*time.Timer)
}
