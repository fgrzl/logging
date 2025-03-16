//go:build windows

package logging

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"golang.org/x/sys/windows"
	"golang.org/x/term"
)

// ConfigureLogging sets up logging with colored output for Windows
func ConfigureLogging() {
	// Enable ANSI support on Windows
	enableWindowsANSI()

	baseHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	handler := &WindowsCustomHandler{handler: baseHandler}
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

// enableWindowsANSI enables ANSI escape code support on Windows
func enableWindowsANSI() {
	handle, err := windows.GetStdHandle(windows.STD_OUTPUT_HANDLE)
	if err != nil {
		return
	}
	var mode uint32
	err = windows.GetConsoleMode(handle, &mode)
	if err != nil {
		return
	}
	_ = windows.SetConsoleMode(handle, mode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
}

// Detect if output is a terminal
func isTerminalOutput() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// Colored levels for Windows
var windowsLevelColors = map[slog.Level]string{
	slog.LevelDebug: "\x1b[34m[DEBUG]\x1b[0m", // Blue
	slog.LevelInfo:  "\x1b[32m[INFO]\x1b[0m",  // Green
	slog.LevelWarn:  "\x1b[33m[WARN]\x1b[0m",  // Yellow
	slog.LevelError: "\x1b[31m[ERROR]\x1b[0m", // Red
}

// Plain levels for structured logging
var plainLevelStrings = map[slog.Level]string{
	slog.LevelDebug: "[DEBUG]",
	slog.LevelInfo:  "[INFO]",
	slog.LevelWarn:  "[WARN]",
	slog.LevelError: "[ERROR]",
}

// WindowsCustomHandler for Windows systems
type WindowsCustomHandler struct {
	handler slog.Handler
}

func (h *WindowsCustomHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level) // Delegate to wrapped handler
}

func (h *WindowsCustomHandler) Handle(ctx context.Context, record slog.Record) error {
	// Determine the appropriate log level string
	plainLevelStr := plainLevelStrings[record.Level]
	coloredLevelStr := plainLevelStr
	if isTerminalOutput() {
		coloredLevelStr = windowsLevelColors[record.Level]
	}

	// Clone record for structured logging
	newRecord := record.Clone()

	// Instead of modifying `Message`, store it as an additional attribute
	newRecord.AddAttrs(slog.String("level_str", plainLevelStr))

	// Print to terminal (colorized)
	if isTerminalOutput() {
		fmt.Fprintf(os.Stdout, "%s %s\n", coloredLevelStr, record.Message)
	}

	// Pass clean structured log to the original handler
	return h.handler.Handle(ctx, newRecord)
}

func (h *WindowsCustomHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &WindowsCustomHandler{handler: h.handler.WithAttrs(attrs)}
}

func (h *WindowsCustomHandler) WithGroup(name string) slog.Handler {
	return &WindowsCustomHandler{handler: h.handler.WithGroup(name)}
}
