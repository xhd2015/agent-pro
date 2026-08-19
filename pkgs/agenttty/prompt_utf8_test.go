package agenttty

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// incidentInvalidUTF8Prompt reproduces the seatalk local-bot → agent-run → grok
// crash: a truncated multi-byte sequence (0xE5 0x92, incomplete U+548C 和 / similar)
// embedded in an open inject body. Real grok panics at:
//
//	library/std/src/env.rs — std::env::args() unwrap on OsString::into_string
//
// Repro binary: grok $'hello\xe5\x92world'
func incidentInvalidUTF8Prompt() string {
	// Mid-character cut as seen in meta initial_prompt / panic Err value:
	// "...逻辑" + incomplete UTF-8 + "edgment (1s)" from "acknowledgment".
	return "SeaTalk local-bot session open\n" +
		"session-id: seatalk-local-bot-test\n" +
		"First message from master:\n" +
		"在checkout场景 SPL推荐PC rule的逻辑" + string([]byte{0xe5, 0x92}) + "edgment (1s)\n" +
		"message_idmessage_id alice@example.com: check why the bot is broken\n"
}

func TestNormalizeRunnerPrompt_InvalidUTF8MustBeValidUTF8(t *testing.T) {
	bad := incidentInvalidUTF8Prompt()
	if utf8.ValidString(bad) {
		t.Fatal("fixture setup: expected invalid UTF-8 in incident prompt")
	}
	got := normalizeRunnerPrompt(bad)
	if !utf8.ValidString(got) {
		t.Fatalf("BUG REPRO: normalizeRunnerPrompt leaves invalid UTF-8 "+
			"(grok std::env::args panics):\n  %q", got)
	}
	if !strings.Contains(got, "\uFFFD") {
		t.Fatalf("expected U+FFFD replacement, got %q", got)
	}
	if !strings.Contains(got, "在checkout场景") {
		t.Fatalf("lost valid Chinese text: %q", got)
	}
}

func TestAppendNewSessionPrompt_GrokInvalidUTF8MustBeValidUTF8(t *testing.T) {
	bad := incidentInvalidUTF8Prompt()
	if utf8.ValidString(bad) {
		t.Fatal("fixture setup: expected invalid UTF-8 in incident prompt")
	}

	argv := appendNewSessionPrompt([]string{"grok", "--always-approve"}, "grok-tty", bad, false, false)
	if len(argv) < 3 {
		t.Fatalf("expected prompt on argv, got %q", argv)
	}
	got := argv[len(argv)-1]
	if !utf8.ValidString(got) {
		b := []byte(got)
		start := len(b) - 40
		if start < 0 {
			start = 0
		}
		t.Fatalf("BUG REPRO: grok-tty new-session argv PROMPT is invalid UTF-8 "+
			"(rust std::env::args panics at env.rs):\n  %q\n  hex suffix: %x",
			got, b[start:])
	}
	// After normalize: incomplete sequence becomes U+FFFD; surrounding text kept.
	if !strings.Contains(got, "\uFFFD") {
		t.Fatalf("expected U+FFFD replacement for truncated multi-byte sequence, got %q", got)
	}
	if !strings.Contains(got, "SeaTalk local-bot session open") {
		t.Fatalf("normalized prompt lost header: %q", got)
	}
	if !strings.Contains(got, "在checkout场景") {
		t.Fatalf("normalized prompt lost valid Chinese text: %q", got)
	}
}

func TestAppendNewSessionPrompt_CommandcodeInvalidUTF8MustBeValidUTF8(t *testing.T) {
	bad := "hello" + string([]byte{0xe5, 0x92}) + "world"
	if utf8.ValidString(bad) {
		t.Fatal("fixture setup: expected invalid UTF-8")
	}
	// headless commandcode uses -p <prompt>
	argv := appendNewSessionPrompt([]string{"commandcode"}, "commandcode-tty", bad, false, false)
	if len(argv) != 3 || argv[1] != "-p" {
		t.Fatalf("expected -p prompt, got %q", argv)
	}
	if !utf8.ValidString(argv[2]) {
		t.Fatalf("BUG REPRO: commandcode -p value is invalid UTF-8: %q", argv[2])
	}
	// open omits argv prompt (inject after ready)
	openArgv := appendNewSessionPrompt([]string{"commandcode"}, "commandcode-tty", "Hello", false, true)
	if len(openArgv) != 1 || openArgv[0] != "commandcode" {
		t.Fatalf("commandcode --open must omit argv prompt, got %q", openArgv)
	}
}

func TestAppendNewSessionPrompt_ValidChineseUnchanged(t *testing.T) {
	good := "在checkout场景 SPL推荐PC rule的逻辑和prioritization是怎么样的？"
	if !utf8.ValidString(good) {
		t.Fatal(good)
	}
	argv := appendNewSessionPrompt([]string{"grok"}, "grok-tty", good, false, true)
	if len(argv) != 2 || argv[1] != good {
		t.Fatalf("valid UTF-8 must pass through unchanged: got %q", argv)
	}
}

func TestAppendNewSessionPrompt_CodexSkipsArgv(t *testing.T) {
	bad := "x" + string([]byte{0xe5, 0x92}) + "y"
	argv := appendNewSessionPrompt([]string{"codex"}, "codex-tty", bad, false, false)
	if len(argv) != 1 {
		t.Fatalf("codex-tty must not put prompt on argv, got %q", argv)
	}
}

func TestAppendNewSessionPrompt_NoSubmitSkipsArgv(t *testing.T) {
	argv := appendNewSessionPrompt([]string{"grok"}, "grok-tty", "hi", true, true)
	if len(argv) != 1 {
		t.Fatalf("NoSubmit must not put prompt on argv, got %q", argv)
	}
}
