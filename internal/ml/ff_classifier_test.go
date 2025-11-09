package ml

import (
	"context"
	"testing"
	"time"

	"sandstorm-tracker/internal/database"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
)

func TestFFClassifier_Predict(t *testing.T) {
	classifier := NewDefaultClassifier()

	tests := []struct {
		name              string
		features          *Features
		wantClass         string
		wantConfidenceMin float64
	}{
		{
			name: "Accidental - Explosives with low rate",
			features: &Features{
				FFRate:             0.03,
				ExplosiveFFPercent: 0.80,
				VehicleFFPercent:   0.10,
				AvgTimeBetweenFF:   200.0,
				FFInFirstMinute:    0,
				TotalFFKills:       10,
				RapidFFCount:       1,
			},
			wantClass:         "likely_accident",
			wantConfidenceMin: 0.70,
		},
		{
			name: "Intentional - High rate with direct fire",
			features: &Features{
				FFRate:             0.18,
				ExplosiveFFPercent: 0.20,
				VehicleFFPercent:   0.05,
				AvgTimeBetweenFF:   40.0,
				FFInFirstMinute:    5,
				TotalFFKills:       25,
				RapidFFCount:       10,
			},
			wantClass:         "likely_intentional",
			wantConfidenceMin: 0.70,
		},
		{
			name: "Possibly intentional - Mixed signals",
			features: &Features{
				FFRate:             0.10,
				ExplosiveFFPercent: 0.50,
				VehicleFFPercent:   0.10,
				AvgTimeBetweenFF:   90.0,
				FFInFirstMinute:    2,
				TotalFFKills:       15,
				RapidFFCount:       4,
			},
			wantClass:         "possibly_intentional",
			wantConfidenceMin: 0.50,
		},
		{
			name: "No FF kills",
			features: &Features{
				TotalFFKills: 0,
			},
			wantClass:         "unclassified",
			wantConfidenceMin: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pred := classifier.Predict(tt.features)

			if pred.Classification != tt.wantClass {
				t.Errorf("Predict() classification = %v, want %v\nReasoning: %v",
					pred.Classification, tt.wantClass, pred.Reasoning)
			}

			if pred.Confidence < tt.wantConfidenceMin {
				t.Errorf("Predict() confidence = %.2f, want >= %.2f",
					pred.Confidence, tt.wantConfidenceMin)
			}

			if len(pred.Reasoning) == 0 && tt.wantClass != "unclassified" {
				t.Error("Predict() reasoning is empty")
			}
		})
	}
}

func TestFFClassifier_ExtractFeatures(t *testing.T) {
	app, ctx, _, match := setupMLTest(t)

	// Check if friendly_fire_incidents collection exists
	_, err := app.FindCollectionByNameOrId("friendly_fire_incidents")
	if err != nil {
		t.Skip("Skipping test: friendly_fire_incidents collection not found (migration not run)")
	}

	classifier := NewDefaultClassifier()

	// Create test data
	player := createMLTestPlayer(t, ctx, app, "76561198000000001", "TestPlayer")

	// Add player stats
	createMatchPlayerStats(t, ctx, app, match, player, 100, 3) // 100 kills, 3 FF = 3% rate

	// Add FF incidents
	createFFIncident(t, ctx, app, match, player, player, "M67 Frag", true, false, 200.0, 0)
	createFFIncident(t, ctx, app, match, player, player, "M67 Frag", true, false, 250.0, 50.0)
	createFFIncident(t, ctx, app, match, player, player, "M4A1", false, false, 500.0, 250.0)

	// Extract features
	features, err := classifier.ExtractFeatures(ctx, app, player.ID)
	if err != nil {
		t.Fatalf("ExtractFeatures() error = %v", err)
	}

	// Validate features
	if features.TotalFFKills != 3 {
		t.Errorf("TotalFFKills = %d, want 3", features.TotalFFKills)
	}

	expectedRate := 3.0 / 100.0
	if features.FFRate < expectedRate-0.01 || features.FFRate > expectedRate+0.01 {
		t.Errorf("FFRate = %.3f, want %.3f", features.FFRate, expectedRate)
	}

	expectedExplosive := 2.0 / 3.0
	if features.ExplosiveFFPercent < expectedExplosive-0.01 || features.ExplosiveFFPercent > expectedExplosive+0.01 {
		t.Errorf("ExplosiveFFPercent = %.2f, want %.2f", features.ExplosiveFFPercent, expectedExplosive)
	}
}

