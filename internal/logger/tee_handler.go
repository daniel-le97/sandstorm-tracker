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
	attrs          []slog.Attr // Store attributes added via WithAttrs
	groups         []string    // Store group names added via WithGroup
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

	// Write to file first (most important)
	if h.fileWriter != nil {
		// Create a new record with our stored attributes added
		r2 := r.Clone()

		// Add groups as nested attributes if any
		if len(h.groups) > 0 {
			// Groups need special handling - for now just add them as attributes
			for _, group := range h.groups {
				r2.AddAttrs(slog.String("group", group))
			}
		}

		// Add stored attributes
		r2.AddAttrs(h.attrs...)

		if err := h.fileWriter.WriteRecord(r2); err != nil {
			// Don't fail the entire log operation if file write fails
			_ = err
		}
	}

	// Write to primary handler (console/PocketBase)
	if err := h.primaryHandler.Handle(ctx, r); err != nil {
		return err
	}

	return nil
}

// WithAttrs implements slog.Handler
func (h *TeeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// Combine existing attrs with new ones
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)

	return &TeeHandler{
		primaryHandler: h.primaryHandler.WithAttrs(attrs),
		fileWriter:     h.fileWriter,
		attrs:          newAttrs,
		groups:         h.groups,
	}
}

// WithGroup implements slog.Handler
func (h *TeeHandler) WithGroup(name string) slog.Handler {
	// Combine existing groups with new one
	newGroups := make([]string, len(h.groups)+1)
	copy(newGroups, h.groups)
	newGroups[len(h.groups)] = name

	return &TeeHandler{
		primaryHandler: h.primaryHandler.WithGroup(name),
		fileWriter:     h.fileWriter,
		attrs:          h.attrs,
		groups:         newGroups,
	}
}

// With returns a new Logger with the given attributes added to every log call
// This is a convenience method that wraps the handler
func (h *TeeHandler) With(args ...any) *slog.Logger {
	return slog.New(h).With(args...)
}
