package raven

import "sync/atomic"

// TeeLogger fans out every log event to two loggers simultaneously —
// a Primary and a Secondary. This is useful for writing to both a
// terminal and a file at the same time, for example

type TeeLogger struct {
	Primary   Logger // anchor-capable logger (e.g. Buffered)
	Secondary Logger // plain output logger (e.g. Unbuffered or file)
	thresh    atomic.Int32
}

// NewTee creates a TeeLogger that fans output to both a and b
// Returns the logger and a stop function that closes both loggers
// The caller must call stop() when done
func NewTee(a RootLogger, b RootLogger) (*TeeLogger, func()) {
	l := &TeeLogger{
		Primary:   a,
		Secondary: b,
	}
	l.thresh.Store(int32(Transient))

	stop := func() {
		a.Close()
		b.Close()
	}

	return l, stop
}

// Parent returns the Primary logger
// Anchor traversal relies on a single root, so Primary is the authority
func (l *TeeLogger) Parent() Logger {
	return l.Primary
}

// Threshold returns the minimum severity level for this logger
func (l *TeeLogger) Threshold() Level {
	return Level(l.thresh.Load())
}

// SetThreshold updates the minimum severity level
func (l *TeeLogger) SetThreshold(severity Level) Logger {
	l.thresh.Store(int32(severity))
	return l
}

// forward sends the log event to both Primary and Secondary.
// The Secondary always receives a zeroed AnchorID since it does
// not support anchored lines.
func (l *TeeLogger) forward(severity Level, msg string, fields []Fielder, opts []PrinterOption, ctx LogContext) {
	ctx.RaiseSeverity(l.Threshold())

	// primary receives the full context including any anchor ID
	l.Primary.forward(severity, msg, fields, opts, ctx)

	// secondary never handles anchored lines
	ctx.AnchorID = 0
	l.Secondary.forward(severity, msg, fields, opts, ctx)
}

func (l *TeeLogger) Transient(msg string, fields ...Fielder) Logger {
	l.forward(Transient, msg, fields, nil, LogContext{})
	return l
}

func (l *TeeLogger) Verbose(msg string, fields ...Fielder) Logger {
	l.forward(Verbose, msg, fields, nil, LogContext{})
	return l
}

func (l *TeeLogger) Info(msg string, fields ...Fielder) Logger {
	l.forward(Info, msg, fields, nil, LogContext{})
	return l
}

func (l *TeeLogger) Warning(msg string, fields ...Fielder) Logger {
	l.forward(Warning, msg, fields, nil, LogContext{})
	return l
}

func (l *TeeLogger) Error(msg string, fields ...Fielder) Logger {
	l.forward(Error, msg, fields, nil, LogContext{})
	return l
}

func (l *TeeLogger) Emit(severity Level, msg string, fields ...Fielder) Logger {
	l.forward(severity, msg, fields, nil, LogContext{})
	return l
}