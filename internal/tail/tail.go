package tail

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
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

// TailFileMulti is like tailFile but prints the filename as a prefix for each line/event
func TailFileMulti(filename string, numLines int, follow bool) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Seek to last offset if available, otherwise start at beginning
	offset := loadOffset(filename)
	if offset > 0 {
		file.Seek(offset, io.SeekStart)
	} else {
		file.Seek(0, io.SeekStart)
	}

	if !follow {
		// If not following, just print all unread lines and exit
		reader := bufio.NewReader(file)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			if len(line) > 0 && line[len(line)-1] == '\n' {
				line = line[:len(line)-1]
			}
			fmt.Printf("[%s] %s\n", filename, line)
		}
		return nil
	}

	fmt.Printf("--- Following file %s (Ctrl+C to exit) ---\n", filename)
	return followFileMulti(file, filename)
}

func showLastLinesMulti(file *os.File, numLines int, filename string) error {
	stat, err := file.Stat()
	if err != nil {
		return err
	}
	var lines []string
	var currentLine []byte
	for pos := stat.Size() - 1; pos >= 0 && len(lines) < numLines; pos-- {
		file.Seek(pos, 0)
		char := make([]byte, 1)
		file.Read(char)
		if char[0] == '\n' && len(currentLine) > 0 {
			line := make([]byte, len(currentLine))
			for i, b := range currentLine {
				line[len(currentLine)-1-i] = b
			}
			lines = append([]string{string(line)}, lines...)
			currentLine = nil
		} else if char[0] != '\n' {
			currentLine = append(currentLine, char[0])
		}
	}
	if len(currentLine) > 0 {
		line := make([]byte, len(currentLine))
		for i, b := range currentLine {
			line[len(currentLine)-1-i] = b
		}
		lines = append([]string{string(line)}, lines...)
	}
	for _, line := range lines {
		fmt.Printf("[%s] %s\n", filename, line)
	}
	return nil
}

func followFileMulti(origFile *os.File, filename string) error {
	// Seek to last offset if available
	offset := loadOffset(filename)
	if offset > 0 {
		origFile.Seek(offset, io.SeekStart)
	} else {
		origFile.Seek(0, io.SeekEnd)
	}
	file := origFile
	reader := bufio.NewReader(file)
	for {
		pos, _ := file.Seek(0, io.SeekCurrent)
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// Check if file has grown
				stat, statErr := file.Stat()
				if statErr == nil && stat.Size() < pos {
					// File truncated/rotated, reopen and reset
					file.Close()
					file, _ = os.Open(filename)
					reader = bufio.NewReader(file)
					saveOffset(filename, 0)
				}
				time.Sleep(200 * time.Millisecond)
				continue
			}
			return err
		}
		if len(line) > 0 && line[len(line)-1] == '\n' {
			line = line[:len(line)-1]
		}
		fmt.Printf("[%s] %s\n", filename, line)
		// Save offset after each line
		offset, _ := file.Seek(0, io.SeekCurrent)
		saveOffset(filename, offset)
	}
}

// onLogReplacedMulti is called when a tailed file is replaced (multi-file version)
func onLogReplacedMulti(filename string) {
	fmt.Printf("[HOOK] Log file replaced: %s\n", filename)
	// Add your custom logic here
}

// onLogTruncatedMulti is called when a tailed file is truncated (multi-file version)
func onLogTruncatedMulti(filename string) {
	fmt.Printf("[HOOK] Log file truncated: %s\n", filename)
	// Add your custom logic here
}
