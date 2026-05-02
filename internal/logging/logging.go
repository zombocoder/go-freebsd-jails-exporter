package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

func New(level, format string) (*slog.Logger, error) {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "info":
		lvl = slog.LevelInfo
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		return nil, fmt.Errorf("unknown log level: %s", level)
	}

	var handler slog.Handler
	w := io.Writer(os.Stderr)
	opts := &slog.HandlerOptions{Level: lvl}

	switch format {
	case "logfmt":
		handler = slog.NewTextHandler(w, opts)
	case "json":
		handler = slog.NewJSONHandler(w, opts)
	default:
		return nil, fmt.Errorf("unknown log format: %s", format)
	}

	return slog.New(handler), nil
}
