
// Reference: https://en.wikipedia.org/wiki/ANSI_escape_code
package terminal

import "fmt"

const (
    // esc is the escape character that begins all ANSI sequences
    esc = "\x1b"
    // csi is the Control Sequence Introducer
    csi = esc + "["
)

// Reset returns terminal output to default colors and styles
const Reset = csi + "0m"

// Foreground colors
const (
    FgBlack       = csi + "30m"
    FgDarkRed     = csi + "31m"
    FgDarkGreen   = csi + "32m"
    FgDarkYellow  = csi + "33m"
    FgDarkBlue    = csi + "34m"
    FgDarkMagenta = csi + "35m"
    FgDarkCyan    = csi + "36m"
    FgLightGray   = csi + "37m"
    FgDarkGray    = csi + "90m"
    FgRed         = csi + "91m"
    FgGreen       = csi + "92m"
    FgYellow      = csi + "93m"
    FgBlue        = csi + "94m"
    FgMagenta     = csi + "95m"
    FgCyan        = csi + "96m"
    FgWhite       = csi + "97m"
)

// Cursor and line control
const (
    // CursorToLineStart moves cursor to column 1 of the current line
    CursorToLineStart = csi + "1G"
    // EraseLine clears the entire current line without moving the cursor
    EraseLine = csi + "2K"
)

// CursorUp returns a sequence that moves the cursor up n lines.
func CursorUp(n int) string {
    if n <= 0 {
        return ""
    }
    return fmt.Sprintf(csi+"%dA", n)
}