package main

import (
    "flag"
    "log"
    "os"
    "os/signal"
    "strings"
    "syscall"

    "sandstorm-tracker/db"
    "sandstorm-tracker/internal/dbutil"
    "sandstorm-tracker/internal/watcher"
)

func main() {
    var (
        pathsStr = flag.String("paths", "", "Comma-separated list of paths to watch (files or directories)")
        dbPath   = flag.String("db", "sandstorm_stats.db", "Path to SQLite database file")
        checkDB  = flag.Bool("check", false, "Check database contents and exit")
    )
    flag.Parse()

    if *checkDB {
        dbutil.CheckDatabase(*dbPath)
        return
    }

    if *pathsStr == "" {
        log.Fatal("Please provide at least one path to watch using -paths flag")
    }

    paths := strings.Split(*pathsStr, ",")
    for i, path := range paths {
        paths[i] = strings.TrimSpace(path)
    }

    log.Printf("Starting Sandstorm log watcher")
    log.Printf("Watching paths: %v", paths)
    log.Printf("Database: %s", *dbPath)

    dbService, err := db.NewDatabaseService(*dbPath)
    if err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    defer dbService.Close()

    log.Println("Database initialized successfully")

    fw, err := watcher.NewFileWatcher(dbService)
    if err != nil {
        log.Fatalf("Failed to create file watcher: %v", err)
    }

    for _, path := range paths {
        if err := fw.AddPath(path); err != nil {
            log.Printf("Warning: Failed to add path %s: %v", path, err)
        }
    }

    fw.Start()

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    log.Println("File watcher started. Press Ctrl+C to stop.")

    <-sigChan
    log.Println("Shutting down...")

    fw.Stop()
    log.Println("File watcher stopped.")
}