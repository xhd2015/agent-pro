# Scenario

**Feature**: pin/list/show/remove Grok session bookmarks via injectable homes

```
# fixture agent-pro home + grok home (never real ~/.grok or AGENT_PRO_HOME)
writeBookmarkSession(summary) under grokHome
  optional writeStoreJSON / corrupt marker under agentProHome
  -> BookmarkGrok | ListBookmarks | GetBookmark | RemoveBookmark | Format*
  -> Bookmark / BookmarkView / Output / store file
```

## Preconditions

- Package will export (RED until implementer): `Bookmark`, `BookmarkView`,
  `PinOptions`, `ListFilter`, `BookmarkGrok`, `ListBookmarks`, `GetBookmark`,
  `RemoveBookmark`, `FormatBookmarksTable`, `FormatBookmarkShow`,
  `FormatBookmarkJSON`.
- Catalog path: `{agentProHome}/session_bookmarks.json`.
- Session layout matches existing Find:
  `{grokHome}/sessions/{url.PathEscape(abs_cwd)}/{uuid}/summary.json`.
- Title resolution matches List: `generated_title` then `session_summary`.
- Tests never read real `~/.grok` or real `AGENT_PRO_HOME`.
- Parallel-safe: no `t.Setenv` / `os.Setenv` / `Chdir`; homes only via `req`.

## Steps

1. Root `Setup` allocates `req.TempDir`, `req.AgentProHome`, `req.GrokHome`.
2. Leaf seeds session dir and/or store JSON; sets `Op`, pin opts, filters.
3. Root `Run` dispatches package API by `req.Op`.
4. Leaf `Assert` checks structured result, store file, warnings, or errors.

## Context

- Canonical fixture session id: `019f283a-bbbb-7bbb-bbbb-bbbbbbbbbb01`
- Secondary id (filters / multi): `019f283a-cccc-7ccc-cccc-cccccccccc02`
- Colliding id for multi-runner: `019f283a-dddd-7ddd-dddd-dddddddddd03`
- Fixture cwd: `/tmp/grok-bookmark-fixture-project`
- Store version: `1`

```go
import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/agent/grok/sessions"
)

const (
	fixtureBookmarkSessionID  = "019f283a-bbbb-7bbb-bbbb-bbbbbbbbbb01"
	fixtureBookmarkSessionID2 = "019f283a-cccc-7ccc-cccc-cccccccccc02"
	fixtureBookmarkCollideID  = "019f283a-dddd-7ddd-dddd-dddddddddd03"
	fixtureBookmarkCWD        = "/tmp/grok-bookmark-fixture-project"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.TempDir = t.TempDir()
	req.AgentProHome = filepath.Join(req.TempDir, "agent-pro-home")
	req.GrokHome = filepath.Join(req.TempDir, ".grok")
	if err := os.MkdirAll(req.AgentProHome, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(req.GrokHome, "sessions"), 0o755); err != nil {
		return err
	}
	if req.FixtureCWD == "" {
		req.FixtureCWD = fixtureBookmarkCWD
	}
	if req.Title == "" {
		req.Title = "bookmark fixture title"
	}
	if req.NumChatMessages == 0 {
		req.NumChatMessages = 42
	}
	return nil
}

func storePath(agentProHome string) string {
	return filepath.Join(agentProHome, "session_bookmarks.json")
}

func absPath(t *testing.T, p string) string {
	t.Helper()
	abs, err := filepath.Abs(p)
	if err != nil {
		t.Fatalf("abs %q: %v", p, err)
	}
	return abs
}

func encodeCWD(t *testing.T, cwd string) string {
	t.Helper()
	return url.PathEscape(absPath(t, cwd))
}

func writeJSON(t *testing.T, path string, v any) {
	t.Helper()
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatalf("marshal %s: %v", path, err)
	}
	b = append(b, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir parent %s: %v", path, err)
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func mustWriteFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir parent %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// writeBookmarkSession seeds summary.json under sessions/<encode(cwd)>/<id>/.
// Returns absolute session directory.
func writeBookmarkSession(t *testing.T, grokHome, sessionID, cwd, title string, numChat int) string {
	t.Helper()
	absCwd := absPath(t, cwd)
	dir := filepath.Join(grokHome, "sessions", url.PathEscape(absCwd), sessionID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir session: %v", err)
	}
	writeSessionSummary(t, dir, sessionID, absCwd, title, numChat)
	return dir
}

// writeSessionSummary rewrites summary.json under an existing session directory
// (used by enrich leaves to mutate live title/msgs after store preseed).
func writeSessionSummary(t *testing.T, sessionDir, sessionID, absCwd, title string, numChat int) {
	t.Helper()
	if absCwd == "" {
		absCwd = absPath(t, fixtureBookmarkCWD)
	}
	summary := map[string]any{
		"info": map[string]any{
			"id":  sessionID,
			"cwd": absCwd,
		},
		"generated_title":   title,
		"created_at":        "2026-07-01T10:00:00.000Z",
		"updated_at":        "2026-07-01T11:00:00.000Z",
		"last_active_at":    "2026-07-01T11:00:00.000Z",
		"num_messages":      numChat + 2,
		"num_chat_messages": numChat,
	}
	writeJSON(t, filepath.Join(sessionDir, "summary.json"), summary)
}

// bookmarkEntry is a store row for preseed helpers (JSON-shaped).
type bookmarkEntry struct {
	AgentRunner     string   `json:"agent_runner"`
	SessionID       string   `json:"session_id"`
	SessionDir      string   `json:"session_dir"`
	Title           string   `json:"title"`
	NumChatMessages int      `json:"num_chat_messages"`
	Tags            []string `json:"tags"`
	Description     string   `json:"description"`
	CreatedAt       string   `json:"created_at"`
	UpdatedAt       string   `json:"updated_at"`
}

func writeStore(t *testing.T, agentProHome string, entries []bookmarkEntry) string {
	t.Helper()
	path := storePath(agentProHome)
	doc := map[string]any{
		"version":    1,
		"bookmarks":  entries,
	}
	writeJSON(t, path, doc)
	return path
}

func readStoreRaw(t *testing.T, agentProHome string) string {
	t.Helper()
	b, err := os.ReadFile(storePath(agentProHome))
	if err != nil {
		t.Fatalf("read store: %v", err)
	}
	return string(b)
}

func readStoreEntries(t *testing.T, agentProHome string) []bookmarkEntry {
	t.Helper()
	b, err := os.ReadFile(storePath(agentProHome))
	if err != nil {
		t.Fatalf("read store: %v", err)
	}
	var doc struct {
		Version    int             `json:"version"`
		Bookmarks  []bookmarkEntry `json:"bookmarks"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatalf("unmarshal store: %v\nraw=%s", err, string(b))
	}
	return doc.Bookmarks
}

