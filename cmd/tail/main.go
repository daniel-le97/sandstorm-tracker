package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

func main() {
	var (
		filename = flag.String("file", "", "File to tail (required)")
		lines    = flag.Int("lines", 10, "Number of lines to show initially")
		follow   = flag.Bool("f", false, "Follow the file (like tail -f)")
	)
	flag.Parse()

	if *filename == "" {
		fmt.Println("Usage: tail -file=path/to/file [-lines=10] [-f]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if err := tailFile(*filename, *lines, *follow); err != nil {
		log.Fatal(err)
	}
}

func tailFile(filename string, numLines int, follow bool) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Show initial lines
	if err := showLastLines(file, numLines); err != nil {
		return err
	}

	if !follow {
		return nil
	}

	// Follow mode - continuously read new lines
	fmt.Println("--- Following file (Ctrl+C to exit) ---")
	return followFile(file)
}

func showLastLines(file *os.File, numLines int) error {
	// Get file size
	stat, err := file.Stat()
	if err != nil {
		return err
	}

	// Read from end to find last N lines
	var lines []string
	var currentLine []byte

	// Start from end and read backwards
	for pos := stat.Size() - 1; pos >= 0 && len(lines) < numLines; pos-- {
		file.Seek(pos, 0)
		char := make([]byte, 1)
		file.Read(char)

		if char[0] == '\n' && len(currentLine) > 0 {
			// Reverse the line since we read backwards
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

	// Handle first line if we reached beginning
	if len(currentLine) > 0 {
		line := make([]byte, len(currentLine))
		for i, b := range currentLine {
			line[len(currentLine)-1-i] = b
		}
		lines = append([]string{string(line)}, lines...)
	}

	// Print the lines
	for _, line := range lines {
		fmt.Println(line)
	}

	return nil
}

// onLogReplaced is called when the log file is replaced (inode/file ID changes)
func onLogReplaced(filename string) {
	fmt.Printf("[HOOK] Log file replaced: %s\n", filename)
	// Add your custom logic here
}

// onLogTruncated is called when the log file is truncated (inode same, size shrinks)
func onLogTruncated(filename string) {
	fmt.Printf("[HOOK] Log file truncated: %s\n", filename)
	// Add your custom logic here
}

func followFile(origFile *os.File) error {
	// Seek to end of file
	origFile.Seek(0, io.SeekEnd)

	file := origFile
	reader := bufio.NewReader(file)
	lastStat, _ := file.Stat()
	lastInode := getFileID(lastStat)
	lastSize := lastStat.Size()
	filename := file.Name()
	checkInterval := time.Second
	lastCheck := time.Now()

	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			// Periodically check for truncation or replacement
			if time.Since(lastCheck) > checkInterval {
				stat, err := os.Stat(filename)
				if err == nil {
					inode := getFileID(stat)
					size := stat.Size()
					// File replaced (inode changed) or truncated (size < lastSize)
					if inode != lastInode {
						onLogReplaced(filename)
						file.Close()
						file, err = os.Open(filename)
						if err != nil {
							return fmt.Errorf("failed to reopen file: %w", err)
						}
						reader = bufio.NewReader(file)
						file.Seek(0, io.SeekEnd)
						lastStat = stat
						lastInode = inode
						lastSize = size
						fmt.Println("--- File replaced, resuming tail ---")
					} else if size < lastSize {
						onLogTruncated(filename)
						file.Close()
						file, err = os.Open(filename)
						if err != nil {
							return fmt.Errorf("failed to reopen file: %w", err)
						}
						reader = bufio.NewReader(file)
						file.Seek(0, io.SeekEnd)
						lastStat = stat
						lastInode = inode
						lastSize = size
						fmt.Println("--- File truncated, resuming tail ---")
					} else {
						lastSize = size
					}
				}
				lastCheck = time.Now()
			}
			time.Sleep(100 * time.Millisecond)
			continue
		}
		if err != nil {
			return err
		}

		// Print new line (remove trailing newline since Println adds one)
		if len(line) > 0 && line[len(line)-1] == '\n' {
			line = line[:len(line)-1]
		}
		fmt.Println(line)
	}
}
