package database

import (
	"context"
	"testing"
	"time"

	_ "sandstorm-tracker/migrations"

	"github.com/pocketbase/pocketbase/tests"
)

// Helper to create pointers for test values

func int64Ptr(i int64) *int64        { return &i }
func timePtr(t time.Time) *time.Time { return &t }

func setupTestApp(t *testing.T) (*tests.TestApp, func()) {
	// Use temp directory for test database
	tempDir := t.TempDir()

	testApp, err := tests.NewTestApp(tempDir)
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}

	// Verify collections exist
	collections := []string{"servers", "players", "matches", "match_player_stats", "match_weapon_stats"}
	for _, name := range collections {
		if _, err := testApp.FindCollectionByNameOrId(name); err != nil {
			t.Fatalf("Collection %s not found after migration: %v", name, err)
		}
	}

	return testApp, func() { testApp.Cleanup() }
}

func TestGetOrCreateServer(t *testing.T) {
	testApp, cleanup := setupTestApp(t)
	defer cleanup()

	ctx := context.Background()

	// Test creating a new server
	serverID, err := GetOrCreateServer(ctx, testApp, "test-server-1", "Test Server", "/test/path")
	if err != nil {
		t.Fatalf("GetOrCreateServer() error = %v", err)
	}

	if serverID == "" {
		t.Error("GetOrCreateServer() returned empty server ID")
	}

	// Test getting existing server (should return same ID)
	serverID2, err := GetOrCreateServer(ctx, testApp, "test-server-1", "Test Server", "/test/path")
	if err != nil {
		t.Fatalf("GetOrCreateServer() second call error = %v", err)
	}

	if serverID != serverID2 {
		t.Errorf("GetOrCreateServer() returned different ID on second call: %s != %s", serverID, serverID2)
	}

	// Verify server record exists
	serverRecord, err := testApp.FindRecordById("servers", serverID)
	if err != nil {
		t.Fatalf("Server record not found: %v", err)
	}

	if serverRecord.GetString("external_id") != "test-server-1" {
		t.Errorf("Server external_id = %s, want test-server-1", serverRecord.GetString("external_id"))
	}

	if serverRecord.GetString("name") != "Test Server" {
		t.Errorf("Server name = %s, want Test Server", serverRecord.GetString("name"))
	}
}

func TestCreateMatch(t *testing.T) {
	testApp, cleanup := setupTestApp(t)
	defer cleanup()

	ctx := context.Background()

	// Create a test server first
	serverID, err := GetOrCreateServer(ctx, testApp, "test-server-1", "Test Server", "/test/path")
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Create a match using external_id
	match, err := CreateMatch(ctx, testApp, "test-server-1", stringPtr("Crossing"), stringPtr("Push"), nil)
	if err != nil {
		t.Fatalf("CreateMatch() error = %v", err)
	}

	if match == nil || match.ID == "" {
		t.Error("CreateMatch() returned empty match")
	}

	// Verify match record
	matchRecord, err := testApp.FindRecordById("matches", match.ID)
	if err != nil {
		t.Fatalf("Match record not found: %v", err)
	}

	if matchRecord.GetString("server") != serverID {
		t.Errorf("Match server = %s, want %s", matchRecord.GetString("server"), serverID)
	}

	if matchRecord.GetString("map") != "Crossing" {
		t.Errorf("Match map = %s, want Crossing", matchRecord.GetString("map"))
	}

	if matchRecord.GetString("mode") != "Push" {
		t.Errorf("Match mode = %s, want Push", matchRecord.GetString("mode"))
	}

	// end_time should be empty for new match
	if matchRecord.GetString("end_time") != "" {
		t.Error("New match should have empty end_time")
	}
}

