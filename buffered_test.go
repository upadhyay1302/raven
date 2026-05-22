package raven

import (
	"bytes"
	"strings"
	"sync"
	"testing"
)

// compile-time interface checks
var (
	_ RootLogger = &Buffered{}
	_ AnchorAdder = &Buffered{}
)

func Test_Buffered_Interfaces(t *testing.T) {
	t.Log("Buffered satisfies RootLogger and AnchorAdder")
}

func Test_NewBuffered_NilWriter(t *testing.T) {
	l := NewBuffered(nil, false, &TextPrinter{})
	if l != nil {
		t.Error("expected NewBuffered to return nil when writer is nil")
	}
}

func Test_NewBuffered_ValidWriter(t *testing.T) {
	l := NewBuffered(&bytes.Buffer{}, false, &TextPrinter{})
	if l == nil {
		t.Error("expected NewBuffered to return a valid logger")
	}
	defer l.Close()
}

func Test_Buffered_DefaultThreshold(t *testing.T) {
	l := NewBuffered(&bytes.Buffer{}, false, &TextPrinter{})
	defer l.Close()

	if l.Threshold() != Info {
		t.Errorf("expected default threshold to be Info, got %v", l.Threshold())
	}
}

func Test_Buffered_SetThreshold(t *testing.T) {
	l := NewBuffered(&bytes.Buffer{}, false, &TextPrinter{})
	defer l.Close()

	l.SetThreshold(Warning)
	if l.Threshold() != Warning {
		t.Errorf("expected threshold Warning, got %v", l.Threshold())
	}
}

func Test_Buffered_Close_MultipleTimes(t *testing.T) {
	l := NewBuffered(&bytes.Buffer{}, false, &TextPrinter{})

	// calling Close multiple times should never panic or deadlock
	l.Close()
	l.Close()
	l.Close()
}

func Test_Buffered_WritesOutput(t *testing.T) {
	var buf bytes.Buffer
	l := NewBuffered(&buf, false, &TextPrinter{showLevel: true})
	defer l.Close()

	l.Info("hello from raven")
	l.Close()

	output := buf.String()
	if !strings.Contains(output, "hello from raven") {
		t.Errorf("expected output to contain message, got: %q", output)
	}
}

func Test_Buffered_ThresholdFiltersMessages(t *testing.T) {
	var buf bytes.Buffer
	l := NewBuffered(&buf, false, &TextPrinter{})
	defer l.Close()

	l.SetThreshold(Error)
	l.Info("this should be filtered")
	l.Close()

	output := buf.String()
	if strings.Contains(output, "this should be filtered") {
		t.Error("expected Info message to be filtered when threshold is Error")
	}
}

func Test_Buffered_MethodChaining(t *testing.T) {
	l := NewBuffered(&bytes.Buffer{}, false, &TextPrinter{})
	defer l.Close()

	// all log methods should return the same logger for chaining
	if l.Info("test") != l {
		t.Error("expected Info() to return the same logger")
	}
	if l.Warning("test") != l {
		t.Error("expected Warning() to return the same logger")
	}
	if l.SetThreshold(Info) != l {
		t.Error("expected SetThreshold() to return the same logger")
	}
}

func Test_Buffered_ThreadSafety(t *testing.T) {
	l := NewBuffered(&bytes.Buffer{}, false, &TextPrinter{})
	defer l.Close()

	// hammer the logger from many goroutines simultaneously
	// run with: go test -race ./...
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			l.Info("concurrent message")
			l.Warning("another concurrent message")
		}()
	}
	wg.Wait()
}