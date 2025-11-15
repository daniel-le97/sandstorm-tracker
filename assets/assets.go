package assets

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed templates/*.html configs/* static/*
var webFS embed.FS

//go:embed static/*
var staticFS embed.FS

// WebAssets holds the embedded web UI files
type WebAssets struct {
	fs embed.FS
}

// NewWebAssets creates a new WebAssets instance with the provided embed.FS
func NewWebAssets(embedFS embed.FS) *WebAssets {
	return &WebAssets{fs: embedFS}
}

// GetWebAssets returns the default WebAssets instance with embedded files
func GetWebAssets() *WebAssets {
	return &WebAssets{fs: webFS}
}

// FS returns the embedded filesystem
func (w *WebAssets) FS() embed.FS {
	return w.fs
}

// Sub returns a sub-filesystem rooted at dir
func (w *WebAssets) Sub(dir string) (fs.FS, error) {
	return fs.Sub(w.fs, dir)
}

// ReadFile reads and returns the content of the named file
func (w *WebAssets) ReadFile(name string) ([]byte, error) {
	return w.fs.ReadFile(name)
}

// Open opens the named file for reading
func (w *WebAssets) Open(name string) (fs.File, error) {
	return w.fs.Open(name)
}

// GetExampleConfig returns the content of an example config file
// format can be "yml" or "toml"
func (w *WebAssets) GetExampleConfig(format string) ([]byte, error) {
	return w.fs.ReadFile("configs/sandstorm-tracker.example." + format)
}

// WriteExampleConfig writes an example config file to the specified path
// format can be "yml" or "toml"
func (w *WebAssets) WriteExampleConfig(path string, format string) error {
	content, err := w.GetExampleConfig(format)
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(path, content, 0644)
}

// StaticFS returns the embedded static filesystem
func StaticFS() fs.FS {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err)
	}
	return sub
}

// ListConfigs returns a list of all embedded config files
func (w *WebAssets) ListConfigs() ([]string, error) {
	var configs []string
	entries, err := w.fs.ReadDir("configs")
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			configs = append(configs, entry.Name())
		}
	}

	return configs, nil
}
