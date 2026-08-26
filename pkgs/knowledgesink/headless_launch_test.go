package knowledgesink

import "testing"

func TestAgentLogTag(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"codex", "codex"},
		{"codex-tty", "codex"},
		{"grok-tty", "grok"},
		{"fake-codex", "fake-codex"},
		{"", "agent"},
		{"  grok  ", "grok"},
	}
	for _, tc := range cases {
		if got := agentLogTag(tc.in); got != tc.want {
			t.Fatalf("agentLogTag(%q)=%q want %q", tc.in, got, tc.want)
		}
	}
}
