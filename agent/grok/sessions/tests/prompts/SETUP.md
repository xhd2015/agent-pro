# Scenario

**Feature**: user-prompt history for Grok sessions via injectable package API

```
# fixture grok home under t.TempDir (never real ~/.grok)
writePromptSession(summary + updates.jsonl with wire timestamps)
  -> Prompts | ListPrompts | ParseRecentWindow
  -> optional FormatPromptsText | FormatPromptsListText

# fixed clock for relative windows and format
req.Now = 2026-07-03T15:00:00Z ; Location = time.UTC
```

## Preconditions

- Package will export (RED until implementer):
  - `UserPrompt`, `SessionPrompts`, `ListPromptsOptions`, `FormatPromptsOptions`
  - `ParseRecentWindow`, `Prompts`, `ListPrompts`
  - `FormatPromptsText`, `FormatPromptsListText`
- Session layout matches existing Find rules:
  `{grokHome}/sessions/{url.PathEscape(abs_cwd)}/{uuid}/`
- Primary prompt source is `updates.jsonl` `user_message_chunk` lines with
  wire timestamps (top-level `timestamp` ms, or `_meta` / `agentTimestampMs`
  when implementer supports them). Coalesced user message uses **first** chunk
  timestamp.
- Tests never call `t.Setenv`, `os.Chdir`, or hijack `os.Stdout`.
- Tests never read real `~/.grok`.

## Steps

1. Root `Setup` allocates temp `.grok` home and injects fixed `Now` + UTC location.
2. Leaf Setup writes session fixtures and sets `Op` / selection flags.
3. Root `Run` calls the intended package API for `req.Op`.
4. Leaf `Assert` checks structured results and/or formatted text.

## Context

- Fixed clock: **2026-07-03 15:00:00 UTC** (`fixedNow`)
- Default fixture cwd: `/tmp/grok-prompts-fixture-project`
- Session id patterns use `019f283a-dddd-7ddd-dddd-ddddddddddNN`
- Multi fixtures use distinct `last_active_at` so newest-first order is stable
- Soft-truncate body: **~200 runes** + Unicode ellipsis `…` (U+2026)
- Compact timestamp layout: `2006-01-02 15:04:05` in `FormatPromptsOptions.Location`
- Empty friendly phrase: `No user prompts found`
- Missing timestamp marker: `[—]` (em dash U+2014)

