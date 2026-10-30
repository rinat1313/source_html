package logging

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

const defaultRingSize = 500

type ringBuffer struct {
	mu    sync.RWMutex
	lines []string
	size  int
	next  int
	full  bool
}

var defaultRing = newRing(defaultRingSize)

func newRing(size int) *ringBuffer {
	if size < 1 {
		size = defaultRingSize
	}
	return &ringBuffer{
		lines: make([]string, size),
		size:  size,
	}
}

func (r *ringBuffer) append(line string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lines[r.next] = line
	r.next = (r.next + 1) % r.size
	if r.next == 0 {
		r.full = true
	}
}

func (r *ringBuffer) snapshot(limit int) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []string
	if r.full {
		out = append(out, r.lines[r.next:]...)
		out = append(out, r.lines[:r.next]...)
	} else {
		out = append(out, r.lines[:r.next]...)
	}
	if limit > 0 && len(out) > limit {
		out = out[len(out)-limit:]
	}
	return out
}

func Recent(limit int) []string {
	return defaultRing.snapshot(limit)
}

type ringHandler struct {
	level slog.Leveler
	inner slog.Handler
}

func newRingHandler(level slog.Leveler, inner slog.Handler) *ringHandler {
	return &ringHandler{level: level, inner: inner}
}

func (h *ringHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level.Level() && h.inner.Enabled(ctx, level)
}

func (h *ringHandler) Handle(ctx context.Context, r slog.Record) error {
	line := r.Time.UTC().Format(time.RFC3339) + " " + r.Level.String() + " " + r.Message
	r.Attrs(func(a slog.Attr) bool {
		line += " " + a.Key + "=" + a.Value.String()
		return true
	})
	defaultRing.append(line)
	return h.inner.Handle(ctx, r)
}

func (h *ringHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ringHandler{level: h.level, inner: h.inner.WithAttrs(attrs)}
}

func (h *ringHandler) WithGroup(name string) slog.Handler {
	return &ringHandler{level: h.level, inner: h.inner.WithGroup(name)}
}