func TestFFClassifier_ClassifyPlayer(t *testing.T) {
	app, ctx, _, match := setupMLTest(t)

	// Check if friendly_fire_incidents collection exists
	_, err := app.FindCollectionByNameOrId("friendly_fire_incidents")
	if err != nil {
		t.Skip("Skipping test: friendly_fire_incidents collection not found (migration not run)")
	}

	classifier := NewDefaultClassifier()

	// Create accidental player (low rate, explosives)
	accidentalPlayer := createMLTestPlayer(t, ctx, app, "76561198000000001", "AccidentalPlayer")
	createMatchPlayerStats(t, ctx, app, match, accidentalPlayer, 200, 5) // 2.5% FF rate
	for i := 0; i < 5; i++ {
		createFFIncident(t, ctx, app, match, accidentalPlayer, accidentalPlayer, "M67 Frag", true, false, float64(300+i*180), float64(i*180))
	}

	pred, err := classifier.ClassifyPlayer(ctx, app, accidentalPlayer.ID)
	if err != nil {
		t.Fatalf("ClassifyPlayer() error = %v", err)
	}

	if pred.Classification != "likely_accident" {
		t.Errorf("Classification = %v, want likely_accident\nReasoning: %v", pred.Classification, pred.Reasoning)
	}
}

func TestFFClassifier_BatchClassifyPlayers(t *testing.T) {
	app, ctx, _, match := setupMLTest(t)

	// Check if friendly_fire_incidents collection exists
	_, err := app.FindCollectionByNameOrId("friendly_fire_incidents")
	if err != nil {
		t.Skip("Skipping test: friendly_fire_incidents collection not found (migration not run)")
	}

	classifier := NewDefaultClassifier()

	// Create 3 players with different risk levels
	players := []struct {
		id        string
		name      string
		kills     int
		ff        int
		explosive bool
	}{
		{"76561198000000001", "SafePlayer", 300, 4, true},         // 1.3% FF, explosives
		{"76561198000000002", "SuspiciousPlayer", 100, 15, false}, // 15% FF, direct fire
		{"76561198000000003", "AveragePlayer", 150, 8, true},      // 5.3% FF, mixed
	}

	playerIDs := make([]string, len(players))
	for i, p := range players {
		player := createMLTestPlayer(t, ctx, app, p.id, p.name)
		playerIDs[i] = player.ID
		createMatchPlayerStats(t, ctx, app, match, player, p.kills, p.ff)

		for j := 0; j < p.ff; j++ {
			weapon := "M4A1"
			if p.explosive {
				weapon = "M67 Frag"
			}
			createFFIncident(t, ctx, app, match, player, player, weapon, p.explosive, false, float64(100+j*60), float64(j*60))
		}
	}

	risks, err := classifier.BatchClassifyPlayers(ctx, app, playerIDs)
	if err != nil {
		t.Fatalf("BatchClassifyPlayers() error = %v", err)
	}

	if len(risks) != 3 {
		t.Errorf("BatchClassifyPlayers() returned %d results, want 3", len(risks))
	}

	// Highest risk should be first
	if risks[0].Classification != "likely_intentional" {
		t.Errorf("First risk classification = %v, want likely_intentional", risks[0].Classification)
	}
}

func TestCalculateRiskScore(t *testing.T) {
	tests := []struct {
		name         string
		prediction   *Prediction
		wantScoreMin float64
		wantScoreMax float64
	}{
		{
			name: "High risk intentional",
			prediction: &Prediction{
				Classification: "likely_intentional",
				Confidence:     0.90,
			},
			wantScoreMin: 70.0,
			wantScoreMax: 80.0,
		},
		{
			name: "Low risk accident",
			prediction: &Prediction{
				Classification: "likely_accident",
				Confidence:     0.80,
			},
			wantScoreMin: 15.0,
			wantScoreMax: 17.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculateRiskScore(tt.prediction)
			if score < tt.wantScoreMin || score > tt.wantScoreMax {
				t.Errorf("calculateRiskScore() = %.1f, want between %.1f and %.1f",
					score, tt.wantScoreMin, tt.wantScoreMax)
			}
		})
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		value float64
		minIn float64
		maxIn float64
		want  float64
	}{
		{0.70, 0.70, 0.95, 0.50},
		{0.95, 0.70, 0.95, 1.00},
		{0.825, 0.70, 0.95, 0.75},
		{0.60, 0.70, 0.95, 0.50},
		{1.00, 0.70, 0.95, 1.00},
	}

	for _, tt := range tests {
		got := normalize(tt.value, tt.minIn, tt.maxIn)
		if got < tt.want-0.01 || got > tt.want+0.01 {
			t.Errorf("normalize(%.2f, %.2f, %.2f) = %.2f, want %.2f",
				tt.value, tt.minIn, tt.maxIn, got, tt.want)
		}
	}
}

