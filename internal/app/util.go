package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"sandstorm-tracker/internal/db"
	generated "sandstorm-tracker/internal/db/generated"
)

// CheckDatabase checks the contents of the database and prints statistics
func CheckDatabase(dbPath string) {
	dbService, err := db.NewDatabaseService(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer dbService.Close()

	queries := dbService.GetQueries()
	ctx := context.Background()

	// Get player count
	players, err := queries.ListPlayers(ctx)
	if err != nil {
		log.Printf("Error getting players: %v", err)
		return
	}

	fmt.Printf("Total players: %d\n\n", len(players))

	// Get match count
	matches, err := queries.ListMatches(ctx)
	if err != nil {
		log.Printf("Error getting matches: %v", err)
	} else {
		fmt.Printf("Total matches: %d\n\n", len(matches))
	}

	// List all players and their all-time stats (aggregated from matches)
	fmt.Println("Players in database:")
	for _, player := range players {
		fmt.Printf("ID: %d, Name: %s, SteamID: %s\n", player.ID, player.Name, player.ExternalID)

		// Get player's best match
		bestMatch, err := queries.GetPlayerBestMatch(ctx, player.ID)
		if err == nil {
			fmt.Printf("  Best Match: %v kills on %s (%s)\n",
				derefInt64(bestMatch.Kills),
				derefString(bestMatch.Map),
				bestMatch.Mode)
		}

		// Get last 10 matches
		matchHistory, err := queries.GetPlayerMatchHistory(ctx, generated.GetPlayerMatchHistoryParams{
			PlayerID: player.ID,
			Limit:    10,
		})
		if err == nil && len(matchHistory) > 0 {
			fmt.Printf("  Last %d matches:\n", len(matchHistory))
			for _, match := range matchHistory {
				fmt.Printf("    Match %d: %v kills, %v deaths, %v assists on %s\n",
					match.MatchID,
					derefInt64(match.Kills),
					derefInt64(match.Deaths),
					derefInt64(match.Assists),
					derefString(match.Map))
			}
		}
	}
}

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

func GetServerIdFromPath(path string) (string, error) {
	// Example path: C:\Games\Steam\steamapps\common\Sandstorm Dedicated Server\Server1\Logs
	entries, err := os.ReadDir(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("path does not exist: %s", path)
		}
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.Contains(name, "backup") {
			continue
		}
		return strings.Trim(name, ".log"), nil
	}
	return "", fmt.Errorf("could not determine server ID from path: %s", path)
}
