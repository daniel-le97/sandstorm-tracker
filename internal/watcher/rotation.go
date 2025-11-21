package watcher

import (
	"log/slog"
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
	logger *slog.Logger,
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
		logger.Debug("Could not extract log file creation time", "filePath", filePath, "error", err)
	}

	// Check if log file was rotated by comparing creation times
	rotationDetected := false
	if savedLogFileTime != "" && !currentLogFileTime.IsZero() {
		savedTime, err := time.Parse(time.RFC3339, savedLogFileTime)
		if err == nil && !currentLogFileTime.Equal(savedTime) {
			logger.Debug("Log rotation detected", "serverID", serverID, "oldTime", savedTime.Format("2006-01-02 15:04:05"), "newTime", currentLogFileTime.Format("2006-01-02 15:04:05"))
			rotationDetected = true
			offset = 0
		}
	}

	// Fallback: If current size is less than offset, log file was rotated (truncated/reset)
	if !rotationDetected && currentSize < int64(offset) {
		logger.Debug("Log rotation detected", "serverID", serverID, "reason", "file size < offset", "fileSize", currentSize, "offset", offset)
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
	logger *slog.Logger,
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
		logger.Debug("New server detected", "serverID", serverID, "offset", currentSize)
		serverRecord.Set("offset", currentSize)
		if !currentLogFileTime.IsZero() {
			serverRecord.Set("log_file_creation_time", currentLogFileTime.Format(time.RFC3339))
		}
		if err := pbApp.Save(serverRecord); err != nil {
			logger.Debug("Error saving initial offset", "serverID", serverID, "error", err)
		}
		return true, "new server - waiting for first rotation"
	}

	// If offset equals current size, we're at the end of the file - wait for new data
	if int64(offset) == currentSize {
		logger.Debug("Server at end of file", "serverID", serverID, "fileSize", currentSize)
		return true, "at end of file - waiting for new data"
	}

	return false, ""
}
