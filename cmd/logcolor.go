package cmd

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"

	"charm.land/lipgloss/v2"
)

// Colored log level styles, initialized lazily from the terminal-adaptive
// palette in style.go.
var (
	logStylesOnce sync.Once
	debugLogStyle lipgloss.Style
	infoLogStyle  lipgloss.Style
	warnLogStyle  lipgloss.Style
	errorLogStyle lipgloss.Style
)

func initLogStyles() {
	logStylesOnce.Do(func() {
		initStyles()
		debugLogStyle = lipgloss.NewStyle().Foreground(colorMuted)
		infoLogStyle = lipgloss.NewStyle().Foreground(colorGood)
		warnLogStyle = lipgloss.NewStyle().Foreground(colorWarn)
		errorLogStyle = lipgloss.NewStyle().Foreground(colorError).Bold(true)
	})
}

func styleForLevel(level slog.Level) lipgloss.Style {
	initLogStyles()
	switch {
	case level >= slog.LevelError:
		return errorLogStyle
	case level >= slog.LevelWarn:
		return warnLogStyle
	case level >= slog.LevelInfo:
		return infoLogStyle
	default:
		return debugLogStyle
	}
}

// coloredTextHandler renders a colored level prefix before delegating to a
// plain text handler. It is for interactive TTY output only; JSON sinks must
// not use it (raw ANSI escapes would break machine-readable output).
type coloredTextHandler struct {
	inner slog.Handler
	w     io.Writer
	mu    sync.Mutex
}

func (h *coloredTextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *coloredTextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &coloredTextHandler{inner: h.inner.WithAttrs(attrs), w: h.w}
}

func (h *coloredTextHandler) WithGroup(name string) slog.Handler {
	return &coloredTextHandler{inner: h.inner.WithGroup(name), w: h.w}
}

func (h *coloredTextHandler) Handle(ctx context.Context, r slog.Record) error {
	if !h.Enabled(ctx, r.Level) {
		return nil
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, err := fmt.Fprintf(h.w, "%s ", styleForLevel(r.Level).Render(r.Level.String())); err != nil {
		return err
	}
	return h.inner.Handle(ctx, r)
}
