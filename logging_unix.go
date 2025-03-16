//go:build !windows

package logging

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

// ConfigureLogging sets up logging with colored output for Unix
func ConfigureLogging() {
	baseHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug, // Minimum level
	})
	handler := &UnixCustomHandler{handler: baseHandler}
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

// Colored levels for Unix
var unixLevelColors = map[slog.Level]string{
	slog.LevelDebug: "\x1b[34m[DEBUG]\x1b[0m", // Blue
	slog.LevelInfo:  "\x1b[32m[INFO]\x1b[0m",  // Green
	slog.LevelWarn:  "\x1b[33m[WARN]\x1b[0m",  // Yellow
	slog.LevelError: "\x1b[31m[ERROR]\x1b[0m", // Red
}

// UnixCustomHandler for Unix systems
type UnixCustomHandler struct {
	handler slog.Handler
}

func (h *UnixCustomHandler) Enabled(ctx context.Context, _ slog.Level) bool {
	return true
}

func (h *UnixCustomHandler) Handle(ctx context.Context, record slog.Record) error {
	levelStr := unixLevelColors[record.Level]
	msg := fmt.Sprintf("%s %s", levelStr, record.Message)

	newRecord := slog.Record{
		Time:    record.Time,
		Level:   record.Level,
		Message: msg,
	}

	record.Attrs(func(a slog.Attr) bool {
		newRecord.AddAttrs(a)
		return true
	})

	return h.handler.Handle(ctx, newRecord)
}

func (h *UnixCustomHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &UnixCustomHandler{handler: h.handler.WithAttrs(attrs)}
}

func (h *UnixCustomHandler) WithGroup(name string) slog.Handler {
	return &UnixCustomHandler{handler: h.handler.WithGroup(name)}
}
