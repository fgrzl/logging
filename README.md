[![CI](https://github.com/fgrzl/logging/actions/workflows/ci.yml/badge.svg)](https://github.com/fgrzl/logging/actions/workflows/ci.yml)
[![Dependabot Updates](https://github.com/fgrzl/logging/actions/workflows/dependabot/dependabot-updates/badge.svg)](https://github.com/fgrzl/logging/actions/workflows/dependabot/dependabot-updates)

# Go Logging Package

This Go package provides a customizable logging solution with colored terminal output and structured logging for non-terminal outputs. It is designed to work seamlessly across Windows, macOS (Darwin), and Linux, leveraging the `log/slog` package with platform-specific enhancements.



## Features

- **Colored Terminal Output**: Log levels like `[DEBUG]`, `[INFO]` are color-coded (blue, green, yellow, red) when output is a terminal.
- **Structured Logging**: Clean structured logs with a `level_str` attribute when redirected to files or other non-terminal outputs.
- **Cross-Platform Support**:
  - **Windows**: Enables ANSI escape codes.
  - **macOS/Linux**: Uses native ANSI support.
- **Configurable Output**: Supports any `io.Writer` (e.g., `os.Stdout`, `os.Stderr`, or a file).
- **Robust Error Handling**: Returns errors on configuration/logging failures.
- **Safe Log Level Handling**: Gracefully handles unknown/custom levels with fallback formatting.



## Installation

Requires Go 1.21 or later:

```bash
go get github.com/yourusername/logging
go get golang.org/x/term
````

Replace `github.com/yourusername/logging` with the actual import path.



## Usage

```go
package main

import (
	"fmt"
	"os"
	"log/slog"

	"github.com/yourusername/logging"
)

func main() {
	if err := logging.ConfigureLogging(os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to configure logging: %v\n", err)
		os.Exit(1)
	}

	slog.Debug("This is a debug message")
	slog.Info("This is an info message")
	slog.Warn("This is a warning message")
	slog.Error("This is an error message")
}
```



## Example Output

### Terminal (Windows/macOS/Linux):

```
[DEBUG] This is a debug message   // in blue
[INFO]  This is an info message   // in green
[WARN]  This is a warning message // in yellow
[ERROR] This is an error message  // in red
```

### Redirected to a File:

```
time=2025-06-06T14:20:00Z level=DEBUG level_str=[DEBUG] msg="This is a debug message"
time=2025-06-06T14:20:00Z level=INFO  level_str=[INFO]  msg="This is an info message"
time=2025-06-06T14:20:00Z level=WARN  level_str=[WARN]  msg="This is a warning message"
time=2025-06-06T14:20:00Z level=ERROR level_str=[ERROR] msg="This is an error message"
```



## Customizing Output

You can direct logs to a file:

```go
package main

import (
	"fmt"
	"os"
	"log/slog"

	"github.com/yourusername/logging"
)

func main() {
	file, err := os.Create("app.log")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create log file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	if err := logging.ConfigureLogging(file); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to configure logging: %v\n", err)
		os.Exit(1)
	}

	slog.Info("Logging to a file")
}
```



## Platform-Specific Notes

* **Windows**: Uses `golang.org/x/sys/windows` to enable ANSI color support.
* **macOS/Linux**: Relies on native terminal support for ANSI color codes.
* Uses Go build tags (`//go:build windows`, `//go:build darwin || linux`) for platform-specific implementations.



## Dependencies

* Standard library: `log/slog`, `os`, `fmt`, `io`, `context`
* `golang.org/x/term`
* `golang.org/x/sys/windows` (Windows only)



## Building and Testing

```bash
go build ./...
go test ./...
```

Test on all supported platforms to validate both colored terminal output and structured log files.
