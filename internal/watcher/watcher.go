package watcher

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"sandstorm-tracker/internal/a2s"
	"sandstorm-tracker/internal/config"
	"sandstorm-tracker/internal/database"
	"sandstorm-tracker/internal/parser"
	"sandstorm-tracker/internal/rcon"
	"sandstorm-tracker/internal/util"

	"github.com/fsnotify/fsnotify"
	"github.com/pocketbase/pocketbase/core"
)

// A2SQuerier is an interface for querying A2S server status
// This allows for mocking in tests
type A2SQuerier interface {
	QueryServer(ctx context.Context, address string) (*a2s.ServerStatus, error)
}

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
	a2sPool          A2SQuerier
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
// All dependencies are injected: parser, rconPool, and a2sPool
func NewWatcher(pbApp core.App, logParser *parser.LogParser, rconPool *rcon.ClientPool, a2sPool *a2s.ServerPool, serverConfigs []config.ServerConfig) (*Watcher, error) {
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
		a2sPool:         a2sPool,
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
	savedLogFileTime := serverRecord.GetString("log_file_creation_time")

	// On first run (offset == 0 and no saved log time), check if we should do startup catch-up
	if offset == 0 && savedLogFileTime == "" {
		catchupOffset, shouldCatchup := w.checkStartupCatchup(filePath, serverID)
		if shouldCatchup {
			offset = catchupOffset
			// Save the offset so we start from here
			serverRecord.Set("offset", offset)
			if err := w.pbApp.Save(serverRecord); err != nil {
				log.Printf("Error saving catch-up offset: %v", err)
			}
		}
	}

	// Extract current log file creation time
	currentLogFileTime, err := w.parser.ExtractLogFileCreationTime(filePath)
	if err != nil {
		log.Printf("Warning: Could not extract log file creation time for %s: %v", filePath, err)
		// Continue with size-based rotation detection
	}

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Error opening file %s: %v", filePath, err)
		return
	}
	defer file.Close()

	// Get current file size
	fileInfo, err := file.Stat()
	if err != nil {
		log.Printf("Error getting file info for %s: %v", filePath, err)
		return
	}
	currentSize := fileInfo.Size()

	// Check if log file was rotated by comparing creation times
	rotationDetected := false
	if savedLogFileTime != "" && !currentLogFileTime.IsZero() {
		savedTime, err := time.Parse(time.RFC3339, savedLogFileTime)
		if err == nil && !currentLogFileTime.Equal(savedTime) {
			log.Printf("Log rotation detected for %s (creation time changed: %v → %v)",
				serverID, savedTime.Format("2006-01-02 15:04:05"), currentLogFileTime.Format("2006-01-02 15:04:05"))
			rotationDetected = true
			offset = 0
		}
	}

	// Fallback: If current size is less than offset, log file was rotated (truncated/reset)
	if !rotationDetected && currentSize < int64(offset) {
		log.Printf("Log rotation detected for %s (file size %d < offset %d), resetting to 0", serverID, currentSize, offset)
		rotationDetected = true
		offset = 0
	}

	// If offset is 0 (new server or just after rotation), save current state
	if offset == 0 && !rotationDetected {
		log.Printf("New server %s detected, setting offset to %d bytes (waiting for first log rotation)", serverID, currentSize)
		serverRecord.Set("offset", currentSize)
		if !currentLogFileTime.IsZero() {
			serverRecord.Set("log_file_creation_time", currentLogFileTime.Format(time.RFC3339))
		}
		if err := w.pbApp.Save(serverRecord); err != nil {
			log.Printf("Error saving initial offset: %v", err)
		}
		return // Don't process until first rotation
	}

	// If offset equals current size, this is the first time we're watching - wait for new data
	if int64(offset) == currentSize {
		log.Printf("Server %s at current end of file (%d bytes), waiting for new log entries", serverID, currentSize)
		return // Don't process until new data is written
	}

	if _, err := file.Seek(int64(offset), 0); err != nil {
		log.Printf("Error seeking to offset %d in %s: %v", offset, filePath, err)
		return
	}

	scanner := bufio.NewScanner(file)
	linesProcessed := 0

	for scanner.Scan() {
		line := scanner.Text()

		// Parse and process directly - pass serverID (external_id), not serverDBID
		if err := w.parser.ParseAndProcess(w.ctx, line, serverID, filePath); err != nil {
			log.Printf("Error processing line: %v", err)
		}

		linesProcessed++
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error scanning file %s: %v", filePath, err)
		return
	}

	newOffset, _ := file.Seek(0, 1)

	// Save new offset and log file creation time to database
	if linesProcessed > 0 {
		serverRecord.Set("offset", newOffset)
		if !currentLogFileTime.IsZero() {
			serverRecord.Set("log_file_creation_time", currentLogFileTime.Format(time.RFC3339))
		}
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

// checkStartupCatchup determines if we should do catch-up processing on tracker startup
// Strategy:
// 1. Query A2S to check if server is online and get current map
// 2. If server offline, skip catch-up (can't run listplayers to get scores)
// 3. If server online, find last map event matching A2S current map
// 4. Only do catch-up if map matches (ensures we can get scores for this match)
// Returns the offset to start watching from and whether catch-up was performed
func (w *Watcher) checkStartupCatchup(filePath, serverID string) (int, bool) {
	// Get server config to access query address
	serverConfig, exists := w.serverConfigs[serverID]
	if !exists {
		log.Printf("[Catchup] No config found for server %s, skipping catch-up", serverID)
		return 0, false
	}

	// Check 1: Query A2S to verify server is online and get current map
	if serverConfig.QueryAddress == "" {
		log.Printf("[Catchup] No query address configured for %s, skipping catch-up", serverID)
		return 0, false
	}

	ctx, cancel := context.WithTimeout(w.ctx, 5*time.Second)
	defer cancel()

	serverStatus, err := w.a2sPool.QueryServer(ctx, serverConfig.QueryAddress)
	if err != nil {
		log.Printf("[Catchup] Server %s appears offline (A2S query failed: %v), skipping catch-up", serverID, err)
		return 0, false
	}

	if serverStatus == nil || serverStatus.Info == nil {
		log.Printf("[Catchup] Server %s returned no info, skipping catch-up", serverID)
		return 0, false
	}

	currentMap := serverStatus.Info.Map
	log.Printf("[Catchup] Server %s is online, current map: %s", serverID, currentMap)

	// Check 2: Is file recently modified?
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		log.Printf("Failed to stat file %s: %v", filePath, err)
		return 0, false
	}

	fileModTime := fileInfo.ModTime()
	timeSinceModification := time.Since(fileModTime)

	// Detect if SAW is active by checking for recent RCON logs
	sawActive := w.hasRecentRconLogs(filePath, 30*time.Second)

	// Use adaptive threshold based on whether SAW is active
	var fileModThreshold time.Duration
	if sawActive {
		fileModThreshold = 1 * time.Minute // SAW keeps file fresh with polling
		log.Printf("[Catchup] SAW detected for %s, using 1-minute threshold", serverID)
	} else {
		fileModThreshold = 9 * time.Hour // Servers restart every 8 hours, allow some buffer
		log.Printf("[Catchup] No SAW detected for %s, using 9-hour threshold", serverID)
	}

	fileRecentlyModified := timeSinceModification < fileModThreshold

	if !fileRecentlyModified {
		log.Printf("[Catchup] File %s not recently modified (%.1f minutes ago), skipping catch-up",
			serverID, timeSinceModification.Minutes())
		return 0, false
	}

	// Check 3: Find last map event in log file
	mapName, scenario, mapTime, startLineNum, err := w.parser.FindLastMapEvent(filePath, time.Now())
	if err != nil {
		log.Printf("[Catchup] No map event found for %s: %v, skipping catch-up", serverID, err)
		return 0, false
	}

	timeSinceMap := time.Since(mapTime)
	recentMapEvent := timeSinceMap < 30*time.Minute

	if !recentMapEvent {
		log.Printf("[Catchup] Map event too old for %s (%.1f minutes ago), skipping catch-up",
			serverID, timeSinceMap.Minutes())
		return 0, false
	}

	// Check 4: Does the log map match the current server map?
	// This is critical - we only process events if we can run listplayers to get scores
	if !strings.EqualFold(mapName, currentMap) {
		log.Printf("[Catchup] Map mismatch for %s: log has '%s' but server is on '%s', skipping catch-up (can't get scores for old match)",
			serverID, mapName, currentMap)
		return 0, false
	}

	// All conditions met - do catch-up!
	log.Printf("[Catchup] Starting catch-up for %s: server online, map matches (%s), file modified %.1fs ago, map loaded %.1fs ago",
		serverID, mapName, timeSinceModification.Seconds(), timeSinceMap.Seconds())

	// Get current file size as the catch-up end point
	catchupEndOffset := fileInfo.Size()

	// Extract player team from scenario
	var playerTeam *string
	if strings.Contains(scenario, "_Security") {
		team := "Security"
		playerTeam = &team
	} else if strings.Contains(scenario, "_Insurgents") {
		team := "Insurgents"
		playerTeam = &team
	}

	// Create match in database
	_, err = w.pbApp.FindFirstRecordByFilter(
		"matches",
		"server = {:server} && end_time = ''",
		map[string]any{"server": serverID},
	)

	// Only create match if one doesn't already exist
	if err != nil {
		_, err = database.CreateMatch(w.ctx, w.pbApp, serverID, &mapName, &scenario, &mapTime, playerTeam)
		if err != nil {
			log.Printf("[Catchup] Failed to create match for %s: %v", serverID, err)
			return 0, false
		}
		log.Printf("[Catchup] Created match for %s: %s (%s) at %v", serverID, mapName, scenario, mapTime)
	} else {
		log.Printf("[Catchup] Active match already exists for %s, using existing match", serverID)
	}

	// Process historical events from map event to current position
	linesProcessed := w.processHistoricalEvents(filePath, serverID, startLineNum, catchupEndOffset)

	log.Printf("[Catchup] Completed for %s: processed %d lines from %d to %d",
		serverID, linesProcessed, startLineNum, catchupEndOffset)

	return int(catchupEndOffset), true
}

