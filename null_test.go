package raven

import (
	"sync"
	"testing"
)

var (
	_ RootLogger = &NullLogger{}
	_ Logger     = &NullLogger{}
)

func Test_NullLogger_Interfaces(t *testing.T) {
	t.Log("NullLogger satisfies RootLogger and Logger")
}

func Test_NullLogger_DefaultThreshold(t *testing.T) {
	l := NewNullLogger()

	if l.Threshold() != Transient {
		t.Errorf("expected default threshold Transient, got %v", l.Threshold())
	}
}

func Test_NullLogger_SetThreshold(t *testing.T) {
	l := NewNullLogger()

	l.SetThreshold(Error)
	if l.Threshold() != Error {
		t.Errorf("expected threshold Error after SetThreshold, got %v", l.Threshold())
	}
}

func Test_NullLogger_Close_IsNoOp(t *testing.T) {
	l := NewNullLogger()

	// should never panic or deadlock
	l.Close()
	l.Close()
	l.Close()
}

func Test_NullLogger_MethodChaining(t *testing.T) {
	l := NewNullLogger()

	if l.Transient("test") != l {
		t.Error("expected Transient() to return the same logger")
	}
	if l.Verbose("test") != l {
		t.Error("expected Verbose() to return the same logger")
	}
	if l.Info("test") != l {
		t.Error("expected Info() to return the same logger")
	}
	if l.Warning("test") != l {
		t.Error("expected Warning() to return the same logger")
	}
	if l.Error("test") != l {
		t.Error("expected Error() to return the same logger")
	}
	if l.Emit(Info, "test") != l {
		t.Error("expected Emit() to return the same logger")
	}
	if l.SetThreshold(Info) != l {
		t.Error("expected SetThreshold() to return the same logger")
	}
}

func Test_NullLogger_DiscardsAllOutput(t *testing.T) {
	l := NewNullLogger()

	// none of these should panic, block, or produce output
	l.Transient("transient message")
	l.Verbose("verbose message")
	l.Info("info message")
	l.Warning("warning message")
	l.Error("error message")
	l.Emit(Info, "emit message")
}

func Test_NullLogger_DiscardsFields(t *testing.T) {
	l := NewNullLogger()

	l.Info("with fields",
		String("key", "value"),
		Int("count", 42),
		Bool("active", true),
	)
}

func Test_NullLogger_ForwardIsNoOp(t *testing.T) {
	l := NewNullLogger()

	// forward should silently discard without panic
	l.forward(Info, "test", nil, nil, LogContext{})
	l.forward(Error, "test", []Fielder{String("k", "v")}, nil, LogContext{})
}

func Test_NullLogger_ThreadSafety(t *testing.T) {
	l := NewNullLogger()

	// hammer threshold reads/writes and log calls from many goroutines
	// run with: go test -race ./...
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(3)
		go func() {
			defer wg.Done()
			l.Info("concurrent log")
			l.Warning("another concurrent log")
		}()
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

func Test_NullLogger_UsableAsDefault(t *testing.T) {
	// common pattern: accept a Logger, fall back to NullLogger if nil
	newService := func(log Logger) Logger {
		if log == nil {
			return NewNullLogger()
		}
		return log
	}

	svc := newService(nil)
	if svc == nil {
		t.Error("expected service to use NullLogger as default")
	}

	// should work without panicking
	svc.Info("service started")
}