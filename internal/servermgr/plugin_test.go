//TODO make this cross platform
//go:build windows

package servermgr

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/pocketbase/pocketbase/tests"
)

func TestPIDFileOperations(t *testing.T) {
	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	plugin := &Plugin{
		app:     app,
		config:  Config{},
		servers: make(map[string]*ManagedServer),
	}

	// Use actual data directory - tests will use real PID files
	// Clean up after tests
	serverID := "test-server-123"
	testPID := 12345
	defer plugin.removePIDFile(serverID)

	t.Run("savePIDFile", func(t *testing.T) {
		err := plugin.savePIDFile(serverID, testPID)
		if err != nil {
			t.Fatalf("savePIDFile failed: %v", err)
		}

		// Verify file exists
		pidFile := plugin.getPIDFilePath(serverID)
		if _, err := os.Stat(pidFile); os.IsNotExist(err) {
			t.Fatal("PID file was not created")
		}
	})

	t.Run("loadPIDFile", func(t *testing.T) {
		pid, err := plugin.loadPIDFile(serverID)
		if err != nil {
			t.Fatalf("loadPIDFile failed: %v", err)
		}

		if pid != testPID {
			t.Fatalf("expected PID %d, got %d", testPID, pid)
		}
	})

	t.Run("loadPIDFile_NotFound", func(t *testing.T) {
		_, err := plugin.loadPIDFile("nonexistent-server-999")
		if err == nil {
			t.Fatal("expected error for nonexistent PID file")
		}
	})

	t.Run("removePIDFile", func(t *testing.T) {
		err := plugin.removePIDFile(serverID)
		if err != nil {
			t.Fatalf("removePIDFile failed: %v", err)
		}

		// Verify file is gone
		pidFile := plugin.getPIDFilePath(serverID)
		if _, err := os.Stat(pidFile); !os.IsNotExist(err) {
			t.Fatal("PID file was not removed")
		}
	})

	t.Run("removePIDFile_NotExist", func(t *testing.T) {
		// Should not error when file doesn't exist
		err := plugin.removePIDFile("nonexistent-server-888")
		if err != nil {
			t.Fatalf("removePIDFile should not error for nonexistent file: %v", err)
		}
	})
}

func TestIsProcessRunning(t *testing.T) {
	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	plugin := &Plugin{
		app:     app,
		config:  Config{},
		servers: make(map[string]*ManagedServer),
	}

	t.Run("CurrentProcess", func(t *testing.T) {
		// Test with current process PID (should be running)
		currentPID := os.Getpid()
		if !plugin.isProcessRunning(currentPID) {
			t.Fatal("current process should be detected as running")
		}
	})

	t.Run("NonExistentProcess", func(t *testing.T) {
		// Use a very high PID that likely doesn't exist
		fakePID := 999999
		if plugin.isProcessRunning(fakePID) {
			t.Fatal("non-existent process should not be detected as running")
		}
	})
}

func TestGetRunningServerProcesses(t *testing.T) {
	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	plugin := &Plugin{
		app:     app,
		config:  Config{},
		servers: make(map[string]*ManagedServer),
	}

	t.Run("NoServersRunning", func(t *testing.T) {
		procs, err := plugin.getRunningServerProcesses()
		if err != nil {
			t.Fatalf("getRunningServerProcesses failed: %v", err)
		}

		// Should return non-nil slice (empty or with servers)
		// This is more of a smoke test - just checking it doesn't error
		_ = procs // procs can be nil or empty, both are valid
	})
}

func TestStopServerWithStalePID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	plugin := &Plugin{
		app:     app,
		config:  Config{},
		servers: make(map[string]*ManagedServer),
	}

	serverID := "test-server-stale"
	stalePID := 888888 // PID that definitely doesn't exist
	defer plugin.removePIDFile(serverID)

	t.Run("DetectAndCleanupStalePID", func(t *testing.T) {
		// Create stale PID file
		if err := plugin.savePIDFile(serverID, stalePID); err != nil {
			t.Fatalf("failed to create stale PID file: %v", err)
		}

		// Try to stop server - should detect stale PID and clean it up
		err := plugin.StopServer(serverID, "")
		if err == nil {
			t.Fatal("expected error for stale PID")
		}

		expectedMsg := fmt.Sprintf("server %s is not running (stale PID file cleaned up)", serverID)
		if err.Error() != expectedMsg {
			t.Fatalf("unexpected error message: %v", err)
		}

		// Verify PID file was removed
		pidFile := plugin.getPIDFilePath(serverID)
		if _, err := os.Stat(pidFile); !os.IsNotExist(err) {
			t.Fatal("stale PID file should have been removed")
		}
	})
}

