package sessions

import (
	"errors"
	"strings"
	"testing"
)

func TestPreferAgentRunOpen_NoAgentRunSkips(t *testing.T) {
	called := false
	hit, warn, err := preferAgentRunOpen(agentRunOpenHooks{
		NoAgentRun: true,
		AgentRunOpen: func(string, string) (*AgentRunOpenResult, error) {
			called = true
			return &AgentRunOpenResult{AgentRunSessionID: "ar1", Mode: AgentRunOpenModeSend, Delivered: true}, nil
		},
	}, "g1", "hi")
	if err != nil || hit != nil || warn != "" {
		t.Fatalf("NoAgentRun: hit=%v warn=%q err=%v", hit, warn, err)
	}
	if called {
		t.Fatal("AgentRunOpen must not be called with NoAgentRun")
	}
}

func TestPreferAgentRunOpen_HookHit(t *testing.T) {
	hit, warn, err := preferAgentRunOpen(agentRunOpenHooks{
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
	if err != nil || hit == nil || warn != "" || hit.AgentRunSessionID != "ar1" {
		t.Fatalf("got hit=%v warn=%q err=%v", hit, warn, err)
	}
}

func TestPreferAgentRunOpen_AmbiguousWarn(t *testing.T) {
	hit, warn, err := preferAgentRunOpen(agentRunOpenHooks{
		AgentRunOpen: func(string, string) (*AgentRunOpenResult, error) {
			return nil, errAmbiguousForTest{}
		},
	}, "x", "y")
	if err != nil || hit != nil {
		t.Fatalf("ambiguous should soft-miss: hit=%v err=%v", hit, err)
	}
	if !strings.Contains(warn, "warning:") || !strings.Contains(warn, "ambiguous") {
		t.Fatalf("warn=%q", warn)
	}
}

type errAmbiguousForTest struct{}

func (errAmbiguousForTest) Error() string {
	return "ambiguous grok-session-id x: multiple matches: a, b"
}

func TestPreferAgentRunOpen_DeliverFailHardError(t *testing.T) {
	hit, warn, err := preferAgentRunOpen(agentRunOpenHooks{
		AgentRunOpen: func(string, string) (*AgentRunOpenResult, error) {
			return nil, errors.New("terminal unreachable")
		},
	}, "g1", "hi")
	if err == nil || hit != nil {
		t.Fatalf("deliver fail must hard-error: hit=%v err=%v", hit, err)
	}
	if warn != "" {
		t.Fatalf("warn=%q", warn)
	}
	if !strings.Contains(err.Error(), "agent-run deliver failed") {
		t.Fatalf("err=%v", err)
	}
}

func TestPreferAgentRunOpen_SkipProduction(t *testing.T) {
	hit, warn, err := preferAgentRunOpen(agentRunOpenHooks{
		SkipProduction: true,
	}, "any", "p")
	if err != nil || hit != nil || warn != "" {
		t.Fatalf("SkipProduction: hit=%v warn=%q err=%v", hit, warn, err)
	}
}

func TestPreferAgentRunOpen_SoftMissNil(t *testing.T) {
	hit, warn, err := preferAgentRunOpen(agentRunOpenHooks{
		AgentRunOpen: func(string, string) (*AgentRunOpenResult, error) {
			return nil, nil
		},
	}, "miss", "p")
	if err != nil || hit != nil || warn != "" {
		t.Fatalf("soft miss: hit=%v warn=%q err=%v", hit, warn, err)
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
