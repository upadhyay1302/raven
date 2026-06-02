package main

import (
	"fmt"
	"os"
	"time"

	"github.com/upadhyay1302/raven"
	"github.com/upadhyay1302/raven/internal/terminal"
)

func main() {
	if !raven.HasTerminal(os.Stdout) {
		fmt.Println("no terminal detected — size unavailable when output is piped")
		return
	}

	start := time.Now()
	cols, rows, err := terminal.GetSize()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error detecting terminal size: %v\n", err)
		os.Exit(1)
	}
	elapsed := time.Since(start)

	fmt.Printf("terminal size: %d columns x %d rows\n", cols, rows)
	fmt.Printf("detection took: %v\n", elapsed)
}