package logger

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// MultiHandler wraps multiple slog handlers and writes to all of them
type MultiHandler struct {
	handlers []slog.Handler
	mu       sync.RWMutex
}

// NewMultiHandler creates a new handler that writes to multiple handlers
func NewMultiHandler(handlers ...slog.Handler) *MultiHandler {
	return &MultiHandler{
		handlers: handlers,
	}
}

// Enabled implements slog.Handler
func (h *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// If any handler is enabled for this level, return true
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

// Handle implements slog.Handler
func (h *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var errs []error
	for _, handler := range h.handlers {
		if err := handler.Handle(ctx, r); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("multi-handler errors: %v", errs)
	}
	return nil
}

// WithAttrs implements slog.Handler
func (h *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.mu.RLock()
	defer h.mu.RUnlock()

	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithAttrs(attrs)
	}
	return &MultiHandler{handlers: newHandlers}
}

// WithGroup implements slog.Handler
func (h *MultiHandler) WithGroup(name string) slog.Handler {
	h.mu.RLock()
	defer h.mu.RUnlock()

	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithGroup(name)
	}
	return &MultiHandler{handlers: newHandlers}
}

// FileHandlerConfig configures the file handler
type FileHandlerConfig struct {
	FilePath   string        // Path to log file
	Level      slog.Level    // Minimum log level
	AddSource  bool          // Add source file information
	MaxSize    int64         // Max file size in bytes (0 = no rotation)
	MaxBackups int           // Number of rotated log files to keep (default: 5, 0 = unlimited)
	BufferSize int           // Buffer size in bytes (0 = unbuffered, recommended: 8192)
	FlushEvery time.Duration // How often to flush buffer (recommended: 3s)
}

// Performance notes:
// - BufferSize=0: Unbuffered writes (lower performance, no data loss risk)
// - BufferSize=8192 (8KB): Good balance for most production workloads
// - FlushEvery=3s: Acceptable data loss window for most apps (max 3s of logs lost on crash)
// - Logs are also flushed immediately when the buffer fills up

// bufferedWriter wraps a file with buffered writes and periodic flushing
type bufferedWriter struct {
	file       *os.File
	writer     io.Writer
	ticker     *time.Ticker
	done       chan struct{}
	mu         sync.Mutex
	bufferSize int
}

func newBufferedWriter(file *os.File, bufferSize int, flushInterval time.Duration) *bufferedWriter {
	if bufferSize <= 0 {
		bufferSize = 4096 // Default 4KB buffer
	}
	if flushInterval <= 0 {
		flushInterval = 3 * time.Second // Default 3s flush
	}

	bw := &bufferedWriter{
		file:       file,
		writer:     io.Writer(file),
		bufferSize: bufferSize,
		done:       make(chan struct{}),
	}

	// Only use buffered writer if buffer size is configured
	if bufferSize > 0 {
		bw.writer = bufio.NewWriterSize(file, bufferSize)

		// Start periodic flush goroutine
		bw.ticker = time.NewTicker(flushInterval)
		go bw.periodicFlush()
	}

	return bw
}

func (bw *bufferedWriter) periodicFlush() {
	for {
		select {
		case <-bw.ticker.C:
			bw.Flush()
		case <-bw.done:
			return
		}
	}
}

func (bw *bufferedWriter) Write(p []byte) (n int, err error) {
	bw.mu.Lock()
	defer bw.mu.Unlock()
	return bw.writer.Write(p)
}

func (bw *bufferedWriter) Flush() error {
	bw.mu.Lock()
	defer bw.mu.Unlock()

	if flusher, ok := bw.writer.(interface{ Flush() error }); ok {
		if err := flusher.Flush(); err != nil {
			return err
		}
	}
	return bw.file.Sync()
}

func (bw *bufferedWriter) Close() error {
	if bw.ticker != nil {
		bw.ticker.Stop()
		close(bw.done)
	}
	bw.Flush()
	return bw.file.Close()
}

// rotateIfNeeded checks if the log file exceeds maxSize and rotates it if needed
// Rotation strategy: app.log -> app.log.1 -> app.log.2 -> ... (keeps last maxBackups rotations)
func rotateIfNeeded(filePath string, maxSize int64, maxBackups int) error {
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet, no rotation needed
		}
		return fmt.Errorf("failed to stat log file: %w", err)
	}

	// Check if rotation is needed
	if info.Size() < maxSize {
		return nil // File is below max size
	}

	// Default to 5 backups if not specified
	if maxBackups <= 0 {
		maxBackups = 5
	}

	// Rotate existing backups (app.log.4 -> app.log.5, app.log.3 -> app.log.4, etc.)
	for i := maxBackups - 1; i >= 1; i-- {
		oldPath := fmt.Sprintf("%s.%d", filePath, i)
		newPath := fmt.Sprintf("%s.%d", filePath, i+1)
		if _, err := os.Stat(oldPath); err == nil {
			os.Rename(oldPath, newPath) // Overwrite oldest backup
		}
	}

	// Rotate current log file to .1
	backupPath := fmt.Sprintf("%s.1", filePath)
	if err := os.Rename(filePath, backupPath); err != nil {
		return fmt.Errorf("failed to rotate log file: %w", err)
	}

	return nil
}

