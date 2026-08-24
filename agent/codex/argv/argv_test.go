package argv

import (
	"reflect"
	"strings"
	"testing"
)

func TestStatusInspectArgv(t *testing.T) {
	opts := StatusInspect()
	opts.Bin = "/tmp/codex"
	got, err := Argv(opts)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"/tmp/codex",
		"--dangerously-bypass-approvals-and-sandbox",
		"-c", "check_for_update_on_startup=false",
		"--dangerously-bypass-hook-trust",
		"-c", "mcp_servers={}",
		"--disable", "plugins",
		"--disable", "computer_use",
		"--disable", "in_app_updates",
		"--disable", "hooks",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("StatusInspect argv\n got: %#v\nwant: %#v", got, want)
	}
}

func TestInteractiveArgv_ModelResume(t *testing.T) {
	opts := Interactive()
	opts.Bin = "/tmp/codex"
	opts.Model = "gpt-5.5"
	opts.ResumeSession = "sess-1"
	got, err := Argv(opts)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"/tmp/codex",
		"--dangerously-bypass-approvals-and-sandbox",
		"--model", "gpt-5.5",
		"resume", "sess-1",
		"-c", "check_for_update_on_startup=false",
		"--dangerously-bypass-hook-trust",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Interactive argv\n got: %#v\nwant: %#v", got, want)
	}
}

func TestArgv_CommandOverride_HookOverlay(t *testing.T) {
	got, err := Argv(Options{
		CommandOverride:    "sh -c 'fake'",
		BypassHookTrust:    true,
		DisableUpdateCheck: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"sh", "-c", "fake",
		"-c", "check_for_update_on_startup=false",
		"--dangerously-bypass-hook-trust",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("override argv\n got: %#v\nwant: %#v", got, want)
	}
	// Status-style override: no knobs
	got2, err := Argv(Options{CommandOverride: "/usr/bin/python3 /tmp/fake.py"})
	if err != nil {
		t.Fatal(err)
	}
	want2 := []string{"/usr/bin/python3", "/tmp/fake.py"}
	if !reflect.DeepEqual(got2, want2) {
		t.Fatalf("bare override\n got: %#v\nwant: %#v", got2, want2)
	}
}

func TestArgv_RequiresBin(t *testing.T) {
	_, err := Argv(StatusInspect())
	if err == nil || !strings.Contains(err.Error(), "Bin is required") {
		t.Fatalf("want Bin required error, got %v", err)
	}
}

func TestEnsureIdempotent(t *testing.T) {
	in := []string{"/tmp/codex", "--dangerously-bypass-hook-trust", "-c", "mcp_servers={}", "--disable", "plugins"}
	out := EnsureBoolFlag(in, "--dangerously-bypass-hook-trust")
	out = EnsureConfigFlag(out, "mcp_servers", "{}")
	out = EnsureDisableFeature(out, "plugins")
	if !reflect.DeepEqual(out, in) {
		t.Fatalf("idempotent ensure mutated: %#v -> %#v", in, out)
	}
	if n := countToken(out, "--dangerously-bypass-hook-trust"); n != 1 {
		t.Fatalf("dup bool flag count=%d", n)
	}
}

func countToken(args []string, tok string) int {
	n := 0
	for _, a := range args {
		if a == tok {
			n++
		}
	}
	return n
}