func TestGetActiveMatch(t *testing.T) {
	testApp, cleanup := setupTestApp(t)
	defer cleanup()

	ctx := context.Background()

	// Create a test server
	_, err := GetOrCreateServer(ctx, testApp, "test-server-1", "Test Server", "/test/path")
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Initially no active match
	_, err = GetActiveMatch(ctx, testApp, "test-server-1")
	if err == nil {
		t.Error("GetActiveMatch() should return error when no active match exists")
	}

	// Create a match using external_id
	match, err := CreateMatch(ctx, testApp, "test-server-1", stringPtr("Crossing"), stringPtr("Push"), nil)
	if err != nil {
		t.Fatalf("Failed to create match: %v", err)
	}

	// Now should find active match
	activeMatch, err := GetActiveMatch(ctx, testApp, "test-server-1")
	if err != nil {
		t.Fatalf("GetActiveMatch() error = %v", err)
	}

	if activeMatch.ID != match.ID {
		t.Errorf("GetActiveMatch() ID = %s, want %s", activeMatch.ID, match.ID)
	}

	if activeMatch.Mode != "Push" {
		t.Errorf("GetActiveMatch() Mode = %s, want Push", activeMatch.Mode)
	}

	if activeMatch.Map == nil || *activeMatch.Map != "Crossing" {
		t.Errorf("GetActiveMatch() Map = %v, want Crossing", activeMatch.Map)
	}
}

func TestGetOrCreatePlayerByExternalID(t *testing.T) {
	testApp, cleanup := setupTestApp(t)
	defer cleanup()

	ctx := context.Background()

	// Create a new player
	player, err := CreatePlayer(ctx, testApp, "76561198012345678", "TestPlayer")
	if err != nil {
		t.Fatalf("CreatePlayer() error = %v", err)
	}

	if player == nil || player.ID == "" {
		t.Error("CreatePlayer() returned empty player")
	}

	// Get existing player
	player2, err := GetPlayerByExternalID(ctx, testApp, "76561198012345678")
	if err != nil {
		t.Fatalf("GetPlayerByExternalID() error = %v", err)
	}

	if player.ID != player2.ID {
		t.Errorf("GetPlayerByExternalID() returned different ID: %s != %s", player.ID, player2.ID)
	}

	// Verify player record
	playerRecord, err := testApp.FindRecordById("players", player.ID)
	if err != nil {
		t.Fatalf("Player record not found: %v", err)
	}

	if playerRecord.GetString("external_id") != "76561198012345678" {
		t.Errorf("Player external_id = %s, want 76561198012345678", playerRecord.GetString("external_id"))
	}

	if playerRecord.GetString("name") != "TestPlayer" {
		t.Errorf("Player name = %s, want TestPlayer", playerRecord.GetString("name"))
	}
}

func TestUpsertMatchPlayerStats(t *testing.T) {
	testApp, cleanup := setupTestApp(t)
	defer cleanup()

	ctx := context.Background()

	// Create server, match, and player
	_, _ = GetOrCreateServer(ctx, testApp, "test-server-1", "Test Server", "/test/path")
	match, _ := CreateMatch(ctx, testApp, "test-server-1", stringPtr("Crossing"), stringPtr("Push"), nil)
	player, _ := CreatePlayer(ctx, testApp, "76561198012345678", "TestPlayer")

	// Create initial stats
	err := UpsertMatchPlayerStats(ctx, testApp, match.ID, player.ID, int64Ptr(0), nil)
	if err != nil {
		t.Fatalf("UpsertMatchPlayerStats() error = %v", err)
	}

	// Verify stats record exists
	statsRecord, err := testApp.FindFirstRecordByFilter(
		"match_player_stats",
		"match = {:match} && player = {:player}",
		map[string]any{"match": match.ID, "player": player.ID},
	)
	if err != nil {
		t.Fatalf("Stats record not found: %v", err)
	}

	if statsRecord.GetString("match") != match.ID {
		t.Errorf("Stats match = %s, want %s", statsRecord.GetString("match"), match.ID)
	}

	if statsRecord.GetString("player") != player.ID {
		t.Errorf("Stats player = %s, want %s", statsRecord.GetString("player"), player.ID)
	}

	if statsRecord.GetInt("team") != 0 {
		t.Errorf("Stats team = %d, want 0", statsRecord.GetInt("team"))
	}
}

