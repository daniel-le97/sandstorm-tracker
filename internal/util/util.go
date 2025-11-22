package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func derefInt64(ptr *int64) int64 {
	if ptr == nil {
		return 0
	}
	return *ptr
}

func derefString(ptr *string) string {
	if ptr == nil {
		return "Unknown"
	}
	return *ptr
}

// GetServerIdFromPath determines the server ID from a config path
// Supports both file paths (e.g., /logs/abc-uuid.log) and directory paths (e.g., /logs)
func GetServerIdFromPath(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("path does not exist: %s - %w", path, err)
	}

	// If it's a file path, extract server ID from filename
	if !info.IsDir() {
		filename := filepath.Base(path)
		serverID := strings.TrimSuffix(filename, ".log")
		if serverID == "" {
			return "", fmt.Errorf("invalid log file name: %s", filename)
		}
		return serverID, nil
	}

	// If it's a directory, find the first .log file (excluding backups)
	entries, err := os.ReadDir(path)
	if err != nil {
		return "", fmt.Errorf("failed to read directory %s: %w", path, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".log") && !strings.Contains(name, "backup") {
			return strings.TrimSuffix(name, ".log"), nil
		}
	}

	return "", fmt.Errorf("no log files found in directory: %s", path)
}

// ExtractGameMode extracts the game mode from a scenario string or simple mode string
// Examples:
//
//	"Scenario_Ministry_Checkpoint_Security" -> "Checkpoint"
//	"Scenario_Refinery_Push_Insurgents" -> "Push"
//	"Scenario_Town_Skirmish" -> "Skirmish"
//	"Checkpoint" -> "Checkpoint"
//	"Push" -> "Push"
//	"Skirmish" -> "Skirmish"
func ExtractGameMode(scenario string) string {
	if scenario == "" {
		return "Unknown"
	}

	// Check if it's already a simple mode string
	switch scenario {
	case "Checkpoint", "Push", "Skirmish":
		return scenario
	}

	// Remove "Scenario_" prefix if present
	if strings.HasPrefix(scenario, "Scenario_") {
		scenario = strings.TrimPrefix(scenario, "Scenario_")
	}

	// Split by underscore
	parts := strings.Split(scenario, "_")
	if len(parts) < 2 {
		return "Unknown"
	}

	// For Checkpoint and Push modes, the mode is the second part
	// (first is the map name, second is the mode, third is the team)
	// For Skirmish, there's no team, so it's just the second part
	mode := parts[1]

	// Validate it's a known mode
	switch mode {
	case "Checkpoint", "Push", "Skirmish":
		return mode
	default:
		return "Unknown"
	}
}
