package argv

import (
	"reflect"
	"strings"
	"testing"

	codexcfg "github.com/xhd2015/agent-pro/agent/codex/config"
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

func TestArgv_ProjectsTrustLevel(t *testing.T) {
	opts := Interactive()
	opts.Bin = "/tmp/codex"
	opts.Projects = TrustProject("/Users/me/hub")
	got, err := Argv(opts)
	if err != nil {
		t.Fatal(err)
	}
	wantKey := `projects."/Users/me/hub".trust_level`
	wantVal := "trusted"
	found := false
	for i := 0; i+1 < len(got); i++ {
		if got[i] == "-c" && got[i+1] == wantKey+"="+wantVal {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("missing %s=%s in %#v", wantKey, wantVal, got)
	}
}

func TestArgv_ProjectsWinOverExtraConfig(t *testing.T) {
	path := "/tmp/ws"
	key := projectTrustConfigKey(path)
	opts := Options{
		Bin: "/tmp/codex",
		Projects: map[string]codexcfg.Project{
			path: {TrustLevel: codexcfg.TrustTrusted},
		},
		ExtraConfig: map[string]string{
			key: "untrusted",
		},
	}
	got, err := Argv(opts)
	if err != nil {
		t.Fatal(err)
	}
	var values []string
	for i := 0; i+1 < len(got); i++ {
		if got[i] != "-c" {
			continue
		}
		if strings.HasPrefix(got[i+1], key+"=") {
			values = append(values, strings.TrimPrefix(got[i+1], key+"="))
		}
	}
	if len(values) != 1 || values[0] != "trusted" {
		t.Fatalf("want single trusted override, got %v in %#v", values, got)
	}
}

func TestProjectTrustConfigKeyEscapesQuotes(t *testing.T) {
	got := projectTrustConfigKey(`/tmp/odd"path`)
	want := `projects."/tmp/odd\"path".trust_level`
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestTrustProjectEmpty(t *testing.T) {
	if TrustProject("  ") != nil {
		t.Fatal("empty path should yield nil map")
	}
}

func TestEnsureTrustedProject(t *testing.T) {
	got := EnsureTrustedProject([]string{"/tmp/codex"}, "/hub")
	want := []string{"/tmp/codex", "-c", `projects."/hub".trust_level=trusted`}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v want %#v", got, want)
	}
	// idempotent
	got2 := EnsureTrustedProject(got, "/hub")
	if !reflect.DeepEqual(got2, got) {
		t.Fatalf("idempotent failed: %#v", got2)
	}
	got3 := EnsureTrustedProject(got, "  ")
	if !reflect.DeepEqual(got3, got) {
		t.Fatalf("empty path should leave argv unchanged: %#v", got3)
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