func TestIncrementMatchPlayerStat(t *testing.T) {
	testApp, cleanup := setupTestApp(t)
	defer cleanup()

	ctx := context.Background()

	// Setup: create server, match, player, and initial stats
	_, _ = GetOrCreateServer(ctx, testApp, "test-server-1", "Test Server", "/test/path")
	match, _ := CreateMatch(ctx, testApp, "test-server-1", stringPtr("Crossing"), stringPtr("Push"), nil)
	player, _ := CreatePlayer(ctx, testApp, "76561198012345678", "TestPlayer")
	UpsertMatchPlayerStats(ctx, testApp, match.ID, player.ID, int64Ptr(0), nil)

	// Increment kills
	err := IncrementMatchPlayerStat(ctx, testApp, match.ID, player.ID, "kills")
	if err != nil {
		t.Fatalf("IncrementMatchPlayerStat() error = %v", err)
	}

	// Verify kills incremented
	statsRecord, err := testApp.FindFirstRecordByFilter(
		"match_player_stats",
		"match = {:match} && player = {:player}",
		map[string]any{"match": match.ID, "player": player.ID},
	)
	if err != nil {
		t.Fatalf("Stats record not found: %v", err)
	}

	if statsRecord.GetInt("kills") != 1 {
		t.Errorf("Kills = %d, want 1", statsRecord.GetInt("kills"))
	}

	// Increment again
	IncrementMatchPlayerStat(ctx, testApp, match.ID, player.ID, "kills")

	statsRecord, _ = testApp.FindRecordById("match_player_stats", statsRecord.Id)
	if statsRecord.GetInt("kills") != 2 {
		t.Errorf("Kills = %d, want 2 after second increment", statsRecord.GetInt("kills"))
	}
}

func TestEndMatch(t *testing.T) {
	testApp, cleanup := setupTestApp(t)
	defer cleanup()

	ctx := context.Background()

	// Create server and match
	_, _ = GetOrCreateServer(ctx, testApp, "test-server-1", "Test Server", "/test/path")
	match, _ := CreateMatch(ctx, testApp, "test-server-1", stringPtr("Crossing"), stringPtr("Push"), nil)

	// End the match
	endTime := time.Now()
	err := EndMatch(ctx, testApp, match.ID, timePtr(endTime), int64Ptr(1), nil)
	if err != nil {
		t.Fatalf("EndMatch() error = %v", err)
	}

	// Verify match ended
	matchRecord, err := testApp.FindRecordById("matches", match.ID)
	if err != nil {
		t.Fatalf("Match record not found: %v", err)
	}

	if matchRecord.GetString("end_time") == "" {
		t.Error("Match end_time should be set")
	}

	if matchRecord.GetInt("winner_team") != 1 {
		t.Errorf("Match winner_team = %d, want 1", matchRecord.GetInt("winner_team"))
	}

	// Verify there's no active match anymore
	_, err = GetActiveMatch(ctx, testApp, "test-server-1")
	if err == nil {
		t.Error("GetActiveMatch() should return error after match ended")
	}
}

func TestUpsertMatchWeaponStats(t *testing.T) {
	testApp, cleanup := setupTestApp(t)
	defer cleanup()

	ctx := context.Background()

	// Setup
	_, _ = GetOrCreateServer(ctx, testApp, "test-server-1", "Test Server", "/test/path")
	match, _ := CreateMatch(ctx, testApp, "test-server-1", stringPtr("Crossing"), stringPtr("Push"), nil)
	player, _ := CreatePlayer(ctx, testApp, "76561198012345678", "TestPlayer")

	// Create weapon stats
	err := UpsertMatchWeaponStats(ctx, testApp, match.ID, player.ID, "M4A1", int64Ptr(1), nil)
	if err != nil {
		t.Fatalf("UpsertMatchWeaponStats() error = %v", err)
	}

	// Verify weapon stats record
	weaponRecord, err := testApp.FindFirstRecordByFilter(
		"match_weapon_stats",
		"match = {:match} && player = {:player} && weapon_name = {:weapon}",
		map[string]any{"match": match.ID, "player": player.ID, "weapon": "M4A1"},
	)
	if err != nil {
		t.Fatalf("Weapon stats record not found: %v", err)
	}

	if weaponRecord.GetString("weapon_name") != "M4A1" {
		t.Errorf("Weapon = %s, want M4A1", weaponRecord.GetString("weapon_name"))
	}

	if weaponRecord.GetInt("kills") != 1 {
		t.Errorf("Weapon kills = %d, want 1", weaponRecord.GetInt("kills"))
	}

	// Upsert again (should increment)
	UpsertMatchWeaponStats(ctx, testApp, match.ID, player.ID, "M4A1", int64Ptr(1), nil)

	weaponRecord, _ = testApp.FindRecordById("match_weapon_stats", weaponRecord.Id)
	if weaponRecord.GetInt("kills") != 2 {
		t.Errorf("Weapon kills = %d, want 2 after upsert", weaponRecord.GetInt("kills"))
	}
}

