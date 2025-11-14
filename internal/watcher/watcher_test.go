package watcher

import (
	"os"
	"path/filepath"
	"testing"

	"sandstorm-tracker/internal/parser"

	"github.com/pocketbase/pocketbase/tests"
)

// TestLogRotationScenario provides a manual test scenario for log rotation
// This documents how to test the tracker after Insurgency servers have restarted
func TestLogRotationScenario(t *testing.T) {
	t.Skip("This is a documentation test - run manually to simulate log rotation")

	// SCENARIO: Testing tracker restart after Insurgency server has rotated logs
	//
	// 1. Initial state:
	//    - Tracker running and tracking server at /path/to/logs/Insurgency.log
	//    - Database has offset = 50000 bytes, log_file_creation_time = "2025-11-10T20:00:00Z"
	//
	// 2. Server restarts while tracker is offline:
	//    - Insurgency server rotates log file
	//    - Old log becomes Insurgency.log.1 (or deleted)
	//    - New Insurgency.log created with "Log file open, 11/11/25 13:30:00"
	//    - New file starts at 0 bytes
	//
	// 3. Tracker starts:
	//    - Reads offset = 50000 from database
	//    - Opens Insurgency.log
	//    - Extracts log file creation time from first line
	//    - Compares: "2025-11-11T13:30:00Z" != "2025-11-10T20:00:00Z" → ROTATION DETECTED
	//    - Resets offset to 0
	//    - Processes log from beginning
	//
	// 4. Verification:
	//    - Check database: offset should be reset to 0 initially, then increase
	//    - Check database: log_file_creation_time should be updated
	//    - Check logs: should see "Log rotation detected" message
	//    - Check match creation: should create new match for current map
	//
	// FALLBACK: If timestamp parsing fails:
	//    - Compare file size (e.g., 5000 bytes) < saved offset (50000 bytes)
	//    - Rotation detected via size check
	//    - Reset to 0 and process from beginning
}

// TestExtractLogFileCreationTime verifies timestamp extraction works correctly
func TestExtractLogFileCreationTime(t *testing.T) {
	testCases := []struct {
		name     string
		content  string
		expected string // Format: "MM/DD/YY HH:MM:SS"
	}{
		{
			name:     "Standard format",
			content:  "Log file open, 11/10/25 20:58:31\n[2025.11.10-20.58.35:123][1]LogWorld: Test\n",
			expected: "11/10/25 20:58:31",
		},
		{
			name:     "Different time",
			content:  "Log file open, 01/15/25 08:30:45\n[2025.01.15-08.30.45:123][1]LogWorld: Test\n",
			expected: "01/15/25 08:30:45",
		},
		{
			name:     "Midnight",
			content:  "Log file open, 12/31/24 00:00:00\n[2024.12.31-00.00.00:123][1]LogWorld: Test\n",
			expected: "12/31/24 00:00:00",
		},
		{
			name:     "Trailing spaces",
			content:  "Log file open, 11/11/25 13:45:22  \n[2025.11.11-13.45.25:123][1]LogWorld: Test\n",
			expected: "11/11/25 13:45:22",
		},
		{
			name:     "Trailing tab",
			content:  "Log file open, 11/11/25 13:45:22\t\n[2025.11.11-13.45.25:123][1]LogWorld: Test\n",
			expected: "11/11/25 13:45:22",
		},
		{
			name:     "UTF-8 BOM at start",
			content:  "\ufeffLog file open, 11/11/25 13:47:42\n[2025.11.11-13.47.45:123][1]LogWorld: Test\n",
			expected: "11/11/25 13:47:42",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create temporary log file
			tmpDir := t.TempDir()
			temp := t.TempDir()
			testApp, err := tests.NewTestApp(temp)
			if err != nil {
				t.Fatalf("Failed to create test app: %v", err)
			}
			defer testApp.Cleanup()

			logFile := filepath.Join(tmpDir, "test.log")
			if err := os.WriteFile(logFile, []byte(tc.content), 0644); err != nil {
				t.Fatalf("Failed to create test log file: %v", err)
			}

			// Create parser and extract timestamp
			logParser := parser.NewLogParser(testApp, testApp.Logger()) // nil app OK for this test
			timestamp, err := logParser.ExtractLogFileCreationTime(logFile)
			if err != nil {
				t.Fatalf("Failed to extract timestamp: %v", err)
			}

			// Verify extracted time matches expected
			extracted := timestamp.Format("01/02/06 15:04:05")
			if extracted != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, extracted)
			}

			t.Logf("✅ Extracted timestamp: %s → %v", tc.expected, timestamp)
		})
	}
}

// TestLogRotationDetectionViaTimestamp documents the primary detection method
func TestLogRotationDetectionViaTimestamp(t *testing.T) {
	t.Log("Log Rotation Detection: Timestamp Method")
	t.Log("==========================================")
	t.Log("")
	t.Log("1. First line of log: 'Log file open, MM/DD/YY HH:MM:SS'")
	t.Log("2. Parsed timestamp saved to: servers.log_file_creation_time")
	t.Log("3. On next run: Extract current file's timestamp")
	t.Log("4. Compare: current_time != saved_time → rotation detected")
	t.Log("5. Reset offset to 0 and update saved timestamp")
	t.Log("")
	t.Log("Example:")
	t.Log("  Saved:   '2025-11-10T20:00:00Z'")
	t.Log("  Current: '2025-11-11T13:30:00Z'")
	t.Log("  Result:  Rotation detected! Reset to 0")
}

// TestLogRotationDetectionViaFileSize documents the fallback detection method
func TestLogRotationDetectionViaFileSize(t *testing.T) {
	t.Log("Log Rotation Detection: File Size Fallback")
	t.Log("==========================================")
	t.Log("")
	t.Log("1. If timestamp parsing fails, use file size")
	t.Log("2. Compare: current_file_size < saved_offset")
	t.Log("3. If true → rotation detected (file truncated/reset)")
	t.Log("4. Reset offset to 0")
	t.Log("")
	t.Log("Example:")
	t.Log("  Saved offset: 50,000 bytes")
	t.Log("  Current size: 5,000 bytes")
	t.Log("  Result:       Rotation detected! Reset to 0")
}
