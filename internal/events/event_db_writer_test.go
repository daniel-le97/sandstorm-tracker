package events

import (
	"bufio"
	"context"
	"os"
	"strings"
	"testing"
	"time"

	db "sandstorm-tracker/db"
	gen "sandstorm-tracker/db/generated"
)

func TestParseAndWriteLogToDB_HCLog(t *testing.T) {
	killLogFile, err := os.Create("db_kills.log")
	if err != nil {
		t.Fatalf("failed to create db_kills.log: %v", err)
	}
	defer killLogFile.Close()

	dbPath := "test_parse_and_write_hclog.sqlite"
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

	// Insert a server row so serverID=1 is valid for foreign key constraints
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
	scanner := bufio.NewScanner(file)
	var matchID *int64 = nil

	// Parse and write all events
	var totalParsedKills int
	for scanner.Scan() {
		line := scanner.Text()
		event, err := parser.ParseLine(line, "hc")
		if err != nil {
			continue
		}
		if event == nil {
			continue
		}
		if event.Type == EventPlayerKill {
			totalParsedKills++
		}
		// Only write kill events to DB (matches WriteEventToDB logic)
		if event.Type == EventPlayerKill || event.Type == EventFriendlyFire || event.Type == EventSuicide {
			// Log the event to the file for inspection
			killLogFile.WriteString(line + "\n")
			if err := WriteEventToDB(ctx, queries, event, serverID, matchID); err != nil {
				t.Fatalf("WriteEventToDB failed: %v", err)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("Error scanning hc.log: %v", err)
	}

	// Check that expected players exist and have weapon stats
	players, err := queries.ListPlayers(ctx)
	if err != nil {
		t.Fatalf("ListPlayers failed: %v", err)
	}
	trackedNames := []string{"Rabbit", "0rigin", "ArmoredBear", "Blue"}
	found := map[string]bool{}
	for _, p := range players {
		for _, name := range trackedNames {
			if strings.Contains(p.Name, name) {
				found[name] = true
				// Check weapon stats exist for this player
				stats, err := queries.GetPlayerStatsByPlayerID(ctx, p.ID)
				if err != nil {
					t.Errorf("No player_stats for player %s: %v", p.Name, err)
				}
				weaponStats, err := queries.GetWeaponStatsForPlayerStats(ctx, stats.ID)
				if err != nil {
					t.Errorf("No weapon_stats for player %s: %v", p.Name, err)
				}
				if len(weaponStats) == 0 {
					t.Errorf("Expected weapon_stats for player %s, got none", p.Name)
				}
			}
		}
	}
	for _, name := range trackedNames {
		if !found[name] {
			t.Errorf("Expected player %s in DB, but not found", name)
		}
	}
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
