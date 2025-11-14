package watcher

import (
	"log"
	"os"
	"time"

	"sandstorm-tracker/internal/parser"

	"github.com/pocketbase/pocketbase/core"
)

// RotationDetector handles log file rotation detection
type RotationDetector struct {
	parser *parser.LogParser
}

// NewRotationDetector creates a new rotation detector
func NewRotationDetector(parser *parser.LogParser) *RotationDetector {
	return &RotationDetector{
		parser: parser,
	}
}

// RotationResult contains the result of rotation detection
type RotationResult struct {
	Rotated     bool
	NewOffset   int
	CurrentSize int64
}

// CheckRotation detects if a log file has been rotated using timestamp or size-based detection
func (r *RotationDetector) CheckRotation(
	filePath string,
	serverID string,
	serverRecord *core.Record,
	fileInfo os.FileInfo,
) RotationResult {
	offset := serverRecord.GetInt("offset")
	savedLogFileTime := serverRecord.GetString("log_file_creation_time")
	currentSize := fileInfo.Size()

	// Extract current log file creation time
	currentLogFileTime, err := r.parser.ExtractLogFileCreationTime(filePath)
	if err != nil {
		log.Printf("Warning: Could not extract log file creation time for %s: %v", filePath, err)
	}

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

	return RotationResult{
		Rotated:     rotationDetected,
		NewOffset:   offset,
		CurrentSize: currentSize,
	}
}

// ShouldSkipProcessing determines if we should skip processing based on file state
func (r *RotationDetector) ShouldSkipProcessing(
	serverID string,
	offset int,
	currentSize int64,
	rotationDetected bool,
	savedLogFileTime string,
	currentLogFileTime time.Time,
	serverRecord *core.Record,
	pbApp core.App,
) (shouldSkip bool, reason string) {
	// If offset is 0 (new server or just after rotation), save current state
	if offset == 0 && !rotationDetected {
		log.Printf("New server %s detected, setting offset to %d bytes (waiting for first log rotation)", serverID, currentSize)
		serverRecord.Set("offset", currentSize)
		if !currentLogFileTime.IsZero() {
			serverRecord.Set("log_file_creation_time", currentLogFileTime.Format(time.RFC3339))
		}
		if err := pbApp.Save(serverRecord); err != nil {
			log.Printf("Error saving initial offset: %v", err)
		}
		return true, "new server - waiting for first rotation"
	}

	// If offset equals current size, we're at the end of the file - wait for new data
	if int64(offset) == currentSize {
		log.Printf("Server %s at current end of file (%d bytes), waiting for new log entries", serverID, currentSize)
		return true, "at end of file - waiting for new data"
	}

	return false, ""
}
