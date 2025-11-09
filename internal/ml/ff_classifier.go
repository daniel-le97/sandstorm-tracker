package ml

import (
	"context"
	"fmt"
	"math"

	"github.com/pocketbase/pocketbase/core"
)

// FFClassifier implements a simple decision tree-based classifier for friendly fire incidents
type FFClassifier struct {
	// Thresholds learned from training data
	FFRateThreshold        float64
	ExplosiveRateThreshold float64
	AvgTimeBetweenFF       float64
	SpawnKillThreshold     int
}

// Features represents the feature vector for classification
type Features struct {
	FFRate               float64 // Friendly fire kills / total kills
	ExplosiveFFPercent   float64 // Percentage of FF with explosives
	VehicleFFPercent     float64 // Percentage of FF with vehicles
	AvgTimeBetweenFF     float64 // Average seconds between FF incidents
	FFInFirstMinute      int     // Number of FF in first minute of matches
	TotalFFKills         int     // Total FF kills
	RapidFFCount         int     // Number of FF < 30 seconds apart
	ConsecutiveFFMatches int     // Matches with multiple FF
}

// Prediction represents the classification result
type Prediction struct {
	Classification string  // likely_accident, possibly_intentional, likely_intentional
	Confidence     float64 // 0.0 to 1.0
	Reasoning      []string
}

// NewDefaultClassifier creates a classifier with default thresholds
func NewDefaultClassifier() *FFClassifier {
	return &FFClassifier{
		FFRateThreshold:        0.08, // 8% FF rate threshold
		ExplosiveRateThreshold: 0.60, // 60% explosive weapons
		AvgTimeBetweenFF:       90.0, // 90 seconds between incidents
		SpawnKillThreshold:     2,    // 2 kills in spawn area
	}
}

// ExtractFeatures extracts features from database for a player
func (c *FFClassifier) ExtractFeatures(ctx context.Context, app core.App, playerID string) (*Features, error) {
	type featureRow struct {
		TotalFF            int     `db:"total_ff"`
		TotalKills         int     `db:"total_kills"`
		ExplosiveFFCount   int     `db:"explosive_ff_count"`
		VehicleFFCount     int     `db:"vehicle_ff_count"`
		AvgTimeBetweenFF   float64 `db:"avg_time_between_ff"`
		FFInFirstMinute    int     `db:"ff_in_first_minute"`
		RapidFFCount       int     `db:"rapid_ff_count"`
		ConsecutiveFFCount int     `db:"consecutive_ff_matches"`
	}

	var row featureRow

	err := app.DB().
		NewQuery(`
			WITH player_ff_data AS (
				SELECT 
					COUNT(*) as total_ff,
					SUM(CASE WHEN is_explosive_weapon = 1 THEN 1 ELSE 0 END) as explosive_ff_count,
					SUM(CASE WHEN is_vehicle_weapon = 1 THEN 1 ELSE 0 END) as vehicle_ff_count,
					AVG(COALESCE(time_since_last_ff_seconds, 0)) as avg_time_between_ff,
					SUM(CASE WHEN time_since_match_start_seconds < 60 THEN 1 ELSE 0 END) as ff_in_first_minute,
					SUM(CASE WHEN time_since_last_ff_seconds < 30 THEN 1 ELSE 0 END) as rapid_ff_count
				FROM friendly_fire_incidents
				WHERE killer = {:player}
			),
			player_kills AS (
				SELECT COALESCE(SUM(kills), 0) as total_kills
				FROM match_player_stats
				WHERE player = {:player}
			),
			consecutive_matches AS (
				SELECT COUNT(DISTINCT match) as consecutive_ff_matches
				FROM (
					SELECT match, COUNT(*) as ff_count
					FROM friendly_fire_incidents
					WHERE killer = {:player}
					GROUP BY match
					HAVING ff_count > 1
				)
			)
			SELECT 
				pfd.total_ff,
				pk.total_kills,
				pfd.explosive_ff_count,
				pfd.vehicle_ff_count,
				pfd.avg_time_between_ff,
				pfd.ff_in_first_minute,
				pfd.rapid_ff_count,
				cm.consecutive_ff_matches
			FROM player_ff_data pfd, player_kills pk, consecutive_matches cm
		`).
		Bind(map[string]any{"player": playerID}).
		One(&row)

	if err != nil {
		return nil, err
	}

	features := &Features{
		TotalFFKills:         row.TotalFF,
		AvgTimeBetweenFF:     row.AvgTimeBetweenFF,
		FFInFirstMinute:      row.FFInFirstMinute,
		RapidFFCount:         row.RapidFFCount,
		ConsecutiveFFMatches: row.ConsecutiveFFCount,
	}

	// Calculate percentages
	if row.TotalKills > 0 {
		features.FFRate = float64(row.TotalFF) / float64(row.TotalKills)
	}
	if row.TotalFF > 0 {
		features.ExplosiveFFPercent = float64(row.ExplosiveFFCount) / float64(row.TotalFF)
		features.VehicleFFPercent = float64(row.VehicleFFCount) / float64(row.TotalFF)
	}

	return features, nil
}

