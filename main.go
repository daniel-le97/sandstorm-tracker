package main

import (
	"log"
	"sandstorm-tracker/internal/app"
	_ "sandstorm-tracker/migrations"
)

// Build-time variables injected by GoReleaser via ldflags
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

func main() {
	// Create application with version information injected at build time
	application, err := app.NewWithVersion(Version, Commit, Date)
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}

	// Bootstrap application (registers hooks, routes, jobs)
	if err := application.Bootstrap(); err != nil {
		log.Fatalf("Failed to bootstrap app: %v", err)
	}

	// Start registers default commands (serve, superuser, version) and executes RootCmd
	if err := application.Start(); err != nil {
		log.Fatal(err)
	}
}
