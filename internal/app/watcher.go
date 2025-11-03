package app

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"sandstorm-tracker/internal/rcon"

	"github.com/fsnotify/fsnotify"
	"github.com/pocketbase/pocketbase/core"
)

// Watcher monitors log files and processes events
type Watcher struct {
	watcher       *fsnotify.Watcher
	parser        *LogParser
	pbApp         core.App
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	mu            sync.RWMutex
	rconClients   map[string]*rcon.RconClient
	rconMu        sync.Mutex
	serverConfigs map[string]ServerConfig
}

// NewWatcher creates a new file watcher
func NewWatcher(pbApp core.App, serverConfigs []ServerConfig) (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	scMap := make(map[string]ServerConfig)
	for _, sc := range serverConfigs {
		serverID, err := GetServerIdFromPath(sc.LogPath)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to get server ID from path %s: %w", sc.LogPath, err)
		}
		scMap[serverID] = sc
	}

	w := &Watcher{
		watcher:       watcher,
		parser:        NewLogParser(pbApp),
		pbApp:         pbApp,
		ctx:           ctx,
		cancel:        cancel,
		rconClients:   make(map[string]*rcon.RconClient),
		serverConfigs: scMap,
	}

	return w, nil
}

// AddPath adds a file or directory to watch
// Supports both:
//   - File path: /path/to/logs/server-uuid.log (for sandstorm-admin-wrapper)
//   - Directory path: /path/to/logs (for standalone servers)
func (w *Watcher) AddPath(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for %s: %w", path, err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("path %s does not exist: %w", absPath, err)
	}

	if info.IsDir() {
		// Watch entire directory for .log files
		err = w.watcher.Add(absPath)
		if err != nil {
			return fmt.Errorf("failed to watch directory %s: %w", absPath, err)
		}
		log.Printf("Watching directory: %s", absPath)

		// Process existing log files in directory
		files, err := os.ReadDir(absPath)
		if err != nil {
			return fmt.Errorf("failed to read directory %s: %w", absPath, err)
		}
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".log") && !strings.Contains(file.Name(), "-backup-") {
				logFilePath := filepath.Join(absPath, file.Name())
				go w.processFile(logFilePath)
			}
		}
	} else {
		// Watch specific file (sandstorm-admin-wrapper use case)
		// Also watch the parent directory to catch file changes
		dir := filepath.Dir(absPath)
		err = w.watcher.Add(dir)
		if err != nil {
			return fmt.Errorf("failed to watch directory %s: %w", dir, err)
		}

		go w.processFile(absPath)
		log.Printf("Watching file: %s", absPath)
	}
	return nil
}

// Start begins watching for file changes
func (w *Watcher) Start() {
	w.wg.Add(1)
	go w.watchLoop()
}

// Stop stops the watcher
func (w *Watcher) Stop() {
	w.cancel()
	w.wg.Wait()
	w.watcher.Close()
}

func (w *Watcher) watchLoop() {
	defer w.wg.Done()
	for {
		select {
		case <-w.ctx.Done():
			return
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			w.handleFileEvent(event)
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("File watcher error: %v", err)
		}
	}
}

func (w *Watcher) handleFileEvent(event fsnotify.Event) {
	if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
		name := filepath.Base(event.Name)
		if strings.HasSuffix(name, ".log") && !strings.Contains(name, "-backup-") {
			go w.processFile(event.Name)
		}
	}
}

func (w *Watcher) processFile(filePath string) {
	serverID := w.extractServerIDFromPath(filePath)

	// Get server DB ID and load offset from database
	serverDBID, err := w.getOrCreateServerDBID(serverID, filePath)
	if err != nil {
		log.Printf("Failed to get server DB ID: %v", err)
		return
	}

	// Load offset from database
	serverRecord, err := w.pbApp.FindRecordById("servers", serverDBID)
	if err != nil {
		log.Printf("Error loading server record: %v", err)
		return
	}

	offset := serverRecord.GetInt("offset")

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Error opening file %s: %v", filePath, err)
		return
	}
	defer file.Close()

	if _, err := file.Seek(int64(offset), 0); err != nil {
		log.Printf("Error seeking to offset %d in %s: %v", offset, filePath, err)
		return
	}

	scanner := bufio.NewScanner(file)
	linesProcessed := 0

	for scanner.Scan() {
		line := scanner.Text()

		// Parse and process directly - no intermediate structs
		if err := w.parser.ParseAndProcess(w.ctx, line, serverDBID, filePath); err != nil {
			log.Printf("Error processing line: %v", err)
		}

		linesProcessed++
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error scanning file %s: %v", filePath, err)
		return
	}

	newOffset, _ := file.Seek(0, 1)

	// Save new offset to database
	if linesProcessed > 0 {
		serverRecord.Set("offset", newOffset)
		if err := w.pbApp.Save(serverRecord); err != nil {
			log.Printf("Error saving server offset: %v", err)
		} else {
			log.Printf("Processed %d lines from %s (new offset: %d)", linesProcessed, filePath, newOffset)
		}
	}
}

func (w *Watcher) extractServerIDFromPath(filePath string) string {
	base := filepath.Base(filePath)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

func (w *Watcher) getOrCreateServerDBID(serverID, logPath string) (string, error) {
	// Normalize the path for comparison
	absPath, err := filepath.Abs(logPath)
	if err != nil {
		absPath = logPath
	}

	// Find server by matching the log path from config
	record, err := w.pbApp.FindFirstRecordByFilter(
		"servers",
		"path = {:path}",
		map[string]any{"path": absPath},
	)

	if err == nil {
		// Server found by path
		return record.Id, nil
	}

	// This should not happen if config is properly set up
	// But log a warning instead of creating a new server
	log.Printf("Warning: No server found in database with path: %s", absPath)
	return "", fmt.Errorf("server not found in database for path: %s", absPath)
}

// GetRconClient returns the RCON client for a server, creating it if needed
func (w *Watcher) GetRconClient(serverID string) (*rcon.RconClient, error) {
	w.rconMu.Lock()
	defer w.rconMu.Unlock()

	if client, exists := w.rconClients[serverID]; exists {
		return client, nil
	}

	serverConfig, exists := w.serverConfigs[serverID]
	if !exists {
		return nil, fmt.Errorf("no config found for server %s", serverID)
	}

	if serverConfig.RconAddress == "" || serverConfig.RconPassword == "" {
		return nil, fmt.Errorf("RCON not configured for server %s", serverID)
	}

	// Connect to RCON server
	conn, err := net.DialTimeout("tcp", serverConfig.RconAddress, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RCON for server %s: %w", serverID, err)
	}

	client := rcon.NewRconClient(conn, rcon.DefaultConfig())

	// Authenticate
	if !client.Auth(serverConfig.RconPassword) {
		conn.Close()
		return nil, fmt.Errorf("RCON authentication failed for server %s", serverID)
	}

	w.rconClients[serverID] = client
	log.Printf("Created RCON client for server %s", serverID)

	return client, nil
}
