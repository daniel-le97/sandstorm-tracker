# Friendly Fire ML Classifier

A native Go machine learning classifier for analyzing friendly fire incidents in Insurgency: Sandstorm. This classifier determines whether teamkills are likely accidental or intentional based on player behavior patterns.

## Overview

The classifier uses a **weighted decision tree algorithm** that analyzes multiple features:

- **FF Rate**: Percentage of friendly fire kills vs total kills
- **Weapon Types**: Usage of explosive/vehicle weapons (accident-prone)
- **Timing Patterns**: Time between incidents and spawn area kills
- **Behavioral Consistency**: Rapid succession kills and repeat offenders

## Quick Start

```go
import "sandstorm-tracker/internal/ml"

// Create classifier with default thresholds
classifier := ml.NewDefaultClassifier()

// Classify a single player
ctx := context.Background()
prediction, err := classifier.ClassifyPlayer(ctx, app, playerID)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Classification: %s (%.0f%% confidence)\n",
    prediction.Classification, prediction.Confidence*100)
fmt.Printf("Reasoning:\n")
for _, reason := range prediction.Reasoning {
    fmt.Printf("  - %s\n", reason)
}
```

## Classification Categories

### likely_accident (confidence 70-100%)

- **Characteristics**: Very low FF rate (<5%), high explosive weapon usage (>70%), long time between incidents (>3 minutes)
- **Action**: No moderation needed
- **Example**: Player has 3% FF rate, 85% from grenades, avg 4 minutes between incidents

### possibly_intentional (confidence 50-70%)

- **Characteristics**: Moderate FF rate (8-10%), mixed weapon usage, inconsistent patterns
- **Action**: Monitor for escalation
- **Example**: Player has 9% FF rate, 50% explosives, some rapid succession kills

### likely_intentional (confidence 70-100%)

- **Characteristics**: High FF rate (>15%), direct-fire weapons, rapid succession (<30s), spawn area kills
- **Action**: Manual review recommended, possible ban
- **Example**: Player has 18% FF rate, 80% from rifles, multiple kills <30s apart, 5 spawn kills

### unclassified (confidence 0%)

- **Characteristics**: Insufficient data (no FF incidents or not enough features)
- **Action**: Wait for more data

## Feature Extraction

The classifier extracts features from your PocketBase database:

```go
features, err := classifier.ExtractFeatures(ctx, app, playerID)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("FF Rate: %.1f%%\n", features.FFRate*100)
fmt.Printf("Explosive FF: %.0f%%\n", features.ExplosiveFFPercent*100)
fmt.Printf("Avg Time Between FF: %.0f seconds\n", features.AvgTimeBetweenFF)
```

**Features Struct:**

```go
type Features struct {
    FFRate              float64 // Friendly fire kills / total kills
    ExplosiveFFPercent  float64 // Percentage of FF with explosives
    VehicleFFPercent    float64 // Percentage of FF with vehicles
    AvgTimeBetweenFF    float64 // Average seconds between FF incidents
    FFInFirstMinute     int     // Number of FF in first minute of matches
    TotalFFKills        int     // Total FF kills
    RapidFFCount        int     // Number of FF < 30 seconds apart
    ConsecutiveFFMatches int    // Matches with multiple FF
}
```

## Batch Processing

Classify multiple players and sort by risk:

```go
playerIDs := []string{"player1", "player2", "player3"}
risks, err := classifier.BatchClassifyPlayers(ctx, app, playerIDs)
if err != nil {
    log.Fatal(err)
}

for i, risk := range risks {
    fmt.Printf("%d. Player %s: %s (Risk Score: %.1f)\n",
        i+1, risk.PlayerID, risk.Classification, risk.RiskScore)
}
```

**Risk Scoring:**

- `likely_intentional` × confidence = 80 × 0.9 = **72.0 risk score**
- `possibly_intentional` × confidence = 50 × 0.6 = **30.0 risk score**
- `likely_accident` × confidence = 20 × 0.8 = **16.0 risk score**

## Custom Thresholds

Create a classifier with custom thresholds:

```go
classifier := &ml.FFClassifier{
    FFRateThreshold:        0.10,  // 10% FF rate threshold
    ExplosiveRateThreshold: 0.65,  // 65% explosive weapons
    AvgTimeBetweenFF:       120.0, // 2 minutes between incidents
    SpawnKillThreshold:     3,     // 3 kills in spawn area
}
```

## Training from Labeled Data

The classifier can learn from manually labeled incidents:

```go
// Admin manually labels incidents in the database
// UPDATE friendly_fire_incidents
// SET accident_classification = 'likely_intentional', confidence_score = 0.95
// WHERE id = 'incident_id'

// Train classifier from labeled data
err := classifier.TrainFromData(ctx, app)
if err != nil {
    log.Fatal(err)
}
// Thresholds are now updated based on labeled data
```

## Feature Importance

Analyze which features are most predictive:

```go
importance, err := classifier.CalculateFeatureImportance(ctx, app)
if err != nil {
    log.Fatal(err)
}

for feature, score := range importance {
    fmt.Printf("%s: %.3f correlation\n", feature, score)
}
```

**Example Output:**

```
ff_rate: 0.852 correlation (most important)
avg_time_between: 0.673 correlation
explosive_percent: 0.421 correlation
ff_in_first_minute: 0.389 correlation
rapid_ff_count: 0.312 correlation
```

## Export Training Data

Export features and labels for external ML training:

