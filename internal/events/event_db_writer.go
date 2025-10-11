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
		createdAt := event.Timestamp

		var killType int64
		switch event.Type {
		case EventPlayerKill:
			killType = 0
		case EventSuicide:
			killType = 1
		case EventFriendlyFire:
			killType = 2
		}

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

            // Calculate multiplier before inserting
            multiplier := calculateKillMultiplier(event, &killer)

			params := gen.InsertKillParams{
				KillerID:   &playerID,
				VictimName: &victimName,
				ServerID:   serverID,
				WeaponName: &weapon,
				KillType:   killType,
				MatchID:    matchID,
				CreatedAt:  &createdAt,
                Multiplier: multiplier,
			}
			if err := queries.InsertKill(ctx, params); err != nil {
				return err
			}
	}
	}
	return nil
}
// calculateKillMultiplier determines the multiplier for a kill event and killer.
// Extend this logic as needed for your game rules (e.g., fire support, headshot, etc.)
func calculateKillMultiplier(event *GameEvent, killer *Killer) float64 {
	// Example logic: you can expand this as needed
	// Default multiplier
	multiplier := 1.0

	// Example: check for fire support (add your own logic)
	if event.Data["fire_support"] == true {
		multiplier = 2.0
	}

	// Example: check for headshot (add your own logic)
	if event.Data["headshot"] == true {
		multiplier *= 1.5
	}

	// Add more rules as needed

	return multiplier

}