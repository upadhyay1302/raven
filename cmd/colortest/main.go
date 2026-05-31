package main

import (
	"fmt"
	"os"

	"github.com/upadhyay1302/raven"
	"github.com/upadhyay1302/raven/internal/terminal"
)

func main() {
	if !raven.HasTerminal(os.Stdout) {
		fmt.Println("no terminal detected — colors disabled")
		fmt.Println("run this directly in a terminal to see colors")
		return
	}

	// --- Raw color swatches ---
	fmt.Println("=== Available Colors ===")
	fmt.Println(terminal.FgBlack + "FgBlack" + terminal.Reset)
	fmt.Println(terminal.FgDarkGray + "FgDarkGray" + terminal.Reset)
	fmt.Println(terminal.FgLightGray + "FgLightGray" + terminal.Reset)
	fmt.Println(terminal.FgWhite + "FgWhite" + terminal.Reset)
	fmt.Println(terminal.FgDarkRed + "FgDarkRed" + terminal.Reset)
	fmt.Println(terminal.FgRed + "FgRed" + terminal.Reset)
	fmt.Println(terminal.FgDarkGreen + "FgDarkGreen" + terminal.Reset)
	fmt.Println(terminal.FgGreen + "FgGreen" + terminal.Reset)
	fmt.Println(terminal.FgDarkYellow + "FgDarkYellow" + terminal.Reset)
	fmt.Println(terminal.FgYellow + "FgYellow" + terminal.Reset)
	fmt.Println(terminal.FgDarkBlue + "FgDarkBlue" + terminal.Reset)
	fmt.Println(terminal.FgBlue + "FgBlue" + terminal.Reset)
	fmt.Println(terminal.FgDarkMagenta + "FgDarkMagenta" + terminal.Reset)
	fmt.Println(terminal.FgMagenta + "FgMagenta" + terminal.Reset)
	fmt.Println(terminal.FgDarkCyan + "FgDarkCyan" + terminal.Reset)
	fmt.Println(terminal.FgCyan + "FgCyan" + terminal.Reset)

	// --- Live logger demos using each built-in palette ---
	fmt.Println("\n=== Default Palette ===")
	showPalette(raven.DefaultPalette)

	fmt.Println("\n=== Muted Palette ===")
	showPalette(raven.MutedPalette)

	fmt.Println("\n=== Bold Palette ===")
	showPalette(raven.BoldPalette)

	// --- Custom palette using WithLevel builder ---
	fmt.Println("\n=== Custom Palette (error=magenta, warning=cyan) ===")
	custom := raven.DefaultPalette.
		WithLevel(raven.Error, raven.Magenta, raven.DarkMagenta).
		WithLevel(raven.Warning, raven.Cyan, raven.DarkCyan)
	showPalette(custom)
}

// showPalette creates a Raven logger with the given palette and logs
// one line at each severity level to show the colors in context.
func showPalette(p raven.Palette) {
	log := raven.New(raven.Auto, raven.OptPalette(p))
	defer log.Close()
	log.SetThreshold(raven.Transient)

	log.Transient("transient message", raven.String("field", "value"))
	log.Verbose("verbose message", raven.String("field", "value"))
	log.Info("info message", raven.String("field", "value"))
	log.Warning("warning message", raven.String("field", "value"))
	log.Error("error message", raven.String("field", "value"))
}