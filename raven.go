package raven

import (
	"io"
	"os"

	"golang.org/x/term"
)

// Style controls how a new Logger formats and delivers its output
type Style byte

const (
	// Auto detects if stdout is a terminal. If so, enables colors and
	// anchored lines. Falls back to Plain if no terminal is detected
	Auto Style = iota

	// Unbuffered includes colors but no anchored lines and no buffering
	Unbuffered

	// Plain outputs plain text with timestamps but no colors or anchoring
	Plain

	// JSON outputs each log line as a single JSON object
	JSON
)

// HasTerminal reports whether the given writer is connected to a terminal.
func HasTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

// New creates a Logger that writes to os.Stdout using the given Style
func New(style Style, opts ...PrinterOption) RootLogger {
	// Auto falls back to Plain when no terminal is detected
	isTerminal := false
	if style == Auto {
		isTerminal = HasTerminal(os.Stdout)
		if !isTerminal {
			style = Plain
		}
	}

	switch style {
	case Auto:
		prn := &TextPrinter{
			colors:       DefaultPalette.toANSI(),
			showTime:     true,
			showLevel:    true,
			columnOffset: 20,
		}
		return NewBuffered(os.Stdout, isTerminal, prn.Configure(opts...))

	case Unbuffered:
		prn := &TextPrinter{
			colors:       DefaultPalette.toANSI(),
			showTime:     true,
			showLevel:    true,
			columnOffset: 20,
		}
		return NewUnbuffered(os.Stdout, prn.Configure(opts...))

	case Plain:
		prn := &TextPrinter{
			showTime:     true,
			showLevel:    true,
			columnOffset: 20,
		}
		return NewUnbuffered(os.Stdout, prn.Configure(opts...))

	case JSON:
		return NewUnbuffered(os.Stdout, &JSONPrinter{})
	}

	return nil
}

// findInChain walks up the Logger parent chain looking for a Logger
// that implements the given interface T. Uses Go generics to avoid
// duplicating the traversal logic for each interface type
func findInChain[T any](log Logger) (T, bool) {
	var zero T
	for log != nil {
		if match, ok := log.(T); ok {
			return match, true
		}
		log = Parent(log)
	}
	return zero, false
}

// Parent returns the parent Logger of a ChildLogger, or nil if log
// has no parent or does not implement ChildLogger
func Parent(log Logger) Logger {
	child, ok := log.(ChildLogger)
	if !ok {
		return nil
	}
	return child.Parent()
}

// AddAnchor walks up the Logger chain to find an AnchorAdder and pins
// a new line to the bottom of the terminal. If no AnchorAdder is found,
// returns a NoAnchorLogger that gracefully degrades anchor behaviour
func AddAnchor(log Logger) Logger {
	adder, ok := findInChain[AnchorAdder](log)
	if !ok || adder == nil {
		// graceful degradation — return a pass-through with anchor interface
		return newNoAnchorLogger(log)
	}

	anchored := adder.AddAnchor(log)
	if anchored == nil {
		return log
	}
	return anchored
}

// RemoveAnchor walks up the Logger chain to find an AnchorRemover and
// releases the pinned terminal line. Safe to call even if the Logger
// was not anchored — it simply does nothing in that case
func RemoveAnchor(log Logger) {
	remover, ok := findInChain[AnchorRemover](log)
	if ok && remover != nil {
		remover.RemoveAnchor()
	}
}

// WithFields returns a new Logger that wraps log and always appends
// the given fields to every log line, without modifying log itself
func WithFields(log Logger, fields ...Fielder) Logger {
	return newCustomizerLogger(log, nil, fields)
}

// WithOptions returns a new Logger that wraps log and always applies
// the given PrinterOptions to every log line, without modifying log itself
func WithOptions(log Logger, opts ...PrinterOption) Logger {
	return newCustomizerLogger(log, opts, nil)
}

// WithContext returns a new Logger that wraps log and always applies
// both the given PrinterOptions and fields to every log line
func WithContext(log Logger, opts []PrinterOption, fields []Fielder) Logger {
	return newCustomizerLogger(log, opts, fields)
}