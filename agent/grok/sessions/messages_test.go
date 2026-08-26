package sessions

import (
	"encoding/json"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func TestCapMessageBody(t *testing.T) {
	t.Parallel()
	body, trunc := capMessageBody("hello", 10)
	if trunc || body != "hello" {
		t.Fatalf("short: got %q trunc=%v", body, trunc)
	}
	long := strings.Repeat("あ", 10)
	body, trunc = capMessageBody(long, 5)
	if !trunc {
		t.Fatal("expected trunc")
	}
	if utf8.RuneCountInString(body) != 5 {
		t.Fatalf("rune count = %d, want 5 (%q)", utf8.RuneCountInString(body), body)
	}
	if !strings.HasSuffix(body, "…") {
		t.Fatalf("want ellipsis suffix: %q", body)
	}
}

func TestPageMessagesFromEnd(t *testing.T) {
	t.Parallel()
	all := []ChatMessage{
		{Kind: MessageKindUser, Text: "0"},
		{Kind: MessageKindUser, Text: "1"},
		{Kind: MessageKindUser, Text: "2"},
		{Kind: MessageKindUser, Text: "3"},
		{Kind: MessageKindUser, Text: "4"},
	}
	page := pageMessagesFromEnd(all, 0, 2)
	if len(page) != 2 || page[0].Text != "3" || page[1].Text != "4" {
		t.Fatalf("limit 2: %+v", page)
	}
	page = pageMessagesFromEnd(all, 2, 2)
	if len(page) != 2 || page[0].Text != "1" || page[1].Text != "2" {
		t.Fatalf("offset 2 limit 2: %+v", page)
	}
	page = pageMessagesFromEnd(all, 1, 32)
	if len(page) != 4 || page[len(page)-1].Text != "3" {
		t.Fatalf("offset 1: %+v", page)
	}
	if page := pageMessagesFromEnd(all, 5, 2); page != nil {
		t.Fatalf("offset past end: %+v", page)
	}
	page = pageMessagesFromEnd(all, 3, 0)
	if len(page) != 2 || page[0].Text != "0" || page[1].Text != "1" {
		t.Fatalf("limit 0 all remaining: %+v", page)
	}
}

func TestEventsToChatMessagesKindsAndCaps(t *testing.T) {
	t.Parallel()
	userLong := strings.Repeat("u", MessagesCapUser+10)
	thinkLong := strings.Repeat("t", MessagesCapThinking+10)
	toolLong := strings.Repeat("x", MessagesCapTool+50)
	respLong := strings.Repeat("r", MessagesCapResponse+10)

	events := []types.AgentEvent{
		{Type: types.ActionMessage, Role: "user", Text: userLong},
		{Type: types.ActionThink, Text: thinkLong},
		{
			Type:      types.ActionToolCall,
			Tool:      "run_terminal_command",
			Text:      "run_terminal_command",
			ToolInput: map[string]any{"command": toolLong},
			Extensions: &types.EventExtensions{
				GrokSession: &types.GrokSessionExtension{Status: "pending"},
			},
		},
		{Type: types.ActionMessage, Role: "assistant", Text: respLong},
		{Type: types.ActionDone},
	}
	msgs := eventsToChatMessages(events)
	if len(msgs) != 4 {
		t.Fatalf("len=%d want 4", len(msgs))
	}
	checks := []struct {
		kind string
		cap  int
	}{
		{MessageKindUser, MessagesCapUser},
		{MessageKindThinking, MessagesCapThinking},
		{MessageKindTool, MessagesCapTool},
		{MessageKindResponse, MessagesCapResponse},
	}
	for i, c := range checks {
		if msgs[i].Kind != c.kind {
			t.Fatalf("[%d] kind=%q want %q", i, msgs[i].Kind, c.kind)
		}
		if !msgs[i].Truncated {
			t.Fatalf("[%d] expected truncated", i)
		}
		if utf8.RuneCountInString(msgs[i].Text) != c.cap {
			t.Fatalf("[%d] runes=%d want %d", i, utf8.RuneCountInString(msgs[i].Text), c.cap)
		}
	}
	if !strings.HasPrefix(msgs[2].Text, "run_terminal_command: ") {
		t.Fatalf("tool body: %q", msgs[2].Text)
	}
}

func TestFormatToolUseBodyPrefersCommand(t *testing.T) {
	t.Parallel()
	got := formatToolUseBody("run_terminal_command", map[string]any{
		"command":     "kck grok list --limit 3",
		"description": "list sessions",
	})
	want := "run_terminal_command: kck grok list --limit 3"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestEventsToChatMessagesSkipsCompletedTool(t *testing.T) {
	t.Parallel()
	events := []types.AgentEvent{
		{
			Type:      types.ActionToolCall,
			Tool:      "run_terminal_command",
			Text:      "run_terminal_command",
			ToolInput: map[string]any{"command": "echo hi"},
			Extensions: &types.EventExtensions{
				GrokSession: &types.GrokSessionExtension{Status: "pending"},
			},
		},
		{
			Type:   types.ActionToolCall,
			Tool:   "run_terminal_command",
			Text:   "run_terminal_command",
			Output: "NAME SERVE SYNC …",
			Extensions: &types.EventExtensions{
				GrokSession: &types.GrokSessionExtension{Status: "completed"},
			},
		},
	}
	msgs := eventsToChatMessages(events)
	if len(msgs) != 1 {
		t.Fatalf("len=%d want 1: %+v", len(msgs), msgs)
	}
	if msgs[0].Text != "run_terminal_command: echo hi" {
		t.Fatalf("text=%q", msgs[0].Text)
	}
}

func TestFormatChatMessagesTextHeader(t *testing.T) {
	t.Parallel()
	loc := time.FixedZone("test", 8*3600)
	ts := time.Date(2026, 7, 31, 18, 17, 51, 0, loc)
	page := []ChatMessage{
		{Kind: MessageKindUser, Text: "hi", Timestamp: ts},
		{Kind: MessageKindResponse, Text: "yo"}, // missing → [—]
	}
	text := formatChatMessagesText(page, 40, loc)
	if !strings.HasPrefix(text, "Chat history (showing 2 of 40):\n") {
		t.Fatalf("header:\n%s", text)
	}
	if !strings.Contains(text, "[2026-07-31 18:17:51] [user] : hi\n") {
		t.Fatalf("missing user line:\n%s", text)
	}
	if !strings.Contains(text, "[—] [assistant] : yo\n") {
		t.Fatalf("missing assistant line:\n%s", text)
	}
}

func TestChatMessagesToJSONOmitsEmptyTimestamp(t *testing.T) {
	t.Parallel()
	loc := time.FixedZone("test", 8*3600)
	ts := time.Date(2026, 7, 31, 18, 17, 51, 0, loc)
	rows := chatMessagesToJSON([]ChatMessage{
		{Kind: MessageKindUser, Text: "hi", Timestamp: ts},
		{Kind: MessageKindThinking, Text: "t"},
	}, loc)
	if rows[0].Timestamp != "2026-07-31T18:17:51+08:00" {
		t.Fatalf("timestamp=%q", rows[0].Timestamp)
	}
	if rows[1].Timestamp != "" {
		t.Fatalf("want omitted empty timestamp, got %q", rows[1].Timestamp)
	}
}

func TestMessagesIntegration(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	sessionID := "019f283a-msg-test-0001"
	cwd := "/tmp/msg-test-proj"
	absCwd, err := filepath.Abs(cwd)
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(home, "sessions", url.PathEscape(absCwd), sessionID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	summary := map[string]any{
		"info":            map[string]any{"id": sessionID, "cwd": absCwd},
		"generated_title": "msg fixture",
		"created_at":      "2026-07-01T10:00:00.000Z",
		"updated_at":      "2026-07-01T11:00:00.000Z",
		"last_active_at":  "2026-07-01T11:00:00.000Z",
		"num_messages":    5,
		"num_chat_messages": 5,
	}
	b, _ := json.MarshalIndent(summary, "", "  ")
	if err := os.WriteFile(filepath.Join(dir, "summary.json"), append(b, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	updates := strings.Join([]string{
		`{"sessionUpdate":"user_message_chunk","content":{"type":"text","text":"m0"}}`,
		`{"sessionUpdate":"agent_message_chunk","content":{"type":"text","text":"a0"}}`,
		`{"sessionUpdate":"turn_completed"}`,
		`{"sessionUpdate":"user_message_chunk","content":{"type":"text","text":"m1"}}`,
		`{"sessionUpdate":"agent_thought_chunk","content":{"type":"text","text":"think1"}}`,
		`{"sessionUpdate":"agent_message_chunk","content":{"type":"text","text":"a1"}}`,
		`{"sessionUpdate":"turn_completed"}`,
		`{"sessionUpdate":"user_message_chunk","content":{"type":"text","text":"m2"}}`,
		`{"sessionUpdate":"agent_message_chunk","content":{"type":"text","text":"a2"}}`,
		`{"sessionUpdate":"turn_completed"}`,
	}, "\n") + "\n"
	if err := os.WriteFile(filepath.Join(dir, "updates.jsonl"), []byte(updates), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := Messages(home, sessionID, &MessagesOpts{Limit: 2, LimitSet: true})
	if err != nil {
		t.Fatal(err)
	}
	if res.Total < 5 {
		t.Fatalf("total=%d want >=5", res.Total)
	}
	if len(res.Messages) != 2 {
		t.Fatalf("page len=%d want 2: %+v", len(res.Messages), res.Messages)
	}
	if res.Messages[0].Text != "m2" || res.Messages[1].Text != "a2" {
		t.Fatalf("newest page: %+v", res.Messages)
	}

	res2, err := Messages(home, sessionID, &MessagesOpts{Limit: 2, LimitSet: true, OffsetFromEnd: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(res2.Messages) != 2 {
		t.Fatalf("page2 len=%d: %+v", len(res2.Messages), res2.Messages)
	}
	if res2.Messages[0].Text != "think1" || res2.Messages[1].Text != "a1" {
		t.Fatalf("second page: %+v", res2.Messages)
	}
	if !strings.Contains(res2.Text, "showing 2 of") {
		t.Fatalf("header: %s", res2.Text)
	}
}

func TestFilterMessagesByGrepAND(t *testing.T) {
	t.Parallel()
	msgs := []ChatMessage{
		{Kind: MessageKindUser, Text: "alpha only"},
		{Kind: MessageKindResponse, Text: "alpha and beta here"},
		{Kind: MessageKindTool, Text: "run_terminal_command: echo hi"},
		{Kind: MessageKindThinking, Text: "beta alone"},
	}
	got := filterMessagesByGrep(msgs, []string{"alpha", "beta"})
	if len(got) != 1 || got[0].Text != "alpha and beta here" {
		t.Fatalf("AND filter: %+v", got)
	}
	got = filterMessagesByGrep(msgs, []string{"ALPHA"}) // CI
	if len(got) != 2 {
		t.Fatalf("CI single: %+v", got)
	}
	got = filterMessagesByGrep(msgs, []string{"run_terminal", "echo"})
	if len(got) != 1 || got[0].Kind != MessageKindTool {
		t.Fatalf("tool AND: %+v", got)
	}
}

func TestValidateMessagesGrepsEmpty(t *testing.T) {
	t.Parallel()
	if err := validateMessagesGreps([]string{"ok", ""}); err == nil {
		t.Fatal("want empty pattern error")
	}
}

func TestParseMessagesArgsGrepAndColor(t *testing.T) {
	t.Parallel()
	got, err := parseMessagesArgs([]string{"sid", "--grep", "a", "--grep=b", "--color"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Greps) != 2 || got.Greps[0] != "a" || got.Greps[1] != "b" {
		t.Fatalf("greps=%v", got.Greps)
	}
	if got.ColorMode != "always" {
		t.Fatalf("color=%q", got.ColorMode)
	}
	_, err = parseMessagesArgs([]string{"--grep", ""})
	if err == nil || !strings.Contains(err.Error(), "--grep") {
		t.Fatalf("empty grep err=%v", err)
	}
	_, err = parseMessagesArgs([]string{"--color", "--no-color"})
	if err == nil || !strings.Contains(err.Error(), "cannot be specified together") {
		t.Fatalf("conflict err=%v", err)
	}
}

func TestMessagesGrepThenLimit(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	sessionID := "019f283a-msg-grep-0001"
	cwd := "/tmp/msg-grep-proj"
	absCwd, err := filepath.Abs(cwd)
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(home, "sessions", url.PathEscape(absCwd), sessionID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	summary := map[string]any{
		"info":              map[string]any{"id": sessionID, "cwd": absCwd},
		"generated_title":   "grep fixture",
		"created_at":        "2026-07-01T10:00:00.000Z",
		"updated_at":        "2026-07-01T11:00:00.000Z",
		"last_active_at":    "2026-07-01T11:00:00.000Z",
		"num_messages":      4,
		"num_chat_messages": 4,
	}
	b, _ := json.MarshalIndent(summary, "", "  ")
	if err := os.WriteFile(filepath.Join(dir, "summary.json"), append(b, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	updates := strings.Join([]string{
		`{"sessionUpdate":"user_message_chunk","content":{"type":"text","text":"hit-0"}}`,
		`{"sessionUpdate":"agent_message_chunk","content":{"type":"text","text":"miss"}}`,
		`{"sessionUpdate":"turn_completed"}`,
		`{"sessionUpdate":"user_message_chunk","content":{"type":"text","text":"hit-1"}}`,
		`{"sessionUpdate":"agent_message_chunk","content":{"type":"text","text":"hit-2"}}`,
		`{"sessionUpdate":"turn_completed"}`,
	}, "\n") + "\n"
	if err := os.WriteFile(filepath.Join(dir, "updates.jsonl"), []byte(updates), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := Messages(home, sessionID, &MessagesOpts{
		Limit: 2, LimitSet: true, Greps: []string{"hit"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Total != 3 {
		t.Fatalf("total=%d want 3 (matched)", res.Total)
	}
	if len(res.Messages) != 2 || res.Messages[0].Text != "hit-1" || res.Messages[1].Text != "hit-2" {
		t.Fatalf("newest 2 hits: %+v", res.Messages)
	}
	if strings.ContainsRune(res.Text, '\x1b') {
		t.Fatalf("Messages.Text must be plain: %q", res.Text)
	}
}

func TestWriteChatMessagesHighlights(t *testing.T) {
	t.Parallel()
	var buf strings.Builder
	err := writeChatMessages(&buf, []ChatMessage{
		{Kind: MessageKindResponse, Text: "fix the Timeout now"},
	}, 1, time.UTC, true, []string{"fix", "timeout"})
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "\x1b[1m\x1b[31m") {
		t.Fatalf("want bold-red highlight:\n%s", out)
	}
	if !strings.Contains(out, "\x1b[2m") {
		t.Fatalf("want dim timestamp:\n%s", out)
	}
}

func TestRunMessagesNoMatching(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	sessionID := "019f283a-msg-nomatch-0001"
	cwd := "/tmp/msg-nomatch-proj"
	absCwd, err := filepath.Abs(cwd)
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(home, "sessions", url.PathEscape(absCwd), sessionID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	summary := map[string]any{
		"info":            map[string]any{"id": sessionID, "cwd": absCwd},
		"generated_title": "nomatch",
		"created_at":      "2026-07-01T10:00:00.000Z",
		"updated_at":      "2026-07-01T11:00:00.000Z",
		"last_active_at":  "2026-07-01T11:00:00.000Z",
	}
	b, _ := json.MarshalIndent(summary, "", "  ")
	_ = os.WriteFile(filepath.Join(dir, "summary.json"), append(b, '\n'), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "updates.jsonl"), []byte(
		`{"sessionUpdate":"user_message_chunk","content":{"type":"text","text":"hello"}}`+"\n"+
			`{"sessionUpdate":"turn_completed"}`+"\n"), 0o644)

	var stdout strings.Builder
	err = RunMessages([]string{sessionID, "--grep", "zzznomatch"}, &stdout, io.Discard, home, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "(no matching messages)") {
		t.Fatalf("stdout=%q", stdout.String())
	}
}
