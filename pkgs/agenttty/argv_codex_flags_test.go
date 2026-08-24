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

func TestFormatArgvLineAndVerboseNotice(t *testing.T) {
	line := FormatArgvLine([]string{"/tmp/codex", "--model", "gpt-5.5", "-c", `projects."/hub".trust_level=trusted`})
	if !strings.Contains(line, "/tmp/codex") || !strings.Contains(line, "trust_level") {
		t.Fatalf("FormatArgvLine = %q", line)
	}
	var buf strings.Builder
	writeVerboseArgvNotice(&buf, "codex-tty", []string{"/tmp/codex", "--model", "m"})
	got := buf.String()
	if !strings.HasPrefix(got, "notice: codex argv:") {
		t.Fatalf("notice = %q", got)
	}
}

func TestApplyCodexTrustedProject(t *testing.T) {
	in := []string{"/tmp/fake-codex"}
	out := ApplyCodexTrustedProject(in, "/Users/me/seatalk-local-bot/knowledge-base-hub")
	joined := strings.Join(out, "\x00")
	want := `projects."/Users/me/seatalk-local-bot/knowledge-base-hub".trust_level=trusted`
	if !strings.Contains(joined, want) {
		t.Fatalf("missing trust override %q in %#v", want, out)
	}
	out2 := ApplyCodexTrustedProject(in, "")
	if len(out2) != len(in) || out2[0] != in[0] {
		t.Fatalf("empty workspace mutated: %#v", out2)
	}
}
