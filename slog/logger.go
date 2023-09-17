// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slog

import (
	"context"
	"log"
	loginternal "log/internal"
	"github.com/sagikazarmark/slog-experiments/slog/internal"
	"runtime"
	"sync/atomic"
	"time"
)

var defaultLogger atomic.Pointer[Logger]

func init() {
	defaultLogger.Store(New(newDefaultHandler(loginternal.DefaultOutput)))
}

// Default returns the default Logger.
func Default() *Logger { return defaultLogger.Load() }

// SetDefault makes l the default Logger.
// After this call, output from the log package's default Logger
// (as with [log.Print], etc.) will be logged at LevelInfo using l's Handler.
func SetDefault(l *Logger) {
	defaultLogger.Store(l)
	// If the default's handler is a defaultHandler, then don't use a handleWriter,
	// or we'll deadlock as they both try to acquire the log default mutex.
	// The defaultHandler will use whatever the log default writer is currently
	// set to, which is correct.
	// This can occur with SetDefault(Default()).
	// See TestSetDefault.
	if _, ok := l.Handler().(*defaultHandler); !ok {
		capturePC := log.Flags()&(log.Lshortfile|log.Llongfile) != 0
		log.SetOutput(&handlerWriter{l.Handler(), LevelInfo, capturePC})
		log.SetFlags(0) // we want just the log message, no time or location
	}
}

// handlerWriter is an io.Writer that calls a Handler.
// It is used to link the default log.Logger to the default slog.Logger.
type handlerWriter struct {
	h         Handler
	level     Level
	capturePC bool
}

func (w *handlerWriter) Write(buf []byte) (int, error) {
	if !w.h.Enabled(context.Background(), w.level) {
		return 0, nil
	}
	var pc uintptr
	if !internal.IgnorePC && w.capturePC {
		// skip [runtime.Callers, w.Write, Logger.Output, log.Print]
		var pcs [1]uintptr
		runtime.Callers(4, pcs[:])
		pc = pcs[0]
	}

	// Remove final newline.
	origLen := len(buf) // Report that the entire buf was written.
	if len(buf) > 0 && buf[len(buf)-1] == '\n' {
		buf = buf[:len(buf)-1]
	}
	r := NewRecord(time.Now(), w.level, string(buf), pc)
	return origLen, w.h.Handle(context.Background(), r)
}

// A Logger records structured information about each call to its
// Log, Debug, Info, Warn, and Error methods.
// For each call, it creates a Record and passes it to a Handler.
//
// To create a new Logger, call [New] or a Logger method
// that begins "With".
type Logger struct {
	handler Handler // for structured logging
}

func (l *Logger) clone() *Logger {
	c := *l
	return &c
}

// Handler returns l's Handler.
func (l *Logger) Handler() Handler { return l.handler }

// With returns a Logger that includes the given attributes
// in each output operation. Arguments are converted to
// attributes as if by [Logger.Log].
func (l *Logger) With(args ...any) *Logger {
	if len(args) == 0 {
		return l
	}
	c := l.clone()
	c.handler = l.handler.WithAttrs(argsToAttrSlice(args))
	return c
}

// WithGroup returns a Logger that starts a group, if name is non-empty.
// The keys of all attributes added to the Logger will be qualified by the given
// name. (How that qualification happens depends on the [Handler.WithGroup]
// method of the Logger's Handler.)
//
// If name is empty, WithGroup returns the receiver.
func (l *Logger) WithGroup(name string) *Logger {
	if name == "" {
		return l
	}
	c := l.clone()
	c.handler = l.handler.WithGroup(name)
	return c

}

// New creates a new Logger with the given non-nil Handler.
func New(h Handler) *Logger {
	if h == nil {
		panic("nil Handler")
	}
	return &Logger{handler: h}
}

// With calls Logger.With on the default logger.
func With(args ...any) *Logger {
	return Default().With(args...)
}

// Enabled reports whether l emits log records at the given context and level.
func (l *Logger) Enabled(ctx context.Context, level Level) bool {
	if ctx == nil {
		ctx = context.Background()
	}
	return l.Handler().Enabled(ctx, level)
}

// NewLogLogger returns a new log.Logger such that each call to its Output method
// dispatches a Record to the specified handler. The logger acts as a bridge from
// the older log API to newer structured logging handlers.
func NewLogLogger(h Handler, level Level) *log.Logger {
	return log.New(&handlerWriter{h, level, true}, "", 0)
}

// Log emits a log record with the current time and the given level and message.
// The Record's Attrs consist of the Logger's attributes followed by
// the Attrs specified by args.
//
// The attribute arguments are processed as follows:
//   - If an argument is an Attr, it is used as is.
//   - If an argument is a string and this is not the last argument,
//     the following argument is treated as the value and the two are combined
//     into an Attr.
//   - Otherwise, the argument is treated as a value with key "!BADKEY".
func (l *Logger) Log(ctx context.Context, level Level, msg string, args ...any) {
	l.log(ctx, level, msg, args...)
}

