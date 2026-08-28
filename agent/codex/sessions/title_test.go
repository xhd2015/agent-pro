package sessions

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractTitle_ResponseItemUser(t *testing.T) {
	t.Parallel()
	line := `{"type":"response_item","payload":{"type":"message","role":"user","content":[{"type":"input_text","text":"fix the auth bug"}]}}`
	got := extractTitle([]string{line})
	if got != "fix the auth bug" {
		t.Fatalf("title=%q", got)
	}
}

func TestExtractTitle_SkipsAgentsPreamble(t *testing.T) {
	t.Parallel()
	agents := `{"type":"response_item","payload":{"type":"message","role":"user","content":[{"type":"input_text","text":"# AGENTS.md instructions for /tmp\n\n<INSTRUCTIONS>\nbig dump` + strings.Repeat("x", 100) + `"}]}}`
	follow := `{"type":"response_item","payload":{"type":"message","role":"user","content":[{"type":"input_text","text":"SeaTalk local-bot session open\nsession-id: abc"}]}}`
	got := extractTitle([]string{agents, follow})
	if !strings.HasPrefix(got, "SeaTalk local-bot session open") {
		t.Fatalf("title=%q", got)
	}
}

func TestExtractTitle_LegacyEventMsg(t *testing.T) {
	t.Parallel()
	line := `{"type":"event_msg","payload":{"type":"user_message","message":"legacy hello"}}`
	got := extractTitle([]string{line})
	if got != "legacy hello" {
		t.Fatalf("title=%q", got)
	}
}

func TestReadLines_LargeToken(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "big.jsonl")
	// >64KiB default scanner limit
	big := strings.Repeat("a", 80*1024)
	body := `{"type":"session_meta","payload":{"id":"019f283a-ffff-7fff-ffff-ffffffffff71","cwd":"/tmp/ws","timestamp":"2026-08-01T12:00:00.000Z","pad":"` + big + `"}}` + "\n" +
		`{"type":"response_item","payload":{"type":"message","role":"user","content":[{"type":"input_text","text":"short title please"}]}}` + "\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	lines, err := readLines(path)
	if err != nil {
		t.Fatalf("readLines: %v", err)
	}
	if len(lines) != 2 {
		t.Fatalf("lines=%d", len(lines))
	}
	sess, err := sessionFromFile(path, lines)
	if err != nil {
		t.Fatal(err)
	}
	if sess.CWD != "/tmp/ws" {
		t.Fatalf("cwd=%q", sess.CWD)
	}
	if sess.Title != "short title please" {
		t.Fatalf("title=%q", sess.Title)
	}
}

func TestMetaFromCodexOpenPath_LargeThenTitle(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	sid := "019f283a-ffff-7fff-ffff-ffffffffff72"
	dir := filepath.Join(home, "sessions", "2026", "08", "01")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "rollout-2026-08-01T12-00-00-"+sid+".jsonl")
	big := strings.Repeat("b", 90*1024)
	body := `{"type":"session_meta","payload":{"id":"` + sid + `","cwd":"/Users/me/proj","timestamp":"2026-08-01T12:00:00.000Z","pad":"` + big + `"}}` + "\n" +
		`{"type":"response_item","payload":{"type":"message","role":"user","content":[{"type":"input_text","text":"# AGENTS.md instructions\n<INSTRUCTIONS>\n` + strings.Repeat("z", 100) + `"}]}}` + "\n" +
		`{"type":"response_item","payload":{"type":"message","role":"user","content":[{"type":"input_text","text":"implement open send list"}]}}` + "\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	// Fixture path under home; also try via .codex-shaped open path remapped
	openPath := filepath.Join(home, ".codex", "sessions", "2026", "08", "01", filepath.Base(path))
	// Use path directly with empty codexHome (no remap)
	meta, ok := metaFromCodexOpenPath(path, "")
	if !ok {
		t.Fatal("metaFromCodexOpenPath failed")
	}
	if meta.Cwd != "/Users/me/proj" {
		t.Fatalf("cwd=%q", meta.Cwd)
	}
	if meta.Title != "implement open send list" {
		t.Fatalf("title=%q", meta.Title)
	}
	_ = openPath
}
