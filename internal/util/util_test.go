package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetServerIdFromPath(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func() (string, func()) // returns path and cleanup function
		wantID      string
		wantErr     bool
		errContains string
	}{
		{
			name: "file path with .log extension",
			setupFunc: func() (string, func()) {
				tmpDir := t.TempDir()
				logFile := filepath.Join(tmpDir, "server-abc123.log")
				f, _ := os.Create(logFile)
				f.Close()
				return logFile, func() {}
			},
			wantID:  "server-abc123",
			wantErr: false,
		},
		{
			name: "file path with UUID",
			setupFunc: func() (string, func()) {
				tmpDir := t.TempDir()
				logFile := filepath.Join(tmpDir, "550e8400-e29b-41d4-a716-446655440000.log")
				f, _ := os.Create(logFile)
				f.Close()
				return logFile, func() {}
			},
			wantID:  "550e8400-e29b-41d4-a716-446655440000",
			wantErr: false,
		},
		{
			name: "directory with single log file",
			setupFunc: func() (string, func()) {
				tmpDir := t.TempDir()
				logFile := filepath.Join(tmpDir, "my-server.log")
				f, _ := os.Create(logFile)
				f.Close()
				return tmpDir, func() {}
			},
			wantID:  "my-server",
			wantErr: false,
		},
		{
			name: "directory with multiple log files returns first non-backup",
			setupFunc: func() (string, func()) {
				tmpDir := t.TempDir()
				// Create backup file first (should be ignored)
				backup := filepath.Join(tmpDir, "server-backup.log")
				f1, _ := os.Create(backup)
				f1.Close()
				// Create actual log file
				logFile := filepath.Join(tmpDir, "server-main.log")
				f2, _ := os.Create(logFile)
				f2.Close()
				return tmpDir, func() {}
			},
			wantID:  "server-main",
			wantErr: false,
		},
		{
			name: "directory with no log files",
			setupFunc: func() (string, func()) {
				tmpDir := t.TempDir()
				// Create non-log file
				txtFile := filepath.Join(tmpDir, "readme.txt")
				f, _ := os.Create(txtFile)
				f.Close()
				return tmpDir, func() {}
			},
			wantErr:     true,
			errContains: "no log files found",
		},
		{
			name: "non-existent path",
			setupFunc: func() (string, func()) {
				return "/path/that/does/not/exist", func() {}
			},
			wantErr:     true,
			errContains: "path does not exist",
		},
		{
			name: "empty filename after trimming .log",
			setupFunc: func() (string, func()) {
				tmpDir := t.TempDir()
				logFile := filepath.Join(tmpDir, ".log")
				f, _ := os.Create(logFile)
				f.Close()
				return logFile, func() {}
			},
			wantErr:     true,
			errContains: "invalid log file name",
		},
		{
			name: "directory ignores backup files",
			setupFunc: func() (string, func()) {
				tmpDir := t.TempDir()
				// Only create backup file
				backup := filepath.Join(tmpDir, "server-backup-2024.log")
				f, _ := os.Create(backup)
				f.Close()
				return tmpDir, func() {}
			},
			wantErr:     true,
			errContains: "no log files found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, cleanup := tt.setupFunc()
			defer cleanup()

			gotID, err := GetServerIdFromPath(path)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetServerIdFromPath() expected error but got nil")
					return
				}
				if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("GetServerIdFromPath() error = %v, should contain %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("GetServerIdFromPath() unexpected error = %v", err)
				return
			}

			if gotID != tt.wantID {
				t.Errorf("GetServerIdFromPath() = %v, want %v", gotID, tt.wantID)
			}
		})
	}
}

func TestExtractGameMode(t *testing.T) {
	tests := []struct {
		name     string
		scenario string
		want     string
	}{
		// Full scenario strings with Scenario_ prefix
		{
			name:     "Ministry Checkpoint Security",
			scenario: "Scenario_Ministry_Checkpoint_Security",
			want:     "Checkpoint",
		},
		{
			name:     "Refinery Push Insurgents",
			scenario: "Scenario_Refinery_Push_Insurgents",
			want:     "Push",
		},
		{
			name:     "Town Skirmish",
			scenario: "Scenario_Town_Skirmish",
			want:     "Skirmish",
		},
		{
			name:     "Hideout Checkpoint Security",
			scenario: "Scenario_Hideout_Checkpoint_Security",
			want:     "Checkpoint",
		},
		{
			name:     "Oilfield Push Insurgents",
			scenario: "Scenario_Oilfield_Push_Insurgents",
			want:     "Push",
		},
		// Simple mode strings (already extracted)
		{
			name:     "Simple Checkpoint",
			scenario: "Checkpoint",
			want:     "Checkpoint",
		},
		{
			name:     "Simple Push",
			scenario: "Push",
			want:     "Push",
		},
		{
			name:     "Simple Skirmish",
			scenario: "Skirmish",
			want:     "Skirmish",
		},
		// Full scenario without Scenario_ prefix
		{
			name:     "Ministry Checkpoint Security without prefix",
			scenario: "Ministry_Checkpoint_Security",
			want:     "Checkpoint",
		},
		{
			name:     "Town Skirmish without prefix",
			scenario: "Town_Skirmish",
			want:     "Skirmish",
		},
		// Edge cases
		{
			name:     "Empty string",
			scenario: "",
			want:     "Unknown",
		},
		{
			name:     "Single part",
			scenario: "Scenario_Ministry",
			want:     "Unknown",
		},
		{
			name:     "Unknown mode",
			scenario: "Scenario_Ministry_Unknown_Security",
			want:     "Unknown",
		},
		{
			name:     "Invalid mode name",
			scenario: "Scenario_Ministry_InvalidMode_Security",
			want:     "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractGameMode(tt.scenario)
			if got != tt.want {
				t.Errorf("ExtractGameMode(%q) = %q, want %q", tt.scenario, got, tt.want)
			}
		})
	}
}

func containsString(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr || len(substr) == 0 ||
		(len(str) > 0 && len(substr) > 0 && findSubstring(str, substr)))
}

func findSubstring(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