// LogAttrs is a more efficient version of [Logger.Log] that accepts only Attrs.
func (l *Logger) LogAttrs(ctx context.Context, level Level, msg string, attrs ...Attr) {
	l.logAttrs(ctx, level, msg, attrs...)
}

// Debug logs at LevelDebug.
func (l *Logger) Debug(msg string, args ...any) {
	l.log(context.Background(), LevelDebug, msg, args...)
}

// DebugContext logs at LevelDebug with the given context.
func (l *Logger) DebugContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LevelDebug, msg, args...)
}

// Info logs at LevelInfo.
func (l *Logger) Info(msg string, args ...any) {
	l.log(context.Background(), LevelInfo, msg, args...)
}

// InfoContext logs at LevelInfo with the given context.
func (l *Logger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LevelInfo, msg, args...)
}

// Warn logs at LevelWarn.
func (l *Logger) Warn(msg string, args ...any) {
	l.log(context.Background(), LevelWarn, msg, args...)
}

// WarnContext logs at LevelWarn with the given context.
func (l *Logger) WarnContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LevelWarn, msg, args...)
}

// Error logs at LevelError.
func (l *Logger) Error(msg string, args ...any) {
	l.log(context.Background(), LevelError, msg, args...)
}

// ErrorContext logs at LevelError with the given context.
func (l *Logger) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LevelError, msg, args...)
}

// log is the low-level logging method for methods that take ...any.
// It must always be called directly by an exported logging method
// or function, because it uses a fixed call depth to obtain the pc.
func (l *Logger) log(ctx context.Context, level Level, msg string, args ...any) {
	if !l.Enabled(ctx, level) {
		return
	}
	var pc uintptr
	if !internal.IgnorePC {
		var pcs [1]uintptr
		// skip [runtime.Callers, this function, this function's caller]
		runtime.Callers(3, pcs[:])
		pc = pcs[0]
	}
	r := NewRecord(time.Now(), level, msg, pc)
	r.Add(args...)
	if ctx == nil {
		ctx = context.Background()
	}
	_ = l.Handler().Handle(ctx, r)
}

// logAttrs is like [Logger.log], but for methods that take ...Attr.
func (l *Logger) logAttrs(ctx context.Context, level Level, msg string, attrs ...Attr) {
	if !l.Enabled(ctx, level) {
		return
	}
	var pc uintptr
	if !internal.IgnorePC {
		var pcs [1]uintptr
		// skip [runtime.Callers, this function, this function's caller]
		runtime.Callers(3, pcs[:])
		pc = pcs[0]
	}
	r := NewRecord(time.Now(), level, msg, pc)
	r.AddAttrs(attrs...)
	if ctx == nil {
		ctx = context.Background()
	}
	_ = l.Handler().Handle(ctx, r)
}

// Debug calls Logger.Debug on the default logger.
func Debug(msg string, args ...any) {
	Default().log(context.Background(), LevelDebug, msg, args...)
}

// DebugContext calls Logger.DebugContext on the default logger.
func DebugContext(ctx context.Context, msg string, args ...any) {
	Default().log(ctx, LevelDebug, msg, args...)
}

// Info calls Logger.Info on the default logger.
func Info(msg string, args ...any) {
	Default().log(context.Background(), LevelInfo, msg, args...)
}

// InfoContext calls Logger.InfoContext on the default logger.
func InfoContext(ctx context.Context, msg string, args ...any) {
	Default().log(ctx, LevelInfo, msg, args...)
}

// Warn calls Logger.Warn on the default logger.
func Warn(msg string, args ...any) {
	Default().log(context.Background(), LevelWarn, msg, args...)
}

// WarnContext calls Logger.WarnContext on the default logger.
func WarnContext(ctx context.Context, msg string, args ...any) {
	Default().log(ctx, LevelWarn, msg, args...)
}

// Error calls Logger.Error on the default logger.
func Error(msg string, args ...any) {
	Default().log(context.Background(), LevelError, msg, args...)
}

// ErrorContext calls Logger.ErrorContext on the default logger.
func ErrorContext(ctx context.Context, msg string, args ...any) {
	Default().log(ctx, LevelError, msg, args...)
}

// Log calls Logger.Log on the default logger.
func Log(ctx context.Context, level Level, msg string, args ...any) {
	Default().log(ctx, level, msg, args...)
}

// LogAttrs calls Logger.LogAttrs on the default logger.
func LogAttrs(ctx context.Context, level Level, msg string, attrs ...Attr) {
	Default().logAttrs(ctx, level, msg, attrs...)
}
