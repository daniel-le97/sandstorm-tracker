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
	testApp, err := tests.NewTestApp(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	ctx := context.Background()

	t.Run("PF940_with_different_IDs_creates_events", func(t *testing.T) {
		serverExternalID := "test-server-weapon"
		_, err := database.GetOrCreateServer(ctx, testApp, serverExternalID, "Weapon Test Server", "test/path")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		parser := NewLogParser(testApp, testApp.Logger())

		// Create a map load event first
		mapLoadLine := `[2025.11.08-13.59.15:803][  0]LogLoad: LoadMap: /Game/Maps/Ministry/Ministry?Name=Player?Scenario=Scenario_Ministry_Checkpoint_Security?MaxPlayers=8?Lighting=Day`
		parser.ParseAndProcess(ctx, mapLoadLine, serverExternalID, "test.log")

		// Process kill events with different weapon IDs
		killLine1 := `[2025.11.08-14.05.58:221][  2]LogGameplayEvents: Display: ArmoredBear[76561198995742987, team 0] killed Observer[INVALID, team 1] with BP_Firearm_PF940_C_2147480339`
		parser.ParseAndProcess(ctx, killLine1, serverExternalID, "test.log")

		killLine2 := `[2025.11.08-14.07.26:741][309]LogGameplayEvents: Display: ArmoredBear[76561198995742987, team 0] killed Suicide Bomber[INVALID, team 1] with BP_Firearm_PF940_C_2147479568`
		parser.ParseAndProcess(ctx, killLine2, serverExternalID, "test.log")

		// Verify kill events were created with weapon data
		events, err := testApp.FindRecordsByFilter("events", "type = 'player_kill'", "-created", 10, 0)
		if err != nil {
			t.Fatalf("failed to query kill events: %v", err)
		}

		if len(events) < 2 {
			t.Fatalf("Expected at least 2 kill events, got %d", len(events))
		}

		// Check that both events contain PF940 weapon
		pf940Count := 0
		for _, event := range events {
			data := event.GetString("data")
			if strings.Contains(data, "PF940") {
				pf940Count++
			}
		}

		if pf940Count != 2 {
			t.Errorf("Expected 2 kill events with PF940, got %d", pf940Count)
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
	testApp, err := tests.NewTestApp(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	ctx := context.Background()

	t.Run("multiple_weapons_create_separate_events", func(t *testing.T) {
		serverExternalID := "test-server-multi"
		_, err := database.GetOrCreateServer(ctx, testApp, serverExternalID, "Multi Weapon Server", "test/path")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		parser := NewLogParser(testApp, testApp.Logger())
		mapLoad := `[2025.11.08-13.59.15:803][  0]LogLoad: LoadMap: /Game/Maps/Ministry/Ministry?Name=Player?Scenario=Scenario_Ministry_Checkpoint_Security?MaxPlayers=8?Lighting=Day`
		parser.ParseAndProcess(ctx, mapLoad, serverExternalID, "test.log")

		parser.ParseAndProcess(ctx, `[2025.11.08-14.00.00:000][  1]LogGameplayEvents: Display: Player1[76561198000000001, team 0] killed Enemy1[INVALID, team 1] with BP_Firearm_PF940_C_123`, serverExternalID, "test.log")
		parser.ParseAndProcess(ctx, `[2025.11.08-14.00.01:000][  2]LogGameplayEvents: Display: Player1[76561198000000001, team 0] killed Enemy2[INVALID, team 1] with BP_Firearm_AKM_C_456`, serverExternalID, "test.log")
		parser.ParseAndProcess(ctx, `[2025.11.08-14.00.02:000][  3]LogGameplayEvents: Display: Player1[76561198000000001, team 0] killed Enemy3[INVALID, team 1] with BP_Firearm_M16A2_C_789`, serverExternalID, "test.log")

		events, err := testApp.FindRecordsByFilter("events", "type = 'player_kill'", "-created", 10, 0)
		if err != nil {
			t.Fatalf("failed to query events: %v", err)
		}

		if len(events) < 3 {
			t.Errorf("Expected 3 kill events, got %d", len(events))
		}

		weaponsSeen := make(map[string]int)
		for _, event := range events {
			data := event.GetString("data")
			if strings.Contains(data, "PF940") {
				weaponsSeen["PF940"]++
			}
			if strings.Contains(data, "AKM") {
				weaponsSeen["AKM"]++
			}
			if strings.Contains(data, "M16A2") {
				weaponsSeen["M16A2"]++
			}
		}

		if len(weaponsSeen) != 3 {
			t.Errorf("Expected 3 different weapons, got %d", len(weaponsSeen))
		}
	})
}
