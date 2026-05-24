package raven

import "sync/atomic"

// ContextLogger wraps a parent Logger and enriches every log line it
// produces with a fixed set of fields and/or printer options
//
// The parent Logger is never modified — each ContextLogger is its own
// independent layer that adds context without affecting anything around it
type ContextLogger struct {
	parent Logger
	opts   []PrinterOption
	fields []Field   
	thresh atomic.Int32 
}

// newCustomizerLogger creates a ContextLogger that enriches log lines
// with the given options and fields before passing them to the parent
func newCustomizerLogger(parent Logger, opts []PrinterOption, fielders []Fielder) *ContextLogger {
	l := &ContextLogger{
		parent: parent,
		opts:   opts,
		fields: Fieldify(fielders), 
	}
	l.thresh.Store(int32(Transient))
	return l
}

// Parent returns the Logger this ContextLogger works with
func (l *ContextLogger) Parent() Logger {
	return l.parent
}

// Threshold returns the minimum severity level for this logger
func (l *ContextLogger) Threshold() Level {
	return Level(l.thresh.Load())
}

// SetThreshold updates the minimum severity level
func (l *ContextLogger) SetThreshold(severity Level) Logger {
	l.thresh.Store(int32(severity))
	return l
}

// forward merges this logger's fields and options into the context,
// then delegates the log event up to the parent
func (l *ContextLogger) forward(severity Level, msg string, fields []Fielder, opts []PrinterOption, ctx LogContext) {
	ctx.RaiseSeverity(l.Threshold())
	ctx.PrependFields(l.fields)

	var mergedOpts []PrinterOption
	if len(l.opts) > 0 || len(opts) > 0 {
		mergedOpts = make([]PrinterOption, 0, len(l.opts)+len(opts))
		mergedOpts = append(mergedOpts, l.opts...)
		mergedOpts = append(mergedOpts, opts...)
	}

	l.parent.forward(severity, msg, fields, mergedOpts, ctx)
}

func (l *ContextLogger) Transient(msg string, fields ...Fielder) Logger {
	l.forward(Transient, msg, fields, nil, LogContext{})
	return l
}

func (l *ContextLogger) Verbose(msg string, fields ...Fielder) Logger {
	l.forward(Verbose, msg, fields, nil, LogContext{})
	return l
}

func (l *ContextLogger) Info(msg string, fields ...Fielder) Logger {
	l.forward(Info, msg, fields, nil, LogContext{})
	return l
}

func (l *ContextLogger) Warning(msg string, fields ...Fielder) Logger {
	l.forward(Warning, msg, fields, nil, LogContext{})
	return l
}

func (l *ContextLogger) Error(msg string, fields ...Fielder) Logger {
	l.forward(Error, msg, fields, nil, LogContext{})
	return l
}

func (l *ContextLogger) Emit(severity Level, msg string, fields ...Fielder) Logger {
	l.forward(severity, msg, fields, nil, LogContext{})
	return l
}