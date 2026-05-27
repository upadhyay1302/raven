package raven

import (
	"bytes"
	"strings"
	"sync"
	"testing"
)

// compile-time interface checks
var (
	_ RootLogger = &Unbuffered{}
	_ Logger     = &Unbuffered{}
)

func Test_Unbuffered_Interfaces(t *testing.T) {
	t.Log("Unbuffered satisfies RootLogger and Logger")
}

func Test_Unbuffered_DefaultThreshold(t *testing.T) {
	l := NewUnbuffered(&bytes.Buffer{}, &TextPrinter{})

	if l.Threshold() != Info {
		t.Errorf("expected default threshold Info, got %v", l.Threshold())
	}
}

func Test_Unbuffered_SetThreshold(t *testing.T) {
	l := NewUnbuffered(&bytes.Buffer{}, &TextPrinter{})

	l.SetThreshold(Warning)
	if l.Threshold() != Warning {
		t.Errorf("expected threshold Warning, got %v", l.Threshold())
	}
}

func Test_Unbuffered_Close_IsNoOp(t *testing.T) {
	l := NewUnbuffered(&bytes.Buffer{}, &TextPrinter{})

	// should never panic or deadlock
	l.Close()
	l.Close()
	l.Close()
}

func Test_Unbuffered_WritesOutput(t *testing.T) {
	var buf bytes.Buffer
	l := NewUnbuffered(&buf, &TextPrinter{showLevel: true})

	l.Info("hello from raven")
	l.Close()

	if !strings.Contains(buf.String(), "hello from raven") {
		t.Errorf("expected output to contain message, got: %q", buf.String())
	}
}

func Test_Unbuffered_ThresholdFiltersMessages(t *testing.T) {
	var buf bytes.Buffer
	l := NewUnbuffered(&buf, &TextPrinter{})

	l.SetThreshold(Error)
	l.Info("this should be filtered")
	l.Warning("this should also be filtered")
	l.Close()

	if strings.Contains(buf.String(), "filtered") {
		t.Errorf("expected messages below threshold to be filtered, got: %q", buf.String())
	}
}

func Test_Unbuffered_AllLevelsWrite(t *testing.T) {
	cases := []struct {
		name    string
		logFunc func(l *Unbuffered)
		msg     string
	}{
		{"transient", func(l *Unbuffered) { l.Transient("transient msg") }, "transient msg"},
		{"verbose", func(l *Unbuffered) { l.Verbose("verbose msg") }, "verbose msg"},
		{"info", func(l *Unbuffered) { l.Info("info msg") }, "info msg"},
		{"warning", func(l *Unbuffered) { l.Warning("warning msg") }, "warning msg"},
		{"error", func(l *Unbuffered) { l.Error("error msg") }, "error msg"},
		{"emit", func(l *Unbuffered) { l.Emit(Info, "emit msg") }, "emit msg"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			l := NewUnbuffered(&buf, &TextPrinter{})
			l.SetThreshold(Transient)
			tc.logFunc(l)
			l.Close()

			if !strings.Contains(buf.String(), tc.msg) {
				t.Errorf("expected output to contain %q, got: %q", tc.msg, buf.String())
			}
		})
	}
}

func Test_Unbuffered_EachLineEndsWithNewline(t *testing.T) {
	var buf bytes.Buffer
	l := NewUnbuffered(&buf, &TextPrinter{})

	l.Info("first line")
	l.Info("second line")
	l.Close()

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d: %q", len(lines), buf.String())
	}
}

func Test_Unbuffered_MethodChaining(t *testing.T) {
	l := NewUnbuffered(&bytes.Buffer{}, &TextPrinter{})

	if l.Info("test") != l {
		t.Error("expected Info() to return the same logger")
	}
	if l.Warning("test") != l {
		t.Error("expected Warning() to return the same logger")
	}
	if l.Error("test") != l {
		t.Error("expected Error() to return the same logger")
	}
	if l.SetThreshold(Info) != l {
		t.Error("expected SetThreshold() to return the same logger")
	}
}

func Test_Unbuffered_WithFields(t *testing.T) {
	var buf bytes.Buffer
	l := NewUnbuffered(&buf, &TextPrinter{})

	l.Info("with fields",
		String("env", "production"),
		Int("port", 8080),
	)
	l.Close()

	output := buf.String()
	if !strings.Contains(output, "env=production") {
		t.Errorf("expected field env=production in output, got: %q", output)
	}
	if !strings.Contains(output, "port=8080") {
		t.Errorf("expected field port=8080 in output, got: %q", output)
	}
}

func Test_Unbuffered_ThreadSafety(t *testing.T) {
	l := NewUnbuffered(&bytes.Buffer{}, &TextPrinter{})

	// hammer the logger from many goroutines simultaneously
	// the internal mutex should prevent interleaved/garbled output
	// run with: go test -race ./...
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			l.Info("concurrent message")
			l.Warning("another concurrent message")
		}()
		go func() {
			defer wg.Done()
			l.SetThreshold(Warning)
			_ = l.Threshold()
		}()
	}
	wg.Wait()
}

func Test_Unbuffered_JSONPrinter(t *testing.T) {
	var buf bytes.Buffer
	l := NewUnbuffered(&buf, &JSONPrinter{})

	l.Info("json output", String("key", "value"))
	l.Close()

	output := strings.TrimSpace(buf.String())
	if !strings.HasPrefix(output, "{") || !strings.HasSuffix(output, "}") {
		t.Errorf("expected valid JSON object, got: %q", output)
	}
	if !strings.Contains(output, `"key":"value"`) {
		t.Errorf("expected JSON to contain key field, got: %q", output)
	}
}