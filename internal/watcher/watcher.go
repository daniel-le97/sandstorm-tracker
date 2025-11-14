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

	"sandstorm-tracker/internal/a2s"
	"sandstorm-tracker/internal/config"
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
	rconPool         *rcon.ClientPool
	a2sPool          A2SQuerier
	serverConfigs    map[string]config.ServerConfig
	serverQueues     map[string]chan string // Per-server work queues for sequential event processing
	serverQueuesMu   sync.RWMutex
	stateTracker     *ServerStateTracker
	rotationDetector *RotationDetector
	catchupProcessor *CatchupProcessor
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
		watcher:          watcher,
		parser:           logParser,
		pbApp:            pbApp,
		ctx:              ctx,
		cancel:           cancel,
		rconPool:         rconPool,
		a2sPool:          a2sPool,
		serverConfigs:    scMap,
		serverQueues:     make(map[string]chan string),
		stateTracker:     NewServerStateTracker(inactivityDuration),
		rotationDetector: NewRotationDetector(logParser),
		catchupProcessor: NewCatchupProcessor(logParser, a2sPool, scMap, pbApp, ctx),
	}

	// Start inactivity monitor
	w.startInactivityMonitor()

	return w, nil
}

// OnServerActive sets a callback to be called when a server becomes active (log rotation detected)
func (w *Watcher) OnServerActive(callback func(serverID string)) {
	w.stateTracker.SetCallbacks(callback, nil)
}

// OnServerInactive sets a callback to be called when a server becomes inactive (no activity for 10s)
func (w *Watcher) OnServerInactive(callback func(serverID string)) {
	var activeCallback func(string)
	w.stateTracker.callbacksMu.RLock()
	activeCallback = w.stateTracker.onServerActive
	w.stateTracker.callbacksMu.RUnlock()
	w.stateTracker.SetCallbacks(activeCallback, callback)
}

// IsServerActive returns whether a server is currently active (has recent log activity)
func (w *Watcher) IsServerActive(serverID string) bool {
	return w.stateTracker.IsActive(serverID)
}

// GetServerLastActivity returns the last activity time for a server
func (w *Watcher) GetServerLastActivity(serverID string) (time.Time, bool) {
	return w.stateTracker.GetLastActivity(serverID)
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
				serverID := w.extractServerIDFromPath(logFilePath)
				w.enqueueFileEvent(serverID, logFilePath)
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

		serverID := w.extractServerIDFromPath(absPath)
		w.enqueueFileEvent(serverID, absPath)
		log.Printf("Watching file: %s", absPath)
	}
	return nil
}

// Start begins watching for file changes
func (w *Watcher) Start() {
	w.wg.Add(1)
	go w.watchLoop()
}

// Stop stops the watcher and all server workers
func (w *Watcher) Stop() {
	// Cancel context to signal all workers to stop
	w.cancel()

	// Close all server queues to unblock workers
	w.serverQueuesMu.Lock()
	for serverID, queue := range w.serverQueues {
		close(queue)
		log.Printf("Closed queue for server %s", serverID)
	}
	w.serverQueuesMu.Unlock()

	// Wait for all workers to finish
	w.wg.Wait()

	// Close the file watcher
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
			serverID := w.extractServerIDFromPath(event.Name)
			w.enqueueFileEvent(serverID, event.Name)
		}
	}
}

// enqueueFileEvent sends a file event to the appropriate server's work queue
// If the server doesn't have a worker yet, one is created
func (w *Watcher) enqueueFileEvent(serverID, filePath string) {
	w.serverQueuesMu.Lock()
	queue, exists := w.serverQueues[serverID]
	if !exists {
		// Create a buffered channel for this server
		queue = make(chan string, 100)
		w.serverQueues[serverID] = queue
		// Start a worker for this server
		w.wg.Add(1)
		go w.serverWorker(serverID, queue)
	}
	w.serverQueuesMu.Unlock()

	// Non-blocking send to avoid blocking the fsnotify loop
	select {
	case queue <- filePath:
		// Successfully enqueued
	default:
		log.Printf("Warning: Queue full for server %s, dropping event for %s", serverID, filePath)
	}
}

// serverWorker processes file events sequentially for a single server
func (w *Watcher) serverWorker(serverID string, queue chan string) {
	defer w.wg.Done()
	log.Printf("Started worker for server %s", serverID)

	for {
		select {
		case <-w.ctx.Done():
			log.Printf("Stopping worker for server %s", serverID)
			return
		case filePath, ok := <-queue:
			if !ok {
				log.Printf("Queue closed for server %s", serverID)
				return
			}
			w.processFile(filePath)
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
		catchupOffset, shouldCatchup := w.catchupProcessor.CheckStartupCatchup(filePath, serverID)
		if shouldCatchup {
			offset = catchupOffset
			// Save the offset so we start from here
			serverRecord.Set("offset", offset)
			if err := w.pbApp.Save(serverRecord); err != nil {
				log.Printf("Error saving catch-up offset: %v", err)
			}
		}
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

	// Check for rotation using rotation detector
	rotationResult := w.rotationDetector.CheckRotation(filePath, serverID, serverRecord, fileInfo)
	offset = rotationResult.NewOffset

	// Extract current log file creation time for saving
	currentLogFileTime, err := w.parser.ExtractLogFileCreationTime(filePath)
	if err != nil {
		log.Printf("Warning: Could not extract log file creation time for %s: %v", filePath, err)
	}

	// Check if we should skip processing
	if shouldSkip, _ := w.rotationDetector.ShouldSkipProcessing(
		serverID, offset, rotationResult.CurrentSize, rotationResult.Rotated,
		savedLogFileTime, currentLogFileTime, serverRecord, w.pbApp,
	); shouldSkip {
		return
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

		// Mark server as active and update activity time
		w.stateTracker.MarkActive(serverID)
		w.stateTracker.UpdateActivity(serverID)
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
	w.stateTracker.CheckInactiveServers()
}
