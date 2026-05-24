package raven

import (
	"bytes"
	"strings"
	"sync"
	"testing"
)

// compile-time interface checks
var (
	_ Logger      = &ContextLogger{}
	_ ChildLogger = &ContextLogger{}
)

func Test_ContextLogger_Interfaces(t *testing.T) {
	t.Log("ContextLogger satisfies Logger and ChildLogger")
}

func Test_ContextLogger_Parent(t *testing.T) {
	parent := NewNullLogger()
	l := newCustomizerLogger(parent, nil, nil)

	if l.Parent() != parent {
		t.Error("expected Parent() to return the logger passed to newCustomizerLogger")
	}
}

func Test_ContextLogger_DefaultThreshold(t *testing.T) {
	l := newCustomizerLogger(NewNullLogger(), nil, nil)

	if l.Threshold() != Transient {
		t.Errorf("expected default threshold Transient, got %v", l.Threshold())
	}
}

func Test_ContextLogger_SetThreshold(t *testing.T) {
	l := newCustomizerLogger(NewNullLogger(), nil, nil)

	l.SetThreshold(Error)
	if l.Threshold() != Error {
		t.Errorf("expected threshold Error, got %v", l.Threshold())
	}
}

func Test_ContextLogger_MethodChaining(t *testing.T) {
	l := newCustomizerLogger(NewNullLogger(), nil, nil)

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

func Test_ContextLogger_AddsFields(t *testing.T) {
	var buf bytes.Buffer
	parent := NewUnbuffered(&buf, &TextPrinter{showLevel: true})

	l := newCustomizerLogger(parent, nil, []Fielder{
		String("service", "raven"),
		Int("version", 2),
	})

	l.Info("hello")
	parent.Close()

	output := buf.String()
	if !strings.Contains(output, "service=raven") {
		t.Errorf("expected output to contain static field service=raven, got: %q", output)
	}
	if !strings.Contains(output, "version=2") {
		t.Errorf("expected output to contain static field version=2, got: %q", output)
	}
}

func Test_ContextLogger_DoesNotModifyParent(t *testing.T) {
	var buf bytes.Buffer
	parent := NewUnbuffered(&buf, &TextPrinter{})

	l := newCustomizerLogger(parent, nil, []Fielder{
		String("only_in_child", "true"),
	})

	// log via child, then via parent directly
	l.Info("child log")
	parent.Info("parent log")
	parent.Close()

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 2 {
		t.Fatalf("expected 2 log lines, got %d: %q", len(lines), output)
	}
	if !strings.Contains(lines[0], "only_in_child") {
		t.Error("expected child log line to contain the static field")
	}
	if strings.Contains(lines[1], "only_in_child") {
		t.Error("expected parent log line to NOT contain the child's static field")
	}
}

func Test_ContextLogger_FieldsPreserveOrder(t *testing.T) {
	var buf bytes.Buffer
	parent := NewUnbuffered(&buf, &TextPrinter{})

	l := newCustomizerLogger(parent, nil, []Fielder{
		String("first", "a"),
	})

	l.Info("ordered", String("second", "b"))
	parent.Close()

	output := buf.String()
	firstIdx := strings.Index(output, "first=a")
	secondIdx := strings.Index(output, "second=b")

	if firstIdx == -1 || secondIdx == -1 {
		t.Fatalf("expected both fields in output, got: %q", output)
	}
	if firstIdx > secondIdx {
		t.Error("expected static fields to appear before per-call fields")
	}
}

func Test_ContextLogger_OptsDoNotMutateStaticOpts(t *testing.T) {
	parent := NewNullLogger()
	staticOpts := []PrinterOption{OptShowTime(true)}

	l := newCustomizerLogger(parent, staticOpts, nil)

	// call forward multiple times with per-call opts
	l.forward(Info, "msg", nil, []PrinterOption{OptShowLevel(true)}, LogContext{})
	l.forward(Info, "msg", nil, []PrinterOption{OptShowLevel(false)}, LogContext{})

	// static opts should still only have one entry
	if len(staticOpts) != 1 {
		t.Errorf("expected staticOpts to remain unchanged, got len=%d", len(staticOpts))
	}
}

func Test_ContextLogger_NilFieldsAndOpts(t *testing.T) {
	var buf bytes.Buffer
	parent := NewUnbuffered(&buf, &TextPrinter{})

	// should not panic with nil opts and fields
	l := newCustomizerLogger(parent, nil, nil)
	l.Info("no extras")
	parent.Close()

	if buf.Len() == 0 {
		t.Error("expected output even with nil opts and fields")
	}
}

func Test_ContextLogger_ThreadSafety(t *testing.T) {
	l := newCustomizerLogger(NewNullLogger(), nil, []Fielder{
		String("key", "value"),
	})

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			l.Info("concurrent log")
		}()
		go func() {
			defer wg.Done()
			l.SetThreshold(Warning)
			_ = l.Threshold()
		}()
	}
	wg.Wait()
}