// Predict classifies friendly fire behavior based on features
func (c *FFClassifier) Predict(features *Features) *Prediction {
	if features.TotalFFKills == 0 {
		return &Prediction{
			Classification: "unclassified",
			Confidence:     0.0,
			Reasoning:      []string{"No friendly fire incidents recorded"},
		}
	}

	reasoning := []string{}
	accidentScore := 0.0
	intentionalScore := 0.0

	// Feature 1: FF Rate (most important)
	if features.FFRate < 0.05 {
		accidentScore += 3.0
		reasoning = append(reasoning, fmt.Sprintf("Very low FF rate (%.1f%%)", features.FFRate*100))
	} else if features.FFRate < c.FFRateThreshold {
		accidentScore += 2.0
		reasoning = append(reasoning, fmt.Sprintf("Low FF rate (%.1f%%)", features.FFRate*100))
	} else if features.FFRate > 0.15 {
		intentionalScore += 3.0
		reasoning = append(reasoning, fmt.Sprintf("High FF rate (%.1f%%)", features.FFRate*100))
	} else if features.FFRate > 0.10 {
		intentionalScore += 2.0
		reasoning = append(reasoning, fmt.Sprintf("Elevated FF rate (%.1f%%)", features.FFRate*100))
	} else {
		// Middle ground - assign small scores to both
		accidentScore += 1.0
		intentionalScore += 1.0
		reasoning = append(reasoning, fmt.Sprintf("Moderate FF rate (%.1f%%)", features.FFRate*100))
	}

	// Feature 2: Explosive/Vehicle weapon usage
	accidentWeaponRate := features.ExplosiveFFPercent + features.VehicleFFPercent
	if accidentWeaponRate > 0.70 {
		accidentScore += 2.5
		reasoning = append(reasoning, fmt.Sprintf("High accident-prone weapon usage (%.0f%%)", accidentWeaponRate*100))
	} else if accidentWeaponRate < 0.30 {
		intentionalScore += 2.0
		reasoning = append(reasoning, fmt.Sprintf("Mostly direct-fire weapons (%.0f%% explosives/vehicles)", accidentWeaponRate*100))
	} else {
		// Mixed weapon usage
		accidentScore += 0.5
		intentionalScore += 0.5
		reasoning = append(reasoning, fmt.Sprintf("Mixed weapon usage (%.0f%% accident-prone)", accidentWeaponRate*100))
	}

	// Feature 3: Time between FF incidents
	if features.AvgTimeBetweenFF > 180 {
		accidentScore += 2.0
		reasoning = append(reasoning, fmt.Sprintf("Long time between incidents (avg %.0fs)", features.AvgTimeBetweenFF))
	} else if features.AvgTimeBetweenFF < 45 {
		intentionalScore += 2.5
		reasoning = append(reasoning, fmt.Sprintf("Rapid succession FF (avg %.0fs)", features.AvgTimeBetweenFF))
	} else if features.AvgTimeBetweenFF > 0 {
		// Moderate timing
		accidentScore += 0.5
		intentionalScore += 0.5
		reasoning = append(reasoning, fmt.Sprintf("Moderate time between incidents (avg %.0fs)", features.AvgTimeBetweenFF))
	}

	// Feature 4: Spawn area kills
	if features.FFInFirstMinute > c.SpawnKillThreshold {
		intentionalScore += 2.5
		reasoning = append(reasoning, fmt.Sprintf("%d kills in spawn areas (first minute)", features.FFInFirstMinute))
	} else if features.FFInFirstMinute == 0 {
		accidentScore += 1.0
		reasoning = append(reasoning, "No spawn area kills")
	}

	// Feature 5: Rapid FF (< 30 seconds)
	rapidFFRate := float64(features.RapidFFCount) / float64(features.TotalFFKills)
	if rapidFFRate > 0.30 {
		intentionalScore += 2.0
		reasoning = append(reasoning, fmt.Sprintf("%.0f%% of FF occurred rapidly (<30s apart)", rapidFFRate*100))
	}

	// Feature 6: Multiple FF per match consistency
	if features.ConsecutiveFFMatches > 3 && features.TotalFFKills > 10 {
		intentionalScore += 1.5
		reasoning = append(reasoning, fmt.Sprintf("Multiple FF in %d different matches", features.ConsecutiveFFMatches))
	}

	// Calculate final classification
	totalScore := accidentScore + intentionalScore
	if totalScore == 0 {
		return &Prediction{
			Classification: "unclassified",
			Confidence:     0.0,
			Reasoning:      []string{"Insufficient data for classification"},
		}
	}

	accidentProbability := accidentScore / totalScore
	intentionalProbability := intentionalScore / totalScore

	// Determine classification
	var classification string
	var confidence float64

	if accidentProbability > 0.70 {
		classification = "likely_accident"
		confidence = normalize(accidentProbability, 0.70, 0.95)
	} else if accidentProbability > 0.55 {
		classification = "likely_accident"
		confidence = normalize(accidentProbability, 0.55, 0.70)
	} else if intentionalProbability > 0.70 {
		classification = "likely_intentional"
		confidence = normalize(intentionalProbability, 0.70, 0.95)
	} else if intentionalProbability > 0.55 {
		classification = "possibly_intentional"
		confidence = normalize(intentionalProbability, 0.55, 0.70)
	} else {
		classification = "possibly_intentional"
		confidence = 0.50
	}

	return &Prediction{
		Classification: classification,
		Confidence:     confidence,
		Reasoning:      reasoning,
	}
}

