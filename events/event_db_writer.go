package events

import (
	"context"

	gen "sandstorm-tracker/db/generated"
)

// WriteEventToDB writes a parsed GameEvent to the database if it is a kill event.
func WriteEventToDB(ctx context.Context, queries *gen.Queries, event *GameEvent, serverID int64, matchID *int64) error {
	if event == nil {
		return nil
	}

	switch event.Type {
	case EventPlayerKill, EventFriendlyFire, EventSuicide:
		killers, _ := event.Data["killers"].([]Killer)
		victimName, _ := event.Data["victim_name"].(string)
		weapon, _ := event.Data["weapon"].(string)
		isTeamKill := event.Type == EventFriendlyFire
		isSuicide := event.Type == EventSuicide
		createdAt := event.Timestamp

		for _, killer := range killers {
			// Upsert player (by SteamID if available, else by name)
			externalID := killer.SteamID
			if externalID == "" {
				externalID = killer.Name // fallback if no SteamID
			}
			upsertParams := gen.UpsertPlayerParams{
				ExternalID: externalID,
				Name:       killer.Name,
			}
			playerID, err := queries.UpsertPlayer(ctx, upsertParams)
			if err != nil {
				return err
			}

			params := gen.InsertKillParams{
				KillerID:   &playerID,
				VictimName: &victimName,
				ServerID:   serverID,
				WeaponName: &weapon,
				IsTeamKill: &isTeamKill,
				IsSuicide:  &isSuicide,
				MatchID:    matchID,
				CreatedAt:  &createdAt,
			}
			if err := queries.InsertKill(ctx, params); err != nil {
				return err
			}
		}
	}
	return nil
}
