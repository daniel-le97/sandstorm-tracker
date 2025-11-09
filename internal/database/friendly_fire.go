package database

import (
	"context"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

// FriendlyFireIncident represents a friendly fire kill incident
type FriendlyFireIncident struct {
	ID                      string
	MatchID                 string
	KillerID                string
	VictimID                string
	Weapon                  string
	Timestamp               time.Time
	KillerTeam              *int
	VictimTeam              *int
	TimeSinceMatchStartSecs *float64
	TimeSinceLastFFSecs     *float64
	KillerTotalKillsInMatch *int
	KillerFFCountInMatch    *int
	IsExplosiveWeapon       bool
	IsVehicleWeapon         bool
	Map                     *string
	GameMode                *string
	AccidentClassification  *string  // likely_accident, possibly_intentional, likely_intentional, unclassified
	ConfidenceScore         *float64 // 0.0 - 1.0
}

// RecordFriendlyFireIncident records a new friendly fire incident
func RecordFriendlyFireIncident(ctx context.Context, pbApp core.App, incident *FriendlyFireIncident) error {
	collection, err := pbApp.FindCollectionByNameOrId("friendly_fire_incidents")
	if err != nil {
		return err
	}

	record := core.NewRecord(collection)
	record.Set("match", incident.MatchID)
	record.Set("killer", incident.KillerID)
	record.Set("victim", incident.VictimID)
	record.Set("weapon", incident.Weapon)
	record.Set("timestamp", incident.Timestamp.Format(time.RFC3339))

	if incident.KillerTeam != nil {
		record.Set("killer_team", *incident.KillerTeam)
	}
	if incident.VictimTeam != nil {
		record.Set("victim_team", *incident.VictimTeam)
	}
	if incident.TimeSinceMatchStartSecs != nil {
		record.Set("time_since_match_start_seconds", *incident.TimeSinceMatchStartSecs)
	}
	if incident.TimeSinceLastFFSecs != nil {
		record.Set("time_since_last_ff_seconds", *incident.TimeSinceLastFFSecs)
	}
	if incident.KillerTotalKillsInMatch != nil {
		record.Set("killer_total_kills_in_match", *incident.KillerTotalKillsInMatch)
	}
	if incident.KillerFFCountInMatch != nil {
		record.Set("killer_ff_count_in_match", *incident.KillerFFCountInMatch)
	}
	if incident.Map != nil {
		record.Set("map", *incident.Map)
	}
	if incident.GameMode != nil {
		record.Set("game_mode", *incident.GameMode)
	}
	if incident.AccidentClassification != nil {
		record.Set("accident_classification", *incident.AccidentClassification)
	}
	if incident.ConfidenceScore != nil {
		record.Set("confidence_score", *incident.ConfidenceScore)
	}

	record.Set("is_explosive_weapon", incident.IsExplosiveWeapon)
	record.Set("is_vehicle_weapon", incident.IsVehicleWeapon)

	return pbApp.Save(record)
}

// GetFriendlyFireStats returns aggregate friendly fire statistics for a player
func GetFriendlyFireStats(ctx context.Context, pbApp core.App, playerID string) (map[string]any, error) {
	type statsRow struct {
		TotalFFKills       int     `db:"total_ff_kills"`
		TotalKills         int     `db:"total_kills"`
		FFPercentage       float64 `db:"ff_percentage"`
		AvgTimeBetweenFF   float64 `db:"avg_time_between_ff"`
		ExplosiveFFCount   int     `db:"explosive_ff_count"`
		VehicleFFCount     int     `db:"vehicle_ff_count"`
		MostCommonFFWeapon string  `db:"most_common_weapon"`
	}

	var row statsRow

	err := pbApp.DB().
		NewQuery(`
			WITH player_ff AS (
				SELECT 
					COUNT(*) as total_ff_kills,
					AVG(time_since_last_ff_seconds) as avg_time_between_ff,
					SUM(CASE WHEN is_explosive_weapon = 1 THEN 1 ELSE 0 END) as explosive_ff_count,
					SUM(CASE WHEN is_vehicle_weapon = 1 THEN 1 ELSE 0 END) as vehicle_ff_count
				FROM friendly_fire_incidents
				WHERE killer = {:player}
			),
			player_total_kills AS (
				SELECT COALESCE(SUM(kills), 0) as total_kills
				FROM match_player_stats
				WHERE player = {:player}
			),
			most_common_weapon AS (
				SELECT weapon
				FROM friendly_fire_incidents
				WHERE killer = {:player}
				GROUP BY weapon
				ORDER BY COUNT(*) DESC
				LIMIT 1
			)
			SELECT 
				pf.total_ff_kills,
				pk.total_kills,
				CASE 
					WHEN pk.total_kills > 0 
					THEN (CAST(pf.total_ff_kills AS REAL) / CAST(pk.total_kills AS REAL)) * 100
					ELSE 0
				END as ff_percentage,
				COALESCE(pf.avg_time_between_ff, 0) as avg_time_between_ff,
				pf.explosive_ff_count,
				pf.vehicle_ff_count,
				COALESCE(mcw.weapon, 'None') as most_common_weapon
			FROM player_ff pf, player_total_kills pk
			LEFT JOIN most_common_weapon mcw
		`).
		Bind(map[string]any{"player": playerID}).
		One(&row)

	if err != nil {
		return nil, err
	}

	return map[string]any{
		"total_ff_kills":        row.TotalFFKills,
		"total_kills":           row.TotalKills,
		"ff_percentage":         row.FFPercentage,
		"avg_time_between_ff":   row.AvgTimeBetweenFF,
		"explosive_ff_count":    row.ExplosiveFFCount,
		"vehicle_ff_count":      row.VehicleFFCount,
		"most_common_ff_weapon": row.MostCommonFFWeapon,
	}, nil
}

// GetFriendlyFirePattern analyzes patterns to determine if FF is likely accidental
func GetFriendlyFirePattern(ctx context.Context, pbApp core.App, playerID string) (string, float64, error) {
	type patternRow struct {
		TotalFF            int     `db:"total_ff"`
		TotalKills         int     `db:"total_kills"`
		ExplosiveFFPercent float64 `db:"explosive_ff_percent"`
		AvgTimeBetweenFF   float64 `db:"avg_time_between_ff"`
		FFInFirstMinute    int     `db:"ff_in_first_minute"`
	}

	var row patternRow

	err := pbApp.DB().
		NewQuery(`
			SELECT 
				COUNT(*) as total_ff,
				(SELECT COALESCE(SUM(kills), 0) FROM match_player_stats WHERE player = {:player}) as total_kills,
				(SUM(CASE WHEN is_explosive_weapon = 1 THEN 1 ELSE 0 END) * 100.0 / COUNT(*)) as explosive_ff_percent,
				AVG(COALESCE(time_since_last_ff_seconds, 0)) as avg_time_between_ff,
				SUM(CASE WHEN time_since_match_start_seconds < 60 THEN 1 ELSE 0 END) as ff_in_first_minute
			FROM friendly_fire_incidents
			WHERE killer = {:player}
		`).
		Bind(map[string]any{"player": playerID}).
		One(&row)

	if err != nil {
		return "unclassified", 0.0, err
	}

	// Simple heuristic-based classification
	// This could be replaced with ML model in the future
	confidence := 0.0
	classification := "unclassified"

	if row.TotalFF == 0 {
		return "unclassified", 0.0, nil
	}

	ffRate := float64(row.TotalFF) / float64(row.TotalKills)

	// Likely accident indicators:
	// - Low FF rate (< 5%)
	// - High explosive weapon usage (> 70%)
	// - Reasonable time between incidents (> 120 seconds)
	// - Few kills in first minute (spawn area)
	if ffRate < 0.05 && row.ExplosiveFFPercent > 70 && row.AvgTimeBetweenFF > 120 {
		classification = "likely_accident"
		confidence = 0.85
	} else if ffRate < 0.10 && (row.ExplosiveFFPercent > 50 || row.AvgTimeBetweenFF > 60) {
		classification = "likely_accident"
		confidence = 0.65
	} else if ffRate > 0.20 || row.FFInFirstMinute > 3 || row.AvgTimeBetweenFF < 30 {
		// Possibly intentional indicators:
		// - High FF rate (> 20%)
		// - Multiple kills in spawn area (first minute)
		// - Very frequent FF (< 30 seconds between)
		classification = "likely_intentional"
		confidence = 0.75
	} else {
		classification = "possibly_intentional"
		confidence = 0.50
	}

	return classification, confidence, nil
}

// UpdateFriendlyFireClassification updates the classification of a FF incident
func UpdateFriendlyFireClassification(ctx context.Context, pbApp core.App, incidentID, classification string, confidence float64) error {
	record, err := pbApp.FindRecordById("friendly_fire_incidents", incidentID)
	if err != nil {
		return err
	}

	record.Set("accident_classification", classification)
	record.Set("confidence_score", confidence)

	return pbApp.Save(record)
}

// GetRecentFriendlyFireIncidents returns recent FF incidents for a match or player
func GetRecentFriendlyFireIncidents(ctx context.Context, pbApp core.App, matchID string, killerID *string, limit int) ([]FriendlyFireIncident, error) {
	filter := "match = {:match}"
	params := map[string]any{"match": matchID}

	if killerID != nil {
		filter += " && killer = {:killer}"
		params["killer"] = *killerID
	}

	records, err := pbApp.FindRecordsByFilter(
		"friendly_fire_incidents",
		filter,
		"-timestamp",
		limit,
		0,
		params,
	)

	if err != nil {
		return nil, err
	}

	incidents := make([]FriendlyFireIncident, len(records))
	for i, record := range records {
		timestamp, _ := time.Parse(time.RFC3339, record.GetString("timestamp"))

		incidents[i] = FriendlyFireIncident{
			ID:                     record.Id,
			MatchID:                record.GetString("match"),
			KillerID:               record.GetString("killer"),
			VictimID:               record.GetString("victim"),
			Weapon:                 record.GetString("weapon"),
			Timestamp:              timestamp,
			IsExplosiveWeapon:      record.GetBool("is_explosive_weapon"),
			IsVehicleWeapon:        record.GetBool("is_vehicle_weapon"),
			AccidentClassification: stringPtr(record.GetString("accident_classification")),
			ConfidenceScore:        float64Ptr(record.GetFloat("confidence_score")),
		}

		if team := record.GetInt("killer_team"); team != 0 {
			incidents[i].KillerTeam = intPtr(team)
		}
		if team := record.GetInt("victim_team"); team != 0 {
			incidents[i].VictimTeam = intPtr(team)
		}
		if secs := record.GetFloat("time_since_match_start_seconds"); secs != 0 {
			incidents[i].TimeSinceMatchStartSecs = &secs
		}
		if weapon := record.GetString("weapon"); weapon != "" {
			incidents[i].Weapon = weapon
		}
	}

	return incidents, nil
}

// Helper functions for pointer conversions
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func float64Ptr(f float64) *float64 {
	if f == 0 {
		return nil
	}
	return &f
}

func intPtr(i int) *int {
	if i == 0 {
		return nil
	}
	return &i
}