// normalize scales a value from one range to another
func normalize(value, minIn, maxIn float64) float64 {
	// Scale from [minIn, maxIn] to [0.5, 1.0]
	if value < minIn {
		return 0.5
	}
	if value > maxIn {
		return 1.0
	}
	return 0.5 + (value-minIn)/(maxIn-minIn)*0.5
}

// ClassifyPlayer classifies a player's friendly fire behavior
func (c *FFClassifier) ClassifyPlayer(ctx context.Context, app core.App, playerID string) (*Prediction, error) {
	features, err := c.ExtractFeatures(ctx, app, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to extract features: %w", err)
	}

	prediction := c.Predict(features)
	return prediction, nil
}

// BatchClassifyPlayers classifies multiple players and returns sorted by risk
func (c *FFClassifier) BatchClassifyPlayers(ctx context.Context, app core.App, playerIDs []string) ([]*PlayerRisk, error) {
	risks := make([]*PlayerRisk, 0, len(playerIDs))

	for _, playerID := range playerIDs {
		prediction, err := c.ClassifyPlayer(ctx, app, playerID)
		if err != nil {
			continue // Skip players with errors
		}

		if prediction.Classification != "unclassified" {
			risks = append(risks, &PlayerRisk{
				PlayerID:       playerID,
				Classification: prediction.Classification,
				Confidence:     prediction.Confidence,
				RiskScore:      calculateRiskScore(prediction),
			})
		}
	}

	// Sort by risk score (highest first)
	sortByRisk(risks)

	return risks, nil
}

