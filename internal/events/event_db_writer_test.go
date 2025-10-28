package events

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	db "sandstorm-tracker/internal/db"
	gen "sandstorm-tracker/internal/db/generated"
)

func TestParseAndWriteLogToDB_HCLog(t *testing.T) {
	dbPath := "test_parse_and_write_hclog_v2.sqlite"
	dbService, err := db.NewDatabaseService(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer func() {
		dbService.Close()
		// _ = os.Remove(dbPath)
	}()
	queries := dbService.GetQueries()
	ctx := context.Background()

	// Insert a server row so serverID is valid for foreign key constraints
	serverPath := "test/path"
	serverParams := gen.CreateServerParams{
		ExternalID: "test-server-1",
		Name:       "Test Server",
		Path:       &serverPath,
	}
	server, err := queries.CreateServer(ctx, serverParams)
	if err != nil {
		t.Fatalf("failed to insert server: %v", err)
	}
	serverID := server.ID

	// Open hc.log
	file, err := os.Open("hc.log")
	if err != nil {
		t.Fatalf("Error opening hc.log: %v", err)
	}
	defer file.Close()

	parser := NewEventParser()
	scanner := NewScanner(file)
	var matchID *int64 = nil

	// Parse and write all events, tracking counts like parse_check_test.go
	killEvents := 0
	unparsedKills := 0

	for scanner.Scan() {
		line := scanner.Text()
		event, err := parser.ParseLine(line, "hc")
		if err != nil {
			unparsedKills++
			continue
		}
		if event == nil {
			continue
		}
		if event.Type == EventPlayerKill {
			// Count a kill for each killer in multi-kill lines
			killersData, ok := event.Data["killers"].([]Killer)
			if ok {
				killEvents += len(killersData)
			} else {
				killEvents++ // fallback
			}

			// Write kill events to DB
			if err := WriteEventToDB(ctx, queries, event, serverID, matchID); err != nil {
				t.Fatalf("WriteEventToDB failed: %v", err)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("Error scanning hc.log: %v", err)
	}

	// Now verify the database contains the expected data by querying
	players, err := queries.ListPlayers(ctx)
	if err != nil {
		t.Fatalf("ListPlayers failed: %v", err)
	}

	// Count kills from database for each tracked player
	rabbitKills := 0
	originKills := 0
	armoredBearKills := 0
	blueKills := 0
	foundPlayers := map[string]bool{}

	for _, player := range players {
		playerName := player.Name
		if strings.Contains(playerName, "Rabbit") {
			foundPlayers["Rabbit"] = true
			stats, err := queries.GetPlayerStatsByPlayerID(ctx, player.ID)
			if err != nil {
				t.Fatalf("GetPlayerStatsByPlayerID failed for Rabbit: %v", err)
			}
			rabbitKills = calculateTotalKillsFromWeaponStats(ctx, queries, stats.ID)
		} else if strings.Contains(playerName, "0rigin") {
			foundPlayers["0rigin"] = true
			stats, err := queries.GetPlayerStatsByPlayerID(ctx, player.ID)
			if err != nil {
				t.Fatalf("GetPlayerStatsByPlayerID failed for 0rigin: %v", err)
			}
			originKills = calculateTotalKillsFromWeaponStats(ctx, queries, stats.ID)
		} else if strings.Contains(playerName, "ArmoredBear") {
			foundPlayers["ArmoredBear"] = true
			stats, err := queries.GetPlayerStatsByPlayerID(ctx, player.ID)
			if err != nil {
				t.Fatalf("GetPlayerStatsByPlayerID failed for ArmoredBear: %v", err)
			}
			armoredBearKills = calculateTotalKillsFromWeaponStats(ctx, queries, stats.ID)
		} else if strings.Contains(playerName, "Blue") {
			foundPlayers["Blue"] = true
			stats, err := queries.GetPlayerStatsByPlayerID(ctx, player.ID)
			if err != nil {
				t.Fatalf("GetPlayerStatsByPlayerID failed for Blue: %v", err)
			}
			blueKills = calculateTotalKillsFromWeaponStats(ctx, queries, stats.ID)
		}
	}

	// Run the same assertions as parse_check_test.go but using DB data
	t.Run("kill event count", func(t *testing.T) {
		if killEvents != 119 {
			t.Errorf("Expected 119 kill events, got %d", killEvents)
		}
	})
	t.Run("unparsed kill events", func(t *testing.T) {
		if unparsedKills != 0 {
			t.Errorf("Expected 0 unparsed kill events, got %d", unparsedKills)
		}
	})
	t.Run("Rabbit kills from DB", func(t *testing.T) {
		if rabbitKills != 53 {
			t.Errorf("Expected 53 Rabbit kills from DB, got %d", rabbitKills)
		}
	})
	t.Run("0rigin kills from DB", func(t *testing.T) {
		if originKills != 36 {
			t.Errorf("Expected 36 0rigin kills from DB, got %d", originKills)
		}
	})
	t.Run("ArmoredBear kills from DB", func(t *testing.T) {
		if armoredBearKills != 19 {
			t.Errorf("Expected 19 ArmoredBear kills from DB, got %d", armoredBearKills)
		}
	})
	t.Run("Blue kills from DB", func(t *testing.T) {
		if blueKills != 9 {
			t.Errorf("Expected 9 Blue kills from DB, got %d", blueKills)
		}
	})
	t.Run("total parsed kills from DB", func(t *testing.T) {
		total := rabbitKills + originKills + armoredBearKills + blueKills
		if total != 117 {
			t.Errorf("Expected 117 total parsed kills from DB, got %d", total)
		}
	})
	t.Run("all expected players found", func(t *testing.T) {
		expectedPlayers := []string{"Rabbit", "0rigin", "ArmoredBear", "Blue"}
		for _, name := range expectedPlayers {
			if !foundPlayers[name] {
				t.Errorf("Expected player %s in DB, but not found", name)
			}
		}
	})
}

func TestWriteEventToDB_KillEvent(t *testing.T) {
	dbPath := "test_event_db_writer.sqlite"
	_ = os.Remove(dbPath)
	dbService, err := db.NewDatabaseService(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer func() {
		dbService.Close()
		_ = os.Remove(dbPath)
	}()
	queries := dbService.GetQueries()
	ctx := context.Background()

	// Insert a server row so serverID is valid for foreign key constraints
	serverPath := "test/path"
	serverParams := gen.CreateServerParams{
		ExternalID: "test-server-2",
		Name:       "Test Server 2",
		Path:       &serverPath,
	}
	server, err := queries.CreateServer(ctx, serverParams)
	if err != nil {
		t.Fatalf("failed to insert server: %v", err)
	}
	serverID := server.ID

	// Simulate a kill event
	event := &GameEvent{
		Type:      EventPlayerKill,
		Timestamp: time.Now(),
		Data: map[string]any{
			"killers":     []Killer{{Name: "TestKiller", SteamID: "123456"}},
			"victim_name": "TestVictim",
			"weapon":      "TestWeapon",
		},
	}
	var matchID *int64 = nil

	err = WriteEventToDB(ctx, queries, event, serverID, matchID)
	if err != nil {
		t.Fatalf("WriteEventToDB failed: %v", err)
	}

	// Check that the player was upserted
	player, err := queries.GetPlayerByExternalID(ctx, "123456")
	if err != nil {
		t.Fatalf("GetPlayerByExternalID failed: %v", err)
	}
	if player.Name != "TestKiller" {
		t.Errorf("expected player name 'TestKiller', got '%s'", player.Name)
	}
	// Check that player_stats and weapon_stats were created
	stats, err := queries.GetPlayerStatsByPlayerID(ctx, player.ID)
	if err != nil {
		t.Fatalf("GetPlayerStatsByPlayerID failed: %v", err)
	}
	weaponStats, err := queries.GetWeaponStatsForPlayerStats(ctx, stats.ID)
	if err != nil {
		t.Fatalf("GetWeaponStatsForPlayerStats failed: %v", err)
	}
	if len(weaponStats) == 0 {
		t.Fatalf("expected weapon_stats for player, got none")
	}
}

func TestWriteEventToDB_BotFiltering(t *testing.T) {
	dbPath := "test_bot_filtering.sqlite"
	_ = os.Remove(dbPath)
	dbService, err := db.NewDatabaseService(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer func() {
		dbService.Close()
		_ = os.Remove(dbPath)
	}()
	queries := dbService.GetQueries()
	ctx := context.Background()

	// Insert a server row so serverID is valid for foreign key constraints
	serverPath := "test/path"
	serverParams := gen.CreateServerParams{
		ExternalID: "test-server-bot",
		Name:       "Test Server Bot",
		Path:       &serverPath,
	}
	server, err := queries.CreateServer(ctx, serverParams)
	if err != nil {
		t.Fatalf("failed to insert server: %v", err)
	}
	serverID := server.ID

	// Simulate a bot kill event (bot kills player)
	botKillEvent := &GameEvent{
		Type:      EventPlayerKill,
		Timestamp: time.Now(),
		Data: map[string]any{
			"killers":     []Killer{{Name: "Rifleman", SteamID: "INVALID"}}, // Bot killer
			"victim_name": "TestVictim",
			"weapon":      "TestWeapon",
		},
	}
	var matchID *int64 = nil

	err = WriteEventToDB(ctx, queries, botKillEvent, serverID, matchID)
	if err != nil {
		t.Fatalf("WriteEventToDB failed: %v", err)
	}

	// Check that no players were created (bot should be filtered out)
	players, err := queries.ListPlayers(ctx)
	if err != nil {
		t.Fatalf("ListPlayers failed: %v", err)
	}
	if len(players) != 0 {
		t.Errorf("Expected 0 players in DB (bot should be filtered), got %d", len(players))
		for _, p := range players {
			t.Errorf("Unexpected player in DB: %s (ID: %s)", p.Name, p.ExternalID)
		}
	}
}

func TestGetTotalKillsForPlayerStats(t *testing.T) {
	dbPath := "test_sql_aggregation.sqlite"
	_ = os.Remove(dbPath)
	dbService, err := db.NewDatabaseService(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer func() {
		dbService.Close()
		_ = os.Remove(dbPath)
	}()
	queries := dbService.GetQueries()
	ctx := context.Background()

	// Insert a server and player for testing
	serverPath := "test/path"
	server, err := queries.CreateServer(ctx, gen.CreateServerParams{
		ExternalID: "test-server-sql",
		Name:       "Test Server SQL",
		Path:       &serverPath,
	})
	if err != nil {
		t.Fatalf("failed to insert server: %v", err)
	}

	player, err := queries.CreatePlayer(ctx, gen.CreatePlayerParams{
		ExternalID: "12345",
		Name:       "TestPlayer",
	})
	if err != nil {
		t.Fatalf("failed to create player: %v", err)
	}

	stats, err := queries.CreatePlayerStats(ctx, gen.CreatePlayerStatsParams{
		ID:       "12345",
		PlayerID: player.ID,
		ServerID: server.ID,
	})
	if err != nil {
		t.Fatalf("failed to create player stats: %v", err)
	}

	// Add some weapon stats
	weapons := map[string]int64{
		"AK-74": 5,
		"M16A4": 3,
		"M9":    1,
	}

	expectedTotal := int64(0)
	for weapon, kills := range weapons {
		expectedTotal += kills
		_, err = queries.UpsertWeaponStats(ctx, gen.UpsertWeaponStatsParams{
			PlayerStatsID: stats.ID,
			WeaponName:    weapon,
			Kills:         &kills,
			Assists:       &[]int64{0}[0],
		})
		if err != nil {
			t.Fatalf("failed to insert weapon stats: %v", err)
		}
	}

	// Test the SQL aggregation
	totalKills, err := queries.GetTotalKillsForPlayerStats(ctx, stats.ID)
	if err != nil {
		t.Fatalf("GetTotalKillsForPlayerStats failed: %v", err)
	}

	if totalKills != expectedTotal {
		t.Errorf("Expected %d total kills, got %d", expectedTotal, totalKills)
	}

	// Test with player that has no weapon stats
	player2, err := queries.CreatePlayer(ctx, gen.CreatePlayerParams{
		ExternalID: "54321",
		Name:       "NoKillsPlayer",
	})
	if err != nil {
		t.Fatalf("failed to create player2: %v", err)
	}

	stats2, err := queries.CreatePlayerStats(ctx, gen.CreatePlayerStatsParams{
		ID:       "54321",
		PlayerID: player2.ID,
		ServerID: server.ID,
	})
	if err != nil {
		t.Fatalf("failed to create player2 stats: %v", err)
	}

	totalKills2, err := queries.GetTotalKillsForPlayerStats(ctx, stats2.ID)
	if err != nil {
		t.Fatalf("GetTotalKillsForPlayerStats failed for player2: %v", err)
	}

	if totalKills2 != 0 {
		t.Errorf("Expected 0 total kills for player with no weapon stats, got %d", totalKills2)
	}
}

// Helper function to calculate total kills from weapon stats using SQL aggregation
func calculateTotalKillsFromWeaponStats(ctx context.Context, queries *gen.Queries, playerStatsID string) int {
	totalKills, err := queries.GetTotalKillsForPlayerStats(ctx, playerStatsID)
	if err != nil {
		return 0
	}
	return int(totalKills)
}
