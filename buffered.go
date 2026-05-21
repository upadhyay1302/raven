package raven

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	"github.com/upadhyay1302/raven/internal/terminal"
)

// msgKind says what operation the processor goroutine should perform
type msgKind byte

const (
	kindPrint        msgKind = iota // render and write a log line
	kindAnchorAdd                   // register a new anchored line at the bottom
	kindAnchorRemove                // deregister and clean up an anchored line
)

type message struct {
	kind     msgKind
	anchorID int32
	severity Level
	text     string
}

// anchoredLine tracks a single live-updating line pinned to the terminal bottom
type anchoredLine struct {
	id  int32
	str string
}

// Buffered is Raven's goroutine-safe root Logger

// It serialises all terminal writes through a channel to a single processor
// goroutine. This means cursor movement and redraws never race with each other,
// even when dozens of goroutines are logging simultaneously

type Buffered struct {
	writer    io.Writer
	printer   Printer
	ch        chan message
	wg        sync.WaitGroup
	closed    atomic.Int32 // 0=open, 1+=closed; guards against double-close
	anchorSeq atomic.Int32 // monotonically increasing ID generator for anchors
	thresh    atomic.Int32 // minimum severity level
}

// NewBuffered creates a Buffered logger that writes to w using the given Printer

// If detectWidth is true, Raven queries the terminal width at startup and crops
// transient lines to fit, preventing them from wrapping and scrambling output
// Returns nil if w is nil.
func NewBuffered(w io.Writer, detectWidth bool, prn Printer) *Buffered {
	if w == nil {
		return nil
	}

	l := &Buffered{
		writer:  w,
		printer: prn,
		ch:      make(chan message),
	}
	l.thresh.Store(int32(Info))

	l.wg.Add(1)
	go func() {
		defer l.wg.Done()
		if detectWidth {
			cols, _, err := terminal.GetSize()
			if err != nil || cols <= 0 {
				cols = -1
			}
			l.printer = l.printer.Configure(OptMaxLineWidth(cols))
		}
		l.processor()
	}()

	return l
}

// Close flushes all pending output and shuts down the processor goroutin
 func (l *Buffered) Close() {
	if l.closed.Add(1) != 1 {
		l.wg.Wait()
		return
	}
	close(l.ch)
	l.wg.Wait()
}

// AddAnchor pins a new line to the bottom of the terminal and returns a Logger
// whose Transient messages update that line in place
func (l *Buffered) AddAnchor(parent Logger) Logger {
	id := l.anchorSeq.Add(1)
	l.ch <- message{kind: kindAnchorAdd, anchorID: id}
	onRemove := func() {
		l.ch <- message{kind: kindAnchorRemove, anchorID: id}
	}
	return newAnchor(parent, id, onRemove)
}

// processor is the sole goroutine permitted to write to the terminal
// It reads messages from the channel until Close() shuts it down
func (l *Buffered) processor() {
	lines := make([]anchoredLine, 0, 8)

	// findIdx locates an anchor by ID, returning its index and whether it was found
	findIdx := func(id int32) (int, bool) {
		for i := range lines {
			if lines[i].id == id {
				return i, true
			}
		}
		return -1, false
	}

	// range over channel exits cleanly when Close() calls close(l.ch)
	for msg := range l.ch {
		switch msg.kind {

		case kindAnchorAdd:
			// scroll the terminal down to make room for the new pinned line
			fmt.Fprint(l.writer, "\n")
			lines = append(lines, anchoredLine{id: msg.anchorID})

		case kindAnchorRemove:
			idx, ok := findIdx(msg.anchorID)
			if !ok {
				continue // already removed, nothing to do
			}

			// remove from slice without allocating a new one
			lines = append(lines[:idx], lines[idx+1:]...)

			// redraw lines that shifted up due to the removal
			fmt.Fprint(l.writer, terminal.PrevLine(1+len(lines)-idx))
			for i := idx; i < len(lines); i++ {
				fmt.Fprint(l.writer, lines[i].str)
				fmt.Fprint(l.writer, terminal.EraseToLineEnd)
				fmt.Fprint(l.writer, terminal.NextLine(1))
			}
			fmt.Fprint(l.writer, terminal.EraseToLineEnd)

		case kindPrint:
			// fast path: no anchors active, just write the line directly
			if len(lines) == 0 {
				fmt.Fprintf(l.writer, "%s\n", msg.text)
				continue
			}

			// non-transient messages and unanchored messages scroll above
			// the pinned lines rather than updating one in place
			if msg.anchorID <= 0 || msg.severity > Transient {
				fmt.Fprint(l.writer, "\n")
				fmt.Fprint(l.writer, terminal.PrevLine(1+len(lines)))
				fmt.Fprintf(l.writer, "%s%s\n", msg.text, terminal.EraseToLineEnd)

				for _, line := range lines {
					fmt.Fprintf(l.writer, "%s%s\n", line.str, terminal.EraseToLineEnd)
				}

				if msg.anchorID <= 0 {
					continue
				}
			}

			// update the specific anchored line in place
			idx, ok := findIdx(msg.anchorID)
			if !ok {
				continue
			}
			lines[idx].str = msg.text
			offset := len(lines) - idx

			fmt.Fprint(l.writer, terminal.PrevLine(offset))
			fmt.Fprint(l.writer, msg.text)
			fmt.Fprint(l.writer, terminal.EraseToLineEnd)
			fmt.Fprint(l.writer, terminal.NextLine(offset))
		}
	}
}

// Threshold returns the current minimum severity level
func (l *Buffered) Threshold() Level {
	return Level(l.thresh.Load())
}

// SetThreshold updates the minimum severity level
func (l *Buffered) SetThreshold(severity Level) Logger {
	l.thresh.Store(int32(severity))
	return l
}

// forward is called by child loggers to pass a log event up to this root logger.=
func (l *Buffered) forward(severity Level, msg string, fields []Fielder, opts []PrinterOption, ctx LogContext) {
	ctx.RaiseSeverity(l.Threshold())

	// anchored transient lines always pass through regardless of threshold
	keep := severity >= ctx.Severity
	if ctx.AnchorID != 0 && severity == Transient {
		keep = true
	}
	if !keep {
		return
	}

	l.ch <- message{
		kind:     kindPrint,
		anchorID: ctx.AnchorID,
		severity: severity,
		text:     l.printer.Render(severity, opts, msg, FieldifyAndAppend(ctx.CachedFields, fields)),
	}
}

func (l *Buffered) Transient(msg string, fields ...Fielder) Logger {
	l.forward(Transient, msg, fields, nil, LogContext{})
	return l
}

func (l *Buffered) Verbose(msg string, fields ...Fielder) Logger {
	l.forward(Verbose, msg, fields, nil, LogContext{})
	return l
}

func (l *Buffered) Info(msg string, fields ...Fielder) Logger {
	l.forward(Info, msg, fields, nil, LogContext{})
	return l
}

func (l *Buffered) Warning(msg string, fields ...Fielder) Logger {
	l.forward(Warning, msg, fields, nil, LogContext{})
	return l
}

func (l *Buffered) Error(msg string, fields ...Fielder) Logger {
	l.forward(Error, msg, fields, nil, LogContext{})
	return l
}

func (l *Buffered) Emit(severity Level, msg string, fields ...Fielder) Logger {
	l.forward(severity, msg, fields, nil, LogContext{})
	return l
}