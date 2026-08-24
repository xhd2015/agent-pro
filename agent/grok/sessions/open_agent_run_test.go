package sessions

import (
	"strings"
	"testing"
)

func TestPreferAgentRunOpen_NoAgentRunSkips(t *testing.T) {
	called := false
	hit, warn, ok := preferAgentRunOpen(agentRunOpenHooks{
		NoAgentRun: true,
		AgentRunOpen: func(string, string) (*AgentRunOpenResult, error) {
			called = true
			return &AgentRunOpenResult{AgentRunSessionID: "ar1", Mode: AgentRunOpenModeSend, Delivered: true}, nil
		},
	}, "g1", "hi")
	if ok || hit != nil || warn != "" {
		t.Fatalf("NoAgentRun: hit=%v warn=%q ok=%v", hit, warn, ok)
	}
	if called {
		t.Fatal("AgentRunOpen must not be called with NoAgentRun")
	}
}

func TestPreferAgentRunOpen_HookHit(t *testing.T) {
	hit, warn, ok := preferAgentRunOpen(agentRunOpenHooks{
		AgentRunOpen: func(id, prompt string) (*AgentRunOpenResult, error) {
			if id != "g1" || prompt != "hi" {
				t.Fatalf("hook args %q %q", id, prompt)
			}
			return &AgentRunOpenResult{
				AgentRunSessionID: "ar1",
				Mode:              AgentRunOpenModeSend,
				Delivered:         true,
			}, nil
		},
	}, "g1", "hi")
	if !ok || hit == nil || warn != "" || hit.AgentRunSessionID != "ar1" {
		t.Fatalf("got hit=%v warn=%q ok=%v", hit, warn, ok)
	}
}

func TestPreferAgentRunOpen_AmbiguousWarn(t *testing.T) {
	hit, warn, ok := preferAgentRunOpen(agentRunOpenHooks{
		AgentRunOpen: func(string, string) (*AgentRunOpenResult, error) {
			return nil, errAmbiguousForTest{}
		},
	}, "x", "y")
	if ok || hit != nil {
		t.Fatalf("ambiguous should soft-miss: hit=%v ok=%v", hit, ok)
	}
	if !strings.Contains(warn, "warning:") || !strings.Contains(warn, "ambiguous") {
		t.Fatalf("warn=%q", warn)
	}
}

type errAmbiguousForTest struct{}

func (errAmbiguousForTest) Error() string {
	return "ambiguous grok-session-id x: multiple matches: a, b"
}

func TestPreferAgentRunOpen_SkipProduction(t *testing.T) {
	hit, warn, ok := preferAgentRunOpen(agentRunOpenHooks{
		SkipProduction: true,
	}, "any", "p")
	if ok || hit != nil || warn != "" {
		t.Fatalf("SkipProduction: hit=%v warn=%q ok=%v", hit, warn, ok)
	}
}

func TestPreferAgentRunOpen_SoftMissNil(t *testing.T) {
	hit, warn, ok := preferAgentRunOpen(agentRunOpenHooks{
		AgentRunOpen: func(string, string) (*AgentRunOpenResult, error) {
			return nil, nil
		},
	}, "miss", "p")
	if ok || hit != nil || warn != "" {
		t.Fatalf("soft miss: hit=%v warn=%q ok=%v", hit, warn, ok)
	}
}

func TestBuildAgentRunAutoOpenCommand(t *testing.T) {
	got := buildAgentRunAutoOpenCommand("/usr/local/bin/agent-run", "ar-1", "/tmp/ws", "hello world", true)
	for _, want := range []string{
		"/usr/local/bin/agent-run",
		"run",
		"--session-id",
		"ar-1",
		"--auto-send-or-resume",
		"--open",
		"--dir",
		"--no-submit",
		"hello world",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("command missing %q: %s", want, got)
		}
	}
}
