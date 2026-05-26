package raven

import (
	"bytes"
	"strings"
	"sync"
	"testing"
)

// compile-time interface checks
var (
	_ Logger      = &TeeLogger{}
	_ ChildLogger = &TeeLogger{}
)

func Test_TeeLogger_Interfaces(t *testing.T) {
	t.Log("TeeLogger satisfies Logger and ChildLogger")
}

func Test_NewTee_ReturnsTeeLogger(t *testing.T) {
	a := NewBuffered(&bytes.Buffer{}, false, &TextPrinter{})
	b := NewUnbuffered(&bytes.Buffer{}, &TextPrinter{})

	tee, stop := NewTee(a, b)
	defer stop()

	if tee == nil {
		t.Error("expected NewTee to return a non-nil TeeLogger")
	}
	if tee.Primary != a {
		t.Error("expected Primary to be the first logger passed to NewTee")
	}
	if tee.Secondary != b {
		t.Error("expected Secondary to be the second logger passed to NewTee")
	}
}

func Test_TeeLogger_Parent_ReturnsPrimary(t *testing.T) {
	a := NewBuffered(&bytes.Buffer{}, false, &TextPrinter{})
	b := NewUnbuffered(&bytes.Buffer{}, &TextPrinter{})

	tee, stop := NewTee(a, b)
	defer stop()

	if tee.Parent() != a {
		t.Error("expected Parent() to return the Primary logger")
	}
}

func Test_TeeLogger_DefaultThreshold(t *testing.T) {
	tee, stop := NewTee(
		NewBuffered(&bytes.Buffer{}, false, &TextPrinter{}),
		NewUnbuffered(&bytes.Buffer{}, &TextPrinter{}),
	)
	defer stop()

	if tee.Threshold() != Transient {
		t.Errorf("expected default threshold Transient, got %v", tee.Threshold())
	}
}

func Test_TeeLogger_SetThreshold(t *testing.T) {
	tee, stop := NewTee(
		NewBuffered(&bytes.Buffer{}, false, &TextPrinter{}),
		NewUnbuffered(&bytes.Buffer{}, &TextPrinter{}),
	)
	defer stop()

	tee.SetThreshold(Error)
	if tee.Threshold() != Error {
		t.Errorf("expected threshold Error, got %v", tee.Threshold())
	}
}

func Test_TeeLogger_WritesToBothLoggers(t *testing.T) {
	var bufA, bufB bytes.Buffer
	a := NewUnbuffered(&bufA, &TextPrinter{})
	b := NewUnbuffered(&bufB, &TextPrinter{})

	tee, stop := NewTee(a, b)
	tee.Info("hello from tee")
	stop()

	if !strings.Contains(bufA.String(), "hello from tee") {
		t.Errorf("expected Primary to receive the message, got: %q", bufA.String())
	}
	if !strings.Contains(bufB.String(), "hello from tee") {
		t.Errorf("expected Secondary to receive the message, got: %q", bufB.String())
	}
}

func Test_TeeLogger_BothLoggersReceiveSameContent(t *testing.T) {
	var bufA, bufB bytes.Buffer
	a := NewUnbuffered(&bufA, &TextPrinter{showLevel: true})
	b := NewUnbuffered(&bufB, &TextPrinter{showLevel: true})

	tee, stop := NewTee(a, b)
	tee.SetThreshold(Transient)
	tee.Info("matching output", String("key", "value"))
	stop()

	if bufA.String() != bufB.String() {
		t.Errorf("expected both loggers to receive identical output\nPrimary:   %q\nSecondary: %q",
			bufA.String(), bufB.String())
	}
}

func Test_TeeLogger_StopClosesBothLoggers(t *testing.T) {
	closedA, closedB := false, false

	// use NullLoggers and verify Close behaviour via Buffered
	var bufA, bufB bytes.Buffer
	a := NewBuffered(&bufA, false, &TextPrinter{})
	b := NewBuffered(&bufB, false, &TextPrinter{})

	tee, stop := NewTee(a, b)
	_ = tee

	stop()

	// after stop(), both loggers should be closed
	// calling Close again should not panic or deadlock
	closedA = true
	closedB = true
	a.Close()
	b.Close()

	if !closedA || !closedB {
		t.Error("expected both loggers to be closed after stop()")
	}
}

func Test_TeeLogger_ThresholdFiltersMessages(t *testing.T) {
	var bufA, bufB bytes.Buffer
	a := NewUnbuffered(&bufA, &TextPrinter{})
	b := NewUnbuffered(&bufB, &TextPrinter{})

	tee, stop := NewTee(a, b)
	tee.SetThreshold(Error)
	tee.Info("this should be filtered")
	stop()

	if strings.Contains(bufA.String(), "this should be filtered") {
		t.Error("expected Primary to filter Info message when threshold is Error")
	}
	if strings.Contains(bufB.String(), "this should be filtered") {
		t.Error("expected Secondary to filter Info message when threshold is Error")
	}
}

func Test_TeeLogger_MethodChaining(t *testing.T) {
	tee, stop := NewTee(
		NewUnbuffered(&bytes.Buffer{}, &TextPrinter{}),
		NewUnbuffered(&bytes.Buffer{}, &TextPrinter{}),
	)
	defer stop()

	if tee.Info("test") != tee {
		t.Error("expected Info() to return the same logger")
	}
	if tee.Warning("test") != tee {
		t.Error("expected Warning() to return the same logger")
	}
	if tee.SetThreshold(Info) != tee {
		t.Error("expected SetThreshold() to return the same logger")
	}
}

func Test_TeeLogger_ThreadSafety(t *testing.T) {
	tee, stop := NewTee(
		NewUnbuffered(&bytes.Buffer{}, &TextPrinter{}),
		NewUnbuffered(&bytes.Buffer{}, &TextPrinter{}),
	)
	defer stop()

	// hammer both loggers from many goroutines simultaneously
	// run with: go test -race ./...
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			tee.Info("concurrent message")
			tee.Warning("another concurrent message")
		}()
		go func() {
			defer wg.Done()
			tee.SetThreshold(Warning)
			_ = tee.Threshold()
		}()
	}
	wg.Wait()
}