```go
import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/xhd2015/agent-pro/agent/grok/sessions"
)

// fixedNow is the injectable clock for all window / relative format leaves.
var fixedNow = time.Date(2026, 7, 3, 15, 0, 0, 0, time.UTC)

const (
	fixturePromptsCWD = "/tmp/grok-prompts-fixture-project"

	// Canonical single-session ids
	idKnownTwo       = "019f283a-dddd-7ddd-dddd-dddddddddd01"
	idCoalesce       = "019f283a-dddd-7ddd-dddd-dddddddddd02"
	idAssistantOnly  = "019f283a-dddd-7ddd-dddd-dddddddddd03"
	idWhitespace     = "019f283a-dddd-7ddd-dddd-dddddddddd04"
	idMissingUpdates = "019f283a-dddd-7ddd-dddd-dddddddddd05"
	idUnknown        = "019f283a-eeee-7eee-eeee-eeeeeeeeee99"
	idFormatSingle   = "019f283a-dddd-7ddd-dddd-dddddddddd10"
	idFormatMultiA   = "019f283a-dddd-7ddd-dddd-dddddddddd11"
	idFormatMultiB   = "019f283a-dddd-7ddd-dddd-dddddddddd12"
	idFullHistory    = "019f283a-dddd-7ddd-dddd-dddddddddd20"
	idLastActiveOnly = "019f283a-dddd-7ddd-dddd-dddddddddd30"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.TempDir = t.TempDir()
	req.GrokHome = filepath.Join(req.TempDir, ".grok")
	if err := os.MkdirAll(filepath.Join(req.GrokHome, "sessions"), 0o755); err != nil {
		return err
	}
	if req.Now.IsZero() {
		req.Now = fixedNow
	}
	if req.Location == nil {
		req.Location = time.UTC
	}
	if req.Home == "" {
		req.Home = req.GrokHome
	}
	return nil
}

// --- time helpers ---

func atFixed(offset time.Duration) time.Time {
	return fixedNow.Add(offset)
}

func ms(t time.Time) int64 {
	return t.UnixMilli()
}

// --- wire line builders ---

// userChunkAt builds a flat user_message_chunk with top-level timestamp ms.
func userChunkAt(text string, ts time.Time) string {
	line, _ := json.Marshal(map[string]any{
		"sessionUpdate": "user_message_chunk",
		"content":       map[string]any{"type": "text", "text": text},
		"timestamp":     ts.UnixMilli(),
	})
	return string(line)
}

// userChunkNoTS builds a user chunk without any timestamp field.
func userChunkNoTS(text string) string {
	line, _ := json.Marshal(map[string]any{
		"sessionUpdate": "user_message_chunk",
		"content":       map[string]any{"type": "text", "text": text},
	})
	return string(line)
}

func assistantChunk(text string) string {
	line, _ := json.Marshal(map[string]any{
		"sessionUpdate": "agent_message_chunk",
		"content":       map[string]any{"type": "text", "text": text},
	})
	return string(line)
}

func toolCallPending(id, title string) string {
	line, _ := json.Marshal(map[string]any{
		"sessionUpdate": "tool_call",
		"toolCallId":    id,
		"kind":          "execute",
		"title":         title,
		"status":        "pending",
	})
	return string(line)
}

func turnCompleted() string {
	return `{"sessionUpdate":"turn_completed"}`
}

// updatesJSONL joins wire lines with trailing newline after each.
func updatesJSONL(lines ...string) string {
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

// --- session fixtures ---

type promptSessionOpts struct {
	ID           string
	CWD          string
	Title        string
	LastActiveAt time.Time
	Updates      string // full updates.jsonl body; empty means omit file when OmitUpdates
	OmitUpdates  bool
}

// writePromptSession seeds summary.json (+ optional updates.jsonl).
// Returns absolute session directory.
func writePromptSession(t *testing.T, grokHome string, o promptSessionOpts) string {
	t.Helper()
	if o.CWD == "" {
		o.CWD = fixturePromptsCWD
	}
	if o.Title == "" {
		o.Title = "prompt fixture " + o.ID[len(o.ID)-2:]
	}
	if o.LastActiveAt.IsZero() {
		o.LastActiveAt = fixedNow
	}
	absCwd, err := filepath.Abs(o.CWD)
	if err != nil {
		t.Fatalf("abs cwd: %v", err)
	}
	dir := filepath.Join(grokHome, "sessions", url.PathEscape(absCwd), o.ID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir session: %v", err)
	}
	// Use RFC3339Milli-like strings for last_active_at.
	lastActive := o.LastActiveAt.UTC().Format("2006-01-02T15:04:05.000Z")
	summary := map[string]any{
		"info": map[string]any{
			"id":  o.ID,
			"cwd": absCwd,
		},
		"generated_title":   o.Title,
		"created_at":        "2026-07-01T10:00:00.000Z",
		"updated_at":        lastActive,
		"last_active_at":    lastActive,
		"num_messages":      2,
		"num_chat_messages": 1,
	}
	writeJSON(t, filepath.Join(dir, "summary.json"), summary)
	if !o.OmitUpdates {
		body := o.Updates
		if body == "" {
			body = updatesJSONL(userChunkAt("default prompt", atFixed(-time.Hour)))
		}
		if err := os.WriteFile(filepath.Join(dir, "updates.jsonl"), []byte(body), 0o644); err != nil {
			t.Fatalf("write updates: %v", err)
		}
	}
	return dir
}

// writeNSessions writes n sessions with last_active_at = fixedNow - i*minute
// (i=0 newest). Each has one in-window user prompt at last_active time.
// IDs are idPrefix + zero-padded index (must yield unique full ids).
func writeNSessions(t *testing.T, grokHome string, n int, idBase string, promptOffset time.Duration) []string {
	t.Helper()
	ids := make([]string, 0, n)
	for i := 0; i < n; i++ {
		// idBase is a 36-char uuid-like prefix; we rewrite last 2 hex digits.
		id := fmt.Sprintf("%s%02d", idBase, i)
		if len(idBase) >= 34 {
			id = idBase[:34] + fmt.Sprintf("%02d", i)
		}
		la := atFixed(-time.Duration(i) * time.Minute)
		pt := la.Add(promptOffset)
		writePromptSession(t, grokHome, promptSessionOpts{
			ID:           id,
			Title:        fmt.Sprintf("session-%02d", i),
			LastActiveAt: la,
			Updates:      updatesJSONL(userChunkAt(fmt.Sprintf("prompt-%02d", i), pt)),
		})
		ids = append(ids, id)
	}
	return ids
}

func writeJSON(t *testing.T, path string, v any) {
	t.Helper()
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatalf("marshal %s: %v", path, err)
	}
	b = append(b, '\n')
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// multiSessionID returns a stable uuid-like id for multi fixtures (index 0..99).
func multiSessionID(i int) string {
	return fmt.Sprintf("019f283a-aaaa-7aaa-aaaa-aaaaaaaaaa%02d", i)
}

// --- assertion helpers ---

func assertNoHarnessErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Run harness error: %v", err)
	}
}

func assertNoError(t *testing.T, resp *Response) {
	t.Helper()
	if resp.Err != nil {
		t.Fatalf("unexpected error: %v", resp.Err)
	}
}

func assertError(t *testing.T, resp *Response) {
	t.Helper()
	if resp.Err == nil {
		t.Fatal("expected error, got nil")
	}
}

func assertErrorContains(t *testing.T, resp *Response, substrs ...string) {
	t.Helper()
	assertError(t, resp)
	msg := resp.Err.Error()
	for _, s := range substrs {
		if !strings.Contains(msg, s) {
			t.Fatalf("error %q missing %q", msg, s)
		}
	}
}

func assertContains(t *testing.T, got, substr string) {
	t.Helper()
	if !strings.Contains(got, substr) {
		t.Fatalf("output missing %q:\n%s", substr, got)
	}
}

func assertNotContains(t *testing.T, got, substr string) {
	t.Helper()
	if strings.Contains(got, substr) {
		t.Fatalf("output unexpectedly contains %q:\n%s", substr, got)
	}
}

func assertTrailingNewline(t *testing.T, s string) {
	t.Helper()
	if s == "" || !strings.HasSuffix(s, "\n") {
		t.Fatalf("expected trailing newline, got %q", s)
	}
}

func assertPromptCount(t *testing.T, sp *sessions.SessionPrompts, want int) {
	t.Helper()
	if sp == nil {
		t.Fatal("SessionPrompts is nil")
	}
	if got := len(sp.UserPrompts); got != want {
		t.Fatalf("UserPrompts len=%d want %d: %+v", got, want, sp.UserPrompts)
	}
}

func assertPromptText(t *testing.T, sp *sessions.SessionPrompts, idx int, want string) {
	t.Helper()
	if sp == nil || idx < 0 || idx >= len(sp.UserPrompts) {
		t.Fatalf("prompt index %d out of range (sp=%v)", idx, sp)
	}
	if got := sp.UserPrompts[idx].Text; got != want {
		t.Fatalf("prompt[%d].Text=%q want %q", idx, got, want)
	}
}

func assertPromptTimeUTC(t *testing.T, sp *sessions.SessionPrompts, idx int, want time.Time) {
	t.Helper()
	if sp == nil || idx < 0 || idx >= len(sp.UserPrompts) {
		t.Fatalf("prompt index %d out of range", idx)
	}
	got := sp.UserPrompts[idx].Timestamp.UTC()
	// Allow second-level equality (wire may be ms-precise).
	if !got.Equal(want.UTC()) && got.Unix() != want.Unix() {
		// Prefer exact ms match when both non-zero.
		if got.UnixMilli() != want.UnixMilli() {
			t.Fatalf("prompt[%d].Timestamp=%v (ms=%d) want %v (ms=%d)",
				idx, got, got.UnixMilli(), want.UTC(), want.UnixMilli())
		}
	}
}

func assertListLen(t *testing.T, list []sessions.SessionPrompts, want int) {
	t.Helper()
	if got := len(list); got != want {
		ids := make([]string, 0, len(list))
		for _, sp := range list {
			ids = append(ids, sp.ID)
		}
		t.Fatalf("list len=%d want %d ids=%v", got, want, ids)
	}
}

func assertListIDsNewestFirst(t *testing.T, list []sessions.SessionPrompts, wantIDs []string) {
	t.Helper()
	assertListLen(t, list, len(wantIDs))
	for i, want := range wantIDs {
		if list[i].ID != want {
			t.Fatalf("list[%d].ID=%q want %q", i, list[i].ID, want)
		}
	}
}

func assertSessionOrderByLastActive(t *testing.T, list []sessions.SessionPrompts) {
	t.Helper()
	for i := 1; i < len(list); i++ {
		if list[i].LastActiveAt.After(list[i-1].LastActiveAt) {
			t.Fatalf("list not newest-first: [%d]=%v after [%d]=%v",
				i, list[i].LastActiveAt, i-1, list[i-1].LastActiveAt)
		}
	}
}

// longPromptRunes returns a string of n 'x' runes (ASCII).
func longPromptRunes(n int) string {
	return strings.Repeat("x", n)
}

// runeLen is utf8.RuneCountInString for truncate asserts.
func runeLen(s string) int {
	return utf8.RuneCountInString(s)
}

// compactLinePrefix builds expected `[YYYY-MM-DD HH:MM:SS]` for ts in loc.
func compactLinePrefix(ts time.Time, loc *time.Location) string {
	if loc == nil {
		loc = time.UTC
	}
	if ts.IsZero() {
		return "[—]"
	}
	return "[" + ts.In(loc).Format("2006-01-02 15:04:05") + "]"
}
```
