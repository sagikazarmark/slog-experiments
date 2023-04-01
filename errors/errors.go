package errors

import (
	"errors"
	"fmt"

	"github.com/sagikazarmark/slog-experiments/slog"
)

// WithAttrs annotates err with with arbitrary key-value pairs.
func WithAttrs(err error, attrs ...Attr) error {
	if err == nil {
		return nil
	}

	if len(attrs) == 0 {
		return err
	}

	var w *withAttrs
	if !errors.As(err, &w) {
		w = &withAttrs{
			error: err,
		}

		err = w
	}

	// Limiting the capacity of the stored keyvals ensures that a new
	// backing array is created if the slice must grow in WithAttrs.
	// Using the extra capacity without copying risks a data race.
	d := append(w.attrs, attrs...) // nolint:gocritic
	w.attrs = d[:len(d):len(d)]

	return err
}

// withAttrs annotates an error with arbitrary key-value pairs.
type withAttrs struct {
	error error
	attrs []Attr
}

func (w *withAttrs) Error() string { return w.error.Error() }
func (w *withAttrs) Cause() error  { return w.error }
func (w *withAttrs) Unwrap() error { return w.error }

// Attrs returns the appended attributes.
func (w *withAttrs) Attrs() []Attr {
	return w.attrs
}

func (w *withAttrs) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			_, _ = fmt.Fprintf(s, "%+v", w.error)

			return
		}

		_, _ = fmt.Fprintf(s, "%v", w.error)

	case 's':
		_, _ = fmt.Fprintf(s, "%s", w.error)

	case 'q':
		_, _ = fmt.Fprintf(s, "%q", w.error)
	}
}

// LogAttr is an alias for log attributes.
type LogAttr = slog.Attr

// WithLogAttrs annotates err with with arbitrary key-value pairs.
func WithLogAttrs(err error, attrs ...LogAttr) error {
	if err == nil {
		return nil
	}

	if len(attrs) == 0 {
		return err
	}

	var w *withLogAttrs
	if !errors.As(err, &w) {
		w = &withLogAttrs{
			error: err,
		}

		err = w
	}

	// Limiting the capacity of the stored keyvals ensures that a new
	// backing array is created if the slice must grow in WithAttrs.
	// Using the extra capacity without copying risks a data race.
	d := append(w.attrs, attrs...) // nolint:gocritic
	w.attrs = d[:len(d):len(d)]

	return err
}

// withLogAttrs annotates an error with arbitrary key-value pairs.
type withLogAttrs struct {
	error error
	attrs []LogAttr
}

func (w *withLogAttrs) Error() string { return w.error.Error() }
func (w *withLogAttrs) Cause() error  { return w.error }
func (w *withLogAttrs) Unwrap() error { return w.error }

// Attrs returns the appended attributes.
func (w *withLogAttrs) Attrs() []LogAttr {
	return w.attrs
}

func (w *withLogAttrs) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			_, _ = fmt.Fprintf(s, "%+v", w.error)

			return
		}

		_, _ = fmt.Fprintf(s, "%v", w.error)

	case 's':
		_, _ = fmt.Fprintf(s, "%s", w.error)

	case 'q':
		_, _ = fmt.Fprintf(s, "%q", w.error)
	}
}
