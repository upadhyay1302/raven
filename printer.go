package raven

import (
	"fmt"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/upadhyay1302/raven/internal/terminal"
)

// colorDisabled is set at startup by checking for the NO_COLOR environment variable.
// See https://no-color.org for the standard this follows
var colorDisabled = false

func init() {
	_, exists := os.LookupEnv("NO_COLOR")
	colorDisabled = exists
}

// Printer is the interface for formatting a log event into a string.
type Printer interface {
	// Render formats a log event into a string ready for output.
	Render(severity Level, overrides []PrinterOption, msg string, fields []Field) string
	// Configure applies the given options to this Printer.
	Configure(...PrinterOption) Printer
}

// TextPrinter formats log lines as human-readable text, with optional ANSI colors
type TextPrinter struct {
	colors      ansiPalette
	showTime    bool
	showLevel   bool

	// columnOffset controls where fields begin relative to the message
	// Fields are always at least 3 spaces from the end of the message,
	// and snapped to the nearest multiple of 5
	// Example with columnOffset=40:
	//   [nfo] a long message that overflows the column offset        key=val
	//   [nfo] short message                                 key=val
	columnOffset int

	// fieldsOnLeft renders fields before the message instead of after.
	// Example:
	//   [nfo] key=val      message here
	fieldsOnLeft bool

	// maxLineWidth crops transient/anchor lines to prevent terminal wrapping.
	// 0 means no cropping.
	maxLineWidth int
}

func (p *TextPrinter) Configure(opts ...PrinterOption) Printer {
	for _, opt := range opts {
		switch o := opt.(type) {
		case optPalette:           
			p.colors = o.ANSIColors
		case optShowTime:          
			p.showTime = o.Visible
		case optShowLevel:         
			p.showLevel = o.Visible
		case optColumnOffset:      
			p.columnOffset = o.Indent
		case optMsgThenFields:     
			p.fieldsOnLeft = false
		case optFieldsThenMsg:     
			p.fieldsOnLeft = true
		case optMaxLineWidth:      
			p.maxLineWidth = o.Cols
		}
	}
	return p
}

func stripNewlines(s string) string {
	for strings.HasPrefix(s, "\n") {
		s = s[1:]
	}
	for strings.HasSuffix(s, "\n") {
		s = s[:len(s)-len("\n")]
	}
	return s
}

func (p *TextPrinter) Render(severity Level, overrides []PrinterOption, msg string, fields []Field) string {
	if len(overrides) > 0 {
		tmp := *p
		tmp.Configure(overrides...)
		return tmp.Render(severity, nil, msg, fields)
	}

	var primary, secondary string
	var useColor bool

	if !colorDisabled {
		primary = p.colors[severity][0]
		secondary = p.colors[severity][1]
		useColor = len(primary) > 0 && len(secondary) > 0
	}

	msg = sanitizeMessage(stripNewlines(msg))

	var buf strings.Builder
	buf.Grow(256)

	if useColor {
		buf.WriteString(secondary)
	}

	if p.showTime {
		buf.WriteString(fmt.Sprintf("%s ", time.Now().Format("2006.01.02-15:04:05")))
	}

	if p.showLevel {
		switch severity {
		case Transient:
			buf.WriteString("[~~~] ")
		case Verbose:
			buf.WriteString("[dbg] ")
		case Info:
			buf.WriteString("[inf] ")
		case Warning:
			buf.WriteString("[WRN] ")
		case Error:
			buf.WriteString("[ERR] ")
		default:
			buf.WriteString("[???] ")
		}
	}

	writeMsg := func() int {
		if useColor {
			buf.WriteString(primary)
		}
		buf.WriteString(msg)
		return utf8.RuneCountInString(msg)
	}

	writeFields := func() int {
		written := 0
		for i, f := range fields {
			if i != 0 {
				buf.WriteByte(' ')
				written++
			}
			val := f.Value
			if f.IsJSONString {
				if !f.IsJSONSafe {
					val = sanitizeFieldValue(val)
				}
				if len(val) == 0 || strings.ContainsAny(val, " \\") {
					val = `"` + val + `"`
				}
			}
			if useColor {
				buf.WriteString(secondary)
			}
			buf.WriteString(f.Name)
			written += utf8.RuneCountInString(f.Name)
			buf.WriteByte('=')
			written++
			if useColor {
				buf.WriteString(primary)
			}
			buf.WriteString(val)
			written += utf8.RuneCountInString(val)
		}
		return written
	}

	// write left side and determine if there is a right side
	var leftWidth int
	var hasRight bool
	if p.fieldsOnLeft {
		leftWidth = writeFields()
		hasRight = len(msg) > 0
	} else {
		leftWidth = writeMsg()
		hasRight = len(fields) > 0
	}

	// write spacing between left and right
	if leftWidth > 0 && hasRight {
		const minGap = 3
		const snapTo = 5

		gap := minGap
		if leftWidth+gap < p.columnOffset {
			gap = p.columnOffset - leftWidth
		}
		snapped := (((leftWidth+gap-1)/snapTo)+1)*snapTo
		for i := 0; i < snapped-leftWidth; i++ {
			buf.WriteByte(' ')
		}
	}

	// write right side
	if hasRight {
		if p.fieldsOnLeft {
			writeMsg()
		} else {
			writeFields()
		}
	}

	if useColor {
		buf.WriteString(terminal.Reset)
	}

	result := buf.String()

	// crop transient lines to prevent terminal wrapping
	if severity == Transient && p.maxLineWidth > 0 {
		if len([]rune(result)) > p.maxLineWidth {
			result = terminal.CropPreservingANSI(result, p.maxLineWidth)

		}
	}

	return result
}

