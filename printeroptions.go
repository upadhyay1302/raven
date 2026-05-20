package raven

// PrinterOption is the interface for configuring a Printer
// The unexported marker method ensures only options defined within
// this package can be used
type PrinterOption interface {
	isRavenOption() // unexported marker, acts as a compile-time guard
	String() string
}

// Color Palette

// OptPalette sets the color palette used for terminal output
func OptPalette(p Palette) PrinterOption {
	return optPalette{ANSIColors: p.toANSI()}
}

type optPalette struct {
    ANSIColors ansiPalette
}

func (optPalette) isRavenOption() {}
func (optPalette) String() string { return "OptPalette" }

// Timestamp 

// OptShowTime controls whether a timestamp is shown on each log line
func OptShowTime(visible bool) optShowTime {
	return optShowTime{Visible: visible}
}

type optShowTime struct {
	Visible bool
}

func (optShowTime) isRavenOption() {}
func (optShowTime) String() string { return "OptShowTime" }

// Level Label 

// OptShowLevel controls whether the severity level label is shown on each log line
func OptShowLevel(visible bool) optShowLevel {
	return optShowLevel{Visible: visible}
}

type optShowLevel struct {
	Visible bool
}

func (optShowLevel) isRavenOption() {}
func (optShowLevel) String() string { return "OptShowLevel" }

// Column Offset

// OptColumnOffset sets how far from the left fields begin rendering
// Fields are snapped to the nearest multiple of 5, with a minimum gap of 3
func OptColumnOffset(cols int) optColumnOffset {
	return optColumnOffset{Indent: cols}
}

type optColumnOffset struct {
	Indent int
}

func (optColumnOffset) isRavenOption() {}
func (optColumnOffset) String() string { return "OptColumnOffset" }

// Layout Order 

// OptMsgThenFields renders the message on the left and fields on the right
// This is the default layout
var OptMsgThenFields optMsgThenFields

type optMsgThenFields struct{}

func (optMsgThenFields) isRavenOption() {}
func (optMsgThenFields) String() string { return "OptMsgThenFields" }

// OptFieldsThenMsg renders fields on the left and the message on the right
var OptFieldsThenMsg optFieldsThenMsg

type optFieldsThenMsg struct{}

func (optFieldsThenMsg) isRavenOption() {}
func (optFieldsThenMsg) String() string { return "OptFieldsThenMsg" }

// Max Line Width 

// OptMaxLineWidth crops transient/anchor lines at the given column count
// to prevent them wrapping and breaking the terminal layout
// Pass 0 to disable cropping
func OptMaxLineWidth(cols int) optMaxLineWidth {
	return optMaxLineWidth{Cols: cols}
}

type optMaxLineWidth struct {
	Cols int
}

func (optMaxLineWidth) isRavenOption() {}
func (optMaxLineWidth) String() string { return "OptMaxLineWidth" }