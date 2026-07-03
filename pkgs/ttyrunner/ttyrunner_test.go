package ttyrunner_test

import (
	"os"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/ttyrunner"
)

func TestIDsIncludeBuiltins(t *testing.T) {
	ids := ttyrunner.IDs()
	hasGrok, hasCodex := false, false
	for _, id := range ids {
		if id == "grok-tty" {
			hasGrok = true
		}
		if id == "codex-tty" {
			hasCodex = true
		}
	}
	if !hasGrok || !hasCodex {
		t.Fatalf("IDs() = %v, want grok-tty and codex-tty", ids)
	}
}

func TestIsTTYRunner(t *testing.T) {
	if !ttyrunner.IsTTYRunner("grok-tty") || !ttyrunner.IsTTYRunner("codex-tty") {
		t.Fatal("expected grok/codex tty runners")
	}
	if ttyrunner.IsTTYRunner("fake-codex") {
		t.Fatal("fake-codex should not be tty runner")
	}
}

func TestRegistryDirConvention(t *testing.T) {
	p, ok := ttyrunner.Get("grok-tty")
	if !ok {
		t.Fatal("grok-tty not registered")
	}
	if p.RegistryDir != "grok-tty-registry" {
		t.Fatalf("RegistryDir = %q", p.RegistryDir)
	}
}

func TestStubAbsentWithoutEnv(t *testing.T) {
	os.Unsetenv("AGENT_RUN_ENABLE_STUB_TTY")
	for _, id := range ttyrunner.IDs() {
		if id == "stub-tty" {
			t.Fatal("stub-tty should not register without env")
		}
	}
}

func TestGrokWritableSealedScrollback(t *testing.T) {
	p, _ := ttyrunner.Get("grok-tty")
	scrollback := []byte("GROK_TTY_BANNER\nGrok > prompt text\nResponse: hello world\n")
	st := p.CheckWritable(scrollback)
	if !st.Ready {
		t.Fatalf("expected writable for sealed scrollback, got %+v", st)
	}
}