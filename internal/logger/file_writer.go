package logger

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

// FileWriter writes log entries to a file with buffering and rotation
type FileWriter struct {
	filePath   string
	maxSize    int64
	maxBackups int
	file       *os.File
	writer     *bufio.Writer
	ticker     *time.Ticker
	done       chan struct{}
	wg         sync.WaitGroup // Wait for goroutine to finish
	mu         sync.Mutex
}

// FileWriterConfig configures the file writer
type FileWriterConfig struct {
	FilePath   string        // Path to log file
	MaxSize    int64         // Max file size in bytes (0 = no rotation)
	MaxBackups int           // Number of rotated log files to keep (default: 5, 0 = unlimited)
	BufferSize int           // Buffer size in bytes (0 = unbuffered, recommended: 8192)
	FlushEvery time.Duration // How often to flush buffer (recommended: 3s)
}

// NewFileWriter creates a new file writer with buffering and rotation
func NewFileWriter(config FileWriterConfig) (*FileWriter, error) {
	// Set defaults
	if config.MaxBackups <= 0 {
		config.MaxBackups = 5
	}
	if config.BufferSize <= 0 {
		config.BufferSize = 8192
	}
	if config.FlushEvery <= 0 {
		config.FlushEvery = 3 * time.Second
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(config.FilePath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Always rotate existing log file at startup if it has content
	if err := rotateOnStartup(config.FilePath, config.MaxBackups); err != nil {
		return nil, fmt.Errorf("failed to rotate log file: %w", err)
	}

	// Open file for appending
	file, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	fw := &FileWriter{
		filePath:   config.FilePath,
		maxSize:    config.MaxSize,
		maxBackups: config.MaxBackups,
		file:       file,
		writer:     bufio.NewWriterSize(file, config.BufferSize),
		done:       make(chan struct{}),
	}

	// Start periodic flush goroutine
	if config.FlushEvery > 0 {
		fw.ticker = time.NewTicker(config.FlushEvery)
		fw.wg.Add(1)
		go fw.periodicFlush()
	}

	return fw, nil
}

// WriteRecord writes an slog.Record to the file
func (fw *FileWriter) WriteRecord(r slog.Record) error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	// Format log as JSON
	entry := map[string]interface{}{
		"time":    r.Time.Format(time.RFC3339Nano),
		"level":   r.Level.String(),
		"message": r.Message,
	}

	// Add attributes from the record
	attrCount := 0
	r.Attrs(func(a slog.Attr) bool {
		entry[a.Key] = a.Value.Any()
		attrCount++
		return true
	})

	// Debug: check if we're receiving attributes
	if attrCount == 0 && r.NumAttrs() > 0 {
		fmt.Fprintf(os.Stderr, "Warning: Record has %d attrs but Attrs() returned 0\n", r.NumAttrs())
	}

	// Marshal to JSON
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	// Write to buffer
	if _, err := fw.writer.Write(data); err != nil {
		return fmt.Errorf("failed to write log entry: %w", err)
	}

	// Write newline
	if _, err := fw.writer.WriteString("\n"); err != nil {
		return fmt.Errorf("failed to write newline: %w", err)
	}

	// Flush immediately (rely on periodic flush for batching if needed)
	// This ensures logs are visible quickly in development
	return fw.writer.Flush()
}

// WriteLog writes a PocketBase log entry to the file (for backward compatibility)
func (fw *FileWriter) WriteLog(log *core.Log) error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	// Format log as JSON
	entry := map[string]interface{}{
		"time":    log.Created.Time().Format(time.RFC3339Nano),
		"level":   logLevelToString(log.Level),
		"message": log.Message,
	}

	// Add data fields if present
	if len(log.Data) > 0 {
		for k, v := range log.Data {
			entry[k] = v
		}
	}

	// Marshal to JSON
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	// Write to buffer
	if _, err := fw.writer.Write(data); err != nil {
		return fmt.Errorf("failed to write log entry: %w", err)
	}

	// Write newline
	if _, err := fw.writer.WriteString("\n"); err != nil {
		return fmt.Errorf("failed to write newline: %w", err)
	}

	// Flush immediately (rely on periodic flush for batching if needed)
	// This ensures logs are visible quickly in development
	return fw.writer.Flush()
}

// Flush flushes the buffer to disk
func (fw *FileWriter) Flush() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	return fw.writer.Flush()
}

// Close closes the file writer
func (fw *FileWriter) Close() error {
	// Stop ticker and signal goroutine to stop
	if fw.ticker != nil {
		fw.ticker.Stop()
		close(fw.done)
		fw.wg.Wait() // Wait for goroutine to finish
	}

	// Final flush
	fw.mu.Lock()
	if err := fw.writer.Flush(); err != nil {
		fw.mu.Unlock()
		return err
	}
	fw.mu.Unlock()

	return fw.file.Close()
}

// periodicFlush flushes the buffer periodically
func (fw *FileWriter) periodicFlush() {
	defer fw.wg.Done()
	for {
		select {
		case <-fw.ticker.C:
			_ = fw.Flush() // Ignore errors during periodic flush
		case <-fw.done:
			return
		}
	}
}

// logLevelToString converts PocketBase log level to string
func logLevelToString(level int) string {
	switch level {
	case -4:
		return "DEBUG"
	case 0:
		return "INFO"
	case 4:
		return "WARN"
	case 8:
		return "ERROR"
	default:
		return fmt.Sprintf("LEVEL_%d", level)
	}
}

// rotateOnStartup rotates the log file at startup if it exists and has content
// Rotation strategy: app.log -> app.log.1 -> app.log.2 -> ... (keeps last maxBackups rotations)
func rotateOnStartup(filePath string, maxBackups int) error {
	// Check if file exists
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist, nothing to rotate
		}
		return fmt.Errorf("failed to stat log file: %w", err)
	}

	// Don't rotate empty files
	if info.Size() == 0 {
		return nil
	}

	// Default to 5 backups if not specified
	if maxBackups <= 0 {
		maxBackups = 5
	}

	// Remove .log extension for rotation (app.log -> app)
	basePath := filePath[:len(filePath)-4] // Remove ".log"

	// Rotate existing backups (app.4.log -> app.5.log, app.3.log -> app.4.log, etc.)
	for i := maxBackups - 1; i >= 1; i-- {
		oldPath := fmt.Sprintf("%s.%d.log", basePath, i)
		newPath := fmt.Sprintf("%s.%d.log", basePath, i+1)
		if _, err := os.Stat(oldPath); err == nil {
			os.Rename(oldPath, newPath) // Overwrite oldest backup
		}
	}

	// Rotate current log file to .1.log (app.log -> app.1.log)
	backupPath := fmt.Sprintf("%s.1.log", basePath)
	if err := os.Rename(filePath, backupPath); err != nil {
		return fmt.Errorf("failed to rotate log file: %w", err)
	}

	return nil
}
