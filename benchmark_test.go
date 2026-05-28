package raven

import (
	"io"
	"math"
	"testing"
	"time"
)

// bench runs a standard suite of benchmarks against any Logger constructor.
// This allows fair comparison across Buffered, Unbuffered, Text, and JSON.
func bench(b *testing.B, newLogger func(fields ...Fielder) (Logger, func())) {
	b.Helper()

	b.Run("fields=none/threshold=error/threads=1", func(b *testing.B) {
		log, stop := newLogger()
		defer stop()
		log.SetThreshold(Error)
		runMsg(b, log, benchMessage)
	})

	b.Run("fields=none/threshold=info/threads=1", func(b *testing.B) {
		log, stop := newLogger()
		defer stop()
		log.SetThreshold(Info)
		runMsg(b, log, benchMessage)
	})

	b.Run("fields=static/threshold=error/threads=1", func(b *testing.B) {
		log, stop := newLogger(sampleFields()...)
		defer stop()
		log.SetThreshold(Error)
		runMsg(b, log, benchMessage)
	})

	b.Run("fields=static/threshold=info/threads=1", func(b *testing.B) {
		log, stop := newLogger(sampleFields()...)
		defer stop()
		log.SetThreshold(Info)
		runMsg(b, log, benchMessage)
	})

	b.Run("fields=dynamic/threshold=error/threads=1", func(b *testing.B) {
		log, stop := newLogger()
		defer stop()
		log.SetThreshold(Error)
		runMsgWithFields(b, log, benchMessage, sampleFields)
	})

	b.Run("fields=dynamic/threshold=info/threads=1", func(b *testing.B) {
		log, stop := newLogger()
		defer stop()
		log.SetThreshold(Info)
		runMsgWithFields(b, log, benchMessage, sampleFields)
	})

	b.Run("fields=halfandhalf/threshold=error/threads=1", func(b *testing.B) {
		log, stop := newLogger(firstHalfFields()...)
		defer stop()
		log.SetThreshold(Error)
		runMsgWithFields(b, log, benchMessage, secondHalfFields)
	})

	b.Run("fields=halfandhalf/threshold=info/threads=1", func(b *testing.B) {
		log, stop := newLogger(firstHalfFields()...)
		defer stop()
		log.SetThreshold(Info)
		runMsgWithFields(b, log, benchMessage, secondHalfFields)
	})

	// parallel benchmarks simulate real-world concurrent logging
	b.Run("fields=static/threshold=error/threads=8", func(b *testing.B) {
		log, stop := newLogger(sampleFields()...)
		defer stop()
		log.SetThreshold(Error)
		b.SetParallelism(8)
		runParallelMsg(b, log, benchMessage)
	})

	b.Run("fields=static/threshold=info/threads=8", func(b *testing.B) {
		log, stop := newLogger(sampleFields()...)
		defer stop()
		log.SetThreshold(Info)
		b.SetParallelism(8)
		runParallelMsg(b, log, benchMessage)
	})
}

// --- Benchmark entry points ---

func Benchmark_UnbufferedJSON(b *testing.B) {
	bench(b, newUnbufferedJSON)
}

func Benchmark_UnbufferedText_NoColor(b *testing.B) {
	bench(b, newUnbufferedTextNoColor)
}

func Benchmark_UnbufferedText_Color(b *testing.B) {
	bench(b, newUnbufferedTextColor)
}

func Benchmark_BufferedText_Color(b *testing.B) {
	bench(b, newBufferedTextColor)
}

// --- Logger constructors for benchmarks ---

func newUnbufferedJSON(fields ...Fielder) (Logger, func()) {
	l := NewUnbuffered(io.Discard, &JSONPrinter{})
	var log Logger = l
	if len(fields) > 0 {
		log = WithFields(log, fields...)
	}
	return log, l.Close
}

func newUnbufferedTextNoColor(fields ...Fielder) (Logger, func()) {
	l := NewUnbuffered(io.Discard, (&TextPrinter{}).Configure(
		OptShowTime(true),
		OptShowLevel(true),
		OptColumnOffset(20),
	))
	var log Logger = l
	if len(fields) > 0 {
		log = WithFields(log, fields...)
	}
	return log, l.Close
}

func newUnbufferedTextColor(fields ...Fielder) (Logger, func()) {
	l := NewUnbuffered(io.Discard, (&TextPrinter{}).Configure(
		OptPalette(DefaultPalette),
		OptShowTime(true),
		OptShowLevel(true),
		OptColumnOffset(20),
	))
	var log Logger = l
	if len(fields) > 0 {
		log = WithFields(log, fields...)
	}
	return log, l.Close
}

func newBufferedTextColor(fields ...Fielder) (Logger, func()) {
	l := NewBuffered(io.Discard, false, (&TextPrinter{}).Configure(
		OptPalette(DefaultPalette),
		OptShowTime(true),
		OptShowLevel(true),
		OptColumnOffset(20),
	))
	var log Logger = l
	if len(fields) > 0 {
		log = WithFields(log, fields...)
	}
	return log, l.Close
}

// --- Benchmark runners ---

func runMsg(b *testing.B, log Logger, msg string) {
	b.Helper()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Info(msg)
	}
}

func runMsgWithFields(b *testing.B, log Logger, msg string, fields func() []Fielder) {
	b.Helper()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Info(msg, fields()...)
	}
}

func runParallelMsg(b *testing.B, log Logger, msg string) {
	b.Helper()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			log.Info(msg)
		}
	})
}

// --- Test data ---

const benchMessage = "This is an example log line, medium length"

// sampleFields returns a representative mix of field types.
// Using only field types implemented in Raven's fields.go.
func sampleFields() []Fielder {
	return []Fielder{
		Bool("active", true),
		Dur("latency", time.Duration(2903458)*time.Microsecond),
		Err(io.EOF),
		Float64("ratio", math.Pi),
		Int64("count", math.MinInt64),
		String("service", "raven"),
		Time("ts", time.Date(1999, 11, 30, 23, 59, 59, 2391, time.UTC)),
	}
}

// firstHalfFields and secondHalfFields simulate a mix of static (pre-set)
// and dynamic (per-call) fields — a common real-world pattern.
func firstHalfFields() []Fielder {
	return []Fielder{
		Bool("active", true),
		String("env", "production"),
		Int("pid", 12345),
		Dur("uptime", 72*time.Hour),
	}
}

func secondHalfFields() []Fielder {
	return []Fielder{
		String("request_id", "abc-123"),
		Int64("bytes", math.MaxInt64),
		Float64("score", math.Pi),
		Err(io.ErrUnexpectedEOF),
	}
}

// allFields returns every field type Raven supports — used for
// comprehensive allocation profiling.
func allFields() []Fielder {
	return []Fielder{
		Bool("bool", true),
		Dur("dur", time.Duration(2903458)*time.Microsecond),
		Err(io.EOF),
		Float64("float64", math.Pi),
		Int("int", math.MinInt),
		Int64("int64", math.MinInt64),
		String("string", "flargenblargen"),
		Time("time", time.Date(1999, 11, 30, 23, 59, 59, 2391, time.UTC)),
		Uint("uint", math.MaxUint),
		Uint64("uint64", math.MaxUint64),
		Path("/var/log/raven"),
	}
}