package parser

import (
	"context"
	"sandstorm-tracker/internal/database"
	"strings"
	"testing"

	_ "sandstorm-tracker/migrations"

	"github.com/pocketbase/pocketbase/tests"
)

func TestWeaponNameStandardization(t *testing.T) {
	// Create test PocketBase app with temp directory
	tempDir := t.TempDir()
	testApp, err := tests.NewTestApp(tempDir)
	if err != nil {
		t.Fatalf("failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	// Create collections
	setupTestCollections(t, testApp)

	ctx := context.Background()

	t.Run("PF940_with_different_IDs_creates_single_weapon_stat", func(t *testing.T) {
		serverExternalID := "test-server-weapon"
		_, err := database.GetOrCreateServer(ctx, testApp, serverExternalID, "Weapon Test Server", "test/path")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		// Create parser
		parser := NewLogParser(testApp)

		// Create a map load event first to establish a match
		mapLoadLine := `[2025.11.08-13.59.15:803][  0]LogLoad: LoadMap: /Game/Maps/Ministry/Ministry?Name=Player?Scenario=Scenario_Ministry_Checkpoint_Security?MaxPlayers=8?Lighting=Day`
		err = parser.ParseAndProcess(ctx, mapLoadLine, serverExternalID, "test.log")
		if err != nil {
			t.Fatalf("failed to process map load: %v", err)
		}

		// Process kill event with BP_Firearm_PF940_C_2147480339
		killLine1 := `[2025.11.08-14.05.58:221][  2]LogGameplayEvents: Display: ArmoredBear[76561198995742987, team 0] killed Observer[INVALID, team 1] with BP_Firearm_PF940_C_2147480339`
		err = parser.ParseAndProcess(ctx, killLine1, serverExternalID, "test.log")
		if err != nil {
			t.Fatalf("failed to process kill line 1: %v", err)
		}

		// Process kill event with BP_Firearm_PF940_C_2147479568
		killLine2 := `[2025.11.08-14.07.26:741][309]LogGameplayEvents: Display: ArmoredBear[76561198995742987, team 0] killed Suicide Bomber[INVALID, team 1] with BP_Firearm_PF940_C_2147479568`
		err = parser.ParseAndProcess(ctx, killLine2, serverExternalID, "test.log")
		if err != nil {
			t.Fatalf("failed to process kill line 2: %v", err)
		}

		// Query weapon stats from database
		weaponStats, err := testApp.FindRecordsByFilter(
			"match_weapon_stats",
			"",
			"",
			-1,
			0,
			map[string]any{},
		)
		if err != nil {
			t.Fatalf("failed to query weapon stats: %v", err)
		}

		// Should have exactly 1 weapon stat record for "Firearm PF940" (not 2 separate records)
		pf940Count := 0
		var pf940WeaponName string
		var pf940Kills int

		for _, stat := range weaponStats {
			weaponName := stat.GetString("weapon_name")
			if weaponName == "PF940" {
				pf940Count++
				pf940WeaponName = weaponName
				pf940Kills = stat.GetInt("kills")
			} else if weaponName != "" {
				// Make sure we don't have any variants with IDs
				t.Errorf("Unexpected weapon name with ID: '%s'", weaponName)
			}
		}

		if pf940Count != 1 {
			t.Errorf("Expected exactly 1 weapon stat record for 'PF940', got %d", pf940Count)
			// Debug: show all weapon names
			for _, stat := range weaponStats {
				t.Logf("Found weapon: '%s' with %d kills", stat.GetString("weapon_name"), stat.GetInt("kills"))
			}
		}

		if pf940WeaponName != "PF940" {
			t.Errorf("Expected weapon name 'PF940', got '%s'", pf940WeaponName)
		}

		// Both kills should be tracked under the same weapon
		if pf940Kills != 2 {
			t.Errorf("Expected 2 kills for 'PF940', got %d", pf940Kills)
		}
	})

	t.Run("cleanWeaponName_removes_IDs_correctly", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected string
		}{
			{"BP_Firearm_PF940_C_2147480339", "PF940"},
			{"BP_Firearm_PF940_C_2147479568", "PF940"},
			{"BP_Firearm_M16A2_C_2147481522", "M16A2"},
			{"BP_Weapon_AK74_C_123456789", "AK74"},
			{"ODCheckpoint A", "ODCheckpoint"},                // Standardized - remove letter suffix
			{"ODCheckpoint_B", "ODCheckpoint"},                // Standardized - remove letter suffix
			{"ODCheckpoint B", "ODCheckpoint"},                // Standardized - remove letter suffix
			{"BP_Melee_Knife_C", "Knife"},                     // No numeric ID
			{"BP_Projectile_Molotov_C_2147480261", "Molotov"}, // Projectile weapon
			{"BP_Projectile_GAU8_C_2147477120", "GAU8"},       // Projectile weapon
			{"BP_Projectile_F1_C_2147467410", "F1"},           // Projectile weapon
		}

		for _, tc := range testCases {
			result := cleanWeaponName(tc.input)
			if result != tc.expected {
				t.Errorf("cleanWeaponName(%q) = %q, expected %q", tc.input, result, tc.expected)
			}
		}
	})
}

