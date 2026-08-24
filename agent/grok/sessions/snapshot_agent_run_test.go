package sessions

import (
	"fmt"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

func TestPreferAgentRunSnapshot_ForceITermSkips(t *testing.T) {
	called := false
	opts := &SnapshotOpts{
		ForceITerm: true,
		AgentRunSnapshot: func(string) (*AgentRunSnapshotResult, error) {
			called = true
			return &AgentRunSnapshotResult{Contents: "x"}, nil
		},
	}
	hit, warn, ok := preferAgentRunSnapshot(opts, "sid")
	if ok || hit != nil || warn != "" {
		t.Fatalf("ForceITerm: got hit=%v warn=%q ok=%v", hit, warn, ok)
	}
	if called {
		t.Fatal("AgentRunSnapshot must not be called with ForceITerm")
	}
}

func TestPreferAgentRunSnapshot_HookHit(t *testing.T) {
	opts := &SnapshotOpts{
		AgentRunSnapshot: func(id string) (*AgentRunSnapshotResult, error) {
			if id != "g1" {
				t.Fatalf("id=%q", id)
			}
			return &AgentRunSnapshotResult{AgentRunSessionID: "ar1", Contents: "frame"}, nil
		},
	}
	hit, warn, ok := preferAgentRunSnapshot(opts, "g1")
	if !ok || warn != "" || hit == nil || hit.Contents != "frame" || hit.AgentRunSessionID != "ar1" {
		t.Fatalf("got hit=%+v warn=%q ok=%v", hit, warn, ok)
	}
}

func TestPreferAgentRunSnapshot_AmbiguousWarn(t *testing.T) {
	opts := &SnapshotOpts{
		AgentRunSnapshot: func(string) (*AgentRunSnapshotResult, error) {
			return nil, fmt.Errorf("ambiguous grok-session-id x: multiple matches: a, b")
		},
	}
	hit, warn, ok := preferAgentRunSnapshot(opts, "x")
	if ok || hit != nil {
		t.Fatalf("want soft miss, got hit=%v ok=%v", hit, ok)
	}
	if len(warn) < 8 || warn[:8] != "warning:" {
		t.Fatalf("warn=%q", warn)
	}
}

func TestPreferAgentRunSnapshot_ListITermSkipsProduction(t *testing.T) {
	opts := &SnapshotOpts{
		ListITerm: func() ([]iterm2.SessionRef, error) { return nil, nil },
	}
	hit, warn, ok := preferAgentRunSnapshot(opts, "any")
	if ok || hit != nil || warn != "" {
		t.Fatalf("ListITerm inject must skip production: hit=%v warn=%q ok=%v", hit, warn, ok)
	}
}

func TestPreferAgentRunSnapshot_SoftMissNil(t *testing.T) {
	opts := &SnapshotOpts{
		AgentRunSnapshot: func(string) (*AgentRunSnapshotResult, error) {
			return nil, nil
		},
	}
	hit, warn, ok := preferAgentRunSnapshot(opts, "miss")
	if ok || hit != nil || warn != "" {
		t.Fatalf("soft miss: hit=%v warn=%q ok=%v", hit, warn, ok)
	}
}
