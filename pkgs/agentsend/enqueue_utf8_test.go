package agentsend

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// Same truncated multi-byte sequence as the seatalk local-bot agent-run / grok
// env::args panic (0xE5 0x92 without the third byte of the rune).
//
// Note: Enqueue → JSONL currently “repairs” invalid UTF-8 as a side effect of
// encoding/json (U+FFFD). That is accidental and does not cover argv open or
// direct inject. These tests lock an explicit normalizeSendText choke point
// that Enqueue (and inject drain) must use.
func incidentInvalidUTF8Followup() string {
	return "[SeaTalk group follow-up group=G1 thread=T1 from user@example.com]\n" +
		"Messages since last delivery (2):\n" +
		"message_id=m1  [2026-08-04T10:43:32+08:00] user@example.com: " +
		"逻辑" + string([]byte{0xe5, 0x92}) + "edgment (1s)\n" +
		"message_id=m2  [2026-08-04T11:10:27+08:00] alice@example.com: check why the bot is broken\n"
}

func TestNormalizeSendText_InvalidUTF8MustBeValidUTF8(t *testing.T) {
	bad := incidentInvalidUTF8Followup()
	if utf8.ValidString(bad) {
		t.Fatal("fixture setup: expected invalid UTF-8 in follow-up body")
	}

	got := normalizeSendText(bad)
	if !utf8.ValidString(got) {
		t.Fatalf("BUG REPRO: normalizeSendText leaves invalid UTF-8 for follow-up send/inject:\n  %q\n  hex: %x",
			got, []byte(got))
	}
	if !strings.Contains(got, "\uFFFD") {
		t.Fatalf("expected U+FFFD replacement for truncated multi-byte sequence, got %q", got)
	}
	if !strings.Contains(got, "[SeaTalk group follow-up") {
		t.Fatalf("normalized follow-up lost header: %q", got)
	}
	if !strings.Contains(got, "check why the bot is broken") {
		t.Fatalf("normalized follow-up lost valid tail text: %q", got)
	}
	if !strings.Contains(got, "逻辑") {
		t.Fatalf("normalized follow-up lost valid Chinese prefix: %q", got)
	}
}

func TestNormalizeSendText_ShortInvalidUTF8(t *testing.T) {
	bad := "hello" + string([]byte{0xe5, 0x92}) + "world"
	if utf8.ValidString(bad) {
		t.Fatal("fixture setup: expected invalid UTF-8")
	}
	got := normalizeSendText(bad)
	if !utf8.ValidString(got) {
		t.Fatalf("BUG REPRO: short invalid UTF-8 not repaired: %q", got)
	}
	if !strings.Contains(got, "\uFFFD") {
		t.Fatalf("expected U+FFFD, got %q", got)
	}
	if !strings.HasPrefix(got, "hello") || !strings.HasSuffix(got, "world") {
		t.Fatalf("surrounding ASCII lost: %q", got)
	}
}

func TestNormalizeSendText_ValidUTF8Unchanged(t *testing.T) {
	good := "在checkout场景 SPL推荐PC rule的逻辑和prioritization是怎么样的？"
	if !utf8.ValidString(good) {
		t.Fatal(good)
	}
	if got := normalizeSendText(good); got != good {
		t.Fatalf("valid UTF-8 must pass through unchanged:\n  got  %q\n  want %q", got, good)
	}
}

func TestNormalizeSendText_Empty(t *testing.T) {
	if got := normalizeSendText(""); got != "" {
		t.Fatalf("got %q", got)
	}
}

// Enqueue must apply normalizeSendText before persisting so repair is intentional
// (not only an encoding/json side effect) and matches inject/argv behavior.
func TestEnqueue_AppliesNormalizeSendText(t *testing.T) {
	home := t.TempDir()
	sess := Session{
		Home:              home,
		Runner:            "grok-tty",
		TerminalSessionID: "seatalk-local-bot-test",
		ListenAddr:        "127.0.0.1:9",
	}
	bad := incidentInvalidUTF8Followup()
	want := normalizeSendText(bad)

	_, err := Enqueue(home, sess, bad)
	if err != nil {
		t.Fatalf("Enqueue: %v", err)
	}
	entries, err := readEntries(queuePath(home, sess.Runner, sess.TerminalSessionID))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("entries=%d", len(entries))
	}
	// When normalize is still a passthrough, want is invalid UTF-8 and JSON will
	// rewrite it — so this assertion fails until normalize actually repairs and
	// Enqueue stores the repaired string (identical before/after JSON).
	if entries[0].Text != want {
		t.Fatalf("Enqueue did not store normalizeSendText(input):\n  stored %q\n  want   %q",
			entries[0].Text, want)
	}
	if !utf8.ValidString(entries[0].Text) {
		t.Fatalf("stored queue text must be valid UTF-8: %q", entries[0].Text)
	}
}