```go
data, err := classifier.ExportTrainingData(ctx, app)
if err != nil {
    log.Fatal(err)
}

// Save to JSON for Python/R analysis
jsonData, _ := json.MarshalIndent(data, "", "  ")
os.WriteFile("ff_training_data.json", jsonData, 0644)
```

## Integration with Chat Commands

```go
// In your RCON command handler
func handleFFCheck(app core.App, playerName string) string {
    player, err := database.GetPlayerByName(ctx, app, playerName)
    if err != nil {
        return "Player not found"
    }

    classifier := ml.NewDefaultClassifier()
    pred, err := classifier.ClassifyPlayer(ctx, app, player.ID)
    if err != nil {
        return "Classification failed"
    }

    response := fmt.Sprintf("%s FF Analysis: %s (%.0f%% confidence)\n",
        playerName, pred.Classification, pred.Confidence*100)

    for _, reason := range pred.Reasoning {
        response += fmt.Sprintf("- %s\n", reason)
    }

    return response
}
```

## Integration with Log Parser

```go
// In your event parser, after recording a friendly fire kill
func handleFriendlyFireKill(app core.App, event *FriendlyFireEvent) {
    // Record incident in database
    // ... existing code ...

    // Classify player after recording incident
    classifier := ml.NewDefaultClassifier()
    pred, err := classifier.ClassifyPlayer(ctx, app, event.KillerID)
    if err != nil {
        return
    }

    // Auto-moderate high-risk players
    if pred.Classification == "likely_intentional" && pred.Confidence > 0.85 {
        // Flag for admin review
        log.Printf("High-risk player detected: %s (%.0f%% confidence)",
            event.KillerName, pred.Confidence*100)

        // Optionally: Auto-kick after X intentional FF
        // SendRCONCommand(app, fmt.Sprintf("kick %s Excessive teamkilling", event.KillerName))
    }
}
```

## Automated Moderation Workflow

```go
// Run periodic checks for high-risk players
func checkHighRiskPlayers(app core.App) {
    // Get all players with FF incidents
    var players []string
    app.DB().NewQuery(`
        SELECT DISTINCT killer FROM friendly_fire_incidents
    `).All(&players)

    classifier := ml.NewDefaultClassifier()
    risks, _ := classifier.BatchClassifyPlayers(ctx, app, players)

    for _, risk := range risks {
        if risk.RiskScore > 70.0 {
            // Send warning to player
            SendWarning(app, risk.PlayerID,
                "Your friendly fire behavior has been flagged for review")

            // Notify admins
            NotifyAdmins(app, fmt.Sprintf(
                "Player %s flagged as high risk (score: %.1f)",
                risk.PlayerID, risk.RiskScore))
        }
    }
}
```

## Algorithm Details

### Scoring System

Each feature contributes a score to either "accident" or "intentional":

**FF Rate (most important):**

- <5%: +3.0 accident
- 5-8%: +2.0 accident
- 8-10%: +1.0 each (neutral)
- 10-15%: +2.0 intentional
- > 15%: +3.0 intentional

**Weapon Usage:**

- > 70% explosive/vehicle: +2.5 accident
- 30-70% mixed: +0.5 each (neutral)
- <30% explosive/vehicle: +2.0 intentional

**Timing:**

- > 180s between: +2.0 accident
- 45-180s between: +0.5 each (neutral)
- <45s between: +2.5 intentional

**Spawn Kills:**

- 0 in first minute: +1.0 accident
- > 2 in first minute: +2.5 intentional

**Final Classification:**

- accident_score / (accident_score + intentional_score) > 0.70 → **likely_accident**
- accident_score / (accident_score + intentional_score) > 0.55 → **likely_accident** (lower confidence)
- intentional_score / (accident_score + intentional_score) > 0.70 → **likely_intentional**
- intentional_score / (accident_score + intentional_score) > 0.55 → **possibly_intentional**
- else → **possibly_intentional** (50% confidence)

### Why Not External ML?

This native Go implementation:

- ✅ **No external dependencies** (Python, TensorFlow, API calls)
- ✅ **Fast inference** (<1ms per classification)
- ✅ **Simple deployment** (single binary)
- ✅ **Explainable** (clear reasoning for each decision)
- ✅ **Tunable** (adjust thresholds without retraining)

For more complex models (neural networks, ensemble methods), consider:

- **GoLearn**: Go ML library with decision trees, random forests
- **Gorgonia**: Deep learning library for Go
- **External API**: Train in Python, serve via REST API

## Testing

Run tests:

```bash
go test ./internal/ml/... -v
```

**Test Coverage:**

- ✅ Prediction logic with multiple scenarios
- ✅ Feature extraction from database
- ✅ Player classification
- ✅ Batch processing and risk scoring
- ✅ Utility functions (normalize, median)

**Note**: Database-dependent tests will skip if migrations haven't been run.

## Requirements

- PocketBase v0.31.0+
- Go 1.24.0+
- Database migrations must be run (see `/internal/database/DAILY_STATS_USAGE.md`)

## Future Enhancements

- [ ] **Online learning**: Update thresholds as new incidents are labeled
- [ ] **Player profiles**: Track behavior changes over time
- [ ] **Team analysis**: Detect coordinated griefing
- [ ] **Map-specific patterns**: Different thresholds per map
- [ ] **Weapon-specific models**: Classify by weapon type
- [ ] **Ensemble methods**: Combine multiple classifiers

## License

MIT License - Part of sandstorm-trackerv2 project
