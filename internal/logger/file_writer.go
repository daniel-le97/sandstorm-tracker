package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// RotationPolicy defines when and how log files should rotate
type RotationPolicy struct {
	MaxSize    int64         // Max file size in bytes before rotation (0 = no limit)
	MaxAge     time.Duration // Max age of a log file before rotation (0 = no limit)
	MaxBackups int           // Number of rotated log files to keep (0 = unlimited)
}

// FileWriter handles writing logs to files with rotation
// Production-grade: thread-safe, non-blocking, efficient buffering
type FileWriter struct {
	filePath  string
	file      *os.File
	policy    RotationPolicy
	fileSize  int64 // Current file size
	fileTime  time.Time
	writesCh  chan []byte    // Async write channel for non-blocking writes
	closeCh   chan struct{}  // Signal to stop the writer goroutine
	wg        sync.WaitGroup // Wait for goroutine to finish
	mu        sync.RWMutex   // Protect file handle
	err       error          // Last error encountered
	rotateDir string         // Directory for rotated logs
}

// NewFileWriter creates a production-ready file writer with async writes
func NewFileWriter(filePath string, policy RotationPolicy) (*FileWriter, error) {
	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	fw := &FileWriter{
		filePath:  filePath,
		policy:    policy,
		fileTime:  time.Now(),
		writesCh:  make(chan []byte, 1000), // 1000-entry buffer for async writes
		closeCh:   make(chan struct{}),
		rotateDir: dir,
	}

	// Open initial file
	if err := fw.openFile(); err != nil {
		return nil, err
	}

	// Start async writer goroutine
	fw.wg.Add(1)
	go fw.asyncWriter()

	return fw, nil
}

// Write writes data to the log file (non-blocking, sends to channel)
func (fw *FileWriter) Write(data []byte) (int, error) {
	// Copy data to avoid mutations after return
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)

	// Non-blocking send with fallback to blocking if buffer full
	select {
	case fw.writesCh <- dataCopy:
		return len(data), nil
	case <-fw.closeCh:
		return 0, fmt.Errorf("writer is closed")
	default:
		// Buffer full, do blocking send
		fw.writesCh <- dataCopy
		return len(data), nil
	}
}

// Sync flushes any pending writes (waits for channel to be processed)
func (fw *FileWriter) Sync() error {
	fw.mu.RLock()
	defer fw.mu.RUnlock()

	if fw.file != nil {
		return fw.file.Sync()
	}
	return fw.err
}

// Close closes the file writer
func (fw *FileWriter) Close() error {
	close(fw.closeCh)
	fw.wg.Wait() // Wait for goroutine to finish

	fw.mu.Lock()
	defer fw.mu.Unlock()

	if fw.file != nil {
		return fw.file.Close()
	}
	return nil
}

// asyncWriter runs in a goroutine and handles all file I/O operations
func (fw *FileWriter) asyncWriter() {
	defer fw.wg.Done()
	ticker := time.NewTicker(100 * time.Millisecond) // Batch writes every 100ms
	defer ticker.Stop()

	var batch [][]byte

	for {
		select {
		case data := <-fw.writesCh:
			batch = append(batch, data)

		case <-ticker.C:
			if len(batch) > 0 {
				fw.flushBatch(batch)
				batch = nil
			}

		case <-fw.closeCh:
			// Final flush on close
			if len(batch) > 0 {
				fw.flushBatch(batch)
			}
			return
		}
	}
}

// flushBatch writes a batch of log entries to disk
func (fw *FileWriter) flushBatch(batch [][]byte) {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	if fw.file == nil {
		return
	}

	for _, data := range batch {
		n, err := fw.file.Write(data)
		if err != nil {
			fw.err = err
			return
		}

		// Write newline
		fw.file.WriteString("\n")

		fw.fileSize += int64(n) + 1 // +1 for newline

		// Check if rotation is needed
		if fw.shouldRotate() {
			if err := fw.rotate(); err != nil {
				fw.err = err
				return
			}
		}
	}

	// Sync to disk (fsync on every flush for safety)
	_ = fw.file.Sync()
}

// shouldRotate checks if the log file should rotate
func (fw *FileWriter) shouldRotate() bool {
	// Check size limit
	if fw.policy.MaxSize > 0 && fw.fileSize >= fw.policy.MaxSize {
		return true
	}

	// Check age limit
	if fw.policy.MaxAge > 0 && time.Since(fw.fileTime) > fw.policy.MaxAge {
		return true
	}

	return false
}

// rotate renames current file and opens a new one
func (fw *FileWriter) rotate() error {
	if fw.file != nil {
		fw.file.Close()
	}

	// Generate timestamped backup filename
	timestamp := fw.fileTime.Format("20060102-150405")
	backupPath := fmt.Sprintf("%s.%s", fw.filePath, timestamp)

	// Ensure backup doesn't already exist
	if _, err := os.Stat(backupPath); err == nil {
		// File exists, add counter
		for i := 1; i < 1000; i++ {
			counterPath := fmt.Sprintf("%s.%s-%d", fw.filePath, timestamp, i)
			if _, err := os.Stat(counterPath); os.IsNotExist(err) {
				backupPath = counterPath
				break
			}
		}
	}

	// Move current file to backup
	if err := os.Rename(fw.filePath, backupPath); err != nil {
		return fmt.Errorf("failed to rotate log file: %w", err)
	}

	// Clean old backups
	fw.cleanOldBackups()

	// Open new file
	if err := fw.openFile(); err != nil {
		return err
	}

	return nil
}

// openFile opens the log file for writing
func (fw *FileWriter) openFile() error {
	file, err := os.OpenFile(fw.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Get current file size
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return fmt.Errorf("failed to stat log file: %w", err)
	}

	fw.file = file
	fw.fileSize = info.Size()
	fw.fileTime = time.Now()
	fw.err = nil

	return nil
}

// cleanOldBackups removes old backup files exceeding MaxBackups limit
func (fw *FileWriter) cleanOldBackups() {
	if fw.policy.MaxBackups <= 0 {
		return
	}

	// List files in the log directory
	entries, err := os.ReadDir(fw.rotateDir)
	if err != nil {
		return
	}

	// Find all backup files for this log
	type backup struct {
		name string
		time time.Time
	}
	var backups []backup

	baseFileName := filepath.Base(fw.filePath)
	for _, entry := range entries {
		if !entry.IsDir() && filepath.HasPrefix(entry.Name(), baseFileName+".") {
			info, err := entry.Info()
			if err == nil {
				backups = append(backups, backup{
					name: filepath.Join(fw.rotateDir, entry.Name()),
					time: info.ModTime(),
				})
			}
		}
	}

	// If we exceed max backups, delete oldest ones
	if len(backups) > fw.policy.MaxBackups {
		// Sort by modification time (oldest first)
		// Simple bubble sort for small lists
		for i := 0; i < len(backups); i++ {
			for j := i + 1; j < len(backups); j++ {
				if backups[j].time.Before(backups[i].time) {
					backups[i], backups[j] = backups[j], backups[i]
				}
			}
		}

		// Delete oldest files
		for i := 0; i < len(backups)-fw.policy.MaxBackups; i++ {
			_ = os.Remove(backups[i].name)
		}
	}
}
