//go:build darwin || linux

package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"golang.org/x/term"
)

// ConfigureLogging sets up logging with colored output for terminals
// and structured logging for non-terminal outputs on Darwin/Linux.
func ConfigureLogging(out io.Writer) error {
	baseHandler := slog.NewTextHandler(out, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	handler := &CustomHandler{
		handler: baseHandler,
		writer:  out,
	}
	logger := slog.New(handler)
	slog.SetDefault(logger)
	return nil
}

// isTerminalOutput checks if the output is a terminal.
func isTerminalOutput(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}

// Colored levels for terminal output.
var levelColors = map[slog.Level]string{
	slog.LevelDebug: "\x1b[34m[DEBUG]\x1b[0m", // Blue
	slog.LevelInfo:  "\x1b[32m[INFO]\x1b[0m",  // Green
	slog.LevelWarn:  "\x1b[33m[WARN]\x1b[0m",  // Yellow
	slog.LevelError: "\x1b[31m[ERROR]\x1b[0m", // Red
}

// Plain levels for structured logging.
var plainLevelStrings = map[slog.Level]string{
	slog.LevelDebug: "[DEBUG]",
	slog.LevelInfo:  "[INFO]",
	slog.LevelWarn:  "[WARN]",
	slog.LevelError: "[ERROR]",
}

// getLevelString retrieves the appropriate level string, with a fallback for unknown levels.
func getLevelString(level slog.Level, useColors bool) string {
	if useColors {
		if color, ok := levelColors[level]; ok {
			return color
		}
		return fmt.Sprintf("\x1b[0m[%s]\x1b[0m", level)
	}
	if plain, ok := plainLevelStrings[level]; ok {
		return plain
	}
	return fmt.Sprintf("[%s]", level)
}

// CustomHandler customizes logging for Darwin/Linux systems.
type CustomHandler struct {
	handler slog.Handler
	writer  io.Writer
}

func (h *CustomHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *CustomHandler) Handle(ctx context.Context, record slog.Record) error {
	if isTerminalOutput(h.writer) {
		levelStr := getLevelString(record.Level, true)

		// Write level and message
		fmt.Fprintf(h.writer, "%s %s", levelStr, record.Message)

		// Print all structured attributes
		record.Attrs(func(a slog.Attr) bool {
			fmt.Fprintf(h.writer, " %s=%v", a.Key, a.Value.Any())
			return true
		})

		fmt.Fprint(h.writer, "\n")
		return nil
	}

	// Clone and add plain level string to the record
	newRecord := record.Clone()
	newRecord.AddAttrs(slog.String("level_str", getLevelString(record.Level, false)))
	return h.handler.Handle(ctx, newRecord)
}

func (h *CustomHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &CustomHandler{
		handler: h.handler.WithAttrs(attrs),
		writer:  h.writer,
	}
}

func (h *CustomHandler) WithGroup(name string) slog.Handler {
	return &CustomHandler{
		handler: h.handler.WithGroup(name),
		writer:  h.writer,
	}
}
