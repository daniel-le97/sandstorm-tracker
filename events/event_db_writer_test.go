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
	serverParams := gen.UpsertServerParams{
		ExternalID: "test-server-1",
		Name:       "Test Server",
		Path:       &serverPath,
	}
	serverID, err := queries.UpsertServer(ctx, serverParams)
	if err != nil {
		t.Fatalf("failed to insert server: %v", err)
	}

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

	// Check total real kills in DB (exclude suicides and friendly fire)
	kills, err := queries.ListAllKills(ctx)
	if err != nil {
		t.Fatalf("ListAllKills failed: %v", err)
	}
	realKills := 0
	for _, k := range kills {
		if k.KillType == 0 {
			realKills++
		}
	}
	if realKills != 119 {
		t.Errorf("Expected 119 real kill events in DB, got %d", realKills)
	}

	// Check kills per player (by name substring, as in parse_check_test.go)
	players, err := queries.ListAllPlayers(ctx)
	if err != nil {
		t.Fatalf("ListAllPlayers failed: %v", err)
	}
	nameToKills := map[string]int{
		"Rabbit":      0,
		"0rigin":      0,
		"ArmoredBear": 0,
		"Blue":        0,
	}
	for _, kill := range kills {
		if kill.KillerID == nil {
			continue
		}
		// Find player by ID
		var killerName string
		for _, p := range players {
			if p.ID == *kill.KillerID {
				killerName = p.Name
				break
			}
		}
		for key := range nameToKills {
			if strings.Contains(killerName, key) {
				nameToKills[key]++
			}
		}
	}
	if nameToKills["Rabbit"] != 53 {
		t.Errorf("Expected 53 Rabbit kills, got %d", nameToKills["Rabbit"])
	}
	if nameToKills["0rigin"] != 36 {
		t.Errorf("Expected 36 0rigin kills, got %d", nameToKills["0rigin"])
	}
	if nameToKills["ArmoredBear"] != 19 {
		t.Errorf("Expected 19 ArmoredBear kills, got %d", nameToKills["ArmoredBear"])
	}
	if nameToKills["Blue"] != 9 {
		t.Errorf("Expected 9 Blue kills, got %d", nameToKills["Blue"])
	}
	total := nameToKills["Rabbit"] + nameToKills["0rigin"] + nameToKills["ArmoredBear"] + nameToKills["Blue"]
	if total != 117 {
		t.Errorf("Expected 117 total parsed kills for tracked players, got %d", total)
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
	serverParams := gen.UpsertServerParams{
		ExternalID: "test-server-2",
		Name:       "Test Server 2",
		Path:       &serverPath,
	}
	serverID, err := queries.UpsertServer(ctx, serverParams)
	if err != nil {
		t.Fatalf("failed to insert server: %v", err)
	}

	// Simulate a kill event
	event := &GameEvent{
		Type:      EventPlayerKill,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
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
	player, err := queries.GetPlayer(ctx, "123456")
	if err != nil {
		t.Fatalf("GetPlayer failed: %v", err)
	}
	if player.Name != "TestKiller" {
		t.Errorf("expected player name 'TestKiller', got '%s'", player.Name)
	}

	// Check that the kill was inserted
	kills, err := queries.ListAllKills(ctx)
	if err != nil {
		t.Fatalf("ListAllKills failed: %v", err)
	}
	if len(kills) != 1 {
		t.Fatalf("expected 1 kill, got %d", len(kills))
	}
	if kills[0].KillerID == nil || *kills[0].KillerID != player.ID {
		t.Errorf("expected kill.KillerID %d, got %v", player.ID, kills[0].KillerID)
	}
	if kills[0].VictimName == nil || *kills[0].VictimName != "TestVictim" {
		t.Errorf("expected kill.VictimName 'TestVictim', got %v", kills[0].VictimName)
	}
	if kills[0].WeaponName == nil || *kills[0].WeaponName != "TestWeapon" {
		t.Errorf("expected kill.WeaponName 'TestWeapon', got %v", kills[0].WeaponName)
	}
}
