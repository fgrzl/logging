//go:build windows

package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"golang.org/x/sys/windows"
	"golang.org/x/term"
)

// ConfigureLogging sets up logging with colored output for Windows terminals
// and structured logging for non-terminal outputs.
func ConfigureLogging(out io.Writer) error {
	// Enable ANSI support on Windows
	if err := enableWindowsANSI(); err != nil {
		return fmt.Errorf("failed to enable ANSI support: %w", err)
	}

	baseHandler := slog.NewTextHandler(out, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	handler := &WindowsCustomHandler{
		handler: baseHandler,
		writer:  out,
	}
	logger := slog.New(handler)
	slog.SetDefault(logger)
	return nil
}

// enableWindowsANSI enables ANSI escape code support on Windows.
func enableWindowsANSI() error {
	handle, err := windows.GetStdHandle(windows.STD_OUTPUT_HANDLE)
	if err != nil {
		return fmt.Errorf("failed to get std handle: %w", err)
	}
	var mode uint32
	if err = windows.GetConsoleMode(handle, &mode); err != nil {
		return fmt.Errorf("failed to get console mode: %w", err)
	}
	if err = windows.SetConsoleMode(handle, mode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING); err != nil {
		return fmt.Errorf("failed to set console mode: %w", err)
	}
	return nil
}

// isTerminalOutput checks if the output is a terminal.
func isTerminalOutput(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}

// Colored levels for Windows terminal output.
var windowsLevelColors = map[slog.Level]string{
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
		if color, ok := windowsLevelColors[level]; ok {
			return color
		}
		return fmt.Sprintf("\x1b[0m[%s]\x1b[0m", level)
	}
	if plain, ok := plainLevelStrings[level]; ok {
		return plain
	}
	return fmt.Sprintf("[%s]", level)
}

// WindowsCustomHandler customizes logging for Windows systems.
type WindowsCustomHandler struct {
	handler slog.Handler
	writer  io.Writer
}

func (h *WindowsCustomHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *WindowsCustomHandler) Handle(ctx context.Context, record slog.Record) error {
	levelStr := getLevelString(record.Level, isTerminalOutput(h.writer))
	plainLevelStr := getLevelString(record.Level, false)

	// Clone record and add plain level string as an attribute
	newRecord := record.Clone()
	newRecord.AddAttrs(slog.String("windows_level_str", plainLevelStr))

	// For terminal output, print colored log directly
	if isTerminalOutput(h.writer) {
		_, err := fmt.Fprintf(h.writer, "%s %s\n", levelStr, record.Message)
		return err
	}

	// For non-terminal output, delegate to the wrapped handler
	return h.handler.Handle(ctx, newRecord)
}

func (h *WindowsCustomHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &WindowsCustomHandler{
		handler: h.handler.WithAttrs(attrs),
		writer:  h.writer,
	}
}

func (h *WindowsCustomHandler) WithGroup(name string) slog.Handler {
	return &WindowsCustomHandler{
		handler: h.handler.WithGroup(name),
		writer:  h.writer,
	}
}