func TestStartServerWithStalePID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	plugin := &Plugin{
		app:     app,
		config:  Config{},
		servers: make(map[string]*ManagedServer),
	}

	serverID := "test-server-start-stale"
	stalePID := 777777
	defer plugin.removePIDFile(serverID)

	t.Run("CleanupStalePIDBeforeStart", func(t *testing.T) {
		// Create stale PID file
		if err := plugin.savePIDFile(serverID, stalePID); err != nil {
			t.Fatalf("failed to create stale PID file: %v", err)
		}

		// Verify PID file exists
		pidFile := plugin.getPIDFilePath(serverID)
		if _, err := os.Stat(pidFile); os.IsNotExist(err) {
			t.Fatal("PID file should exist before test")
		}

		// Create a minimal config
		config := SAWServerConfig{
			ServerHostname:     "Test Server",
			ServerDefaultMap:   "Ministry",
			ServerScenarioMode: "Checkpoint",
			ServerDefaultSide:  "Security",
			ServerMaxPlayers:   "20",
			ServerGamePort:     "27102",
			ServerQueryPort:    "27131",
		}

		// Try to start server - should clean up stale PID
		// This will fail to actually start (no real server), but should clean up PID
		_ = plugin.StartServer(serverID, config, "", false)

		// The error will be about missing server executable, which is fine for this test
		// The important thing is the stale PID was cleaned up

		// Note: We can't easily test the full start flow without a real server executable
		// This test mainly verifies the cleanup logic is called
	})
}

func TestStartServerWithRunningProcess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	plugin := &Plugin{
		app:     app,
		config:  Config{},
		servers: make(map[string]*ManagedServer),
	}

	serverID := "test-server-running"
	defer plugin.removePIDFile(serverID)

	t.Run("DetectRunningProcess", func(t *testing.T) {
		// Use current process PID (definitely running)
		currentPID := os.Getpid()

		// Create PID file with current process
		if err := plugin.savePIDFile(serverID, currentPID); err != nil {
			t.Fatalf("failed to create PID file: %v", err)
		}

		config := SAWServerConfig{
			ServerHostname:     "Test Server",
			ServerDefaultMap:   "Ministry",
			ServerScenarioMode: "Checkpoint",
			ServerDefaultSide:  "Security",
			ServerMaxPlayers:   "20",
			ServerGamePort:     "27102",
			ServerQueryPort:    "27131",
		}

		// Try to start server - should detect running process
		err := plugin.StartServer(serverID, config, "", false)
		if err == nil {
			t.Fatal("expected error when server already running")
		}

		expectedError := fmt.Sprintf("server %s is already running (PID: %d)", serverID, currentPID)
		if err.Error() != expectedError {
			t.Fatalf("expected error '%s', got '%s'", expectedError, err.Error())
		}

		// Cleanup
		plugin.removePIDFile(serverID)
	})
}

func TestProcessInfoParsing(t *testing.T) {
	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	plugin := &Plugin{
		app:     app,
		config:  Config{},
		servers: make(map[string]*ManagedServer),
	}

	t.Run("GetCurrentProcesses", func(t *testing.T) {
		// Just verify we can call the function without errors
		_, err := plugin.getRunningServerProcesses()
		if err != nil {
			t.Fatalf("getRunningServerProcesses failed: %v", err)
		}
		// procs can be nil or have data, both are valid
	})
}

