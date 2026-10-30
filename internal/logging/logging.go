package logging

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"
)

func Setup(service string) {
	level := parseLevel(os.Getenv("LOG_LEVEL"))
	opts := &slog.HandlerOptions{Level: level}
	stderr := slog.NewTextHandler(os.Stderr, opts)
	handlers := []slog.Handler{newRingHandler(level, stderr)}

	if path := strings.TrimSpace(os.Getenv("LOG_FILE")); path != "" {
		_ = os.MkdirAll(parentDir(path), 0o755)
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err == nil {
			handlers = append(handlers, slog.NewTextHandler(f, opts))
		}
	}

	var handler slog.Handler = handlers[0]
	if len(handlers) > 1 {
		handler = multiHandler{handlers: handlers}
	}
	slog.SetDefault(slog.New(handler))
	slog.Info("logging configured", "service", service, "level", level.String())
}

type multiHandler struct {
	handlers []slog.Handler
}

func (m multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m multiHandler) Handle(ctx context.Context, r slog.Record) error {
	var first error
	for _, h := range m.handlers {
		if !h.Enabled(ctx, r.Level) {
			continue
		}
		if err := h.Handle(ctx, r.Clone()); err != nil && first == nil {
			first = err
		}
	}
	return first
}

func (m multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		next[i] = h.WithAttrs(attrs)
	}
	return multiHandler{handlers: next}
}

func (m multiHandler) WithGroup(name string) slog.Handler {
	next := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		next[i] = h.WithGroup(name)
	}
	return multiHandler{handlers: next}
}

func parentDir(path string) string {
	i := strings.LastIndexByte(path, '/')
	if i <= 0 {
		return "."
	}
	return path[:i]
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
