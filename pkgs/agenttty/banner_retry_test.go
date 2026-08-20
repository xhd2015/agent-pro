package agenttty

import "testing"

func TestShouldRetryCodexSubmit(t *testing.T) {
	tests := []struct {
		name                     string
		runnerID, banner, prompt string
		noSubmit                 bool
		want                     bool
	}{
		{name: "codex tty open prompt", runnerID: "codex-tty", prompt: "follow up", want: true},
		{name: "codex banner provider", runnerID: "custom-tty", banner: "codex", prompt: "follow up", want: true},
		{name: "no submit draft", runnerID: "codex-tty", prompt: "follow up", noSubmit: true},
		{name: "empty prompt", runnerID: "codex-tty"},
		{name: "non codex", runnerID: "grok-tty", banner: "grok", prompt: "follow up"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldRetryCodexSubmit(tt.runnerID, tt.banner, tt.prompt, tt.noSubmit); got != tt.want {
				t.Fatalf("shouldRetryCodexSubmit(%q, %q, %q, %t) = %t, want %t", tt.runnerID, tt.banner, tt.prompt, tt.noSubmit, got, tt.want)
			}
		})
	}
}
