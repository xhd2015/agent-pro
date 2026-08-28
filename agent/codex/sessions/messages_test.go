package sessions

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeMessagesRollout(t *testing.T, home, sid string, lines ...string) string {
	t.Helper()
	dir := filepath.Join(home, "sessions", "2026", "08", "01")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "rollout-2026-08-01T12-00-00-"+sid+".jsonl")
	meta := `{"timestamp":"2026-08-01T12:00:00.000Z","type":"session_meta","payload":{"id":"` + sid + `","cwd":"/tmp/proj"}}`
	body := meta + "\n" + strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestMessages_KindsOrderAndSkipPreamble(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	sid := "019f283a-ffff-7fff-ffff-ffffffffff71"
	preamble := `{"timestamp":"2026-08-01T12:00:01.000Z","type":"response_item","payload":{"type":"message","role":"user","content":[{"type":"input_text","text":"# AGENTS.md instructions\n<INSTRUCTIONS>\n` + strings.Repeat("x", 100) + `"}]}}`
	user := `{"timestamp":"2026-08-01T12:00:02.000Z","type":"response_item","payload":{"type":"message","role":"user","content":[{"type":"input_text","text":"hello user"}]}}`
	think := `{"timestamp":"2026-08-01T12:00:03.000Z","type":"response_item","payload":{"type":"reasoning","text":"think hard"}}`
	tool := `{"timestamp":"2026-08-01T12:00:04.000Z","type":"response_item","payload":{"type":"function_call","name":"exec_command","arguments":"{\"cmd\":\"ls\"}"}}`
	asst := `{"timestamp":"2026-08-01T12:00:05.000Z","type":"response_item","payload":{"type":"message","role":"assistant","content":[{"type":"output_text","text":"hello assistant"}]}}`
	writeMessagesRollout(t, home, sid, preamble, user, think, tool, asst)

	res, err := Messages(home, sid, &MessagesOpts{Limit: 10, LimitSet: true, Loc: time.UTC})
	if err != nil {
		t.Fatal(err)
	}
	if res.Total != 4 {
		t.Fatalf("total=%d want 4 (preamble skipped); msgs=%v", res.Total, res.Messages)
	}
	wantKinds := []string{MessageKindUser, MessageKindThinking, MessageKindTool, MessageKindResponse}
	for i, want := range wantKinds {
		if res.Messages[i].Kind != want {
			t.Fatalf("msg[%d].Kind=%q want %q", i, res.Messages[i].Kind, want)
		}
	}
	if !strings.Contains(res.Messages[0].Text, "hello user") {
		t.Fatalf("user text: %q", res.Messages[0].Text)
	}
	if res.Messages[2].Tool != "exec_command" {
		t.Fatalf("tool=%q", res.Messages[2].Tool)
	}
}

func TestMessages_LimitKeepsLatest(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	sid := "019f283a-ffff-7fff-ffff-ffffffffff72"
	var lines []string
	for i := 0; i < 5; i++ {
		lines = append(lines, `{"timestamp":"2026-08-01T12:00:0`+strconvDigit(i)+`.000Z","type":"response_item","payload":{"type":"message","role":"user","content":[{"type":"input_text","text":"msg`+strconvDigit(i)+`"}]}}`)
	}
	writeMessagesRollout(t, home, sid, lines...)
	res, err := Messages(home, sid, &MessagesOpts{Limit: 2, LimitSet: true, Loc: time.UTC})
	if err != nil {
		t.Fatal(err)
	}
	if res.Total != 5 || len(res.Messages) != 2 {
		t.Fatalf("total=%d len=%d", res.Total, len(res.Messages))
	}
	if !strings.Contains(res.Messages[0].Text, "msg3") || !strings.Contains(res.Messages[1].Text, "msg4") {
		t.Fatalf("%v", res.Messages)
	}
}

func TestRunMessages_JSON(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	sid := "019f283a-ffff-7fff-ffff-ffffffffff73"
	writeMessagesRollout(t, home, sid,
		`{"timestamp":"2026-08-01T12:00:02.000Z","type":"response_item","payload":{"type":"message","role":"user","content":[{"type":"input_text","text":"alpha beta"}]}}`,
	)
	var stdout, stderr bytes.Buffer
	err := RunMessages([]string{sid, "--json", "--limit", "5"}, &stdout, &stderr, home, &MessagesOpts{Loc: time.UTC})
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &doc); err != nil {
		t.Fatalf("%v\n%s", err, stdout.String())
	}
	if doc["session_id"] != sid {
		t.Fatalf("%v", doc)
	}
	if strings.Contains(stdout.String(), "\x1b[") {
		t.Fatal("ANSI in JSON")
	}
}

func TestRunMessages_Grep(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	sid := "019f283a-ffff-7fff-ffff-ffffffffff74"
	writeMessagesRollout(t, home, sid,
		`{"timestamp":"2026-08-01T12:00:02.000Z","type":"response_item","payload":{"type":"message","role":"user","content":[{"type":"input_text","text":"keep me zebra"}]}}`,
		`{"timestamp":"2026-08-01T12:00:03.000Z","type":"response_item","payload":{"type":"message","role":"user","content":[{"type":"input_text","text":"drop me"}]}}`,
	)
	var stdout, stderr bytes.Buffer
	err := RunMessages([]string{sid, "--grep", "zebra", "--limit", "10"}, &stdout, &stderr, home, &MessagesOpts{Loc: time.UTC})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "keep me zebra") || strings.Contains(stdout.String(), "drop me") {
		t.Fatalf("%s", stdout.String())
	}
}

func TestRunMessages_Unknown(t *testing.T) {
	t.Parallel()
	err := RunMessages([]string{"019f283a-ffff-7fff-ffff-ffffffffff99"}, ioDiscard(), ioDiscard(), t.TempDir(), nil)
	if err == nil || !strings.Contains(err.Error(), "codex session not found") {
		t.Fatalf("err=%v", err)
	}
}

func strconvDigit(i int) string {
	return string(rune('0' + i))
}