// PlayerRisk represents a player's risk assessment
type PlayerRisk struct {
	PlayerID       string
	Classification string
	Confidence     float64
	RiskScore      float64 // 0-100
}

// calculateRiskScore converts classification and confidence to a 0-100 score
func calculateRiskScore(pred *Prediction) float64 {
	baseScore := 0.0
	switch pred.Classification {
	case "likely_intentional":
		baseScore = 80.0
	case "possibly_intentional":
		baseScore = 50.0
	case "likely_accident":
		baseScore = 20.0
	default:
		baseScore = 0.0
	}

	// Adjust by confidence
	return baseScore * pred.Confidence
}

// sortByRisk sorts player risks by risk score descending
func sortByRisk(risks []*PlayerRisk) {
	for i := 0; i < len(risks); i++ {
		for j := i + 1; j < len(risks); j++ {
			if risks[j].RiskScore > risks[i].RiskScore {
				risks[i], risks[j] = risks[j], risks[i]
			}
		}
	}
}

// TrainFromData allows updating thresholds based on labeled data (simple version)
func (c *FFClassifier) TrainFromData(ctx context.Context, app core.App) error {
	// Query labeled data
	type trainingRow struct {
		FFRate           float64 `db:"ff_rate"`
		ExplosivePercent float64 `db:"explosive_percent"`
		AvgTimeBetween   float64 `db:"avg_time_between"`
		Classification   string  `db:"classification"`
	}

	var rows []trainingRow

	err := app.DB().
		NewQuery(`
			SELECT 
				(SELECT CAST(COUNT(*) AS REAL) / NULLIF(SUM(kills), 0)
				 FROM match_player_stats 
				 WHERE player = ff.killer) as ff_rate,
				AVG(CASE WHEN is_explosive_weapon = 1 THEN 1.0 ELSE 0.0 END) as explosive_percent,
				AVG(time_since_last_ff_seconds) as avg_time_between,
				accident_classification as classification
			FROM friendly_fire_incidents ff
			WHERE accident_classification IS NOT NULL
			  AND accident_classification != 'unclassified'
			GROUP BY killer, accident_classification
			HAVING COUNT(*) >= 5
		`).
		All(&rows)

	if err != nil || len(rows) == 0 {
		// Not enough training data, use defaults
		return nil
	}

	// Simple threshold learning: find median values that separate classes
	accidentRates := []float64{}
	intentionalRates := []float64{}

	for _, row := range rows {
		if row.Classification == "likely_accident" {
			accidentRates = append(accidentRates, row.FFRate)
		} else if row.Classification == "likely_intentional" {
			intentionalRates = append(intentionalRates, row.FFRate)
		}
	}

	if len(accidentRates) > 0 && len(intentionalRates) > 0 {
		// Set threshold as midpoint between median accident and median intentional rates
		accidentMedian := median(accidentRates)
		intentionalMedian := median(intentionalRates)
		c.FFRateThreshold = (accidentMedian + intentionalMedian) / 2.0
	}

	return nil
}

// median calculates the median of a slice
func median(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	// Simple bubble sort for small datasets
	sorted := make([]float64, len(values))
	copy(sorted, values)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j] < sorted[i] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2.0
	}
	return sorted[mid]
}

