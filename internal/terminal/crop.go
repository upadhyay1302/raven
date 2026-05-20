package terminal

import (
    "strings"
    "github.com/mattn/go-runewidth"
)

// CropPreservingANSI crops the string to at most maxCols visible terminal
// columns, while keeping all ANSI escape sequences intact.
// Wide characters (e.g. CJK) count as 2 columns.
// Once the column limit is reached, all further visible runes are dropped.
func CropPreservingANSI(s string, maxCols int) string {
    var buf strings.Builder
    buf.Grow(len(s))

    visibleCols := 0
    inEscape := false

    for _, r := range s {
        if inEscape {
            buf.WriteRune(r)
            // escape sequences end at a letter that isn't a digit, semicolon, or bracket
            if r != '[' && !(r >= '0' && r <= '9') && r != ';' && r != '?' {
                inEscape = false
            }
            continue
        }

        if r == '\x1b' {
            buf.WriteRune(r)
            inEscape = true
            continue
        }

        w := runewidth.RuneWidth(r)
        if visibleCols+w > maxCols {
            // mark as cropped — skip all subsequent visible runes
            visibleCols = maxCols + 1
            continue
        }

        buf.WriteRune(r)
        visibleCols += w
    }

    return buf.String()
}