func TestMedian(t *testing.T) {
	tests := []struct {
		name   string
		values []float64
		want   float64
	}{
		{"Empty", []float64{}, 0.0},
		{"Single", []float64{5.0}, 5.0},
		{"Odd count", []float64{1.0, 3.0, 5.0, 7.0, 9.0}, 5.0},
		{"Even count", []float64{2.0, 4.0, 6.0, 8.0}, 5.0},
		{"Unsorted", []float64{9.0, 1.0, 5.0, 3.0, 7.0}, 5.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := median(tt.values)
			if got != tt.want {
				t.Errorf("median() = %.1f, want %.1f", got, tt.want)
			}
		})
	}
}

// Test helpers

func setupMLTest(t *testing.T) (core.App, context.Context, string, *database.Match) {
	t.Helper()

	testApp, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(testApp.Cleanup)

	ctx := context.Background()

	// Check if required collections exist (migrations must be run)
	if _, err := testApp.FindCollectionByNameOrId("servers"); err != nil {
		t.Skip("Skipping test: database collections not found (migrations not run)")
	}

	serverExternalID := "test-server"

	_, err = database.GetOrCreateServer(ctx, testApp, serverExternalID, "Test Server", "/test/path")
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	mapName := "Ministry"
	mode := "Push"
	startTime := time.Now()
	match, err := database.CreateMatch(ctx, testApp, serverExternalID, &mapName, &mode, &startTime)
	if err != nil {
		t.Fatalf("Failed to create match: %v", err)
	}

	return testApp, ctx, serverExternalID, match
}

func createMLTestPlayer(t *testing.T, ctx context.Context, app core.App, externalID, name string) *database.Player {
	t.Helper()

	player, err := database.CreatePlayer(ctx, app, externalID, name)
	if err != nil {
		t.Fatalf("Failed to create player: %v", err)
	}

	return player
}

func createMatchPlayerStats(t *testing.T, ctx context.Context, app core.App, match *database.Match, player *database.Player, kills, deaths int) {
	t.Helper()

	joinTime := time.Now()
	err := database.UpsertMatchPlayerStats(ctx, app, match.ID, player.ID, nil, &joinTime)
	if err != nil {
		t.Fatalf("Failed to create match player stats: %v", err)
	}

	// Update kills and deaths using updatePlayerStats pattern
	collection, err := app.FindCollectionByNameOrId("match_player_stats")
	if err != nil {
		t.Fatalf("Failed to find collection: %v", err)
	}

	record, err := app.FindFirstRecordByFilter(
		collection.Id,
		"match = {:match} && player = {:player}",
		map[string]any{"match": match.ID, "player": player.ID},
	)
	if err != nil {
		t.Fatalf("Failed to find match player stats: %v", err)
	}

	record.Set("kills", kills)
	record.Set("deaths", deaths)
	record.Set("score", kills*10)

	if err := app.Save(record); err != nil {
		t.Fatalf("Failed to update match player stats: %v", err)
	}
}

func createFFIncident(t *testing.T, ctx context.Context, app core.App, match *database.Match, killer, victim *database.Player, weapon string, explosive, vehicle bool, timeInMatch, timeSinceLast float64) {
	t.Helper()

	collection, err := app.FindCollectionByNameOrId("friendly_fire_incidents")
	if err != nil {
		t.Fatalf("Failed to find friendly_fire_incidents collection: %v", err)
	}

	record := core.NewRecord(collection)
	record.Set("match", match.ID)
	record.Set("killer", killer.ID)
	record.Set("victim", victim.ID)
	record.Set("weapon", weapon)
	record.Set("is_explosive_weapon", explosive)
	record.Set("is_vehicle_weapon", vehicle)
	record.Set("time_since_match_start_seconds", timeInMatch)
	if timeSinceLast > 0 {
		record.Set("time_since_last_ff_seconds", timeSinceLast)
	}

	if err := app.Save(record); err != nil {
		t.Fatalf("Failed to create FF incident: %v", err)
	}
}
