package watcher

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"sandstorm-tracker/internal/config"
	"sandstorm-tracker/internal/parser"
	"sandstorm-tracker/internal/rcon"
	"sandstorm-tracker/internal/util"

	"github.com/fsnotify/fsnotify"
	"github.com/pocketbase/pocketbase/core"
)

// Watcher monitors log files and processes events
type Watcher struct {
	watcher          *fsnotify.Watcher
	parser           *parser.LogParser
	pbApp            core.App
	ctx              context.Context
	cancel           context.CancelFunc
	wg               sync.WaitGroup
	mu               sync.RWMutex
	rconPool         *rcon.ClientPool
	serverConfigs    map[string]config.ServerConfig
	onServerActive   func(serverID string) // Callback when server becomes active (log rotation detected)
	onServerInactive func(serverID string) // Callback when server becomes inactive (no activity for 10s)
	activeServers    map[string]bool       // Track which servers are active
	activeServersMu  sync.RWMutex
	lastActivity     map[string]time.Time // Track last activity time for each server
	lastActivityMu   sync.RWMutex
	inactivityTimer  *time.Duration // How long to wait before marking server as inactive (default 10s)
}

// NewWatcher creates a new file watcher
// All dependencies are injected: parser and rconPool
func NewWatcher(pbApp core.App, logParser *parser.LogParser, rconPool *rcon.ClientPool, serverConfigs []config.ServerConfig) (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	scMap := make(map[string]config.ServerConfig)
	for _, sc := range serverConfigs {
		serverID, err := util.GetServerIdFromPath(sc.LogPath)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to get server ID from path %s: %w", sc.LogPath, err)
		}
		scMap[serverID] = sc
	}

	inactivityDuration := 10 * time.Second

	w := &Watcher{
		watcher:         watcher,
		parser:          logParser,
		pbApp:           pbApp,
		ctx:             ctx,
		cancel:          cancel,
		rconPool:        rconPool,
		serverConfigs:   scMap,
		activeServers:   make(map[string]bool),
		lastActivity:    make(map[string]time.Time),
		inactivityTimer: &inactivityDuration,
	}

	// Start inactivity monitor
	w.startInactivityMonitor()

	return w, nil
}

// OnServerActive sets a callback to be called when a server becomes active (log rotation detected)
func (w *Watcher) OnServerActive(callback func(serverID string)) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.onServerActive = callback
}

// OnServerInactive sets a callback to be called when a server becomes inactive (no activity for 10s)
func (w *Watcher) OnServerInactive(callback func(serverID string)) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.onServerInactive = callback
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

		// Mark server as active and trigger callback if this is the first time
		w.markServerActive(serverID)

		// Update last activity time
		w.updateLastActivity(serverID)
	}
}

// updateLastActivity updates the last activity timestamp for a server
func (w *Watcher) updateLastActivity(serverID string) {
	w.lastActivityMu.Lock()
	defer w.lastActivityMu.Unlock()
	w.lastActivity[serverID] = time.Now()
}

// markServerActive marks a server as active and triggers the onServerActive callback
func (w *Watcher) markServerActive(serverID string) {
	w.activeServersMu.Lock()
	defer w.activeServersMu.Unlock()

	// Check if server was already marked as active
	if w.activeServers[serverID] {
		return
	}

	// Mark as active
	w.activeServers[serverID] = true
	log.Printf("[Watcher] Server %s became active (log rotation detected)", serverID)

	// Trigger callback if set
	w.mu.RLock()
	callback := w.onServerActive
	w.mu.RUnlock()

	if callback != nil {
		go callback(serverID)
	}
}

func (w *Watcher) extractServerIDFromPath(filePath string) string {
	base := filepath.Base(filePath)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

func (w *Watcher) getOrCreateServerDBID(serverID, logPath string) (string, error) {
	// Find server by external_id (the UUID from the log filename)
	record, err := w.pbApp.FindFirstRecordByFilter(
		"servers",
		"external_id = {:external_id}",
		map[string]any{"external_id": serverID},
	)

	if err == nil {
		// Server found by external_id
		return record.Id, nil
	}

	// This should not happen if config is properly set up
	// But log a warning instead of creating a new server
	log.Printf("Warning: No server found in database with external_id: %s (from path: %s)", serverID, logPath)
	return "", fmt.Errorf("server not found in database for external_id: %s", serverID)
}

// GetRconClient returns the RCON client for a server, creating it if needed
func (w *Watcher) GetRconClient(serverID string) (*rcon.RconClient, error) {
	return w.rconPool.GetClient(serverID)
}

// startInactivityMonitor starts a goroutine that checks for inactive servers every 5 seconds
func (w *Watcher) startInactivityMonitor() {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-w.ctx.Done():
				return
			case <-ticker.C:
				w.checkInactiveServers()
			}
		}
	}()
}

// checkInactiveServers checks for servers that haven't had activity in 10 seconds and marks them inactive
func (w *Watcher) checkInactiveServers() {
	w.lastActivityMu.RLock()
	now := time.Now()
	inactivityThreshold := *w.inactivityTimer

	var inactiveServers []string
	for serverID, lastTime := range w.lastActivity {
		if now.Sub(lastTime) > inactivityThreshold {
			// Check if server is currently marked as active
			w.activeServersMu.RLock()
			isActive := w.activeServers[serverID]
			w.activeServersMu.RUnlock()

			if isActive {
				inactiveServers = append(inactiveServers, serverID)
			}
		}
	}
	w.lastActivityMu.RUnlock()

	// Mark servers as inactive and trigger callback
	for _, serverID := range inactiveServers {
		w.markServerInactive(serverID)
	}
}

// markServerInactive marks a server as inactive and triggers the onServerInactive callback
func (w *Watcher) markServerInactive(serverID string) {
	w.activeServersMu.Lock()
	defer w.activeServersMu.Unlock()

	// Check if server was actually active
	if !w.activeServers[serverID] {
		return
	}

	// Mark as inactive
	w.activeServers[serverID] = false
	log.Printf("[Watcher] Server %s became inactive (no activity for 10s)", serverID)

	// Trigger callback if set
	w.mu.RLock()
	callback := w.onServerInactive
	w.mu.RUnlock()

	if callback != nil {
		go callback(serverID)
	}
}
