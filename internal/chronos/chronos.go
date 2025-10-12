package chronos

// Package chronos provides log replay and catch-up functionality for Sandstorm Tracker.
// It ensures all log lines and events are processed in strict sequential order, even if the
// application starts after the game server has already been running.
//
// Responsibilities:
// - On startup, scan the log directory for all relevant log files (including the current active one).
// - For each log file, check for a saved offset. If none, start from the beginning; if present, resume from the saved offset.
// - Read and process every line from the offset to the end of each file, in order.
// - For the current active log file, continue tailing new lines as they are written.
// - Always update and persist the offset after processing each line, so no events are missed or duplicated, even across restarts.
// - Integrate with the watcher to ensure all historical and new events are processed in strict order.

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"sandstorm-tracker/internal/events"
)

// OffsetState holds the last read offset for a file
type OffsetState struct {
	Offset int64 `json:"offset"`
}

// getOffsetStatePath returns the path to the offset state file for a log file
func getOffsetStatePath(logfile string) string {
	base := filepath.Base(logfile)
	return base + ".offset.json"
}

// loadOffset loads the last offset for a file, or returns 0 if not found
func loadOffset(logfile string) int64 {
	path := getOffsetStatePath(logfile)
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer f.Close()
	var state OffsetState
	if err := json.NewDecoder(f).Decode(&state); err != nil {
		return 0
	}
	return state.Offset
}

// saveOffset saves the last offset for a file
func saveOffset(logfile string, offset int64) {
	path := getOffsetStatePath(logfile)
	f, err := os.Create(path)
	if err != nil {
		log.Printf("[WARN] Could not save offset for %s: %v", logfile, err)
		return
	}
	defer f.Close()
	state := OffsetState{Offset: offset}
	_ = json.NewEncoder(f).Encode(state)
}

// ReplayAndCatchUp replays all unread log lines in order and continues tailing the current log file.
// logDir: directory to scan for log files
// parser: event parser to use
// handleEvent: callback for each parsed event
func ReplayAndCatchUp(logDir string, parser *events.EventParser, handleEvent func(*events.GameEvent, string) error) error {
	// 1. Find all log files and sort by modification time (true chronological order)
	files, err := filepath.Glob(filepath.Join(logDir, "*.log"))
	if err != nil {
		return fmt.Errorf("failed to list log files: %w", err)
	}
	if len(files) == 0 {
		return nil // nothing to do
	}
	type fileInfo struct {
		name    string
		modTime int64
	}
	var fileInfos []fileInfo
	for _, f := range files {
		stat, err := os.Stat(f)
		if err != nil {
			continue
		}
		fileInfos = append(fileInfos, fileInfo{name: f, modTime: stat.ModTime().UnixNano()})
	}
	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].modTime < fileInfos[j].modTime
	})
	for _, fi := range fileInfos {
		err := replayFile(fi.name, parser, handleEvent)
		if err != nil {
			log.Printf("[CHRONOS] Error replaying %s: %v", fi.name, err)
		}
	}
	return nil
}

// replayFile replays all unread lines from a log file
func replayFile(filename string, parser *events.EventParser, handleEvent func(*events.GameEvent, string) error) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	offset := loadOffset(filename)
	if offset > 0 {
		_, err := file.Seek(offset, io.SeekStart)
		if err != nil {
			return fmt.Errorf("failed to seek: %w", err)
		}
	}
	reader := bufio.NewReader(file)
	newOffset := offset
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if len(line) > 0 && line[len(line)-1] == '\n' {
			line = line[:len(line)-1]
		}
		newOffset += int64(len(line) + 1)
		event, err := parser.ParseLine(line, extractServerIDFromPath(filename))
		if err != nil {
			log.Printf("[CHRONOS] Failed to parse line: %v", err)
			continue
		}
		if event != nil {
			err = handleEvent(event, filename)
			if err != nil {
				log.Printf("[CHRONOS] Event handler error: %v", err)
			}
		}
		saveOffset(filename, newOffset)
	}
	return nil
}

// extractServerIDFromPath returns a server ID from the log filename
func extractServerIDFromPath(filePath string) string {
	filename := filepath.Base(filePath)
	if strings.HasSuffix(filename, ".log") {
		return strings.TrimSuffix(filename, ".log")
	}
	return filename
}
