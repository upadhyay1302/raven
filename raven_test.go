package raven

import (
	"bytes"
	"os"
	"testing"
)

// --- HasTerminal ---

func Test_HasTerminal_WithBuffer(t *testing.T) {
	// a bytes.Buffer is never a terminal
	var buf bytes.Buffer
	if HasTerminal(&buf) {
		t.Error("expected HasTerminal to return false for bytes.Buffer")
	}
}

func Test_HasTerminal_WithStdout(t *testing.T) {
	// stdout is not a terminal in test environments (it's piped)
	// so this should return false when running go test
	result := HasTerminal(os.Stdout)
	t.Logf("HasTerminal(os.Stdout) = %v (false expected in CI/test)", result)
}

// --- Parent ---

func Test_Parent_WithChildLogger(t *testing.T) {
	parent := NewNullLogger()
	child := newNoAnchorLogger(parent)

	if Parent(child) != parent {
		t.Error("expected Parent() to return the parent logger")
	}
}

func Test_Parent_WithNonChildLogger(t *testing.T) {
	// NullLogger is not a ChildLogger, so Parent should return nil
	if Parent(NewNullLogger()) != nil {
		t.Error("expected Parent() to return nil for a non-ChildLogger")
	}
}

// --- AddAnchor / RemoveAnchor ---

func Test_AddAnchor_WithNoAnchorSupport(t *testing.T) {
	// NullLogger doesn't support anchors — should return a NoAnchorLogger
	log := NewNullLogger()
	anchored := AddAnchor(log)

	if anchored == nil {
		t.Error("expected AddAnchor to return a non-nil logger")
	}
	if anchored == log {
		t.Error("expected AddAnchor to return a new logger, not the same one")
	}
}

func Test_RemoveAnchor_WithNoAnchorSupport(t *testing.T) {
	log := NewNullLogger()
	anchored := AddAnchor(log)

	// should not panic even with NoAnchorLogger
	RemoveAnchor(anchored)
}

func Test_AddAnchor_WithBuffered(t *testing.T) {
	var buf bytes.Buffer
	log := NewBuffered(&buf, false, &TextPrinter{})
	defer log.Close()

	anchored := AddAnchor(log)
	if anchored == nil {
		t.Error("expected AddAnchor to return a valid logger")
	}
	RemoveAnchor(anchored)
}

// --- findInChain (generic traversal) ---

func Test_findInChain_FindsInterface(t *testing.T) {
	parent := NewNullLogger()
	child := newNoAnchorLogger(parent)

	// should find AnchorRemover in child
	_, ok := findInChain[AnchorRemover](child)
	if !ok {
		t.Error("expected findInChain to find AnchorRemover in chain")
	}
}

func Test_findInChain_ReturnsNotFound(t *testing.T) {
	// NullLogger doesn't implement AnchorAdder
	_, ok := findInChain[AnchorAdder](NewNullLogger())
	if ok {
		t.Error("expected findInChain to return false for missing interface")
	}
}
