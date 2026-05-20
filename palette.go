package raven

import "github.com/upadhyay1302/raven/internal/terminal"

// Color represents a terminal foreground color.
type Color byte

const (
	Black Color = iota
	DarkRed
	DarkGreen
	DarkYellow
	DarkBlue
	DarkMagenta
	DarkCyan
	LightGray
	DarkGray
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
	colorCount // sentinel — always keep this last
)

// ansiColorMap maps each Color to its ANSI escape sequence.
var ansiColorMap = [colorCount]string{
	Black:       terminal.FgBlack,
	DarkRed:     terminal.FgDarkRed,
	DarkGreen:   terminal.FgDarkGreen,
	DarkYellow:  terminal.FgDarkYellow,
	DarkBlue:    terminal.FgDarkBlue,
	DarkMagenta: terminal.FgDarkMagenta,
	DarkCyan:    terminal.FgDarkCyan,
	LightGray:   terminal.FgLightGray,
	DarkGray:    terminal.FgDarkGray,
	Red:         terminal.FgRed,
	Green:       terminal.FgGreen,
	Yellow:      terminal.FgYellow,
	Blue:        terminal.FgBlue,
	Magenta:     terminal.FgMagenta,
	Cyan:        terminal.FgCyan,
	White:       terminal.FgWhite,
}

// colorNames is used by Color.String() for debugging.
var colorNames = [colorCount]string{
	Black:       "Black",
	DarkRed:     "DarkRed",
	DarkGreen:   "DarkGreen",
	DarkYellow:  "DarkYellow",
	DarkBlue:    "DarkBlue",
	DarkMagenta: "DarkMagenta",
	DarkCyan:    "DarkCyan",
	LightGray:   "LightGray",
	DarkGray:    "DarkGray",
	Red:         "Red",
	Green:       "Green",
	Yellow:      "Yellow",
	Blue:        "Blue",
	Magenta:     "Magenta",
	Cyan:        "Cyan",
	White:       "White",
}

// String implements fmt.Stringer for readable debugging output.
func (c Color) String() string {
	if c >= colorCount {
		return "UnknownColor"
	}
	return colorNames[c]
}

// toANSI converts a Color to its ANSI escape sequence.
func (c Color) toANSI() string {
	if c >= colorCount {
		return ""
	}
	return ansiColorMap[c]
}

// ColorPair holds the primary and secondary colors for a single log level.
type ColorPair struct {
	Primary   Color
	Secondary Color
}

// Palette defines the colors used for each log level.
type Palette [levelMax]ColorPair

// ansiPalette is the internal representation using pre-computed ANSI strings.
type ansiPalette [levelMax][2]string

// toANSI converts a Palette into its internal ansiPalette representation.
func (p Palette) toANSI() ansiPalette {
	var out ansiPalette
	for lvl := levelMin; lvl < levelMax; lvl++ {
		out[lvl][0] = p[lvl].Primary.toANSI()
		out[lvl][1] = p[lvl].Secondary.toANSI()
	}
	return out
}

// WithLevel returns a copy of the Palette with one level's colors replaced.
func (p Palette) WithLevel(lvl Level, primary, secondary Color) Palette {
	p[lvl] = ColorPair{Primary: primary, Secondary: secondary}
	return p
}

// Built-in palettes

var DefaultPalette = Palette{
	Transient: {Primary: DarkGreen, Secondary: DarkGray},
	Verbose:   {Primary: Cyan, Secondary: DarkCyan},
	Info:      {Primary: White, Secondary: LightGray},
	Warning:   {Primary: Yellow, Secondary: DarkYellow},
	Error:     {Primary: Red, Secondary: DarkRed},
}

var MutedPalette = Palette{
	Transient: {Primary: DarkGray, Secondary: DarkGray},
	Verbose:   {Primary: DarkGray, Secondary: DarkGray},
	Info:      {Primary: DarkGray, Secondary: DarkGray},
	Warning:   {Primary: DarkGray, Secondary: DarkGray},
	Error:     {Primary: DarkGray, Secondary: DarkGray},
}

var BoldPalette = Palette{
	Transient: {Primary: Green, Secondary: DarkGreen},
	Verbose:   {Primary: Cyan, Secondary: DarkCyan},
	Info:      {Primary: White, Secondary: LightGray},
	Warning:   {Primary: Yellow, Secondary: DarkYellow},
	Error:     {Primary: Red, Secondary: Magenta},
}