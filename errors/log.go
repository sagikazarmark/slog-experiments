package errors

import (
	"context"
	"errors"

	"github.com/sagikazarmark/slog-experiments/slog"
)

type hasAttrs interface {
	Attrs() []Attr
}

type Handler struct {
	Delegate slog.Handler
}

func (h Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.Delegate.Enabled(ctx, level)
}

func (h Handler) Handle(ctx context.Context, r slog.Record) error {
	var err hasAttrs

	r.Attrs(func(a slog.Attr) {
		if err == nil && a.Key == "err" {
			e, ok := a.Value.Any().(error)
			if ok {
				errors.As(e, &err)
			}
		}
	})

	if attrs := err.Attrs(); len(attrs) > 0 {
		var logAttrs []slog.Attr
		for _, a := range attrs {
			logAttrs = append(logAttrs, slog.Any(a.Key, a.Value.Any()))
		}

		r.AddAttrs(logAttrs...)
	}

	return h.Delegate.Handle(ctx, r)
}

func (h Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return Handler{h.Delegate.WithAttrs(attrs)}
}

func (h Handler) WithGroup(name string) slog.Handler {
	return Handler{h.Delegate.WithGroup(name)}
}

type hasLogAttrs interface {
	Attrs() []slog.Attr
}

type LogHandler struct {
	Delegate slog.Handler
}

func (h LogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.Delegate.Enabled(ctx, level)
}

func (h LogHandler) Handle(ctx context.Context, r slog.Record) error {
	var err hasLogAttrs

	r.Attrs(func(a slog.Attr) {
		if err == nil && a.Key == "err" {
			e, ok := a.Value.Any().(error)
			if ok {
				errors.As(e, &err)
			}
		}
	})

	if attrs := err.Attrs(); len(attrs) > 0 {
		r.AddAttrs(attrs...)
	}

	return h.Delegate.Handle(ctx, r)
}

func (h LogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return LogHandler{h.Delegate.WithAttrs(attrs)}
}

func (h LogHandler) WithGroup(name string) slog.Handler {
	return LogHandler{h.Delegate.WithGroup(name)}
}