// hasRecentRconLogs checks if there are recent RCON log entries (indicates SAW is active)
func (w *Watcher) hasRecentRconLogs(filePath string, threshold time.Duration) bool {
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	// Read last 100 lines to check for recent RCON activity
	scanner := bufio.NewScanner(file)
	var lastLines []string
	maxLines := 100

	for scanner.Scan() {
		lastLines = append(lastLines, scanner.Text())
		if len(lastLines) > maxLines {
			lastLines = lastLines[1:]
		}
	}

	// Check if any of the last lines contain RCON log entries within threshold
	cutoffTime := time.Now().Add(-threshold)
	timestampPattern := regexp.MustCompile(`^\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{3})\]`)

	for _, line := range lastLines {
		if strings.Contains(line, "LogRcon:") {
			// Try to extract timestamp from line
			if matches := timestampPattern.FindStringSubmatch(line); len(matches) >= 2 {
				if ts, err := parseTimestampFromLog(matches[1]); err == nil {
					if ts.After(cutoffTime) {
						return true
					}
				}
			}
		}
	}

	return false
}

// processHistoricalEvents processes events from a specific line number to an offset
// Returns the number of lines processed
func (w *Watcher) processHistoricalEvents(filePath, serverID string, startLine int, endOffset int64) int {
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("[Catchup] Failed to open file for historical processing: %v", err)
		return 0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	linesProcessed := 0

	// Read until we hit the start line
	for scanner.Scan() && lineNum < startLine {
		lineNum++
	}

	// Process lines from startLine until we reach endOffset
	for scanner.Scan() {
		currentPos, _ := file.Seek(0, 1)
		if currentPos > endOffset {
			break
		}

		line := scanner.Text()
		if err := w.parser.ParseAndProcess(w.ctx, line, serverID, filePath); err != nil {
			log.Printf("[Catchup] Error processing line %d: %v", lineNum, err)
		}

		linesProcessed++
		lineNum++
	}

	return linesProcessed
}

// parseTimestampFromLog parses a timestamp from log format (2025.10.04-15.23.38:790)
func parseTimestampFromLog(ts string) (time.Time, error) {
	colonIdx := strings.LastIndex(ts, ":")
	if colonIdx == -1 {
		return time.Time{}, fmt.Errorf("invalid timestamp format: %s", ts)
	}

	dateTimePart := ts[:colonIdx]
	msPart := ts[colonIdx+1:]

	// Parse in local timezone (log timestamps are in server's local time)
	dt, err := time.ParseInLocation("2006.01.02-15.04.05", dateTimePart, time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse datetime part: %w", err)
	}

	ms, err := strconv.Atoi(msPart)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse milliseconds: %w", err)
	}

	return dt.Add(time.Duration(ms) * time.Millisecond), nil
}
