package handlers

import (
	"io"
	"io/fs"

	"sandstorm-tracker/assets"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterPreactUI registers routes for the Preact SPA UI
// This serves the Preact SPA as the primary UI, while keeping legacy template routes intact
func RegisterPreactUI(e *core.ServeEvent) {
	// Get the UI filesystem (points to /ui directory in embedded files)
	uiFS := assets.UIFS()

	// Serve Preact UI assets (built files in /assets subdirectory of UI)
	// We need to get the sub-filesystem for assets
	assetsFS, _ := fs.Sub(uiFS, "assets")
	e.Router.GET("/ui/assets/{path...}", apis.Static(assetsFS, false))

	// Serve map images from the /maps folder
	mapsFS, _ := fs.Sub(uiFS, "maps")
	e.Router.GET("/ui/maps/{path...}", apis.Static(mapsFS, false))

	// Also serve maps at /maps for backward compatibility
	e.Router.GET("/maps/{path...}", apis.Static(mapsFS, false))

	// Serve Preact UI with SPA routing (fallback to index.html for client-side routing)
	// This must be registered AFTER other specific routes so they take precedence

	e.Router.GET("/ui", func(re *core.RequestEvent) error {
		file, err := uiFS.Open("index.html")
		if err != nil {
			return re.NotFoundError("Not found", nil)
		}
		defer file.Close()
		data, _ := io.ReadAll(file)
		return re.HTML(200, string(data))
	})

	e.Router.GET("/ui/{path...}", func(re *core.RequestEvent) error {
		path := re.Request.PathValue("path")
		if path == "" {
			path = "index.html"
		}

		// Try to serve the requested file
		file, err := uiFS.Open(path)
		if err == nil {
			defer file.Close()
			// Check if it's a directory, serve index.html
			if stat, err := file.Stat(); err == nil && stat.IsDir() {
				file, _ = uiFS.Open(path + "/index.html")
				if file != nil {
					defer file.Close()
				}
			}
		}

		// If file doesn't exist, try to serve index.html for client-side routing
		if err != nil || file == nil {
			file, err = uiFS.Open("index.html")
			if err != nil {
				return re.NotFoundError("Not found", nil)
			}
			defer file.Close()
		}

		data, _ := io.ReadAll(file)
		return re.HTML(200, string(data))
	})
}
