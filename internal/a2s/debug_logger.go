package a2s

import (
	"fmt"
	"os"
	"sync"
	"time"
)

var (
	debugLog   *os.File
	debugMutex sync.Mutex
)

// InitDebugLog creates a debug log file for A2S operations
func InitDebugLog() error {
	debugMutex.Lock()
	defer debugMutex.Unlock()

	var err error
	debugLog, err = os.OpenFile("a2s_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to create debug log: %w", err)
	}

	writeLog("=== A2S Debug Log Started ===")
	return nil
}

// CloseDebugLog closes the debug log file
func CloseDebugLog() {
	debugMutex.Lock()
	defer debugMutex.Unlock()

	if debugLog != nil {
		writeLog("=== A2S Debug Log Closed ===")
		debugLog.Close()
		debugLog = nil
	}
}

// logDebug writes a message to the debug log with timestamp (thread-safe)
func logDebug(format string, args ...interface{}) {
	debugMutex.Lock()
	defer debugMutex.Unlock()

	writeLog(format, args...)
}

// writeLog writes to the log file without locking (internal use only)
func writeLog(format string, args ...interface{}) {
	if debugLog == nil {
		return
	}

	timestamp := time.Now().Format("2006/01/02 15:04:05")
	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(debugLog, "%s [A2S] %s\n", timestamp, message)
}
