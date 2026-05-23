package raven

import (
	"sync"
	"sync/atomic"
)

// AnchoredLogger is a Logger that pins Transient messages to a fixed line
// at the bottom of the terminal, updating it in place rather than scrolling
//
// Non-transient messages (Verbose, Info, Warning, Error) are passed through
// to the parent Logger and scroll normally above the anchored line

type AnchoredLogger struct {
	parent    Logger
	thresh    atomic.Int32 // minimum severity, defaults to Transient (0)
	anchorID  atomic.Int32 // 0 after RemoveAnchor is called
	closeOnce sync.Once    // guarantees onRemove is called exactly once
	onRemove  func()       // notifies Buffered to clean up the pinned line
}

// newAnchor creates an AnchoredLogger that routes Transient messages to the
// terminal line identified by id, and calls onRemove when RemoveAnchor is called
func newAnchor(parent Logger, id int32, onRemove func()) *AnchoredLogger {
	l := &AnchoredLogger{
		parent:   parent,
		onRemove: onRemove,
	}
	l.anchorID.Store(id)
	l.thresh.Store(int32(Transient))
	return l
}

// RemoveAnchor releases the pinned terminal line and stops updating it
// Safe to call from any goroutine. Subsequent calls are no-ops
func (l *AnchoredLogger) RemoveAnchor() {
	l.closeOnce.Do(func() {
		l.anchorID.Store(0) 
		if l.onRemove != nil {
			l.onRemove()
		}
	})
}

func (l *AnchoredLogger) Parent() Logger {
	return l.parent
}

// Threshold returns the minimum severity level for this logger
func (l *AnchoredLogger) Threshold() Level {
	return Level(l.thresh.Load())
}

// SetThreshold updates the minimum severity level
func (l *AnchoredLogger) SetThreshold(severity Level) Logger {
	l.thresh.Store(int32(severity))
	return l
}

// forward routes the log event to the correct destination:
// - Transient messages target the pinned anchor line
// - Everything else passes through to the parent
func (l *AnchoredLogger) forward(severity Level, msg string, fields []Fielder, opts []PrinterOption, ctx LogContext) {
	ctx.RaiseSeverity(l.Threshold())

	// only Transient messages target the anchored line
	// after RemoveAnchor, anchorID is 0 so Transient falls through to parent
	if severity == Transient {
		ctx.AnchorID = l.anchorID.Load()
	}

	l.parent.forward(severity, msg, fields, opts, ctx)
}

func (l *AnchoredLogger) Transient(msg string, fields ...Fielder) Logger {
	l.forward(Transient, msg, fields, nil, LogContext{})
	return l
}

func (l *AnchoredLogger) Verbose(msg string, fields ...Fielder) Logger {
	l.forward(Verbose, msg, fields, nil, LogContext{})
	return l
}

func (l *AnchoredLogger) Info(msg string, fields ...Fielder) Logger {
	l.forward(Info, msg, fields, nil, LogContext{})
	return l
}

func (l *AnchoredLogger) Warning(msg string, fields ...Fielder) Logger {
	l.forward(Warning, msg, fields, nil, LogContext{})
	return l
}

func (l *AnchoredLogger) Error(msg string, fields ...Fielder) Logger {
	l.forward(Error, msg, fields, nil, LogContext{})
	return l
}

func (l *AnchoredLogger) Emit(severity Level, msg string, fields ...Fielder) Logger {
	l.forward(severity, msg, fields, nil, LogContext{})
	return l
}