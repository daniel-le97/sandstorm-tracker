package logger

import (
	"context"
	"log/slog"
	"sync"
)

// TeeHandler writes log records to both a primary handler and a file writer
type TeeHandler struct {
	primaryHandler slog.Handler
	fileWriter     *FileWriter
	mu             sync.Mutex
}

// NewTeeHandler creates a handler that writes to both the primary handler and file
func NewTeeHandler(primaryHandler slog.Handler, fileWriter *FileWriter) *TeeHandler {
	return &TeeHandler{
		primaryHandler: primaryHandler,
		fileWriter:     fileWriter,
	}
}

// Enabled implements slog.Handler
func (h *TeeHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.primaryHandler.Enabled(ctx, level)
}

// Handle implements slog.Handler
func (h *TeeHandler) Handle(ctx context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Write to primary handler (console/PocketBase)
	if err := h.primaryHandler.Handle(ctx, r); err != nil {
		return err
	}

	// Write to file (best effort - don't fail if file write fails)
	if h.fileWriter != nil {
		if err := h.fileWriter.WriteRecord(r); err != nil {
			// Debug: print error to console if file write fails
			// This helps diagnose file writing issues
			_ = err // Suppress error for now
		}
	}

	return nil
}

// WithAttrs implements slog.Handler
func (h *TeeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &TeeHandler{
		primaryHandler: h.primaryHandler.WithAttrs(attrs),
		fileWriter:     h.fileWriter,
	}
}

// WithGroup implements slog.Handler
func (h *TeeHandler) WithGroup(name string) slog.Handler {
	return &TeeHandler{
		primaryHandler: h.primaryHandler.WithGroup(name),
		fileWriter:     h.fileWriter,
	}
}