// JSONPrinter formats log lines as single-line JSON objects.
type JSONPrinter struct {
	// FixedTime allows tests to override the timestamp for deterministic output.
	FixedTime time.Time
}

func (p *JSONPrinter) Configure(opts ...PrinterOption) Printer {
	// JSONPrinter does not currently support printer options
	return p
}

func (p *JSONPrinter) Render(severity Level, overrides []PrinterOption, msg string, fields []Field) string {
	var stamp time.Time
	if !p.FixedTime.IsZero() {
		stamp = p.FixedTime
	} else {
		stamp = time.Now()
	}

	var buf strings.Builder
	buf.Grow(70 + len(msg) + len(fields)*50)

	buf.WriteString(`{"ts":"`)
	buf.WriteString(stamp.Format(time.RFC3339))
	buf.WriteString(`","level":"`)
	buf.WriteString(severity.String())
	buf.WriteString(`","msg":"`)
	buf.WriteString(sanitizeForJSON(stripNewlines(msg)))
	buf.WriteString(`"`)

	for _, f := range fields {
		if f.IsJSONString {
			buf.WriteString(`,"`)
			buf.WriteString(f.Name)
			buf.WriteString(`":"`)
			if f.IsJSONSafe {
				buf.WriteString(f.Value)
			} else {
				buf.WriteString(sanitizeForJSON(f.Value))
			}
			buf.WriteString(`"`)
		} else {
			buf.WriteString(`,"`)
			buf.WriteString(f.Name)
			buf.WriteString(`":`)
			buf.WriteString(f.Value)
		}
	}

	buf.WriteString(`}`)
	return buf.String()
}

// sanitizeMessage escapes special characters in a log message for terminal display
func sanitizeMessage(s string) string {
	var buf strings.Builder
	buf.Grow(len(s) * 2)
	for _, r := range s {
		switch r {
		case '\t':
			buf.WriteString(`\t`)
		case '\n':
			buf.WriteString(`\n`)
		case '\r':
			buf.WriteString(`\r`)
		case '\\':
			buf.WriteString(`\\`)
		default:
			buf.WriteRune(r)
		}
	}
	return buf.String()
}

// sanitizeFieldValue escapes special characters in a field value for terminal display
func sanitizeFieldValue(s string) string {
	var buf strings.Builder
	buf.Grow(len(s) * 2)
	for _, r := range s {
		switch r {
		case '\t':
			buf.WriteString(`\t`)
		case '\n':
			buf.WriteString(`\n`)
		case '\r':
			buf.WriteString(`\r`)
		case '"':
			buf.WriteString(`\"`)
		case '\\':
			buf.WriteString(`\\`)
		default:
			buf.WriteRune(r)
		}
	}
	return buf.String()
}

// sanitizeForJSON escapes a string to be safe inside a JSON value
func sanitizeForJSON(s string) string {
	var buf strings.Builder
	buf.Grow(len(s) * 6)
	for _, r := range s {
		switch r {
		case 0x00:
			buf.WriteString(`\u0000`)
		case 0x01:
			buf.WriteString(`\u0001`)
		case 0x02:
			buf.WriteString(`\u0002`)
		case 0x03:
			buf.WriteString(`\u0003`)
		case 0x04:
			buf.WriteString(`\u0004`)
		case 0x05:
			buf.WriteString(`\u0005`)
		case 0x06:
			buf.WriteString(`\u0006`)
		case 0x07:
			buf.WriteString(`\u0007`)
		case 0x08:
			buf.WriteString(`\u0008`)
		case '\t':
			buf.WriteString(`\t`)
		case '\n':
			buf.WriteString(`\n`)
		case 0x0b:
			buf.WriteString(`\u000b`)
		case 0x0c:
			buf.WriteString(`\u000c`)
		case '\r':
			buf.WriteString(`\r`)
		case 0x0e:
			buf.WriteString(`\u000e`)
		case 0x0f:
			buf.WriteString(`\u000f`)
		case 0x10:
			buf.WriteString(`\u0010`)
		case 0x11:
			buf.WriteString(`\u0011`)
		case 0x12:
			buf.WriteString(`\u0012`)
		case 0x13:
			buf.WriteString(`\u0013`)
		case 0x14:
			buf.WriteString(`\u0014`)
		case 0x15:
			buf.WriteString(`\u0015`)
		case 0x16:
			buf.WriteString(`\u0016`)
		case 0x17:
			buf.WriteString(`\u0017`)
		case 0x18:
			buf.WriteString(`\u0018`)
		case 0x19:
			buf.WriteString(`\u0019`)
		case 0x1a:
			buf.WriteString(`\u001a`)
		case 0x1b:
			buf.WriteString(`\u001b`)
		case 0x1c:
			buf.WriteString(`\u001c`)
		case 0x1d:
			buf.WriteString(`\u001d`)
		case 0x1e:
			buf.WriteString(`\u001e`)
		case 0x1f:
			buf.WriteString(`\u001f`)
		case '"':
			buf.WriteString(`\"`)
		case '&':
			buf.WriteString(`\u0026`)
		case '<':
			buf.WriteString(`\u003c`)
		case '>':
			buf.WriteString(`\u003e`)
		case '\\':
			buf.WriteString(`\\`)
		case '\u2028':
			buf.WriteString(`\u2028`)
		case '\u2029':
			buf.WriteString(`\u2029`)
		default:
			buf.WriteRune(r)
		}
	}
	return buf.String()
}