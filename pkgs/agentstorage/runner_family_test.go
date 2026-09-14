package agentstorage

import "testing"

func TestRunnerFamily(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in, want string
	}{
		{"codex-tty", RunnerFamilyCodex},
		{"codex", RunnerFamilyCodex},
		{" grok-tty ", RunnerFamilyGrok},
		{"grok", RunnerFamilyGrok},
		{"commandcode-tty", "commandcode-tty"},
		{"", ""},
	}
	for _, tc := range cases {
		if got := RunnerFamily(tc.in); got != tc.want {
			t.Errorf("RunnerFamily(%q)=%q want %q", tc.in, got, tc.want)
		}
	}
}

func TestSameRunnerFamily(t *testing.T) {
	t.Parallel()
	if !SameRunnerFamily("codex", "codex-tty") {
		t.Fatal("codex aliases")
	}
	if !SameRunnerFamily("grok-tty", "grok") {
		t.Fatal("grok aliases")
	}
	if SameRunnerFamily("codex-tty", "grok-tty") {
		t.Fatal("codex vs grok")
	}
	if SameRunnerFamily("", "codex-tty") || SameRunnerFamily("", "") {
		t.Fatal("empty is not a family")
	}
}

func TestHydrateAndSetAndClearRunnerSessions(t *testing.T) {
	t.Parallel()
	m := SessionMeta{Runner: "codex-tty", RunnerSessionID: "codex-1"}
	m.HydrateRunnerSessions()
	if m.RunnerSessions[RunnerFamilyCodex] != "codex-1" {
		t.Fatalf("hydrate: %+v", m.RunnerSessions)
	}
	m.SetRunnerSessionBind("codex-2")
	if m.RunnerSessionID != "codex-2" || m.RunnerSessions[RunnerFamilyCodex] != "codex-2" {
		t.Fatalf("set: live=%q map=%+v", m.RunnerSessionID, m.RunnerSessions)
	}
	m.RunnerSessions[RunnerFamilyGrok] = "grok-1"
	m.ClearCurrentFamilyBind()
	if m.RunnerSessionID != "" {
		t.Fatal("live should clear")
	}
	if _, ok := m.RunnerSessions[RunnerFamilyCodex]; ok {
		t.Fatal("codex slot should clear")
	}
	if m.RunnerSessions[RunnerFamilyGrok] != "grok-1" {
		t.Fatalf("grok slot lost: %+v", m.RunnerSessions)
	}
}
