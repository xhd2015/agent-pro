package agentui

import "testing"

func TestParseRunOptionsAcceptsCodexRunner(t *testing.T) {
	opts, err := parseRunOptions(Config{Usage: "usage"}, []string{"--agent-runner", "codex", "build tests"})
	if err != nil {
		t.Fatalf("parseRunOptions: %v", err)
	}
	if opts.AgentRunner != "codex" {
		t.Fatalf("AgentRunner = %q, want codex", opts.AgentRunner)
	}
	if opts.Feature != "build tests" {
		t.Fatalf("Feature = %q, want build tests", opts.Feature)
	}
}

func TestParseRunOptionsAcceptsFakeCodexRunner(t *testing.T) {
	opts, err := parseRunOptions(Config{Usage: "usage"}, []string{"--agent-runner", "fake-codex", "build tests"})
	if err != nil {
		t.Fatalf("parseRunOptions: %v", err)
	}
	if opts.AgentRunner != "fake-codex" {
		t.Fatalf("AgentRunner = %q, want fake-codex", opts.AgentRunner)
	}
}

func TestParseRunOptionsRejectsUnknownRunner(t *testing.T) {
	_, err := parseRunOptions(Config{Usage: "usage"}, []string{"--agent-runner", "unknown", "build tests"})
	if err == nil {
		t.Fatal("expected error")
	}
}