func TestWeaponTypeClassification(t *testing.T) {
	testApp, cleanup := setupTestApp(t)
	defer cleanup()

	ctx := context.Background()

	// Setup
	_, _ = GetOrCreateServer(ctx, testApp, "test-server-1", "Test Server", "/test/path")
	match, _ := CreateMatch(ctx, testApp, "test-server-1", stringPtr("Crossing"), stringPtr("Push"), nil)
	player, _ := CreatePlayer(ctx, testApp, "76561198012345678", "TestPlayer")

	testCases := []struct {
		weaponName   string
		expectedType string
	}{
		// Firearms
		{"BP_Firearm_M4A1_C_2147480587", "Firearm"},
		{"BP_Firearm_AKM_C_2147480339", "Firearm"},
		{"BP_Firearm_M16A4_C_2147481419", "Firearm"},

		// Projectiles
		{"BP_Projectile_Molotov_C_2147480055", "Projectile"},
		{"BP_Projectile_F1_C_2147467410", "Projectile"},
		{"BP_Projectile_GAU8_C_2147477120", "Projectile"},

		// Melee
		{"BP_Melee_Knife", "Melee"},
	}

	for _, tc := range testCases {
		t.Run(tc.weaponName, func(t *testing.T) {
			err := UpsertMatchWeaponStats(ctx, testApp, match.ID, player.ID, tc.weaponName, int64Ptr(1), nil)
			if err != nil {
				t.Fatalf("UpsertMatchWeaponStats() error = %v", err)
			}

			// Use cleaned weapon name for the filter (UpsertMatchWeaponStats cleans it before storage)
			cleanedWeaponName := CleanWeaponName(tc.weaponName)
			weaponRecord, err := testApp.FindFirstRecordByFilter(
				"match_weapon_stats",
				"match = {:match} && player = {:player} && weapon_name = {:weapon}",
				map[string]any{"match": match.ID, "player": player.ID, "weapon": cleanedWeaponName},
			)
			if err != nil {
				t.Fatalf("Weapon stats record not found: %v", err)
			}

			actualType := weaponRecord.GetString("type")
			if actualType != tc.expectedType {
				t.Errorf("Weapon %s: type = %s, want %s", tc.weaponName, actualType, tc.expectedType)
			}
		})
	}
}

func TestGetWeaponType(t *testing.T) {
	testCases := []struct {
		weaponName   string
		expectedType string
		description  string
	}{
		// Firearms - rifles, pistols, sniper rifles, shotguns
		{"BP_Firearm_M4A1_C_2147480587", "Firearm", "M4A1 rifle"},
		{"BP_Firearm_AKM_C_2147480339", "Firearm", "AKM rifle"},
		{"BP_Firearm_M16A4_C_2147481419", "Firearm", "M16A4 rifle"},
		{"BP_Firearm_M1911_C_2147481234", "Firearm", "M1911 pistol"},
		{"BP_Firearm_Mosin_Nagant", "Firearm", "Mosin-Nagant sniper rifle"},
		{"BP_Firearm_Remington_870", "Firearm", "Remington 870 shotgun"},

		// Projectiles - grenades, molotovs, c4, etc
		{"BP_Projectile_Molotov_C_2147480055", "Projectile", "Molotov cocktail"},
		{"BP_Projectile_F1_C_2147467410", "Projectile", "F1 grenade"},
		{"BP_Projectile_GAU8_C_2147477120", "Projectile", "GAU8 cluster bomb"},
		{"BP_Projectile_C4_Remote", "Projectile", "C4 explosive"},

		// Melee - knives, melee weapons
		{"BP_Melee_Knife", "Melee", "Combat knife"},
		{"BP_Melee_Crowbar", "Melee", "Crowbar"},

		// Weapons without full naming convention
		{"M4A1", "M4A1", "Weapon name only (no BP_ prefix)"},
		{"Knife", "Knife", "Simple weapon name"},

		// Edge cases
		{"BP_", "", "Only BP_ prefix"},
		{"BP_Unknown_Weapon_C_12345", "Unknown", "Unknown weapon type"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := GetWeaponType(tc.weaponName)
			if result != tc.expectedType {
				t.Errorf("GetWeaponType(%q) = %q, want %q", tc.weaponName, result, tc.expectedType)
			}
		})
	}
}
