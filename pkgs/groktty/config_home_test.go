package groktty

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveAgentRunnerConfigHome_flagWins(t *testing.T) {
	t.Setenv(EnvAgentRunnerConfigHome, "/from-env")
	got := ResolveAgentRunnerConfigHome("/from-flag")
	if got != "/from-flag" {
		t.Fatalf("got %q, want /from-flag", got)
	}
}

func TestResolveAgentRunnerConfigHome_envFallback(t *testing.T) {
	t.Setenv(EnvAgentRunnerConfigHome, "/from-env")
	got := ResolveAgentRunnerConfigHome("")
	if got != "/from-env" {
		t.Fatalf("got %q, want /from-env", got)
	}
}

func TestGrokHomeForRunner_usesConfigHome(t *testing.T) {
	t.Setenv(envGrokHome, "")
	home := filepath.Join(t.TempDir(), "shared-grok-home")
	got := GrokHomeForRunner(home)
	if got != home {
		t.Fatalf("got %q, want %q", got, home)
	}
}

func TestAutoProvisionGrokConfigHome_llmMockBinary(t *testing.T) {
	got, err := AutoProvisionGrokConfigHome("grok-tty", "llm-mock-run-grok", "")
	if err != nil {
		t.Fatalf("AutoProvisionGrokConfigHome: %v", err)
	}
	if got == "" {
		t.Fatal("expected auto-provisioned home")
	}
	if _, err := os.Stat(got); err != nil {
		t.Fatalf("stat auto home: %v", err)
	}
}

func TestPrependCommandEnv_grokTTY(t *testing.T) {
	got := PrependCommandEnv([]string{"llm-mock-run-grok"}, "grok-tty", "/mock/grok")
	want := []string{"env", "GROK_HOME=/mock/grok", "llm-mock-run-grok"}
	if len(got) != len(want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d]=%q want %q (full %#v)", i, got[i], want[i], got)
		}
	}
}

func TestAutoProvisionGrokConfigHome_skipsWhenConfigured(t *testing.T) {
	got, err := AutoProvisionGrokConfigHome("grok-tty", "llm-mock-run-grok", "/configured")
	if err != nil {
		t.Fatalf("AutoProvisionGrokConfigHome: %v", err)
	}
	if got != "" {
		t.Fatalf("got %q, want empty", got)
	}
}