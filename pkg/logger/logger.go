package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
)

var (
	L     *slog.Logger
	level = new(slog.LevelVar)
)

type PlainHandler struct {
	w     io.Writer
	level slog.Leveler
}

func NewPlainHandler(w io.Writer, level slog.Leveler) *PlainHandler {
	return &PlainHandler{w: w, level: level}
}

func (h *PlainHandler) Enabled(_ context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.level != nil {
		minLevel = h.level.Level()
	}
	return level >= minLevel
}

func (h *PlainHandler) Handle(_ context.Context, r slog.Record) error {
	_, err := fmt.Fprintf(h.w, "[%s] %s\n", r.Level, r.Message)
	return err
}

func init() {
	L = slog.New(NewPlainHandler(os.Stdout, level))
}

func (h *PlainHandler) WithAttrs(attrs []slog.Attr) slog.Handler { return h }
func (h *PlainHandler) WithGroup(name string) slog.Handler       { return h }

func SetVerbosity(verbosity int) {
	if verbosity > 3 {
		verbosity = 3
	}

	switch verbosity {
	case 1:
		level.Set(slog.LevelWarn)
	case 2:
		level.Set(slog.LevelInfo)
	case 3:
		level.Set(slog.LevelDebug)
	default:
		level.Set(slog.LevelError)
	}
}
