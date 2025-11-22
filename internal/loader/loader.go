package loader

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"sandstorm-tracker/internal/parser"

	"github.com/pocketbase/pocketbase"
)

// LoadLogsFromPath loads log files from the specified path into the database.
func LoadLogsFromPath(app *pocketbase.PocketBase, logPath string, serverID string, baseTime time.Time) error {
	logger := app.Logger().With("COMPONENT", "LOG_LOADER")

	lineBtes, err := os.ReadFile(logPath)
	if err != nil {
		return fmt.Errorf("failed to read log file: %w", err)
	}

	newParser := parser.NewLogParser(app, logger)

	lines := strings.Split(string(lineBtes), "\n")

	// Update timestamps

	lines, err = parser.UpdateLogTimestamps(lines, baseTime)
	if err != nil {
		return fmt.Errorf("failed to update timestamps: %w", err)
	}

	for _, line := range lines {
		if err := newParser.ParseAndProcess(context.Background(), line, serverID, logPath); err != nil {
			continue
		}
	}

	logger.Info("Log file loaded successfully", "path", logPath)
	return nil
}
