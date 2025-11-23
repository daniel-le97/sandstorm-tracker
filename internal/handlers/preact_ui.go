package handlers

import (
	"io"
	"net/http"

	"sandstorm-tracker/assets"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterPreactUI registers routes for the Preact SPA UI
// This serves the Preact SPA as the primary UI, while keeping legacy template routes intact
func RegisterPreactUI(e *core.ServeEvent) {
	// Serve Preact UI assets (built files in /assets subdirectory)
	e.Router.GET("/ui/assets/{path...}", apis.Static(assets.UIFS(), false))

	// Serve map images from the public folder (copied to /maps in Vite build output)
	e.Router.GET("/maps/{path...}", apis.Static(assets.UIFS(), false))

	// Serve Preact UI with SPA routing (fallback to index.html for client-side routing)
	// This must be registered AFTER other specific routes so they take precedence
	uiFS := assets.UIFS()

	e.Router.GET("/ui", func(re *core.RequestEvent) error {
		file, err := uiFS.Open("index.html")
		if err != nil {
			return re.NotFoundError("Not found", nil)
		}
		defer file.Close()
		data, _ := io.ReadAll(file)
		return re.HTML(http.StatusOK, string(data))
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
		return re.HTML(http.StatusOK, string(data))
	})
}
