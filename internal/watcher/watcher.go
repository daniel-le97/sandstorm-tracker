package watcher

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"sandstorm-tracker/internal/db"
	generated "sandstorm-tracker/internal/db/generated"
	"sandstorm-tracker/internal/config"
	"sandstorm-tracker/internal/events"
	"sandstorm-tracker/internal/rcon"
	"sandstorm-tracker/internal/utils"

	"github.com/fsnotify/fsnotify"
)

type FileWatcher struct {
	watcher       *fsnotify.Watcher
	parser        *events.EventParser
	db            *db.DatabaseService
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	fileOffsets   map[string]int64               // Track file read positions
	mu            sync.RWMutex                   // Protect fileOffsets map
	rconClients   map[string]*rcon.RconClient    // serverID -> RCON client
	rconMu        sync.Mutex                     // Protect rconClients
	serverConfigs map[string]config.ServerConfig // serverID -> ServerConfig
	offsetsPath   string // Path to save offsets
}

func NewFileWatcher(dbService *db.DatabaseService, serverConfigs []config.ServerConfig) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	scMap := make(map[string]config.ServerConfig)
	for _, sc := range serverConfigs {
		// Use the log file stem (without .log) as the key, matching serverID in events
		serverID, err := utils.GetServerIdFromPath(sc.LogPath)
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

	fw := &FileWatcher{
		watcher:       watcher,
		parser:        events.NewEventParser(),
		db:            dbService,
		ctx:           ctx,
		cancel:        cancel,
		offsetsPath:   filepath.Join(cwd, "sandstorm-tracker-offsets.json"),
		fileOffsets:   make(map[string]int64),
		rconClients:   make(map[string]*rcon.RconClient),
		serverConfigs: scMap,
	}
	fw.loadOffsets()

	return fw, nil
}
// Call this in your constructor after initializing fileOffsets
func (fw *FileWatcher) loadOffsets() {
	// check if file exists, or create it
	if _, err := os.Stat(fw.offsetsPath); os.IsNotExist(err) {
		file, err := os.Create(fw.offsetsPath)
		if err != nil {
			log.Printf("Failed to create offsets file: %v", err)
			return
		}
		file.Close()
	}
    data, err := os.ReadFile(fw.offsetsPath)
    if err != nil {
        log.Printf("No previous file offsets found, starting fresh.")
        return
    }
    var offsets map[string]int64
    if err := json.Unmarshal(data, &offsets); err != nil {
        log.Printf("Failed to unmarshal file offsets: %v", err)
        return
    }
    fw.mu.Lock()
    fw.fileOffsets = offsets
    fw.mu.Unlock()
    log.Printf("Loaded file offsets from %s", fw.offsetsPath)
}

func (fw *FileWatcher) saveOffsets() {
    fw.mu.RLock()
    data, err := json.MarshalIndent(fw.fileOffsets, "", "  ")
    fw.mu.RUnlock()
    if err != nil {
        log.Printf("Failed to marshal file offsets: %v", err)
        return
    }
    if err := os.WriteFile(fw.offsetsPath, data, 0644); err != nil {
        log.Printf("Failed to write file offsets: %v", err)
    }
}

func (fw *FileWatcher) AddPath(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for %s: %w", path, err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("path %s does not exist: %w", absPath, err)
	}

	if info.IsDir() {
		err = fw.watcher.Add(absPath)
		if err != nil {
			return fmt.Errorf("failed to watch directory %s: %w", absPath, err)
		}
		log.Printf("Watching directory: %s", absPath)
	} else {
		err = fw.watcher.Add(absPath)
		if err != nil {
			return fmt.Errorf("failed to watch file %s: %w", absPath, err)
		}
		fw.mu.Lock()
		fw.fileOffsets[absPath] = 0
		fw.mu.Unlock()
		go fw.processFile(absPath)
		log.Printf("Watching file: %s (starting from offset %d)", absPath, 0)
	}
	return nil
}

func (fw *FileWatcher) Start() {
	fw.wg.Add(1)
	go fw.watchLoop()
}

func (fw *FileWatcher) Stop() {
	fw.cancel()
	fw.wg.Wait()
	fw.watcher.Close()
}

func (fw *FileWatcher) watchLoop() {
	defer fw.wg.Done()
	for {
		select {
		case <-fw.ctx.Done():
			return
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}
			fw.handleFileEvent(event)
		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("File watcher error: %v", err)
		}
	}
}

func (fw *FileWatcher) handleFileEvent(event fsnotify.Event) {
	if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
		// Only process main log files, ignore backups
		name := filepath.Base(event.Name)
		if strings.HasSuffix(name, ".log") && !strings.Contains(name, "-backup-") {
			go fw.processFile(event.Name)
		}
	}
}

