package raven

import "sync/atomic"

// NullLogger is a Logger that silently discards all log output
// It is useful in tests, as a safe default, or to disable logging
// for a specific component without changing surrounding code.

type NullLogger struct {
	threshold atomic.Uint32 
}

// NewNullLogger returns a NullLogger that discards all log output
// Its default threshold is Transient, meaning it nominally accepts
// all levels, but discards them all silently
func NewNullLogger() *NullLogger {
	n := &NullLogger{}
	n.threshold.Store(uint32(Transient))
	return n
}

func (n *NullLogger) Close() {}

func (n *NullLogger) Threshold() Level {
	return Level(n.threshold.Load())
}

// SetThreshold updates the minimum severity level and returns the logger
// to allow method chaining
func (n *NullLogger) SetThreshold(severity Level) Logger {
	n.threshold.Store(uint32(severity))
	return n
}

// forward silently discards the log event
func (n *NullLogger) forward(severity Level, msg string, fields []Fielder, opts []PrinterOption, data LogContext) {
}

// The following methods all discard their input and return n to
// allow method chaining without breaking the caller's flow

func (n *NullLogger) Transient(msg string, fields ...Fielder) Logger { return n }
func (n *NullLogger) Verbose(msg string, fields ...Fielder) Logger   { return n }
func (n *NullLogger) Info(msg string, fields ...Fielder) Logger      { return n }
func (n *NullLogger) Warning(msg string, fields ...Fielder) Logger   { return n }
func (n *NullLogger) Error(msg string, fields ...Fielder) Logger     { return n }
func (n *NullLogger) Emit(severity Level, msg string, fields ...Fielder) Logger { return n }

