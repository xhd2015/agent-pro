package tty

import (
	"reflect"
	"testing"
)

func TestBuildCodexArgv_CommandOverrideBare(t *testing.T) {
	t.Parallel()
	got, err := buildCodexArgv(newExecEnv(), Options{Command: "/tmp/fake-codex-bin-for-status"})
	if err != nil {
		t.Fatal(err)
	}
	// Fake TUI hooks must not receive StatusInspect knobs.
	want := []string{"/tmp/fake-codex-bin-for-status"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("override argv = %#v, want %#v", got, want)
	}
}

func TestBuildCodexArgv_RealBinUsesStatusInspect(t *testing.T) {
	t.Parallel()
	got, err := buildCodexArgv(newExecEnv(), Options{})
	if err != nil {
		t.Skip(err.Error())
	}
	requiredPairs := [][2]string{
		{"--disable", "plugins"},
		{"--disable", "computer_use"},
		{"--disable", "in_app_updates"},
		{"--disable", "hooks"},
	}
	requiredTokens := []string{
		"--dangerously-bypass-approvals-and-sandbox",
		"--dangerously-bypass-hook-trust",
		"check_for_update_on_startup=false",
		"mcp_servers={}",
	}
	for _, tok := range requiredTokens {
		if !containsToken(got, tok) {
			t.Fatalf("missing %q in argv=%v", tok, got)
		}
	}
	for _, pair := range requiredPairs {
		if !containsPair(got, pair[0], pair[1]) {
			t.Fatalf("missing %q %q in argv=%v", pair[0], pair[1], got)
		}
	}
}

func containsToken(args []string, tok string) bool {
	for _, a := range args {
		if a == tok {
			return true
		}
	}
	return false
}

func containsPair(args []string, a, b string) bool {
	for i := 0; i+1 < len(args); i++ {
		if args[i] == a && args[i+1] == b {
			return true
		}
	}
	return false
}