func TestWeaponStatsAggregation(t *testing.T) {
	// Create test PocketBase app with temp directory
	tempDir := t.TempDir()
	testApp, err := tests.NewTestApp(tempDir)
	if err != nil {
		t.Fatalf("failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	// Create collections
	setupTestCollections(t, testApp)

	ctx := context.Background()

	t.Run("multiple_weapons_tracked_separately", func(t *testing.T) {
		serverExternalID := "test-server-multi-weapon"
		_, err := database.GetOrCreateServer(ctx, testApp, serverExternalID, "Multi Weapon Server", "test/path")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		// Create parser
		parser := NewLogParser(testApp)

		// Create a map load event first
		mapLoadLine := `[2025.11.08-13.59.15:803][  0]LogLoad: LoadMap: /Game/Maps/Ministry/Ministry?Name=Player?Scenario=Scenario_Ministry_Checkpoint_Security?MaxPlayers=8?Lighting=Day`
		err = parser.ParseAndProcess(ctx, mapLoadLine, serverExternalID, "test.log")
		if err != nil {
			t.Fatalf("failed to process map load: %v", err)
		}

		// Process kills with different weapons
		kills := []string{
			`[2025.11.08-14.00.00:000][  0]LogGameplayEvents: Display: Player1[76561198000000001, team 0] killed Enemy1[INVALID, team 1] with BP_Firearm_PF940_C_2147480339`,
			`[2025.11.08-14.00.01:000][  1]LogGameplayEvents: Display: Player1[76561198000000001, team 0] killed Enemy2[INVALID, team 1] with BP_Firearm_PF940_C_2147479568`,
			`[2025.11.08-14.00.02:000][  2]LogGameplayEvents: Display: Player1[76561198000000001, team 0] killed Enemy3[INVALID, team 1] with BP_Firearm_M16A2_C_2147481522`,
			`[2025.11.08-14.00.03:000][  3]LogGameplayEvents: Display: Player1[76561198000000001, team 0] killed Enemy4[INVALID, team 1] with BP_Firearm_M16A2_C_9999999999`,
		}

		for _, killLine := range kills {
			err = parser.ParseAndProcess(ctx, killLine, serverExternalID, "test.log")
			if err != nil {
				t.Fatalf("failed to process kill line: %v", err)
			}
		}

		// Query weapon stats
		weaponStats, err := testApp.FindRecordsByFilter(
			"match_weapon_stats",
			"",
			"",
			-1,
			0,
			map[string]any{},
		)
		if err != nil {
			t.Fatalf("failed to query weapon stats: %v", err)
		}

		// Should have exactly 2 weapon types: PF940 (2 kills) and M16A2 (2 kills)
		weaponKills := make(map[string]int)
		for _, stat := range weaponStats {
			weaponName := stat.GetString("weapon_name")
			if weaponName != "" {
				weaponKills[weaponName] += stat.GetInt("kills")
			}
		}

		if len(weaponKills) != 2 {
			t.Errorf("Expected 2 distinct weapons, got %d", len(weaponKills))
			for weapon, kills := range weaponKills {
				t.Logf("Weapon: '%s', Kills: %d", weapon, kills)
			}
		}

		if weaponKills["PF940"] != 2 {
			t.Errorf("Expected 2 kills for 'PF940', got %d", weaponKills["PF940"])
		}

		if weaponKills["M16A2"] != 2 {
			t.Errorf("Expected 2 kills for 'M16A2', got %d", weaponKills["M16A2"])
		}
	})

	t.Run("ODCheckpoint_variants_aggregate_to_single_weapon", func(t *testing.T) {
		serverExternalID := "test-server-odcheckpoint"
		_, err := database.GetOrCreateServer(ctx, testApp, serverExternalID, "ODCheckpoint Server", "test/path")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		// Create parser
		parser := NewLogParser(testApp)

		// Create a map load event first
		mapLoadLine := `[2025.11.08-13.59.15:803][  0]LogLoad: LoadMap: /Game/Maps/Ministry/Ministry?Name=Player?Scenario=Scenario_Ministry_Checkpoint_Security?MaxPlayers=8?Lighting=Day`
		err = parser.ParseAndProcess(ctx, mapLoadLine, serverExternalID, "test.log")
		if err != nil {
			t.Fatalf("failed to process map load: %v", err)
		}

		// Process kills with ODCheckpoint A and ODCheckpoint B
		kills := []string{
			`[2025.11.08-14.00.00:000][  0]LogGameplayEvents: Display: Player1[76561198000000001, team 0] killed Enemy1[INVALID, team 1] with ODCheckpoint A`,
			`[2025.11.08-14.00.01:000][  1]LogGameplayEvents: Display: Player1[76561198000000001, team 0] killed Enemy2[INVALID, team 1] with ODCheckpoint B`,
			`[2025.11.08-14.00.02:000][  2]LogGameplayEvents: Display: Player1[76561198000000001, team 0] killed Enemy3[INVALID, team 1] with ODCheckpoint A`,
		}

		for _, killLine := range kills {
			err = parser.ParseAndProcess(ctx, killLine, serverExternalID, "test.log")
			if err != nil {
				t.Fatalf("failed to process kill line: %v", err)
			}
		}

		// Query weapon stats
		weaponStats, err := testApp.FindRecordsByFilter(
			"match_weapon_stats",
			"",
			"",
			-1,
			0,
			map[string]any{},
		)
		if err != nil {
			t.Fatalf("failed to query weapon stats: %v", err)
		}

		// Should have exactly 1 weapon stat record for "ODCheckpoint" (not separate A and B)
		odCheckpointCount := 0
		var odCheckpointKills int

		for _, stat := range weaponStats {
			weaponName := stat.GetString("weapon_name")
			if weaponName == "ODCheckpoint" {
				odCheckpointCount++
				odCheckpointKills = stat.GetInt("kills")
			} else if strings.Contains(weaponName, "ODCheckpoint") {
				// Make sure we don't have any variants with A or B
				t.Errorf("Unexpected ODCheckpoint variant: '%s'", weaponName)
			}
		}

		if odCheckpointCount != 1 {
			t.Errorf("Expected exactly 1 weapon stat record for 'ODCheckpoint', got %d", odCheckpointCount)
			// Debug: show all weapon names
			for _, stat := range weaponStats {
				t.Logf("Found weapon: '%s'", stat.GetString("weapon_name"))
			}
		}

		// All 3 kills should be tracked under the same weapon
		if odCheckpointKills != 3 {
			t.Errorf("Expected 3 kills for 'ODCheckpoint', got %d", odCheckpointKills)
		}
	})

	t.Run("Projectile_weapons_with_different_IDs_aggregate_correctly", func(t *testing.T) {
		serverExternalID := "test-server-projectile"
		_, err := database.GetOrCreateServer(ctx, testApp, serverExternalID, "Projectile Server", "test/path")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		// Create parser
		parser := NewLogParser(testApp)

		// Create a map load event first
		mapLoadLine := `[2025.11.08-13.59.15:803][  0]LogLoad: LoadMap: /Game/Maps/Ministry/Ministry?Name=Player?Scenario=Scenario_Ministry_Checkpoint_Security?MaxPlayers=8?Lighting=Day`
		err = parser.ParseAndProcess(ctx, mapLoadLine, serverExternalID, "test.log")
		if err != nil {
			t.Fatalf("failed to process map load: %v", err)
		}

		// Get the active match for this server using the external ID
		match, err := database.GetActiveMatch(ctx, testApp, serverExternalID)
		if err != nil {
			t.Fatalf("failed to get active match: %v", err)
		}

		// Process kills with projectile weapons (different IDs for same weapon type)
		kills := []string{
			`[2025.11.08-14.00.00:000][  0]LogGameplayEvents: Display: Player1[76561198000000001, team 0] killed Enemy1[INVALID, team 1] with BP_Projectile_Molotov_C_2147480261`,
			`[2025.11.08-14.00.01:000][  1]LogGameplayEvents: Display: Player1[76561198000000001, team 0] killed Enemy2[INVALID, team 1] with BP_Projectile_Molotov_C_9999999999`,
			`[2025.11.08-14.00.02:000][  2]LogGameplayEvents: Display: Player1[76561198000000001, team 0] killed Enemy3[INVALID, team 1] with BP_Projectile_GAU8_C_2147477120`,
			`[2025.11.08-14.00.03:000][  3]LogGameplayEvents: Display: Player1[76561198000000001, team 0] killed Enemy4[INVALID, team 1] with BP_Projectile_F1_C_2147467410`,
			`[2025.11.08-14.00.04:000][  4]LogGameplayEvents: Display: Player1[76561198000000001, team 0] killed Enemy5[INVALID, team 1] with BP_Projectile_F1_C_1111111111`,
		}

		for _, killLine := range kills {
			err = parser.ParseAndProcess(ctx, killLine, serverExternalID, "test.log")
			if err != nil {
				t.Fatalf("failed to process kill line: %v", err)
			}
		}

		// Query weapon stats for this match only
		weaponStats, err := testApp.FindRecordsByFilter(
			"match_weapon_stats",
			"match = {:matchId}",
			"",
			-1,
			0,
			map[string]any{
				"matchId": match.ID,
			},
		)
		if err != nil {
			t.Fatalf("failed to query weapon stats: %v", err)
		}

		// Count kills per weapon
		weaponKills := make(map[string]int)
		for _, stat := range weaponStats {
			weaponName := stat.GetString("weapon_name")
			if weaponName != "" {
				weaponKills[weaponName] += stat.GetInt("kills")
			}
		}

		// Should have exactly 3 distinct weapons: Molotov (2), GAU8 (1), F1 (2)
		if len(weaponKills) != 3 {
			t.Errorf("Expected 3 distinct projectile weapons, got %d", len(weaponKills))
			for weapon, kills := range weaponKills {
				t.Logf("Weapon: '%s', Kills: %d", weapon, kills)
			}
		}

		if weaponKills["Molotov"] != 2 {
			t.Errorf("Expected 2 kills for 'Molotov', got %d", weaponKills["Molotov"])
		}

		if weaponKills["GAU8"] != 1 {
			t.Errorf("Expected 1 kill for 'GAU8', got %d", weaponKills["GAU8"])
		}

		if weaponKills["F1"] != 2 {
			t.Errorf("Expected 2 kills for 'F1', got %d", weaponKills["F1"])
		}
	})
}
