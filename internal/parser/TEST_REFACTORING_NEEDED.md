# Parser Tests Refactoring Needed

## Issue

After migrating to event-driven architecture, parser tests are failing because they verify database state after parsing. The parser no longer updates the database directly - it only creates events.

## Architecture Change

**Before:**

- Parser → Parse logs → Update database directly

**After:**

- Parser → Parse logs → Create events
- Handlers → Process events → Update database

## Failing Tests

### 1. `TestProcessFullLog` (full_log_test.go)

- **What it tests:** Verifies players, matches, kills, objectives in database
- **Why it fails:** Parser creates events but doesn't process them
- **Fix options:**
  - Option A: Change test to verify events are created (recommended for parser tests)
  - Option B: Set up handlers and process events through full flow
  - Option C: Remove test (end-to-end tests should cover this)

### 2. `TestParseAndWriteLogToDB_HCLog` (parser_test.go)

- **What it tests:** Verifies kill counts in database after parsing hc.log
- **Why it fails:** Parser creates kill events but handlers don't process them
- **Fix options:** Same as above

### 3. `TestMapLoadEvents` (mapload_test.go)

- **What it tests:** Verifies match properties (player_team field)
- **Why it fails:** Match is created but events aren't processed by handlers
- **Fix options:** Same as above

### 4. `TestWeaponNameStandardization` & `TestWeaponStatsAggregation` (weapon_names_test.go)

- **What they test:** Verifies weapon stats aggregation in database
- **Why they fail:** Parser creates kill events but handlers don't process them
- **Fix options:** Same as above

## Recommended Approach

### For Parser Unit Tests

Test that parser creates correct **events**, not database state:

```go
func TestParserCreatesKillEvents(t *testing.T) {
    // Setup test app with events collection
    testApp, _ := tests.NewTestApp()
    defer testApp.Cleanup()

    parser := NewLogParser(testApp, logger)

    // Parse kill line
    killLine := "[timestamp] Player1[123, team 0] killed Player2[456, team 1] with AKM"
    parser.ParseAndProcess(ctx, killLine, "server-id", "test.log")

    // Verify kill event was created
    events, _ := testApp.FindRecordsByFilter("events", "type = 'player_kill'", "-created", 10, 0)
    assert.Equal(t, 1, len(events))

    // Verify event data
    var data map[string]interface{}
    json.Unmarshal([]byte(events[0].GetString("data")), &data)
    assert.Equal(t, "123", data["killer_steam_id"])
    assert.Equal(t, "456", data["victim_steam_id"])
    assert.Equal(t, "AKM", data["weapon"])
}
```

### For Integration Tests

Move database verification tests to a separate integration test suite that:

1. Sets up parser + handlers
2. Parses logs
3. Verifies database state after event processing

## Quick Fix (Temporary)

Skip failing tests until refactored:

```go
func TestProcessFullLog(t *testing.T) {
    t.Skip("Needs refactoring for event-driven architecture")
    // ... test code
}
```

## Status

- ✅ Chat command tests: PASSING (don't depend on database state)
- ✅ Event extraction tests: PASSING (only verify parsing, not DB)
- ✅ Timestamp tests: PASSING (pure functions)
- ❌ Full log tests: FAILING (expect database updates)
- ❌ Weapon stats tests: FAILING (expect database updates)
- ❌ Map load tests: FAILING (expect database updates)
