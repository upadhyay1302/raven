package terminal

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

// GetSize returns the width and height of the connected terminal in columns and rows.
// Returns an error if stdout is not a terminal (e.g. output is piped to a file).
func GetSize() (cols, rows int, err error) {
	cols, rows, err = term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 0, 0, fmt.Errorf("raven: could not get terminal size: %w", err)
	}
	return cols, rows, nil
}