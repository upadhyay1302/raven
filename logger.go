package raven

// RootLogger is the top-level logger interface
// It owns resources and must be closed when done
type RootLogger interface {
	Logger

	Close()
}

// Logger is the core interface for writing structured log lines
type Logger interface {
	// Threshold returns the minimum severity level this Logger will accept
	Threshold() Level

	// SetThreshold updates the minimum severity level for this Logger
	SetThreshold(severity Level) Logger

	// Transient writes a progress message
	Transient(msg string, fields ...Fielder) Logger

	// Verbose writes a debug message
	Verbose(msg string, fields ...Fielder) Logger

	// Info writes a message about a normal event
	Info(msg string, fields ...Fielder) Logger

	// Warning writes a message about something unexpected
	Warning(msg string, fields ...Fielder) Logger

	// Error writes a message about something that went wrong
	Error(msg string, fields ...Fielder) Logger

	// Emit writes a message at the given severity level
	Emit(severity Level, msg string, fields ...Fielder) Logger

	// forward is called by child Loggers to pass log events up to the root
	forward(severity Level, msg string, fields []Fielder, opts []PrinterOption, data LogContext)
}

// ChildLogger is implemented by any Logger that delegates to a parent
type ChildLogger interface {
	Parent() Logger
}

// AnchorAdder is implemented by Loggers that support pinning a line
// to the bottom of the terminal for live progress updates
type AnchorAdder interface {
	AddAnchor(source Logger) Logger
}

// AnchorRemover is implemented by anchored Loggers to allow
// cleanup of the pinned line when it is no longer needed
type AnchorRemover interface {
	RemoveAnchor()
}