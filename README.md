# Raven 🖤

A fast, structured logging library for Go with support for colored
terminal output, JSON formatting, live anchored progress lines,
and safe concurrent use.

Built from scratch as an original implementation, inspired by the
design of [frog](https://github.com/danbrakeley/frog).

## Features

- **Structured logging** — key-value fields on every log line
- **Five log levels** — Transient, Verbose, Info, Warning, Error
- **Plain text and JSON output** — swap formatters without changing call sites
- **ANSI color support** — three built-in palettes plus full customization via `WithLevel()`
- **Live anchored lines** — pin progress bars to the terminal bottom while logs scroll above
- **Goroutine safe** — `Buffered` serialises all writes through a channel; `Unbuffered` uses a mutex
- **Terminal detection** — automatically disables colors and anchoring when output is piped
- **Respects `NO_COLOR`** — honours the [no-color.org](https://no-color.org) standard
- **Zero personal dependencies** — only uses stdlib and well-maintained official packages

## Installation

```bash
go get github.com/upadhyay1302/raven
```

## Quick Start

```go
log := raven.New(raven.Auto)
defer log.Close()

log.Info("server started", raven.Int("port", 8080))
log.Warning("high memory usage", raven.Float64("percent", 87.3))
log.Error("connection failed", raven.Err(err))
```

Output:
```
2026.05.31-11:00:00 [inf] server started       port=8080
2026.05.31-11:00:00 [WRN] high memory usage    percent=87.3
2026.05.31-11:00:00 [ERR] connection failed    error=connection refused
```

## Logger Styles

```go
raven.New(raven.Auto)    // detects terminal; enables colors + anchors if supported
raven.New(raven.Direct)  // colors, no anchors, no buffering
raven.New(raven.Plain)   // plain text, no colors, no buffering
raven.New(raven.JSON)    // one JSON object per line
```

JSON output:
```json
{"ts":"2026-05-31T11:00:00Z","level":"info","msg":"server started","port":8080}
```

## Structured Fields

```go
raven.Bool("active", true)
raven.String("env", "production")
raven.Int("port", 8080)
raven.Int64("count", math.MaxInt64)
raven.Uint("id", 42)
raven.Uint64("max", math.MaxUint64)
raven.Float64("ratio", 3.14)
raven.Dur("latency", 2500*time.Millisecond)
raven.Time("started_at", time.Now())
raven.Err(err)
raven.Path("/var/log/app.log")
raven.Stringer("level", myStringer) // accepts any fmt.Stringer
```

## Threshold Filtering

```go
log.SetThreshold(raven.Transient) // show everything
log.SetThreshold(raven.Verbose)   // hide Transient only
log.SetThreshold(raven.Info)      // hide Transient and Verbose (default)
log.SetThreshold(raven.Warning)   // show only Warning and Error
log.SetThreshold(raven.Error)     // show only Error
```

Filtered messages cost **~7ns and zero allocations** — safe to leave
verbose logging in production code without performance impact.

## Child Loggers

Child loggers add context without modifying the parent:

```go
// add static fields to every line
dbLog := raven.WithFields(log,
    raven.String("layer", "database"),
    raven.String("driver", "postgres"),
)
dbLog.Info("query executed", raven.Dur("took", 42*time.Millisecond))

// add printer options per child
quietLog := raven.WithOptions(log, raven.OptPalette(raven.MutedPalette))

// add both fields and options together
auditLog := raven.WithContext(log,
    []raven.PrinterOption{raven.OptPalette(raven.BoldPalette)},
    []raven.Fielder{raven.String("audit", "true")},
)
```

## Anchored Lines

Anchored lines stay pinned at the bottom of the terminal while
regular log lines scroll above — ideal for progress bars:

```go
log := raven.New(raven.Auto)
defer log.Close()

var wg sync.WaitGroup
for i := 0; i < 3; i++ {
    wg.Add(1)
    go func(anchor raven.Logger, id int) {
        defer wg.Done()
        defer raven.RemoveAnchor(anchor)
        for pct := 0; pct <= 100; pct++ {
            anchor.Transient("downloading",
                raven.Int("thread", id),
                raven.Int("percent", pct),
            )
            time.Sleep(50 * time.Millisecond)
        }
        log.Info("download complete", raven.Int("thread", id))
    }(raven.AddAnchor(log), i)
}
wg.Wait()
```

When output is piped to a file, anchoring is automatically disabled
and transient lines are suppressed — no ANSI codes in log files.

## Color Palettes

```go
raven.New(raven.Auto, raven.OptPalette(raven.DefaultPalette))
raven.New(raven.Auto, raven.OptPalette(raven.MutedPalette))
raven.New(raven.Auto, raven.OptPalette(raven.BoldPalette))

// customize individual levels using the WithLevel builder
custom := raven.DefaultPalette.
    WithLevel(raven.Error, raven.Magenta, raven.DarkMagenta).
    WithLevel(raven.Warning, raven.Cyan, raven.DarkCyan)
raven.New(raven.Auto, raven.OptPalette(custom))
```

## TeeLogger

Fan out every log event to two loggers simultaneously:

```go
file, _ := os.Create("app.log")
primary   := raven.NewBuffered(os.Stdout, true, &raven.TextPrinter{})
secondary := raven.NewUnbuffered(file, &raven.JSONPrinter{})

tee, stop := raven.NewTee(primary, secondary)
defer stop()

tee.Info("goes to both terminal and log file simultaneously")
```

## Printer Options

```go
raven.New(raven.Auto,
    raven.OptShowTime(true),
    raven.OptShowLevel(true),
    raven.OptColumnOffset(40),
    raven.OptPalette(raven.BoldPalette),
    raven.OptMsgThenFields,
    raven.OptFieldsThenMsg,
    raven.OptMaxLineWidth(120),
)
```

## Design Decisions

**Why a channel in `Buffered`?**
Serialising all writes through a single goroutine means cursor
movement for anchored lines never races with regular log output,
even with hundreds of concurrent goroutines.

**Why `atomic.Int32` throughout?**
All threshold fields use `sync/atomic` types rather than mutexes.
Threshold reads happen on every log call — keeping them lock-free
reduces contention under high concurrency.

**Why `sync.Once` in `AnchoredLogger`?**
`RemoveAnchor` must be safe to call multiple times from any goroutine.
`sync.Once` guarantees the cleanup callback runs exactly once
regardless of how many goroutines call it simultaneously.

**Why `internal/terminal` instead of external ANSI packages?**
Raven has zero personal/third-party ANSI dependencies. The
`internal/terminal` package implements exactly what Raven needs using
raw VT100 escape codes and the official `golang.org/x/term` package.

**Why generics in `findInChain`?**
The parent chain traversal logic is identical whether searching for
an `AnchorAdder` or `AnchorRemover`. A generic function eliminates
the duplication and makes adding future interface searches trivial.

**Why `ContextLogger` pre-converts fields?**
Static fields passed to `WithFields()` are converted from `Fielder`
to `Field` once at creation time and cached. This avoids re-converting
them on every log call, which the benchmarks confirm significantly
reduces allocations.

## Benchmarks

Run on Intel Core i5-1030NG7 @ 1.10GHz (amd64):

```
Benchmark_UnbufferedJSON/fields=none/threshold=error      ~7 ns/op      0 B/op    0 allocs/op
Benchmark_UnbufferedJSON/fields=none/threshold=info    ~1298 ns/op    416 B/op    4 allocs/op
Benchmark_UnbufferedText/fields=none/threshold=error      ~7 ns/op      0 B/op    0 allocs/op
Benchmark_UnbufferedText/fields=none/threshold=info    ~1291 ns/op    432 B/op    6 allocs/op
Benchmark_BufferedText/fields=none/threshold=error        ~7 ns/op      0 B/op    0 allocs/op
Benchmark_BufferedText/fields=none/threshold=info      ~2611 ns/op    432 B/op    6 allocs/op
```

Filtered messages (below threshold) cost ~7ns with **zero allocations**.

Run benchmarks yourself:
```bash
./bench.sh
```

## Running the Demos

```bash
go run ./cmd/demo/        # full featured demo with CLI flags
go run ./cmd/anchors/     # live progress bar demo
go run ./cmd/colortest/   # color palette showcase
go run ./cmd/lengthtest/  # long line cropping test
go run ./cmd/sizetest/    # terminal size detection
```

## Running Tests

```bash
# all tests with race detector
go test -race ./...

# benchmarks with allocation stats
go test -bench=. -benchmem ./...

# or use the benchmark script
./bench.sh
```

## Dependencies

| Package | Purpose |
|---|---|
| `golang.org/x/term` | Terminal size detection and TTY checks |
| `github.com/mattn/go-runewidth` | Correct column width for CJK and wide characters |

## License

MIT