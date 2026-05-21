package raven

import "sync/atomic"

// NoAnchorLogger is returned by AddAnchor when the root Logger does not
// support terminal anchoring
// It behaves identically to its parent Logger, but satisfies the
// AnchorRemover interface so callers never need to check for anchor support
type NoAnchorLogger struct {
	parent    Logger
	threshold atomic.Uint32 
}

func newNoAnchorLogger(parent Logger) *NoAnchorLogger {
	l := &NoAnchorLogger{parent: parent}
	l.threshold.Store(uint32(Transient))
	return l
}

// RemoveAnchor is a no-op, there is no real anchor to remove
// It exists solely to satisfy the AnchorRemover interface.
func (l *NoAnchorLogger) RemoveAnchor() {}

// Parent returns the logger this NoAnchorLogger delegates to
func (l *NoAnchorLogger) Parent() Logger {
	return l.parent
}

// Threshold returns the minimum severity level for this particular logger
func (l *NoAnchorLogger) Threshold() Level {
	return Level(l.threshold.Load())
}

// SetThreshold updates the minimum severity level and returns the logger
// to support method chaining
func (l *NoAnchorLogger) SetThreshold(severity Level) Logger {
	l.threshold.Store(uint32(severity))
	return l
}


func (l *NoAnchorLogger) forward(severity Level, msg string, fields []Fielder, opts []PrinterOption, ctx LogContext) {
	ctx.RaiseSeverity(l.Threshold())
	l.parent.forward(severity, msg, fields, opts, ctx)
}

func (l *NoAnchorLogger) Transient(msg string, fields ...Fielder) Logger {
	l.forward(Transient, msg, fields, nil, LogContext{})
	return l
}

func (l *NoAnchorLogger) Verbose(msg string, fields ...Fielder) Logger {
	l.forward(Verbose, msg, fields, nil, LogContext{})
	return l
}

func (l *NoAnchorLogger) Info(msg string, fields ...Fielder) Logger {
	l.forward(Info, msg, fields, nil, LogContext{})
	return l
}

func (l *NoAnchorLogger) Warning(msg string, fields ...Fielder) Logger {
	l.forward(Warning, msg, fields, nil, LogContext{})
	return l
}

func (l *NoAnchorLogger) Error(msg string, fields ...Fielder) Logger {
	l.forward(Error, msg, fields, nil, LogContext{})
	return l
}

func (l *NoAnchorLogger) Emit(severity Level, msg string, fields ...Fielder) Logger {
	l.forward(severity, msg, fields, nil, LogContext{})
	return l
}