// ExportTrainingData exports features and labels for external ML training
func (c *FFClassifier) ExportTrainingData(ctx context.Context, app core.App) ([]map[string]interface{}, error) {
	type exportRow struct {
		PlayerID         string  `db:"player_id"`
		FFRate           float64 `db:"ff_rate"`
		ExplosivePercent float64 `db:"explosive_percent"`
		VehiclePercent   float64 `db:"vehicle_percent"`
		AvgTimeBetween   float64 `db:"avg_time_between"`
		FFInFirstMinute  int     `db:"ff_in_first_minute"`
		RapidFFCount     int     `db:"rapid_ff_count"`
		TotalFF          int     `db:"total_ff"`
		Classification   string  `db:"classification"`
	}

	var rows []exportRow

	err := app.DB().
		NewQuery(`
			SELECT 
				ff.killer as player_id,
				(SELECT CAST(COUNT(*) AS REAL) / NULLIF(SUM(kills), 0)
				 FROM match_player_stats 
				 WHERE player = ff.killer) as ff_rate,
				AVG(CASE WHEN is_explosive_weapon = 1 THEN 1.0 ELSE 0.0 END) as explosive_percent,
				AVG(CASE WHEN is_vehicle_weapon = 1 THEN 1.0 ELSE 0.0 END) as vehicle_percent,
				AVG(time_since_last_ff_seconds) as avg_time_between,
				SUM(CASE WHEN time_since_match_start_seconds < 60 THEN 1 ELSE 0 END) as ff_in_first_minute,
				SUM(CASE WHEN time_since_last_ff_seconds < 30 THEN 1 ELSE 0 END) as rapid_ff_count,
				COUNT(*) as total_ff,
				MAX(accident_classification) as classification
			FROM friendly_fire_incidents ff
			WHERE accident_classification IS NOT NULL
			GROUP BY killer
			HAVING total_ff >= 5
		`).
		All(&rows)

	if err != nil {
		return nil, err
	}

	// Convert to generic format for export
	data := make([]map[string]interface{}, len(rows))
	for i, row := range rows {
		data[i] = map[string]interface{}{
			"player_id":          row.PlayerID,
			"ff_rate":            row.FFRate,
			"explosive_percent":  row.ExplosivePercent,
			"vehicle_percent":    row.VehiclePercent,
			"avg_time_between":   row.AvgTimeBetween,
			"ff_in_first_minute": row.FFInFirstMinute,
			"rapid_ff_count":     row.RapidFFCount,
			"total_ff":           row.TotalFF,
			"classification":     row.Classification,
		}
	}

	return data, nil
}

// CalculateFeatureImportance returns which features are most predictive
func (c *FFClassifier) CalculateFeatureImportance(ctx context.Context, app core.App) (map[string]float64, error) {
	// Simple correlation-based importance for labeled data
	data, err := c.ExportTrainingData(ctx, app)
	if err != nil {
		return nil, err
	}

	if len(data) < 10 {
		return nil, fmt.Errorf("insufficient training data (need at least 10 samples)")
	}

	importance := map[string]float64{
		"ff_rate":            calculateCorrelation(data, "ff_rate"),
		"explosive_percent":  calculateCorrelation(data, "explosive_percent"),
		"avg_time_between":   calculateCorrelation(data, "avg_time_between"),
		"ff_in_first_minute": calculateCorrelation(data, "ff_in_first_minute"),
		"rapid_ff_count":     calculateCorrelation(data, "rapid_ff_count"),
	}

	return importance, nil
}

// calculateCorrelation calculates simple correlation between feature and classification
func calculateCorrelation(data []map[string]interface{}, feature string) float64 {
	// Convert classification to numeric: accident=0, intentional=1
	var xSum, ySum, xySum, x2Sum, y2Sum float64
	n := float64(len(data))

	for _, row := range data {
		x := toFloat(row[feature])
		y := 0.0
		if row["classification"] == "likely_intentional" {
			y = 1.0
		} else if row["classification"] == "possibly_intentional" {
			y = 0.5
		}

		xSum += x
		ySum += y
		xySum += x * y
		x2Sum += x * x
		y2Sum += y * y
	}

	numerator := n*xySum - xSum*ySum
	denominator := math.Sqrt((n*x2Sum - xSum*xSum) * (n*y2Sum - ySum*ySum))

	if denominator == 0 {
		return 0
	}

	return math.Abs(numerator / denominator) // Return absolute correlation
}

// toFloat converts interface{} to float64
func toFloat(val interface{}) float64 {
	switch v := val.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	default:
		return 0
	}
}
