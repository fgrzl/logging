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

// Detect if we're inside VSCode Debug Console
var forceColor = os.Getenv("VSCODE_PID") != ""

// ConfigureLogging sets up logging with colored output for Windows terminals
// and structured logging for non-terminal outputs.
func ConfigureLogging(out io.Writer) error {
	_ = enableWindowsANSI(out)

	var baseHandler slog.Handler
	if !isTerminalOutput(out) {
		baseHandler = slog.NewTextHandler(out, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}

	slog.SetDefault(slog.New(&WindowsCustomHandler{
		handler: baseHandler,
		writer:  out,
	}))
	return nil
}

// enableWindowsANSI enables ANSI escape code support on Windows.
func enableWindowsANSI(w io.Writer) error {
	if !isTerminalOutput(w) {
		return nil
	}

	_, ok := w.(*os.File)
	if !ok {
		return nil
	}

	handle, err := windows.GetStdHandle(windows.STD_OUTPUT_HANDLE)
	if err != nil || handle == windows.InvalidHandle {
		return nil
	}

	var mode uint32
	if err := windows.GetConsoleMode(handle, &mode); err != nil {
		if err == windows.ERROR_INVALID_HANDLE {
			return nil
		}
		return fmt.Errorf("get console mode: %w", err)
	}

	if err := windows.SetConsoleMode(handle, mode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING); err != nil {
		if err == windows.ERROR_INVALID_HANDLE {
			return nil
		}
		return fmt.Errorf("set console mode: %w", err)
	}

	return nil
}

// isTerminalOutput determines whether to treat the output as a terminal.
func isTerminalOutput(w io.Writer) bool {
	if forceColor {
		return true
	}
	f, ok := w.(*os.File)
	return ok && term.IsTerminal(int(f.Fd()))
}

// colorCode returns the ANSI color code string for a given log level.
func colorCode(level slog.Level) string {
	switch level {
	case slog.LevelDebug:
		return "34" // Blue
	case slog.LevelInfo:
		return "32" // Green
	case slog.LevelWarn:
		return "33" // Yellow
	case slog.LevelError:
		return "31" // Red
	default:
		return "0"
	}
}

// levelText returns the label string for a log level.
func levelText(level slog.Level) string {
	switch level {
	case slog.LevelDebug:
		return "DEBUG"
	case slog.LevelInfo:
		return "INFO"
	case slog.LevelWarn:
		return "WARN"
	case slog.LevelError:
		return "ERROR"
	default:
		return level.String()
	}
}

// WindowsCustomHandler formats log output based on terminal detection.
type WindowsCustomHandler struct {
	handler slog.Handler
	writer  io.Writer
}

func (h *WindowsCustomHandler) Enabled(ctx context.Context, level slog.Level) bool {
	if isTerminalOutput(h.writer) {
		return true
	}
	return h.handler != nil && h.handler.Enabled(ctx, level)
}

func (h *WindowsCustomHandler) Handle(ctx context.Context, record slog.Record) error {

	if isTerminalOutput(h.writer) {
		// Start with level and message
		fmt.Fprintf(h.writer, "\x1b[%sm[%s] %s", colorCode(record.Level), levelText(record.Level), record.Message)

		// Append structured attributes
		record.Attrs(func(a slog.Attr) bool {
			fmt.Fprintf(h.writer, " %s : %v", a.Key, a.Value.Any())
			return true
		})

		fmt.Fprint(h.writer, "\x1b[0m\n") // Reset color and newline
		return nil
	}

	if h.handler != nil {
		record.AddAttrs(slog.String("windows_level_str", "["+levelText(record.Level)+"]"))
		return h.handler.Handle(ctx, record)
	}

	return nil
}

func (h *WindowsCustomHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if h.handler != nil {
		return &WindowsCustomHandler{
			handler: h.handler.WithAttrs(attrs),
			writer:  h.writer,
		}
	}
	return h
}

func (h *WindowsCustomHandler) WithGroup(name string) slog.Handler {
	if h.handler != nil {
		return &WindowsCustomHandler{
			handler: h.handler.WithGroup(name),
			writer:  h.writer,
		}
	}
	return h
}
