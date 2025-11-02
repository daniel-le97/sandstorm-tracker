package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sandstorm-tracker/internal/app"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/ghupdate"
	"github.com/pocketbase/pocketbase/plugins/jsvm"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	"github.com/pocketbase/pocketbase/tools/hook"
	"github.com/pocketbase/pocketbase/tools/osutils"
)

func main() {
	// Initialize app configuration
	appConfig, err := app.InitConfig()
	if err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	pb := pocketbase.New()

	// ---------------------------------------------------------------
	// Optional plugin flags:
	// ---------------------------------------------------------------

	var hooksDir string
	pb.RootCmd.PersistentFlags().StringVar(
		&hooksDir,
		"hooksDir",
		"",
		"the directory with the JS app hooks",
	)

	var hooksWatch bool
	pb.RootCmd.PersistentFlags().BoolVar(
		&hooksWatch,
		"hooksWatch",
		true,
		"auto restart the app on pb_hooks file change; it has no effect on Windows",
	)

	var hooksPool int
	pb.RootCmd.PersistentFlags().IntVar(
		&hooksPool,
		"hooksPool",
		15,
		"the total prewarm goja.Runtime instances for the JS app hooks execution",
	)

	var migrationsDir string
	pb.RootCmd.PersistentFlags().StringVar(
		&migrationsDir,
		"migrationsDir",
		"",
		"the directory with the user defined migrations",
	)

	var automigrate bool
	pb.RootCmd.PersistentFlags().BoolVar(
		&automigrate,
		"automigrate",
		true,
		"enable/disable auto migrations",
	)

	var publicDir string
	pb.RootCmd.PersistentFlags().StringVar(
		&publicDir,
		"publicDir",
		defaultPublicDir(),
		"the directory to serve static files",
	)

	var indexFallback bool
	pb.RootCmd.PersistentFlags().BoolVar(
		&indexFallback,
		"indexFallback",
		true,
		"fallback the request to index.html on missing static path, e.g. when pretty urls are used with SPA",
	)

	// Sandstorm tracker specific flags
	var pathsStr string
	pb.RootCmd.PersistentFlags().StringVar(
		&pathsStr,
		"paths",
		"",
		"comma-separated list of paths to watch (files or directories)",
	)

	pb.RootCmd.ParseFlags(os.Args[1:])

	// ---------------------------------------------------------------
	// Plugins and hooks:
	// ---------------------------------------------------------------

	// load jsvm (pb_hooks and pb_migrations)
	jsvm.MustRegister(pb, jsvm.Config{
		MigrationsDir: migrationsDir,
		HooksDir:      hooksDir,
		HooksWatch:    hooksWatch,
		HooksPoolSize: hooksPool,
	})

	// migrate command (with js templates)
	migratecmd.MustRegister(pb, pb.RootCmd, migratecmd.Config{
		TemplateLang: migratecmd.TemplateLangJS,
		Automigrate:  automigrate,
		Dir:          migrationsDir,
	})

	// GitHub selfupdate
	ghupdate.MustRegister(pb, pb.RootCmd, ghupdate.Config{})

	// ---------------------------------------------------------------
	// Sandstorm Tracker Integration:
	// ---------------------------------------------------------------

	var fileWatcher *app.Watcher

	// Register web UI routes
	app.RegisterWebRoutes(pb)

	// Initialize file watcher after PocketBase is ready
	pb.OnServe().Bind(&hook.Handler[*core.ServeEvent]{
		Func: func(e *core.ServeEvent) error {
			// Set up file watcher if paths were provided
			if pathsStr != "" {
				paths := strings.Split(pathsStr, ",")
				for i, path := range paths {
					paths[i] = strings.TrimSpace(path)
					_, err := app.GetServerIdFromPath(paths[i])
					if err != nil {
						log.Printf("Warning: Failed to get server ID from path %s: %v", paths[i], err)
						continue
					}
				}

				log.Printf("Starting Sandstorm log watcher")
				log.Printf("Watching paths: %v", paths)

				fileWatcher, err = app.NewWatcher(pb, appConfig.Servers)
				if err != nil {
					return err
				}

				for _, path := range paths {
					if err := fileWatcher.AddPath(path); err != nil {
						log.Printf("Warning: Failed to add path %s: %v", path, err)
					}
				}

				fileWatcher.Start()
				log.Println("File watcher started")
			}

			// static route to serves files from the provided public dir
			// (if publicDir exists and the route path is not already defined)
			if !e.Router.HasRoute(http.MethodGet, "/{path...}") {
				e.Router.GET("/{path...}", apis.Static(os.DirFS(publicDir), indexFallback))
			}

			return e.Next()
		},
		Priority: 999, // execute as latest as possible to allow users to provide their own route
	})

	// Handle graceful shutdown of file watcher
	pb.OnTerminate().BindFunc(func(e *core.TerminateEvent) error {
		if fileWatcher != nil {
			log.Println("Stopping file watcher...")
			fileWatcher.Stop()
			log.Println("File watcher stopped")
		}
		return e.Next()
	})

	// ---------------------------------------------------------------
	// A2S Query Cron Job - Update player scores every minute
	// ---------------------------------------------------------------
	pb.OnServe().BindFunc(func(e *core.ServeEvent) error {
		app.RegisterA2SCron(e.App, appConfig)
		return e.Next()
	})

	if err := pb.Start(); err != nil {
		log.Fatal(err)
	}
}

// the default pb_public dir location is relative to the executable
func defaultPublicDir() string {
	if osutils.IsProbablyGoRun() {
		return "./pb_public"
	}

	return filepath.Join(os.Args[0], "../pb_public")
}
