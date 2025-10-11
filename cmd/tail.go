package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sandstorm-tracker/internal/tail"
	"sync"
	"time"
)

type fileList []string

func (f *fileList) String() string { return fmt.Sprint(*f) }
func (f *fileList) Set(value string) error {
	*f = append(*f, value)
	return nil
}

func main() {
	var files fileList
	lines := flag.Int("lines", 0, "Number of lines to show initially")
	follow := flag.Bool("f", false, "Follow the file (like tail -f)")
	writerFile := flag.String("writer", "", "If set, continuously write test lines to this file (for testing)")
	flag.Var(&files, "file", "File to tail (can be specified multiple times)")
	flag.Parse()

	if *writerFile != "" {
		go func(path string) {
			f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				log.Fatalf("[WRITER] Could not open file for writing: %v", err)
			}
			defer f.Close()
			i := 1
			for {
				line := fmt.Sprintf("Test log line %d at %s\n", i, time.Now().Format(time.RFC3339Nano))
				if _, err := f.WriteString(line); err != nil {
					log.Printf("[WRITER] Write error: %v", err)
				}
				f.Sync()
				i++
				time.Sleep(500 * time.Millisecond)
			}
		}(*writerFile)
	}

	if len(files) == 0 {
		fmt.Println("Usage: tail -file=path/to/file [-file=path2] [-lines=10] [-f] [-writer=path]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	var wg sync.WaitGroup
	for _, fname := range files {
		wg.Add(1)
		go func(filename string) {
			defer wg.Done()
			if err := tail.TailFileMulti(filename, *lines, *follow); err != nil {
				log.Printf("[ERROR] %s: %v", filename, err)
			}
		}(fname)
	}
	wg.Wait()
}
