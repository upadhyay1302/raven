package raven

import (
	"sync"
	"testing"
)

// compile-time interface checks
var (
	_ Logger        = &AnchoredLogger{}
	_ ChildLogger   = &AnchoredLogger{}
	_ AnchorRemover = &AnchoredLogger{}
)

func Test_AnchoredLogger_Interfaces(t *testing.T) {
	t.Log("AnchoredLogger satisfies Logger, ChildLogger, and AnchorRemover")
}

func Test_AnchoredLogger_Parent(t *testing.T) {
	parent := NewNullLogger()
	l := newAnchor(parent, 1, func() {})

	if l.Parent() != parent {
		t.Error("expected Parent() to return the logger passed to newAnchor")
	}
}

func Test_AnchoredLogger_DefaultThreshold(t *testing.T) {
	l := newAnchor(NewNullLogger(), 1, func() {})

	if l.Threshold() != Transient {
		t.Errorf("expected default threshold Transient, got %v", l.Threshold())
	}
}

func Test_AnchoredLogger_SetThreshold(t *testing.T) {
	l := newAnchor(NewNullLogger(), 1, func() {})

	l.SetThreshold(Warning)
	if l.Threshold() != Warning {
		t.Errorf("expected threshold Warning, got %v", l.Threshold())
	}
}

func Test_AnchoredLogger_RemoveAnchor_CallsOnRemove(t *testing.T) {
	called := 0
	l := newAnchor(NewNullLogger(), 1, func() { called++ })

	l.RemoveAnchor()

	if called != 1 {
		t.Errorf("expected onRemove to be called once, got %d", called)
	}
}

func Test_AnchoredLogger_RemoveAnchor_CalledOnlyOnce(t *testing.T) {
	called := 0
	l := newAnchor(NewNullLogger(), 1, func() { called++ })

	// call multiple times — sync.Once should guarantee exactly one execution
	l.RemoveAnchor()
	l.RemoveAnchor()
	l.RemoveAnchor()

	if called != 1 {
		t.Errorf("expected onRemove called exactly once, got %d", called)
	}
}

func Test_AnchoredLogger_RemoveAnchor_ZerosAnchorID(t *testing.T) {
	l := newAnchor(NewNullLogger(), 42, func() {})

	if l.anchorID.Load() != 42 {
		t.Error("expected anchorID to be 42 before RemoveAnchor")
	}

	l.RemoveAnchor()

	if l.anchorID.Load() != 0 {
		t.Error("expected anchorID to be 0 after RemoveAnchor")
	}
}

func Test_AnchoredLogger_MethodChaining(t *testing.T) {
	l := newAnchor(NewNullLogger(), 1, func() {})

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

func Test_AnchoredLogger_ThreadSafety(t *testing.T) {
	l := newAnchor(NewNullLogger(), 1, func() {})

	// hammer RemoveAnchor, Threshold reads/writes, and log calls concurrently
	// run with: go test -race ./...
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(3)
		go func() {
			defer wg.Done()
			l.RemoveAnchor()
		}()
		go func() {
			defer wg.Done()
			l.SetThreshold(Warning)
			_ = l.Threshold()
		}()
		go func() {
			defer wg.Done()
			l.Info("concurrent log")
		}()
	}
	wg.Wait()
}

func Test_AnchoredLogger_RemoveAnchor_ConcurrentlySafe(t *testing.T) {
	// specifically test that concurrent RemoveAnchor calls
	// never trigger onRemove more than once
	called := 0
	var mu sync.Mutex
	l := newAnchor(NewNullLogger(), 1, func() {
		mu.Lock()
		called++
		mu.Unlock()
	})

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			l.RemoveAnchor()
		}()
	}
	wg.Wait()

	if called != 1 {
		t.Errorf("expected onRemove called exactly once across 50 goroutines, got %d", called)
	}
}