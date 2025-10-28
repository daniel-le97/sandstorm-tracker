package utils

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"sandstorm-tracker/internal/db"
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

	// List all players and their stats
	fmt.Println("Players in database:")
	for _, player := range players {
		fmt.Printf("ID: %d, Name: %s, SteamID: %s\n", player.ID, player.Name, player.ExternalID)
		stats, err := queries.GetPlayerStatsByPlayerID(ctx, player.ID)
		if err == nil {
			// Get total kills using SQL aggregation
			totalKills, killsErr := queries.GetTotalKillsForPlayerStats(ctx, stats.ID)
			if killsErr != nil {
				totalKills = 0
			}
			fmt.Printf("  Stats: Kills=%v, Deaths=%v, FF Kills=%v, Highest Score=%v\n",
				totalKills, derefInt64(stats.TotalDeaths), derefInt64(stats.FriendlyFireKills), derefInt64(stats.HighestScore))
			weaponStats, err := queries.GetWeaponStatsForPlayerStats(ctx, stats.ID)
			if err == nil && len(weaponStats) > 0 {
				fmt.Println("  Weapon stats:")
				for _, ws := range weaponStats {
					fmt.Printf("    Weapon: %s, Kills: %v, Assists: %v\n", ws.WeaponName, derefInt64(ws.Kills), derefInt64(ws.Assists))
				}
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
