package events

import (
	"context"
	"fmt"
	db "sandstorm-tracker/internal/db/generated"
	"time"
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

		// First killer gets the kill, the rest get assists
		for i, killer := range killers {
			// Skip bots (they have SteamID "INVALID")
			if killer.SteamID == "INVALID" {
				continue
			}
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
					TotalDeaths:       nil,
					FriendlyFireKills: nil,
					HighestScore:      nil,
				})
				if err != nil {
					return err
				}
			}
			// Note: No need to update player_stats for kills since we track kills in weapon_stats

			// Calculate multiplier before inserting (if needed for other logic)
			// multiplier := calculateKillMultiplier(event, &killer)

			// Insert or update weapon stats
			weaponName := weapon
			if weaponName == "" {
				weaponName = "Unknown" // Handle cases where weapon is not specified
			}

			// First player gets the kill, rest get assists
			var kills, assists *int64
			one := int64(1)
			zero := int64(0)

			if i == 0 {
				// First killer gets the kill
				kills = &one
				assists = &zero
			} else {
				// Subsequent killers get assists
				kills = &zero
				assists = &one
			}

			// Update lifetime weapon stats
			_, err = queries.UpsertWeaponStats(ctx, db.UpsertWeaponStatsParams{
				PlayerStatsID: playerStats.ID,
				WeaponName:    weaponName,
				Kills:         kills,
				Assists:       assists,
			})
			if err != nil {
				return fmt.Errorf("failed to upsert weapon stats: %w", err)
			}

			// Update daily weapon stats (for rolling time windows)
			now := time.Now()
			today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
			_, err = queries.UpsertDailyWeaponStats(ctx, db.UpsertDailyWeaponStatsParams{
				PlayerID:   player.ID,
				ServerID:   serverID,
				Date:       today,
				WeaponName: weaponName,
				Kills:      kills,
				Assists:    assists,
			})
			if err != nil {
				return fmt.Errorf("failed to upsert daily weapon stats: %w", err)
			}

			// Update daily player stats (one update per killer, not per weapon)
			if i == 0 {
				// Only count the kill once in daily_player_stats (not per weapon)
				_, err = queries.UpsertDailyPlayerStats(ctx, db.UpsertDailyPlayerStatsParams{
					PlayerID:    player.ID,
					ServerID:    serverID,
					Date:        today,
					Kills:       kills,
					Assists:     &zero,
					Deaths:      &zero,
					GamesPlayed: &zero,
					TotalScore:  &zero,
				})
				if err != nil {
					return fmt.Errorf("failed to upsert daily player stats: %w", err)
				}
			} else {
				// Assist only
				_, err = queries.UpsertDailyPlayerStats(ctx, db.UpsertDailyPlayerStatsParams{
					PlayerID:    player.ID,
					ServerID:    serverID,
					Date:        today,
					Kills:       &zero,
					Assists:     assists,
					Deaths:      &zero,
					GamesPlayed: &zero,
					TotalScore:  &zero,
				})
				if err != nil {
					return fmt.Errorf("failed to upsert daily player stats: %w", err)
				}
			}
		}
	}
	return nil
}
