package jobs

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"

	"sandstorm-tracker/internal/config"
	"sandstorm-tracker/internal/ghupdate"

	"github.com/pocketbase/pocketbase/core"
	// "sandstorm-tracker/internal/parser"
)

// RegisterUpdateChecker sets up a cron job that runs every 30 minutes to check for updates.
// It uses RCON to check if any players are on the servers before checking for updates.
// If an update is available, it exits gracefully.
func RegisterUpdateChecker(app AppInterface, cfg *config.Config, logger *slog.Logger) {
	scheduler := app.Cron()
	updateLogger := logger.With("component", "UPDATE_CHECKER")

	// Run every 30 minutes: 0, 30 minutes past every hour
	scheduler.MustAdd("check_for_updates", "*/30 * * * *", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		checkForUpdatesAndRestart(ctx, app, cfg, updateLogger)
	})

	updateLogger.Info("Registered cron job to check for updates every 30 minutes")
}

// checkForUpdatesAndRestart checks if all servers are idle, then checks for updates and restarts if available
func checkForUpdatesAndRestart(ctx context.Context, app AppInterface, cfg *config.Config, logger *slog.Logger) {
	logger.Debug("Checking for updates...")

	// Get the ghupdate plugin
	updater, err := ghupdate.GetPlugin(app)
	if err != nil {
		logger.Error("Failed to get update plugin", slog.String("error", err.Error()))
		return
	}

	// Check if a new version is available FIRST (before querying servers)
	newVersion, err := updater.CheckForUpdate(ctx)
	if err != nil {
		logger.Error("Failed to check for updates", slog.String("error", err.Error()))
		return
	}

	// If newVersion is empty, we're already on the latest version - no need to check if servers are idle
	if newVersion == "" {
		logger.Debug("Already on the latest version")
		return
	}

	logger.Info("New version available", slog.String("version", newVersion))

	// Only now check if all servers are idle (no need to check if no update available)
	isIdle, err := isAllServersIdle(ctx, app, cfg, logger)
	if err != nil {
		logger.Error("Failed to check server idle status", slog.String("error", err.Error()))
		return
	}

	if !isIdle {
		logger.Info("Servers currently have players, skipping update")
		return
	}

	logger.Info("All servers are idle, proceeding with update")

	// Gracefully shutdown the app
	// This triggers OnTerminate hooks for proper cleanup
	// The PowerShell wrapper will restart and run the update command
	logger.Info("Initiating graceful shutdown for update")
	go func() {
		event := new(core.TerminateEvent)
		event.App = app
		app.OnTerminate().Trigger(event, func(e *core.TerminateEvent) error {
			return e.App.ResetBootstrapState()
		})
		time.Sleep(500 * time.Millisecond)
		os.Exit(0)
	}()
}

// isAllServersIdle checks if all servers have no players using RCON list_players command
func isAllServersIdle(ctx context.Context, app AppInterface, cfg *config.Config, logger *slog.Logger) (bool, error) {
	// Check each enabled server for players
	for _, serverCfg := range cfg.Servers {
		if !serverCfg.Enabled {
			continue
		}

		// Query players via RCON
		// The list_players command returns a list of players on the server
		// If there are players, the response will contain player information
		response, err := app.SendRconCommand(serverCfg.Name, "list_players")
		if err != nil {
			logger.Warn("Failed to query server for players",
				slog.String("server", serverCfg.Name),
				slog.String("error", err.Error()))
			// If we can't query a server, assume it's active (don't update)
			return false, nil
		}

		players := parseRconListPlayers(response)
		if len(players) > 0 {
			logger.Info("Found players on server", slog.Any("players", players))
			return false, nil
		}

		// Check if the response indicates players are on the server
		// Typically, list_players returns player names or "No players online" message
		// If response is not empty and doesn't contain "no players" or similar, assume players present
		if response != "" && !strings.Contains(strings.ToLower(response), "no players") {
			logger.Info("Found players on server, servers not idle",
				slog.String("server", serverCfg.Name))
			return false, nil
		}
	}

	return true, nil
}
