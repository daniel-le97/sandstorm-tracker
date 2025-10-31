package app

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"sandstorm-tracker/internal/db"
	generated "sandstorm-tracker/internal/db/generated"
	"sandstorm-tracker/internal/rcon"

	"github.com/fsnotify/fsnotify"
)

// Watcher monitors log files and processes events
type Watcher struct {
	watcher       *fsnotify.Watcher
	parser        *LogParser
	db            *db.DatabaseService
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	fileOffsets   map[string]int64
	mu            sync.RWMutex
	rconClients   map[string]*rcon.RconClient
	rconMu        sync.Mutex
	serverConfigs map[string]ServerConfig
	offsetsPath   string
}

// NewWatcher creates a new file watcher
func NewWatcher(dbService *db.DatabaseService, serverConfigs []ServerConfig) (*Watcher, error) {
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

	cwd, err := os.Getwd()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}

	w := &Watcher{
		watcher:       watcher,
		parser:        NewLogParser(dbService.GetQueries()),
		db:            dbService,
		ctx:           ctx,
		cancel:        cancel,
		offsetsPath:   filepath.Join(cwd, "sandstorm-tracker-offsets.json"),
		fileOffsets:   make(map[string]int64),
		rconClients:   make(map[string]*rcon.RconClient),
		serverConfigs: scMap,
	}
	w.loadOffsets()

	return w, nil
}

func (w *Watcher) loadOffsets() {
	if _, err := os.Stat(w.offsetsPath); os.IsNotExist(err) {
		file, err := os.Create(w.offsetsPath)
		if err != nil {
			log.Printf("Failed to create offsets file: %v", err)
			return
		}
		file.Close()
	}

	data, err := os.ReadFile(w.offsetsPath)
	if err != nil {
		log.Printf("No previous file offsets found, starting fresh.")
		return
	}

	var offsets map[string]int64
	if err := json.Unmarshal(data, &offsets); err != nil {
		log.Printf("Failed to unmarshal file offsets: %v", err)
		return
	}

	w.mu.Lock()
	w.fileOffsets = offsets
	w.mu.Unlock()
	log.Printf("Loaded file offsets from %s", w.offsetsPath)
}

func (w *Watcher) saveOffsets() {
	w.mu.RLock()
	data, err := json.MarshalIndent(w.fileOffsets, "", "  ")
	w.mu.RUnlock()
	if err != nil {
		log.Printf("Failed to marshal file offsets: %v", err)
		return
	}
	if err := os.WriteFile(w.offsetsPath, data, 0644); err != nil {
		log.Printf("Failed to write file offsets: %v", err)
	}
}

// AddPath adds a file or directory to watch
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
		err = w.watcher.Add(absPath)
		if err != nil {
			return fmt.Errorf("failed to watch directory %s: %w", absPath, err)
		}
		log.Printf("Watching directory: %s", absPath)
	} else {
		err = w.watcher.Add(absPath)
		if err != nil {
			return fmt.Errorf("failed to watch file %s: %w", absPath, err)
		}
		w.mu.Lock()
		w.fileOffsets[absPath] = 0
		w.mu.Unlock()
		go w.processFile(absPath)
		log.Printf("Watching file: %s (starting from offset %d)", absPath, 0)
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
	w.saveOffsets()
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

	w.mu.Lock()
	offset, exists := w.fileOffsets[filePath]
	w.mu.Unlock()
	if !exists {
		offset = 0
		w.mu.Lock()
		w.fileOffsets[filePath] = 0
		w.mu.Unlock()
	}

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Error opening file %s: %v", filePath, err)
		return
	}
	defer file.Close()

	if _, err := file.Seek(offset, 0); err != nil {
		log.Printf("Error seeking to offset %d in %s: %v", offset, filePath, err)
		return
	}

	// Get server DB ID
	serverDBID, err := w.getOrCreateServerDBID(serverID, filePath)
	if err != nil {
		log.Printf("Failed to get server DB ID: %v", err)
		return
	}

	scanner := bufio.NewScanner(file)
	linesProcessed := 0

	for scanner.Scan() {
		line := scanner.Text()

		// Parse and process directly - no intermediate structs
		if err := w.parser.ParseAndProcess(w.ctx, line, serverDBID); err != nil {
			log.Printf("Error processing line: %v", err)
		}

		linesProcessed++
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error scanning file %s: %v", filePath, err)
		return
	}

	newOffset, _ := file.Seek(0, 1)
	w.mu.Lock()
	w.fileOffsets[filePath] = newOffset
	w.mu.Unlock()

	if linesProcessed > 0 {
		w.saveOffsets()
		log.Printf("Processed %d lines from %s (new offset: %d)", linesProcessed, filePath, newOffset)
	}
}

func (w *Watcher) extractServerIDFromPath(filePath string) string {
	base := filepath.Base(filePath)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

func (w *Watcher) getOrCreateServerDBID(serverID, logPath string) (int64, error) {
	ctx := context.Background()
	queries := w.db.GetQueries()

	server, err := queries.GetServerByExternalID(ctx, serverID)
	if err != nil {
		// Create server
		server, err = queries.CreateServer(ctx, generated.CreateServerParams{
			ExternalID: serverID,
			Name:       serverID,
			Path:       &logPath,
		})
		if err != nil {
			return 0, fmt.Errorf("failed to create server: %w", err)
		}
		log.Printf("Created server in DB: %s (ID: %d)", serverID, server.ID)
	}

	return server.ID, nil
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
