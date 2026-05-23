package raven

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"
)

// Unbuffered is a Logger that writes directly to the underlying writer
// without any internal buffering or background goroutines
type Unbuffered struct {
	writer  io.Writer
	printer Printer
	thresh  atomic.Int32 // minimum severity level
	mu      sync.Mutex   // protects writer from concurrent writes
}

// NewUnbuffered creates an Unbuffered logger that writes to w using the given Printer.
// Default threshold is Info
func NewUnbuffered(w io.Writer, prn Printer) *Unbuffered {
	l := &Unbuffered{
		writer:  w,
		printer: prn,
	}
	l.thresh.Store(int32(Info))
	return l
}

func (l *Unbuffered) Close() {}

// Threshold returns the current minimum severity level
func (l *Unbuffered) Threshold() Level {
	return Level(l.thresh.Load())
}

// SetThreshold updates the minimum severity level
func (l *Unbuffered) SetThreshold(severity Level) Logger {
	l.thresh.Store(int32(severity))
	return l
}

// forward renders and writes the log event directly to the writer
// The mutex ensures concurrent calls don't interleave their output
func (l *Unbuffered) forward(severity Level, msg string, fields []Fielder, opts []PrinterOption, ctx LogContext) {
	ctx.RaiseSeverity(l.Threshold())

	if severity < ctx.Severity {
		return
	}

	rendered := l.printer.Render(severity, opts, msg, FieldifyAndAppend(ctx.CachedFields, fields))

	l.mu.Lock()
	fmt.Fprintf(l.writer, "%s\n", rendered)
	l.mu.Unlock()
}

func (l *Unbuffered) Transient(msg string, fields ...Fielder) Logger {
	l.forward(Transient, msg, fields, nil, LogContext{})
	return l
}

func (l *Unbuffered) Verbose(msg string, fields ...Fielder) Logger {
	l.forward(Verbose, msg, fields, nil, LogContext{})
	return l
}

func (l *Unbuffered) Info(msg string, fields ...Fielder) Logger {
	l.forward(Info, msg, fields, nil, LogContext{})
	return l
}

func (l *Unbuffered) Warning(msg string, fields ...Fielder) Logger {
	l.forward(Warning, msg, fields, nil, LogContext{})
	return l
}

func (l *Unbuffered) Error(msg string, fields ...Fielder) Logger {
	l.forward(Error, msg, fields, nil, LogContext{})
	return l
}

func (l *Unbuffered) Emit(severity Level, msg string, fields ...Fielder) Logger {
	l.forward(severity, msg, fields, nil, LogContext{})
	return l
}
