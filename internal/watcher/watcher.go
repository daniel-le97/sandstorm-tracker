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

	"sandstorm-tracker/db"
	generated "sandstorm-tracker/db/generated"
	"sandstorm-tracker/internal/events"

	"github.com/fsnotify/fsnotify"
)

type FileWatcher struct {
	watcher     *fsnotify.Watcher
	parser      *events.EventParser
	db          *db.DatabaseService
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	fileOffsets map[string]int64 // Track file read positions
	mu          sync.RWMutex     // Protect fileOffsets map
}

func NewFileWatcher(dbService *db.DatabaseService) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	fw := &FileWatcher{
		watcher:     watcher,
		parser:      events.NewEventParser(),
		db:          dbService,
		ctx:         ctx,
		cancel:      cancel,
		fileOffsets: make(map[string]int64),
	}

	return fw, nil
}

func boolPtr(b bool) *bool {
	return &b
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
		if strings.HasSuffix(event.Name, ".log") || strings.Contains(event.Name, "log") {
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
	for scanner.Scan() {
		line := scanner.Text()
		newOffset += int64(len(line) + 1)
		event, err := fw.parser.ParseLine(line, serverID)
		if err != nil {
			log.Printf("Failed to parse line: %v", err)
			continue
		}
		if event != nil {
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
	if lineCount > 0 {
		log.Printf("Processed %d new events from %s", lineCount, filepath.Base(filePath))
	}
}

func (fw *FileWatcher) handleGameEvent(event *events.GameEvent, filePath string, serverID string) {
	log.Printf("Event from %s (Server: %s): Type=%v, Time=%v, Player=%s",
		filepath.Base(filePath),
		serverID,
		event.Type,
		event.Timestamp.Format("15:04:05"),
		getPlayerName(event))
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
		if err := fw.handleChatCommand(ctx, event, serverDBID); err != nil {
			log.Printf("Error handling chat command: %v", err)
		}
	default:
		log.Printf("Event details: %+v", event.Data)
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
	return fw.db.GetQueries().UpsertServer(ctx, generated.UpsertServerParams{
		ExternalID: serverID,
		Name:       serverID,
		Path:       &logPath,
	})
}

func (fw *FileWatcher) handleKillEvent(ctx context.Context, event *events.GameEvent, serverDBID int64) error {
	killersData, ok := event.Data["killers"].([]events.Killer)
	if !ok {
		return fmt.Errorf("invalid killers data in event")
	}
	victimName, _ := event.Data["victim_name"].(string)
	victimSteamID, _ := event.Data["victim_steam_id"].(string)
	weapon, _ := event.Data["weapon"].(string)
	killType, _ := event.Data["kill_type"].(string)
	isMultiKill, _ := event.Data["multi_kill"].(bool)
	if victimSteamID != "INVALID" {
		_, err := fw.db.GetQueries().UpsertPlayer(ctx, generated.UpsertPlayerParams{
			ExternalID: victimSteamID,
			Name:       victimName,
		})
		if err != nil {
			return fmt.Errorf("failed to upsert victim: %w", err)
		}
	}
	if victimSteamID == "INVALID" {
		for _, killer := range killersData {
			if killer.SteamID != "INVALID" {
				killerID, err := fw.db.GetQueries().UpsertPlayer(ctx, generated.UpsertPlayerParams{
					ExternalID: killer.SteamID,
					Name:       killer.Name,
				})
				if err != nil {
					return fmt.Errorf("failed to upsert killer %s: %w", killer.Name, err)
				}
				var killTypeInt int64
				switch killType {
				case "regular":
					killTypeInt = 0
				case "suicide":
					killTypeInt = 1
				case "team_kill":
					killTypeInt = 2
				default:
					killTypeInt = 0
				}
				err = fw.db.GetQueries().InsertKill(ctx, generated.InsertKillParams{
					KillerID:   &killerID,
					VictimName: &victimName,
					ServerID:   serverDBID,
					WeaponName: &weapon,
					KillType:   killTypeInt,
					MatchID:    nil,
					CreatedAt:  &event.Timestamp,
				})
				if err != nil {
					return fmt.Errorf("failed to insert kill for killer %s: %w", killer.Name, err)
				}
				log.Printf("AI kill recorded: %s killed %s with %s%s",
					killer.Name, victimName, weapon,
					func() string {
						if isMultiKill {
							return " (multi-kill)"
						} else {
							return ""
						}
					}())
			}
		}
	} else {
		if len(killersData) == 1 {
			killer := killersData[0]
			if killer.SteamID != "INVALID" {
				killerID, err := fw.db.GetQueries().UpsertPlayer(ctx, generated.UpsertPlayerParams{
					ExternalID: killer.SteamID,
					Name:       killer.Name,
				})
				if err != nil {
					return fmt.Errorf("failed to upsert killer %s: %w", killer.Name, err)
				}
				var killTypeInt int64
				switch killType {
				case "regular":
					killTypeInt = 0
				case "suicide":
					killTypeInt = 1
				case "team_kill":
					killTypeInt = 2
				default:
					killTypeInt = 0
				}
				err = fw.db.GetQueries().InsertKill(ctx, generated.InsertKillParams{
					KillerID:   &killerID,
					VictimName: &victimName,
					ServerID:   serverDBID,
					WeaponName: &weapon,
					KillType:   killTypeInt,
					MatchID:    nil,
					CreatedAt:  &event.Timestamp,
				})
				if err != nil {
					return fmt.Errorf("failed to insert kill: %w", err)
				}
				log.Printf("Player kill recorded: %s killed %s with %s", killer.Name, victimName, weapon)
			}
		}
	}
	return nil
}

func (fw *FileWatcher) handlePlayerJoin(ctx context.Context, event *events.GameEvent, serverDBID int64) error {
	playerName, _ := event.Data["player_name"].(string)
	steamID, _ := event.Data["steam_id"].(string)
	if steamID == "INVALID" {
		return nil
	}
	if steamID == "" {
		log.Printf("Skipping player join for %s - no Steam ID provided", playerName)
		return nil
	}
	_, err := fw.db.GetQueries().UpsertPlayer(ctx, generated.UpsertPlayerParams{
		ExternalID: steamID,
		Name:       playerName,
	})
	return err
}

func (fw *FileWatcher) handlePlayerLeave(ctx context.Context, event *events.GameEvent, serverDBID int64) error {
	playerName, _ := event.Data["player_name"].(string)
	steamID, _ := event.Data["steam_id"].(string)
	if steamID == "INVALID" {
		return nil
	}
	if steamID == "" {
		log.Printf("Player %s left - no Steam ID to track", playerName)
		return nil
	}
	log.Printf("Player %s [%s] left the server", playerName, steamID)
	return nil
}

func (fw *FileWatcher) handleChatCommand(ctx context.Context, event *events.GameEvent, serverDBID int64) error {
	playerName, _ := event.Data["player_name"].(string)
	command, _ := event.Data["command"].(string)
	log.Printf("Chat command from %s: %s", playerName, command)
	return nil
}
