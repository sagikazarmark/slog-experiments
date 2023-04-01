package errors

import (
	"errors"
	"io"
	"os"
	"testing"
	"time"

	"github.com/sagikazarmark/slog-experiments/slog"
)

func ExampleAttrs() {
	var err error

	err = errors.New("something went wrong")
	err = WithAttrs(err, Bool("is-it-wrong", true))

	logger := slog.New(Handler{(slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				a.Key = ""
			}

			return a
		},
	}).NewTextHandler(os.Stdout)})

	logger.Error("oops", slog.Any("err", err))
	// Output: level=ERROR msg=oops err="something went wrong" is-it-wrong=true
}

func ExampleLogAttrs() {
	var err error

	err = errors.New("something went wrong")
	err = WithLogAttrs(err, slog.Bool("is-it-wrong", true))

	logger := slog.New(LogHandler{(slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				a.Key = ""
			}

			return a
		},
	}).NewTextHandler(os.Stdout)})

	logger.Error("oops", slog.Any("err", err))
	// Output: level=ERROR msg=oops err="something went wrong" is-it-wrong=true
}

const TestMessage = "Test logging, but use a somewhat realistic message length."

var (
	TestTime     = time.Date(2022, time.May, 1, 0, 0, 0, 0, time.UTC)
	TestString   = "7e3b3b2aaeff56a7108fe11e154200dd/7819479873059528190"
	TestInt      = 32768
	TestDuration = 23 * time.Second
	TestError    = errors.New("fail")
)

func BenchmarkAttrs(b *testing.B) {
	err := errors.New("something went wrong")

	logger := slog.New(Handler{(slog.HandlerOptions{}).NewTextHandler(io.Discard)})

	for _, call := range []struct {
		name string
		f    func()
	}{
		{
			// The number should match nAttrsInline in slog/record.go.
			// This should exercise the code path where no allocations
			// happen in Record or Attr. If there are allocations, they
			// should only be from Duration.String and Time.String.
			"5 args",
			func() {
				logger.Error(TestMessage,
					slog.Any("err", WithAttrs(err,
						String("string", TestString),
						Int("status", TestInt),
						Duration("duration", TestDuration),
						Time("time", TestTime),
						Any("error", TestError),
					)),
				)
			},
		},
		{
			"10 args",
			func() {
				logger.Error(TestMessage,
					slog.Any("err", WithAttrs(err,
						String("string", TestString),
						Int("status", TestInt),
						Duration("duration", TestDuration),
						Time("time", TestTime),
						Any("error", TestError),
						String("string", TestString),
						Int("status", TestInt),
						Duration("duration", TestDuration),
						Time("time", TestTime),
						Any("error", TestError),
					)),
				)
			},
		},
		{
			"40 args",
			func() {
				logger.Error(TestMessage,
					slog.Any("err", WithAttrs(err,
						String("string", TestString),
						Int("status", TestInt),
						Duration("duration", TestDuration),
						Time("time", TestTime),
						Any("error", TestError),
						String("string", TestString),
						Int("status", TestInt),
						Duration("duration", TestDuration),
						Time("time", TestTime),
						Any("error", TestError),
						String("string", TestString),
						Int("status", TestInt),
						Duration("duration", TestDuration),
						Time("time", TestTime),
						Any("error", TestError),
						String("string", TestString),
						Int("status", TestInt),
						Duration("duration", TestDuration),
						Time("time", TestTime),
						Any("error", TestError),
						String("string", TestString),
						Int("status", TestInt),
						Duration("duration", TestDuration),
						Time("time", TestTime),
						Any("error", TestError),
						String("string", TestString),
						Int("status", TestInt),
						Duration("duration", TestDuration),
						Time("time", TestTime),
						Any("error", TestError),
						String("string", TestString),
						Int("status", TestInt),
						Duration("duration", TestDuration),
						Time("time", TestTime),
						Any("error", TestError),
						String("string", TestString),
						Int("status", TestInt),
						Duration("duration", TestDuration),
						Time("time", TestTime),
						Any("error", TestError),
					)),
				)
			},
		},
	} {
		b.Run(call.name, func(b *testing.B) {
			b.ReportAllocs()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					call.f()
				}
			})
		})
	}
}

func BenchmarkLogAttrs(b *testing.B) {
	err := errors.New("something went wrong")

	logger := slog.New(LogHandler{(slog.HandlerOptions{}).NewTextHandler(io.Discard)})

	for _, call := range []struct {
		name string
		f    func()
	}{
		{
			// The number should match nAttrsInline in slog/record.go.
			// This should exercise the code path where no allocations
			// happen in Record or Attr. If there are allocations, they
			// should only be from Duration.String and Time.String.
			"5 args",
			func() {
				logger.Error(TestMessage,
					slog.Any("err", WithLogAttrs(err,
						slog.String("string", TestString),
						slog.Int("status", TestInt),
						slog.Duration("duration", TestDuration),
						slog.Time("time", TestTime),
						slog.Any("error", TestError),
					)),
				)
			},
		},
		{
			"10 args",
			func() {
				logger.Error(TestMessage,
					slog.Any("err", WithLogAttrs(err,
						slog.String("string", TestString),
						slog.Int("status", TestInt),
						slog.Duration("duration", TestDuration),
						slog.Time("time", TestTime),
						slog.Any("error", TestError),
						slog.String("string", TestString),
						slog.Int("status", TestInt),
						slog.Duration("duration", TestDuration),
						slog.Time("time", TestTime),
						slog.Any("error", TestError),
					)),
				)
			},
		},
		{
			"40 args",
			func() {
				logger.Error(TestMessage,
					slog.Any("err", WithLogAttrs(err,
						slog.String("string", TestString),
						slog.Int("status", TestInt),
						slog.Duration("duration", TestDuration),
						slog.Time("time", TestTime),
						slog.Any("error", TestError),
						slog.String("string", TestString),
						slog.Int("status", TestInt),
						slog.Duration("duration", TestDuration),
						slog.Time("time", TestTime),
						slog.Any("error", TestError),
						slog.String("string", TestString),
						slog.Int("status", TestInt),
						slog.Duration("duration", TestDuration),
						slog.Time("time", TestTime),
						slog.Any("error", TestError),
						slog.String("string", TestString),
						slog.Int("status", TestInt),
						slog.Duration("duration", TestDuration),
						slog.Time("time", TestTime),
						slog.Any("error", TestError),
						slog.String("string", TestString),
						slog.Int("status", TestInt),
						slog.Duration("duration", TestDuration),
						slog.Time("time", TestTime),
						slog.Any("error", TestError),
						slog.String("string", TestString),
						slog.Int("status", TestInt),
						slog.Duration("duration", TestDuration),
						slog.Time("time", TestTime),
						slog.Any("error", TestError),
						slog.String("string", TestString),
						slog.Int("status", TestInt),
						slog.Duration("duration", TestDuration),
						slog.Time("time", TestTime),
						slog.Any("error", TestError),
						slog.String("string", TestString),
						slog.Int("status", TestInt),
						slog.Duration("duration", TestDuration),
						slog.Time("time", TestTime),
						slog.Any("error", TestError),
					)),
				)
			},
		},
	} {
		b.Run(call.name, func(b *testing.B) {
			b.ReportAllocs()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					call.f()
				}
			})
		})
	}
}
