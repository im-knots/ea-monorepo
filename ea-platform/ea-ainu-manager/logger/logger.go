package logger

import (
	"log/slog"
	"os"
)

// Slog is the global logger instance.
var Slog *slog.Logger

func init() {
	// Initialize the global logger with default settings
	Slog = slog.New(slog.NewJSONHandler(os.Stdout, nil))
}

// SetLogger allows overriding the global logger instance.
func SetLogger(l *slog.Logger) {
	if l == nil {
		panic("Cannot set a nil logger")
	}
	Slog = l
}
