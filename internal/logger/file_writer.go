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

	// Check if log rotation is needed before opening
	if config.MaxSize > 0 {
		if err := rotateIfNeeded(config.FilePath, config.MaxSize, config.MaxBackups); err != nil {
			return nil, fmt.Errorf("failed to rotate log file: %w", err)
		}
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
		fmt.Fprintf(os.Stderr, "FileWriter: Starting periodic flush goroutine (every %v)\n", config.FlushEvery)
		fw.wg.Add(1)
		go fw.periodicFlush()
	}

	return fw, nil
}

// WriteRecord writes an slog.Record to the file
func (fw *FileWriter) WriteRecord(r slog.Record) error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	// Debug: verify this is being called
	fmt.Fprintf(os.Stderr, "FileWriter.WriteRecord called: %s\n", r.Message)

	// Format log as JSON
	entry := map[string]interface{}{
		"time":    r.Time.Format(time.RFC3339Nano),
		"level":   r.Level.String(),
		"message": r.Message,
	}

	// Add attributes from the record
	r.Attrs(func(a slog.Attr) bool {
		entry[a.Key] = a.Value.Any()
		return true
	})

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

	// Flush if buffer is getting full
	if fw.writer.Available() < 1024 {
		return fw.writer.Flush()
	}

	return nil
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

	// Flush if buffer is getting full
	if fw.writer.Available() < 1024 {
		return fw.writer.Flush()
	}

	return nil
}

// Flush flushes the buffer to disk
func (fw *FileWriter) Flush() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	buffered := fw.writer.Buffered()
	fmt.Fprintf(os.Stderr, "FileWriter.Flush: %d bytes buffered\n", buffered)
	return fw.writer.Flush()
}

// Close closes the file writer
func (fw *FileWriter) Close() error {
	fmt.Fprintf(os.Stderr, "FileWriter: Close() called, stopping flush goroutine\n")
	if fw.ticker != nil {
		fw.ticker.Stop()
		close(fw.done)
		fw.wg.Wait() // Wait for goroutine to finish
		fmt.Fprintf(os.Stderr, "FileWriter: Flush goroutine stopped\n")
	}
	fw.Flush()
	fmt.Fprintf(os.Stderr, "FileWriter: Final flush completed, closing file\n")
	return fw.file.Close()
}

// periodicFlush flushes the buffer periodically
func (fw *FileWriter) periodicFlush() {
	defer fw.wg.Done()
	fmt.Fprintf(os.Stderr, "FileWriter: periodicFlush goroutine started\n")
	for {
		select {
		case <-fw.ticker.C:
			fmt.Fprintf(os.Stderr, "FileWriter: Flushing buffer now\n")
			if err := fw.Flush(); err != nil {
				// Log flush errors to stderr
				fmt.Fprintf(os.Stderr, "FileWriter: periodic flush failed: %v\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "FileWriter: Flush completed successfully\n")
			}
		case <-fw.done:
			fmt.Fprintf(os.Stderr, "FileWriter: periodicFlush goroutine stopping\n")
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