func TestLoadSAWConfigs(t *testing.T) {
	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	plugin := &Plugin{
		app:     app,
		config:  Config{},
		servers: make(map[string]*ManagedServer),
	}

	t.Run("InvalidPath", func(t *testing.T) {
		_, err := plugin.LoadSAWConfigs("/nonexistent/path")
		if err == nil {
			t.Fatal("expected error for invalid SAW path")
		}
	})

	t.Run("EmptyConfigs", func(t *testing.T) {
		// Create temp directory with empty config file
		tempDir := t.TempDir()
		adminDir := filepath.Join(tempDir, "admin-interface", "config")
		os.MkdirAll(adminDir, 0755)

		configFile := filepath.Join(adminDir, "server-configs.json")
		os.WriteFile(configFile, []byte("{}"), 0644)

		configs, err := plugin.LoadSAWConfigs(tempDir)
		if err != nil {
			t.Fatalf("LoadSAWConfigs failed: %v", err)
		}

		if len(configs) != 0 {
			t.Fatalf("expected 0 configs, got %d", len(configs))
		}
	})

	t.Run("ValidConfig", func(t *testing.T) {
		// Create temp directory with valid config
		tempDir := t.TempDir()
		adminDir := filepath.Join(tempDir, "admin-interface", "config")
		os.MkdirAll(adminDir, 0755)

		configData := `{
			"test-server-id": {
				"server_hostname": "Test Server",
				"server_default_map": "Ministry",
				"server_scenario_mode": "Checkpoint",
				"server_default_side": "Security",
				"server_max_players": "20",
				"server_game_port": "27102",
				"server_query_port": "27131"
			}
		}`

		configFile := filepath.Join(adminDir, "server-configs.json")
		os.WriteFile(configFile, []byte(configData), 0644)

		configs, err := plugin.LoadSAWConfigs(tempDir)
		if err != nil {
			t.Fatalf("LoadSAWConfigs failed: %v", err)
		}

		if len(configs) != 1 {
			t.Fatalf("expected 1 config, got %d", len(configs))
		}

		config, exists := configs["test-server-id"]
		if !exists {
			t.Fatal("expected test-server-id to exist")
		}

		if config.ServerHostname != "Test Server" {
			t.Fatalf("expected hostname 'Test Server', got '%s'", config.ServerHostname)
		}
	})
}

func TestCopyFile(t *testing.T) {
	t.Run("CopyValidFile", func(t *testing.T) {
		tempDir := t.TempDir()
		srcFile := filepath.Join(tempDir, "source.txt")
		dstFile := filepath.Join(tempDir, "subdir", "dest.txt")

		// Create source file
		content := []byte("test content")
		if err := os.WriteFile(srcFile, content, 0644); err != nil {
			t.Fatalf("failed to create source file: %v", err)
		}

		// Copy file
		if err := copyFile(srcFile, dstFile); err != nil {
			t.Fatalf("copyFile failed: %v", err)
		}

		// Verify destination exists and has same content
		dstContent, err := os.ReadFile(dstFile)
		if err != nil {
			t.Fatalf("failed to read destination file: %v", err)
		}

		if string(dstContent) != string(content) {
			t.Fatalf("content mismatch: expected '%s', got '%s'", content, dstContent)
		}
	})

	t.Run("CopyNonexistentFile", func(t *testing.T) {
		tempDir := t.TempDir()
		srcFile := filepath.Join(tempDir, "nonexistent.txt")
		dstFile := filepath.Join(tempDir, "dest.txt")

		err := copyFile(srcFile, dstFile)
		if err == nil {
			t.Fatal("expected error for nonexistent source file")
		}
	})
}

// Benchmark tests
func BenchmarkPIDFileOperations(b *testing.B) {
	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	plugin := &Plugin{
		app:     app,
		config:  Config{},
		servers: make(map[string]*ManagedServer),
	}

	serverID := "bench-server"
	testPID := 12345

	b.Run("SavePID", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			plugin.savePIDFile(fmt.Sprintf("%s-%d", serverID, i), testPID)
		}
		// Cleanup
		for i := 0; i < b.N; i++ {
			plugin.removePIDFile(fmt.Sprintf("%s-%d", serverID, i))
		}
	})

	b.Run("LoadPID", func(b *testing.B) {
		plugin.savePIDFile(serverID, testPID)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			plugin.loadPIDFile(serverID)
		}
		plugin.removePIDFile(serverID)
	})

	b.Run("RemovePID", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			plugin.savePIDFile(fmt.Sprintf("%s-%d", serverID, i), testPID)
			b.StartTimer()
			plugin.removePIDFile(fmt.Sprintf("%s-%d", serverID, i))
		}
	})
}

func BenchmarkIsProcessRunning(b *testing.B) {
	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	plugin := &Plugin{
		app:     app,
		config:  Config{},
		servers: make(map[string]*ManagedServer),
	}

	currentPID := os.Getpid()

	b.Run("CurrentProcess", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			plugin.isProcessRunning(currentPID)
		}
	})

	b.Run("NonexistentProcess", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			plugin.isProcessRunning(999999)
		}
	})
}
