package main

import (
	"log"
	"sandstorm-tracker/internal/app"
	_ "sandstorm-tracker/migrations"
)

func main() {
	// Create application (wraps PocketBase + your components)
	application, err := app.New()
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
