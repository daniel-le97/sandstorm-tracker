# Test Organization

All tests are now properly organized in the `/tests` directory with each test file containing a single exported function that `test-runner.ts` calls.

## Test Structure

### `/tests` Directory

```
tests/
├── tests.ts               # Event parsing tests (25 test cases)
├── weapon-test.ts          # Weapon name mapping tests
├── database-tests.ts       # Database functionality tests
├── command-tests.ts        # Chat command response tests
├── watcher-tests.ts        # File watcher functionality tests
└── integration-tests.ts    # End-to-end integration tests
```

### Test Runner

- **`test-runner.ts`** - Main test orchestrator that calls all test functions

## Test Functions

Each test file exports a single function:

- `runTests()` - Comprehensive event parsing (25 tests)
- `runWeaponTests()` - Weapon name mapping verification
- `runDatabaseTests()` - SQLite database operations
- `runCommandTests()` - Chat command responses (!stats, !kdr, !top, !guns)
- `runWatcherTests()` - File monitoring functionality
- `runIntegrationTests()` - Complete end-to-end system test

## Running Tests

### Run All Tests

```bash
bun test-runner.ts
```

### Run Individual Test Suites

```bash
# Import and run specific tests
import { runTests } from './tests/tests';
runTests();
```

## Test Coverage

✅ **Event Parsing** (25 tests)

- Player joins/leaves/disconnects
- Kills, team kills, suicides
- Round events, map changes
- Chat commands with various formats
- Invalid/edge cases

✅ **Database Operations**

- Table creation and initialization
- Player session tracking
- Kill/death recording
- Statistics retrieval
- Chat command logging

✅ **Chat Commands**

- `!stats` - Personal statistics
- `!kdr` - Kill/death ratio
- `!top` - Leaderboard
- `!guns` - Weapon statistics

✅ **File Watching**

- File modification detection
- Change handling
- Setup verification

✅ **Integration**

- Complete log processing pipeline
- Database integration
- Real-time event handling
- Command response system

## Test Results Summary

```
Test Suites Passed: 6/6
🎉 ALL TESTS PASSED! 🎉
```

All tests pass successfully, ensuring the entire statistics tracking system works correctly from log parsing to database storage to chat command responses.