func (fw *FileWatcher) processFile(filePath string) {
	serverID := extractServerIDFromPath(filePath)
	fw.mu.Lock()
	offset, exists := fw.fileOffsets[filePath]
	fw.mu.Unlock()
	if !exists {
		offset = 0
		fw.mu.Lock()
		fw.fileOffsets[filePath] = 0
		fw.mu.Unlock()
		log.Printf("Started watching new server log: %s (Server ID: %s)", filepath.Base(filePath), serverID)
	}
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Failed to open file %s: %v", filePath, err)
		return
	}
	defer file.Close()
	_, err = file.Seek(offset, 0)
	if err != nil {
		log.Printf("Failed to seek to offset %d in file %s: %v", offset, filePath, err)
		return
	}
	scanner := bufio.NewScanner(file)
	newOffset := offset
	lineCount := 0
	isRealtime := offset != 0
	for scanner.Scan() {
		line := scanner.Text()
		newOffset += int64(len(line) + 1)
		event, err := fw.parser.ParseLine(line, serverID)
		if err != nil {
			log.Printf("Failed to parse line: %v", err)
			continue
		}
		if event != nil {
			// TODO - handle non-realtime chat commands
			// we do not want to handle a chat command unless the event is realtime
			// so we only handle it if we are not at the start of the file
			if event.Type == events.EventChatCommand {
				if isRealtime {
					fw.handleGameEvent(event, filePath, serverID)
				}
				lineCount++
				continue
			}
			fw.handleGameEvent(event, filePath, serverID)
			lineCount++
		}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("Error reading file %s: %v", filePath, err)
		return
	}
	fw.mu.Lock()
	fw.fileOffsets[filePath] = newOffset
	fw.mu.Unlock()
	fw.saveOffsets()
	if lineCount > 0 {
		// log.Printf("Processed %d new events from %s", lineCount, filepath.Base(filePath))
	}
}

func (fw *FileWatcher) handleGameEvent(event *events.GameEvent, filePath string, serverID string) {
	// log.Printf("Event from %s (Server: %s): Type=%v, Time=%v, Player=%s",
	// 	filepath.Base(filePath),
	// 	serverID,
	// 	event.Type,
	// 	event.Timestamp.Format("15:04:05"),
	// 	getPlayerName(event))
	ctx := context.Background()
	serverDBID, err := fw.ensureServer(ctx, serverID, filePath)
	if err != nil {
		log.Printf("Error ensuring server exists: %v", err)
		return
	}
	switch event.Type {
	case events.EventPlayerKill, events.EventSuicide, events.EventFriendlyFire:
		if err := fw.handleKillEvent(ctx, event, serverDBID); err != nil {
			log.Printf("Error handling kill event: %v", err)
		}
	case events.EventPlayerJoin:
		if err := fw.handlePlayerJoin(ctx, event, serverDBID); err != nil {
			log.Printf("Error handling player join: %v", err)
		}
	case events.EventPlayerLeave:
		if err := fw.handlePlayerLeave(ctx, event, serverDBID); err != nil {
			log.Printf("Error handling player leave: %v", err)
		}
	case events.EventChatCommand:
		log.Printf("processing chat: %s", event.RawLogLine)
		// log.Printf("Event from %s (Server: %s): Type=%v, Time=%v, Player=%s",
		if err := fw.handleChatCommand(ctx, event, serverDBID); err != nil {
			log.Printf("Error handling chat command: %v", err)
		}
	default:
		return
		// log.Printf("Event details: %+v", event.Data)
	}
}

func getPlayerName(event *events.GameEvent) string {
	switch event.Type {
	case events.EventPlayerKill:
		if killerName, ok := event.Data["killer_name"].(string); ok {
			return killerName
		}
	case events.EventPlayerJoin, events.EventPlayerLeave, events.EventChatCommand:
		if playerName, ok := event.Data["player_name"].(string); ok {
			return playerName
		}
	}
	return "N/A"
}

func extractServerIDFromPath(filePath string) string {
	filename := filepath.Base(filePath)
	if strings.HasSuffix(filename, ".log") {
		return strings.TrimSuffix(filename, ".log")
	}
	return filename
}

func (fw *FileWatcher) ensureServer(ctx context.Context, serverID, logPath string) (int64, error) {
	queries := fw.db.GetQueries()
	// Try to get by external_id first
	server, err := queries.GetServerByExternalID(ctx, serverID)
	if err == nil {
		return server.ID, nil
	}
	// If not found, create
	created, err := queries.CreateServer(ctx, generated.CreateServerParams{
		ExternalID: serverID,
		Name:       serverID,
		Path:       &logPath,
	})
	if err != nil {
		return 0, err
	}
	return created.ID, nil
}