func findEntry(entries []bookmarkEntry, runner, sessionID string) (bookmarkEntry, bool) {
	for _, e := range entries {
		if e.AgentRunner == runner && e.SessionID == sessionID {
			return e, true
		}
	}
	return bookmarkEntry{}, false
}

func defaultPreseedTimes() (created, updated string) {
	return "2026-07-30T08:12:00Z", "2026-07-30T08:12:00Z"
}

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

func assertEqualString(t *testing.T, field, got, want string) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: got %q, want %q", field, got, want)
	}
}

func assertEqualBool(t *testing.T, field string, got, want bool) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: got %v, want %v", field, got, want)
	}
}

func assertEqualInt(t *testing.T, field string, got, want int) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: got %d, want %d", field, got, want)
	}
}

func assertContains(t *testing.T, got, substr string) {
	t.Helper()
	if !strings.Contains(got, substr) {
		t.Fatalf("output missing %q:\n%s", substr, got)
	}
}

func assertNoANSI(t *testing.T, s string) {
	t.Helper()
	if strings.Contains(s, "\x1b[") {
		t.Fatalf("output has ANSI escapes:\n%s", s)
	}
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	st, err := os.Stat(path)
	if err != nil || st.IsDir() {
		t.Fatalf("expected file %q: %v", path, err)
	}
}

func assertPathMissing(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("expected path missing: %q", path)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat %q: %v", path, err)
	}
}

func assertTagsEqual(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("tags len=%d got=%v want=%v", len(got), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("tags[%d]=%q want %q (full got=%v)", i, got[i], want[i], got)
		}
	}
}

func tagsSortedUnique(tags []string) bool {
	if len(tags) == 0 {
		return true
	}
	seen := map[string]bool{}
	prev := tags[0]
	if strings.TrimSpace(prev) != prev || prev == "" {
		return false
	}
	seen[prev] = true
	for i := 1; i < len(tags); i++ {
		if tags[i] == "" || strings.TrimSpace(tags[i]) != tags[i] {
			return false
		}
		if tags[i] < prev {
			return false
		}
		if seen[tags[i]] {
			return false
		}
		seen[tags[i]] = true
		prev = tags[i]
	}
	return true
}

func viewBySession(views []sessions.BookmarkView, sessionID string) (sessions.BookmarkView, bool) {
	for _, v := range views {
		if v.SessionID == sessionID {
			return v, true
		}
	}
	return sessions.BookmarkView{}, false
}

func warningHasSession(warnings []string, sessionID string) bool {
	for _, w := range warnings {
		if strings.Contains(w, sessionID) && strings.Contains(strings.ToLower(w), "not found") {
			return true
		}
	}
	return false
}

// silence unused when leaves do not need every helper in a package shard
var (
	_ = time.Time{}
	_ = strings.Contains
)
```