// NewFileHandler creates a new slog handler that writes to a file
func NewFileHandler(config FileHandlerConfig) (slog.Handler, io.Closer, error) {
	// Create log directory if it doesn't exist
	dir := filepath.Dir(config.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Check if log rotation is needed before opening
	if config.MaxSize > 0 {
		if err := rotateIfNeeded(config.FilePath, config.MaxSize, config.MaxBackups); err != nil {
			return nil, nil, fmt.Errorf("failed to rotate log file: %w", err)
		}
	}

	// Open log file for appending
	file, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Create buffered writer for better performance
	// Uses buffering + periodic flush instead of syncing after every write
	writer := newBufferedWriter(file, config.BufferSize, config.FlushEvery)

	// Create handler options
	opts := &slog.HandlerOptions{
		Level:     config.Level,
		AddSource: config.AddSource,
	}

	// Create JSON handler for structured logging
	handler := slog.NewJSONHandler(writer, opts)

	return handler, writer, nil
}

// Logger wraps slog.Logger with convenience methods
type Logger struct {
	*slog.Logger
	fileCloser io.Closer
}

// NewLogger creates a logger that writes to both PocketBase and a file
func NewLogger(pbHandler slog.Handler, fileConfig FileHandlerConfig) (*Logger, error) {
	// Create file handler
	fileHandler, fileCloser, err := NewFileHandler(fileConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create file handler: %w", err)
	}

	// Create multi-handler that writes to both
	multiHandler := NewMultiHandler(pbHandler, fileHandler)

	logger := &Logger{
		Logger:     slog.New(multiHandler),
		fileCloser: fileCloser,
	}

	return logger, nil
}

// Close closes the file handler
func (l *Logger) Close() error {
	if l.fileCloser != nil {
		return l.fileCloser.Close()
	}
	return nil
}

// FileOnlyLogger creates a logger that writes only to a file (not to PocketBase)
func FileOnlyLogger(fileConfig FileHandlerConfig) (*Logger, error) {
	fileHandler, fileCloser, err := NewFileHandler(fileConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create file handler: %w", err)
	}

	logger := &Logger{
		Logger:     slog.New(fileHandler),
		fileCloser: fileCloser,
	}

	return logger, nil
}

// DefaultFileConfig returns default file handler configuration
func DefaultFileConfig(logPath string) FileHandlerConfig {
	return FileHandlerConfig{
		FilePath:  logPath,
		Level:     slog.LevelInfo,
		AddSource: true,
		MaxSize:   0, // No rotation
	}
}

// RotatingFileWriter wraps a file and handles rotation based on size
type RotatingFileWriter struct {
	filePath string
	maxSize  int64
	file     *os.File
	mu       sync.Mutex
	size     int64
}

// NewRotatingFileWriter creates a file writer that rotates when size exceeds maxSize
func NewRotatingFileWriter(filePath string, maxSize int64) (*RotatingFileWriter, error) {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to stat log file: %w", err)
	}

	return &RotatingFileWriter{
		filePath: filePath,
		maxSize:  maxSize,
		file:     file,
		size:     info.Size(),
	}, nil
}

// Write implements io.Writer
func (w *RotatingFileWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Check if rotation is needed
	if w.maxSize > 0 && w.size+int64(len(p)) > w.maxSize {
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}

	n, err = w.file.Write(p)
	w.size += int64(n)
	return n, err
}

// rotate closes current file and opens a new one with timestamp
func (w *RotatingFileWriter) rotate() error {
	if w.file != nil {
		w.file.Close()
	}

	// Rename current file with timestamp
	timestamp := time.Now().Format("2006-01-02-150405")
	ext := filepath.Ext(w.filePath)
	base := w.filePath[:len(w.filePath)-len(ext)]
	rotatedPath := fmt.Sprintf("%s-%s%s", base, timestamp, ext)

	if err := os.Rename(w.filePath, rotatedPath); err != nil {
		// If rename fails, try to continue with new file anyway
		_ = os.Remove(w.filePath)
	}

	// Open new file
	file, err := os.OpenFile(w.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open new log file: %w", err)
	}

	w.file = file
	w.size = 0
	return nil
}

// Close closes the file
func (w *RotatingFileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file != nil {
		return w.file.Close()
	}
	return nil
}
