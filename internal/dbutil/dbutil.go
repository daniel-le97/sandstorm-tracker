package dbutil

import (
	"context"
	"fmt"
	"log"

	"sandstorm-tracker/db"
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
	players, err := queries.ListAllPlayers(ctx)
	if err != nil {
		log.Printf("Error getting players: %v", err)
		return
	}

	fmt.Printf("Total players: %d\n\n", len(players))

	// List all players
	fmt.Println("Players in database:")
	for _, player := range players {
		fmt.Printf("ID: %d, Name: %s, SteamID: %s\n", player.ID, player.Name, player.ExternalID)
	}

	// Get kill count
	kills, err := queries.ListAllKills(ctx)
	if err != nil {
		log.Printf("Error getting kills: %v", err)
		return
	}

	fmt.Printf("\nTotal kills: %d\n", len(kills))

	// Count kills by player
	killCounts := make(map[string]int)
	for _, kill := range kills {
		if kill.KillerID != nil {
			for _, player := range players {
				if player.ID == *kill.KillerID {
					killCounts[player.Name]++
					break
				}
			}
		}
	}

	fmt.Println("\nKill counts by player:")
	for playerName, count := range killCounts {
		fmt.Printf("%s: %d kills\n", playerName, count)
	}
}
