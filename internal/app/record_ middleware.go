package app

import (
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func BindRecordMiddlewares(app *pocketbase.PocketBase) {
	// Register hooks to restrict player metadata field access to superusers only
	app.OnRecordsListRequest("players").BindFunc(func(e *core.RecordsListRequestEvent) error {
		// If not authenticated as superuser admin, hide metadata from all records
		if !e.HasSuperuserAuth() {
			for _, record := range e.Records {
				record.Hide("metadata")
			}
		}
		return e.Next()
	})

	app.OnRecordViewRequest("players").BindFunc(func(e *core.RecordRequestEvent) error {
		// If not authenticated as superuser admin, hide metadata from record
		if !e.HasSuperuserAuth() {
			e.Record.Hide("metadata")
		}
		return e.Next()
	})

	app.OnRecordAfterUpdateSuccess("matches").BindFunc(func(e *core.RecordEvent) error {
		status := e.Record.GetString("status")
		
		// Only process status changes that affect player stats
		if status != "crashed" && status != "finished" {
			return e.Next()
		}

		matchID := e.Record.Id
		
		// Fetch all match_player_stats for this match in a single query
		matchPlayerStats, err := e.App.FindRecordsByFilter(
			"match_player_stats", 
			"match = {:matchId}", 
			"", // No sorting needed since we're updating all
			-1, 0, 
			map[string]any{"matchId": matchID},
		)
		if err != nil {
			e.App.Logger().Error(
				"Failed to find match player stats",
				"component", "MATCH_HOOKS",
				"match_id", matchID,
				"status", status,
				"error", err,
			)
			return e.Next()
		}

		// Early exit if no records to update
		if len(matchPlayerStats) == 0 {
			return e.Next()
		}

		// Batch update all records and collect errors
		failureCount := 0
		for _, mps := range matchPlayerStats {
			mps.Set("status", "finished")
			if err := e.App.Save(mps); err != nil {
				failureCount++
			}
		}

		// Log aggregate results instead of per-record
		if failureCount > 0 {
			e.App.Logger().Error(
				"Failed to update match player stats",
				"component", "MATCH_HOOKS",
				"match_id", matchID,
				"total_records", len(matchPlayerStats),
				"failed_count", failureCount,
				"status", status,
			)
		}

		return e.Next()
})

}
