package raven

// LogContext carries internal metadata as a log event travels from a child
// Logger up through the chain to the root Logger
type LogContext struct {
	// AnchorID identifies which pinned terminal line this log belongs to.
	// 0 means this is a regular log line with no anchor.
	AnchorID int32

	// Severity tracks the effective minimum log level as it moves up the chain.
	// Each parent Logger raises this to its own threshold if needed, so the
	// most restrictive threshold in the chain always wins.
	Severity Level

	// CachedFields holds Fields that have already been converted from Fielders.
	// Child Loggers cache their static fields here to avoid reprocessing them
	// on every log call.
	CachedFields []Field
}

// RaiseSeverity updates Severity to whichever is higher — the current value
// or the incoming one. This ensures the strictest threshold wins.
func (c *LogContext) RaiseSeverity(incoming Level) {
	if c.Severity < incoming {
		c.Severity = incoming
	}
}

// PrependFields inserts the given fields before any existing cached fields,
// so parent fields always appear before child fields in the output.
func (c *LogContext) PrependFields(incoming []Field) {
	if len(incoming) == 0 {
		return
	}
	c.CachedFields = append(incoming, c.CachedFields...)
}