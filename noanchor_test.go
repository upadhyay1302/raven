package raven

import (
	"sync"
	"testing"
)

// compile-time interface checks — if NoAnchorLogger doesn't implement
// these, this file will not compile, catching the bug immediately
var (
	_ Logger        = &NoAnchorLogger{}
	_ ChildLogger   = &NoAnchorLogger{}
	_ AnchorRemover = &NoAnchorLogger{}
)

func Test_NoAnchorLogger_Interfaces(t *testing.T) {
	t.Log("NoAnchorLogger satisfies Logger, ChildLogger, and AnchorRemover")
}

func Test_NoAnchorLogger_Parent(t *testing.T) {
	parent := NewNullLogger()
	l := newNoAnchorLogger(parent)

	if l.Parent() != parent {
		t.Error("expected Parent() to return the logger passed to newNoAnchorLogger")
	}
}

func Test_NoAnchorLogger_Threshold(t *testing.T) {
	l := newNoAnchorLogger(NewNullLogger())

	if l.Threshold() != Transient {
		t.Errorf("expected default threshold to be Transient, got %v", l.Threshold())
	}

	l.SetThreshold(Warning)
	if l.Threshold() != Warning {
		t.Errorf("expected threshold to be Warning after SetThreshold, got %v", l.Threshold())
	}
}

func Test_NoAnchorLogger_RemoveAnchor_IsNoOp(t *testing.T) {
	l := newNoAnchorLogger(NewNullLogger())
	// should not panic or do anything observable
	l.RemoveAnchor()
}

func Test_NoAnchorLogger_ChainsCorrectly(t *testing.T) {
	l := newNoAnchorLogger(NewNullLogger())

	// all log methods should return the same logger for chaining
	if l.Info("test") != l {
		t.Error("expected Info() to return the same logger")
	}
	if l.Warning("test") != l {
		t.Error("expected Warning() to return the same logger")
	}
}

func Test_NoAnchorLogger_ThreadSafety(t *testing.T) {
	l := newNoAnchorLogger(NewNullLogger())

	// hammer threshold reads and writes from multiple goroutines
	// run with: go test -race ./...
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			l.SetThreshold(Warning)
		}()
		go func() {
			defer wg.Done()
			_ = l.Threshold()
		}()
	}
	wg.Wait()
}