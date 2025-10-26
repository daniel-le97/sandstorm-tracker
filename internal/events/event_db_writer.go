package events

import (
	"context"
	db "sandstorm-tracker/internal/db/generated"
)

// WriteEventToDB writes a parsed GameEvent to the database using the new /db schema and sqlc queries.
func WriteEventToDB(ctx context.Context, queries *db.Queries, event *GameEvent, serverID int64, matchID *int64) error {
	if event == nil {
		return nil
	}

	switch event.Type {
	case EventPlayerKill, EventFriendlyFire, EventSuicide:
		killers, _ := event.Data["killers"].([]Killer)
		weapon, _ := event.Data["weapon"].(string)
		for _, killer := range killers {
			// Upsert player (by SteamID if available, else by name)
			externalID := killer.SteamID
			if externalID == "" {
				externalID = killer.Name // fallback if no SteamID
			}
			player, err := queries.GetPlayerByExternalID(ctx, externalID)
			if err != nil {
				// Player does not exist, create
				player, err = queries.CreatePlayer(ctx, db.CreatePlayerParams{
					ExternalID: externalID,
					Name:       killer.Name,
				})
				if err != nil {
					return err
				}
			}

			// Optionally: add player to match_players if matchID is provided
			if matchID != nil {
				_ = queries.AddPlayerToMatch(ctx, db.AddPlayerToMatchParams{
					MatchID:  *matchID,
					PlayerID: player.ID,
				})
			}

			// Insert or update player_stats (upsert logic)
			playerStats, err := queries.GetPlayerStatsByPlayerID(ctx, player.ID)
			if err != nil {
				// Not found, create new player_stats
				playerStats, err = queries.CreatePlayerStats(ctx, db.CreatePlayerStatsParams{
					ID:                externalID, // Use externalID as unique string ID
					PlayerID:          player.ID,
					ServerID:          serverID,
					GamesPlayed:       nil,
					Wins:              nil,
					Losses:            nil,
					TotalScore:        nil,
					TotalPlayTime:     nil,
					LastLogin:         nil,
					TotalKills:        nil,
					TotalDeaths:       nil,
					FriendlyFireKills: nil,
					HighestScore:      nil,
				})
				if err != nil {
					return err
				}
			}

			// Calculate multiplier before inserting (if needed for other logic)
			// multiplier := calculateKillMultiplier(event, &killer)

			// Insert or update weapon stats
			one := int64(1)
			zero := int64(0)
			_, _ = queries.UpsertWeaponStats(ctx, db.UpsertWeaponStatsParams{
				PlayerStatsID: playerStats.ID,
				WeaponName:    weapon,
				Kills:         &one,
				Assists:       &zero,
			})
		}
	}
	return nil
}

// calculateKillMultiplier determines the multiplier for a kill event and killer.
// Extend this logic as needed for your game rules (e.g., fire support, headshot, etc.)
func calculateKillMultiplier(event *GameEvent, killer *Killer) float64 {
	multiplier := 1.0
	if event.Data["fire_support"] == true {
		multiplier = 2.0
	}
	if event.Data["headshot"] == true {
		multiplier *= 1.5
	}
	return multiplier
}
