package logging

import (
	"log/slog"
	"os"
	"strings"
	"time"
)

func Setup(service string) {
	level := parseLevel(os.Getenv("LOG_LEVEL"))
	var handler slog.Handler
	opts := &slog.HandlerOptions{Level: level}
	path := os.Getenv("LOG_FILE")
	if path == "" {
		path = "logs/" + service + ".log"
	}
	_ = os.MkdirAll("logs", 0o755)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		handler = slog.NewTextHandler(os.Stderr, opts)
	} else {
		handler = slog.NewTextHandler(f, opts)
	}
	slog.SetDefault(slog.New(handler))
	slog.Info("logging configured", "service", service, "level", level.String())
}

func parseLevel(raw string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func Timed(name string, attrs ...any) func(err *error) {
	start := time.Now()
	slog.Debug(name+" start", attrs...)
	return func(err *error) {
		all := append([]any{"duration_ms", time.Since(start).Milliseconds(), "err", derefErr(err)}, attrs...)
		slog.Info(name+" done", all...)
	}
}

func derefErr(err *error) error {
	if err == nil || *err == nil {
		return nil
	}
	return *err
}
