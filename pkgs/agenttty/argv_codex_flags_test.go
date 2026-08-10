package agenttty

import (
	"strings"
	"testing"
)

func TestBuildCodexCommandArgv_HookTrustBypass(t *testing.T) {
	argv, err := BuildCodexCommandArgv(nil, "", "/tmp/fake-codex", "", "")
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, a := range argv {
		if a == "--dangerously-bypass-hook-trust" {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("want 1 --dangerously-bypass-hook-trust, got %d; argv=%v", n, argv)
	}
	// approvals + update check still present
	joined := strings.Join(argv, " ")
	if !strings.Contains(joined, "--dangerously-bypass-approvals-and-sandbox") {
		t.Fatalf("missing approvals bypass: %v", argv)
	}
	if !hasCodexConfigKey(argv, "check_for_update_on_startup") {
		t.Fatalf("missing check_for_update: %v", argv)
	}
}

func TestEnsureCodexBoolFlag_NoDup(t *testing.T) {
	in := []string{"/tmp/fake-codex", "--dangerously-bypass-hook-trust"}
	out := EnsureCodexBoolFlag(in, "--dangerously-bypass-hook-trust")
	n := 0
	for _, a := range out {
		if a == "--dangerously-bypass-hook-trust" {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("want 1, got %d; out=%v", n, out)
	}
}
