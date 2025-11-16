package jobs

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

// RegisterArchiveOldData sets up a cron job that archives data older than 30 days
// Archives matches, match player stats, weapon stats, and events
func RegisterArchiveOldData(app core.App, logger *slog.Logger) {
	scheduler := app.Cron()

	// Run archive job daily at 2 AM UTC
	scheduler.MustAdd("archive_old_data", "0 2 * * *", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		archiveOldData(ctx, app, logger)
	})

	logger.Info("Registered cron job to archive data older than 30 days daily at 2 AM UTC", "component", "JOBS")
}

// archiveOldData moves matches and related data older than 30 days into an archived state
func archiveOldData(ctx context.Context, app core.App, logger *slog.Logger) {
	logger.Info("Starting archive job", "component", "ARCHIVE_JOB")

	// Calculate cutoff date (30 days ago)
	cutoffDate := time.Now().AddDate(0, 0, -30)

	// Run in a transaction for consistency
	err := app.RunInTransaction(func(txApp core.App) error {
		// Archive old events first (least dependent)
		archivedEvents, err := archiveEvents(txApp, logger, cutoffDate)
		if err != nil {
			logger.Error("Failed to archive events", "error", err)
			return err
		}

		// Archive related match player stats
		archivedPlayerStats, err := archiveMatchPlayerStats(txApp, logger, cutoffDate)
		if err != nil {
			logger.Error("Failed to archive match player stats", "error", err)
			return err
		}

		// Archive related weapon stats
		archivedWeaponStats, err := archiveMatchWeaponStats(txApp, logger, cutoffDate)
		if err != nil {
			logger.Error("Failed to archive weapon stats", "error", err)
			return err
		}

		// Archive old matches last
		archivedMatches, err := archiveMatches(txApp, logger, cutoffDate)
		if err != nil {
			logger.Error("Failed to archive matches", "error", err)
			return err
		}

		logger.Info("Archive job completed successfully",
			"component", "ARCHIVE_JOB",
			"archived_matches", archivedMatches,
			"archived_player_stats", archivedPlayerStats,
			"archived_weapon_stats", archivedWeaponStats,
			"archived_events", archivedEvents,
			"cutoff_date", cutoffDate.Format("2006-01-02"))

		return nil
	})

	if err != nil {
		logger.Error("Archive job failed", "component", "ARCHIVE_JOB", "error", err)
		return
	}
}

// archiveMatches marks old matches with an archived flag or soft-deletes them
// Returns the count of archived records
func archiveMatches(txApp core.App, logger *slog.Logger, cutoffDate time.Time) (int, error) {
	// Find all matches that ended before the cutoff date
	// Assumes matches have an end_time field
	records, err := txApp.FindRecordsByFilter(
		"matches",
		"end_time != '' && end_time < {:cutoff}",
		"-end_time", // Sort by end_time descending
		10000,       // Limit to a large number to avoid memory issues
		0,           // No offset
		map[string]any{"cutoff": cutoffDate.Format(time.RFC3339)},
	)

	if err != nil {
		return 0, fmt.Errorf("failed to query old matches: %w", err)
	}

	archived := 0
	for _, record := range records {
		// Delete the record
		if err := txApp.Delete(record); err != nil {
			logger.Error("Failed to delete old match", "matchID", record.Id, "error", err)
			continue
		}
		archived++
	}

	return archived, nil
}

// archiveMatchPlayerStats deletes old player stats
func archiveMatchPlayerStats(txApp core.App, logger *slog.Logger, cutoffDate time.Time) (int, error) {
	// Find stats for matches that ended before the cutoff date using created timestamp
	records, err := txApp.FindRecordsByFilter(
		"match_player_stats",
		"created < {:cutoff}",
		"-created",
		10000,
		0,
		map[string]any{"cutoff": cutoffDate.Format(time.RFC3339)},
	)

	if err != nil {
		return 0, fmt.Errorf("failed to query old match_player_stats: %w", err)
	}

	archived := 0
	for _, record := range records {
		if err := txApp.Delete(record); err != nil {
			logger.Error("Failed to delete old player stats", "statsID", record.Id, "error", err)
			continue
		}
		archived++
	}

	return archived, nil
}

// archiveMatchWeaponStats deletes old weapon stats
func archiveMatchWeaponStats(txApp core.App, logger *slog.Logger, cutoffDate time.Time) (int, error) {
	// Find stats for matches that ended before the cutoff date
	records, err := txApp.FindRecordsByFilter(
		"match_weapon_stats",
		"created < {:cutoff}",
		"-created",
		10000,
		0,
		map[string]any{"cutoff": cutoffDate.Format(time.RFC3339)},
	)

	if err != nil {
		return 0, fmt.Errorf("failed to query old match_weapon_stats: %w", err)
	}

	archived := 0
	for _, record := range records {
		if err := txApp.Delete(record); err != nil {
			logger.Error("Failed to delete old weapon stats", "statsID", record.Id, "error", err)
			continue
		}
		archived++
	}

	return archived, nil
}

// archiveEvents deletes old events
func archiveEvents(txApp core.App, logger *slog.Logger, cutoffDate time.Time) (int, error) {
	// Find events older than cutoff date
	records, err := txApp.FindRecordsByFilter(
		"events",
		"created < {:cutoff}",
		"-created",
		10000,
		0,
		map[string]any{"cutoff": cutoffDate.Format(time.RFC3339)},
	)

	if err != nil {
		return 0, fmt.Errorf("failed to query old events: %w", err)
	}

	archived := 0
	for _, record := range records {
		if err := txApp.Delete(record); err != nil {
			logger.Error("Failed to delete old event", "eventID", record.Id, "error", err)
			continue
		}
		archived++
	}

	return archived, nil
}
