package agenttty

import "testing"

func TestCodexExitFooterPresent_requiresBoth(t *testing.T) {
	real := "Token usage: total=1\nTo continue this session, run codex resume 019fdca1-3893-7fa3-a8aa-ebc1ccc750a0\n"
	if !CodexExitFooterPresent(real) {
		t.Fatal("real Codex footer should match (phrase AND resume cmd)")
	}
	withDash := "To continue this session, run codex --resume 019fdca1-3893-7fa3-a8aa-ebc1ccc750a0"
	if !CodexExitFooterPresent(withDash) {
		t.Fatal("phrase + codex --resume should match")
	}
	// Residual UI + full footer still matches.
	withGlyphs := "› old prompt\n" + real + "\n[Terminal exited]\n"
	if !CodexExitFooterPresent(withGlyphs) {
		t.Fatal("footer with residual › should still match")
	}
}

func TestCodexExitFooterPresent_rejectsStandaloneHalves(t *testing.T) {
	onlyPhrase := "To continue this session, please wait"
	if CodexExitFooterPresent(onlyPhrase) {
		t.Fatal("phrase alone must not match")
	}
	onlyResume := "hint: codex resume 019fdca1-3893-7fa3-a8aa-ebc1ccc750a0"
	if CodexExitFooterPresent(onlyResume) {
		t.Fatal("codex resume alone must not match")
	}
	onlyDash := "use codex --resume abcdef01-2345-6789-abcd-ef0123456789"
	if CodexExitFooterPresent(onlyDash) {
		t.Fatal("codex --resume alone must not match")
	}
	if CodexExitFooterPresent("") {
		t.Fatal("empty must not match")
	}
}

func TestScrollbackSuggestsAgentExited_runnerScoped(t *testing.T) {
	codexFooter := "To continue this session, run codex resume 019fdca1-3893-7fa3-a8aa-ebc1ccc750a0"
	if !ScrollbackSuggestsAgentExited(codexFooter, "codex-tty") {
		t.Fatal("codex-tty should accept real footer")
	}
	if ScrollbackSuggestsAgentExited(codexFooter, "grok-tty") {
		t.Fatal("grok-tty must not treat Codex footer as exit")
	}
	if ScrollbackSuggestsAgentExited(codexFooter, "commandcode-tty") {
		t.Fatal("other runners must not treat Codex footer as exit")
	}

	// Standalone halves never fire even for codex-tty.
	if ScrollbackSuggestsAgentExited("codex resume 019fdca1-3893-7fa3-a8aa-ebc1ccc750a0", "codex-tty") {
		t.Fatal("resume cmd alone must not suggest exit")
	}

	grokFooter := "Resume this session with:\n  grok --resume abc"
	if !ScrollbackSuggestsAgentExited(grokFooter, "grok-tty") {
		t.Fatal("grok-tty should accept grok footer")
	}
	if ScrollbackSuggestsAgentExited(grokFooter, "codex-tty") {
		t.Fatal("codex-tty must not treat grok footer as exit")
	}

	if !ScrollbackSuggestsAgentExited("done\n[Terminal exited]\n", "codex-tty") {
		t.Fatal("[Terminal exited] is shared")
	}
	if !ScrollbackSuggestsAgentExited("done\n[Terminal exited]\n", "grok-tty") {
		t.Fatal("[Terminal exited] is shared for grok")
	}
	if !ScrollbackSuggestsAgentExited("done\n[Terminal exited]\n", "unknown-runner") {
		t.Fatal("[Terminal exited] is shared for unknown")
	}
}
