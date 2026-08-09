# Scenario

**Feature**: list Grok sessions with injectable place / recent / active / role / forked / grep filters and KIND column

```
# synthetic GROK_HOME + summary.json (+ session_kind / parent_session_id / forked_at)
test harness -> writeListSession / writeListSessionOpts + writeActiveSessions + writeChatHistory
  -> ListWithOptions(grokHome, ListOptions{PlaceCWDs, Recent*, Active, MainAgent, SubAgent, Forked, Grep*, Limit, Now})
  -> []Session with Kind tokens (newest first after filters)

# optional format for empty-survivor phrase and KIND column
[]Session -> FormatListTable(now) -> "No sessions found" when empty
  else table header: SESSION ID  KIND  LAST ACTIVE  TITLE  MSGS  CWD
```

## Preconditions

- Package exports (role/KIND RED until implementer):
  - `ListOptions` with Limit, PlaceCWDs, Recent, RecentSet, Active, Now, Grep, GrepSet,
    **MainAgent**, **SubAgent**, **Forked**
  - `ListWithOptions(grokHome string, opts ListOptions) ([]Session, error)`
  - `Session.Kind` display token: `main` | `sub` | `sub+` | `sub-f` | `fork`
  - `FormatListTable` / `FormatListTableWithHits` KIND column after SESSION ID
- Pipeline order: place → recent → active → **role** → **forked** → grep → sort → limit.
- Place match: Abs+Clean equality; OR across PlaceCWDs; empty PlaceCWDs = no place filter.
- Sub-agent class: session_kind ∈ {subagent, subagent_resume, subagent_fork}
  OR (kind empty/absent AND parent_session_id non-empty).
- Main-agent class: complement of sub-agent (includes plain `fork`).
- Forked: session_kind ∈ {fork, subagent_fork} OR forked_at non-empty non-whitespace.
- MainAgent && SubAgent → error; Forked ANDs with role.
- Active uses existing `IsFileActive` / `active_sessions.json` object form.
- Grep reuses case-insensitive literal search over summary + chat_history.
- Tests never call `os.Chdir`, `t.Chdir`, `os.Setenv`, `t.Setenv`, or reassign stdout.
- Tests never read real `~/.grok`.

## Steps

1. Root `Setup` allocates `req.TempDir` / `req.GrokHome = {temp}/.grok` and fixed `Now`.
2. Leaf Setup writes session fixtures (optional kind/parent/forked_at), optional active list / chat history, and filter fields on `Request`.
3. Root `Run` maps Request → `ListOptions` and calls `ListWithOptions`; optional FormatListTable / FormatListTableWithHits.
4. Leaf `Assert` checks session ids / Kind / order / emptiness / errors / table text.

## Context

- Fixed clock: **2026-07-03 15:00:00 UTC** (`fixedNow`)
- Canonical place cwds:
  - `cwdA` = `/tmp/list-filter-proj-a`
  - `cwdB` = `/tmp/list-filter-proj-b`
  - `cwdC` = `/tmp/list-filter-proj-c`
- Session id pattern: `019f283a-aaaa-7aaa-aaaa-aaaaaaaaaaNN` (plus kind-specific ids)
- `last_active_at` strings use RFC3339 with `.000Z` for stable parse
- Empty-list format phrase: `No sessions found` (same as `FormatListTable`)
- summary.json role fields (top-level): `session_kind`, `parent_session_id`, `forked_at`

```go
import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/agent/grok/sessions"
)

// fixedNow is the injectable clock for recent windows and FormatListTable.
var fixedNow = time.Date(2026, 7, 3, 15, 0, 0, 0, time.UTC)

const (
	cwdA = "/tmp/list-filter-proj-a"
	cwdB = "/tmp/list-filter-proj-b"
	cwdC = "/tmp/list-filter-proj-c"

	idA1 = "019f283a-aaaa-7aaa-aaaa-aaaaaaaaaa01"
	idA2 = "019f283a-aaaa-7aaa-aaaa-aaaaaaaaaa02"
	idA3 = "019f283a-aaaa-7aaa-aaaa-aaaaaaaaaa03"
	idB1 = "019f283a-bbbb-7bbb-bbbb-bbbbbbbbbb01"
	idB2 = "019f283a-bbbb-7bbb-bbbb-bbbbbbbbbb02"
	idC1 = "019f283a-cccc-7ccc-cccc-cccccccccc01"
	idD1 = "019f283a-dddd-7ddd-dddd-dddddddddd01"
	idE1 = "019f283a-eeee-7eee-eeee-eeeeeeeeee01"

	// Role / kind fixture ids (stable, unique across leaves).
	idMain     = "019f283a-1111-7111-1111-111111111101" // plain main (no kind)
	idFork     = "019f283a-2222-7222-2222-222222222201" // session_kind=fork
	idSub      = "019f283a-3333-7333-3333-333333333301" // session_kind=subagent
	idSubRes   = "019f283a-3333-7333-3333-333333333302" // session_kind=subagent_resume
	idSubFork  = "019f283a-3333-7333-3333-333333333303" // session_kind=subagent_fork
	idEmptyPar = "019f283a-4444-7444-4444-444444444401" // empty kind + parent_session_id
	idEmptyNo  = "019f283a-4444-7444-4444-444444444402" // empty kind, no parent
	idForkedAt = "019f283a-5555-7555-5555-555555555501" // forked_at set, kind empty
	idParent   = "019f283a-0000-7000-0000-000000000001" // synthetic parent id string only
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.TempDir = t.TempDir()
	req.GrokHome = filepath.Join(req.TempDir, ".grok")
	if err := os.MkdirAll(filepath.Join(req.GrokHome, "sessions"), 0o755); err != nil {
		return err
	}
	req.Now = fixedNow
	return nil
}

func absPath(t *testing.T, p string) string {
	t.Helper()
	abs, err := filepath.Abs(p)
	if err != nil {
		t.Fatalf("abs %q: %v", p, err)
	}
	return filepath.Clean(abs)
}

func encodeCWD(t *testing.T, cwd string) string {
	t.Helper()
	return url.PathEscape(absPath(t, cwd))
}

// atFixed returns fixedNow + d as RFC3339Nano UTC with millisecond zeros.
func atFixed(d time.Duration) string {
	return fixedNow.Add(d).UTC().Format("2006-01-02T15:04:05.000Z")
}

type listSessionOpts struct {
	NumChatMessages int
	// CWDEmpty forces info.cwd to "" (still stored under encode of storageCWD).
	CWDEmpty   bool
	StorageCWD string // when CWDEmpty, path key for on-disk layout

	// Role / fork summary fields (top-level summary.json keys).
	SessionKind     string // session_kind
	ParentSessionID string // parent_session_id
	ForkedAt        string // forked_at
	// OmitKindField forces session_kind absent even when SessionKind=="".
	// Default write omits empty kind/parent/forked_at keys.
	ForceWriteKind     bool
	ForceWriteParent   bool
	ForceWriteForkedAt bool
}

// writeListSession seeds summary.json under sessions/<encode(cwd)>/<id>/.
// Returns absolute session directory path.
func writeListSession(t *testing.T, grokHome, id, lastActiveAt, cwd, title string) string {
	t.Helper()
	return writeListSessionOpts(t, grokHome, id, lastActiveAt, cwd, title, listSessionOpts{})
}

func writeListSessionOpts(t *testing.T, grokHome, id, lastActiveAt, cwd, title string, opts listSessionOpts) string {
	t.Helper()
	storageCWD := cwd
	if opts.CWDEmpty {
		storageCWD = opts.StorageCWD
		if storageCWD == "" {
			storageCWD = "/tmp/list-filter-empty-cwd-storage"
		}
	}
	absCwd, err := filepath.Abs(storageCWD)
	if err != nil {
		t.Fatalf("abs cwd: %v", err)
	}
	dir := filepath.Join(grokHome, "sessions", url.PathEscape(absCwd), id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir session: %v", err)
	}

	infoCWD := absCwd
	if opts.CWDEmpty {
		infoCWD = ""
	}

	summary := map[string]any{
		"info": map[string]any{
			"id":  id,
			"cwd": infoCWD,
		},
		"generated_title":   title,
		"created_at":        lastActiveAt,
		"updated_at":        lastActiveAt,
		"last_active_at":    lastActiveAt,
		"num_messages":      1,
		"num_chat_messages": opts.NumChatMessages,
	}
	if opts.SessionKind != "" || opts.ForceWriteKind {
		summary["session_kind"] = opts.SessionKind
	}
	if opts.ParentSessionID != "" || opts.ForceWriteParent {
		summary["parent_session_id"] = opts.ParentSessionID
	}
	if opts.ForkedAt != "" || opts.ForceWriteForkedAt {
		summary["forked_at"] = opts.ForkedAt
	}
	writeJSON(t, filepath.Join(dir, "summary.json"), summary)
	return dir
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

// writeActiveSessions writes object-form active_sessions.json listing the given ids.
func writeActiveSessions(t *testing.T, grokHome string, sessionIDs ...string) {
	t.Helper()
	entries := make([]map[string]any, 0, len(sessionIDs))
	for _, id := range sessionIDs {
		entries = append(entries, map[string]any{
			"sessionId": id,
			"cwd":       cwdA,
			"openedAt":  "2026-07-03T12:00:00Z",
		})
	}
	writeJSON(t, filepath.Join(grokHome, "active_sessions.json"), map[string]any{
		"sessions": entries,
	})
}

// writeChatUser writes a single user line to chat_history.jsonl next to summary.
func writeChatUser(t *testing.T, sessionDir, text string) {
	t.Helper()
	line, err := json.Marshal(map[string]any{
		"type": "user",
		"content": []map[string]any{
			{"type": "text", "text": text},
		},
	})
	if err != nil {
		t.Fatalf("marshal chat: %v", err)
	}
	path := filepath.Join(sessionDir, "chat_history.jsonl")
	if err := os.WriteFile(path, append(line, '\n'), 0o644); err != nil {
		t.Fatalf("write chat: %v", err)
	}
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

func assertSessionIDs(t *testing.T, got []sessions.Session, want ...string) {
	t.Helper()
	if len(got) != len(want) {
		ids := make([]string, len(got))
		for i, s := range got {
			ids[i] = s.ID
		}
		t.Fatalf("len(sessions) = %d ids=%v, want %d ids=%v", len(got), ids, len(want), want)
	}
	for i, id := range want {
		if got[i].ID != id {
			t.Fatalf("sessions[%d].ID = %q, want %q (full=%v)", i, got[i].ID, id, sessionIDs(got))
		}
	}
}

func sessionIDs(list []sessions.Session) []string {
	out := make([]string, len(list))
	for i, s := range list {
		out[i] = s.ID
	}
	return out
}

// sessionKind reads Session.Kind when the field exists; otherwise "".
// Classic TDD: RED until implementer adds Session.Kind and populates it.
func sessionKind(s sessions.Session) string {
	v := reflect.ValueOf(s).FieldByName("Kind")
	if v.IsValid() && v.Kind() == reflect.String {
		return v.String()
	}
	return ""
}

// assertSessionKinds checks Session.Kind for each session in order.
func assertSessionKinds(t *testing.T, got []sessions.Session, wantKinds ...string) {
	t.Helper()
	if len(got) != len(wantKinds) {
		t.Fatalf("len(sessions)=%d, want %d kinds=%v ids=%v", len(got), len(wantKinds), wantKinds, sessionIDs(got))
	}
	for i, k := range wantKinds {
		gotK := sessionKind(got[i])
		if gotK != k {
			t.Fatalf("sessions[%d] id=%s Kind=%q, want %q", i, got[i].ID, gotK, k)
		}
	}
}

// assertIDsAndKinds checks id and Kind pairs in order: id0, kind0, id1, kind1, ...
func assertIDsAndKinds(t *testing.T, got []sessions.Session, idKindPairs ...string) {
	t.Helper()
	if len(idKindPairs)%2 != 0 {
		t.Fatalf("idKindPairs must be even length, got %d", len(idKindPairs))
	}
	n := len(idKindPairs) / 2
	if len(got) != n {
		t.Fatalf("len(sessions)=%d ids=%v, want %d pairs", len(got), sessionIDs(got), n)
	}
	for i := 0; i < n; i++ {
		wantID := idKindPairs[i*2]
		wantKind := idKindPairs[i*2+1]
		if got[i].ID != wantID {
			t.Fatalf("sessions[%d].ID=%q, want %q (full=%v)", i, got[i].ID, wantID, sessionIDs(got))
		}
		gotK := sessionKind(got[i])
		if gotK != wantKind {
			t.Fatalf("sessions[%d] id=%s Kind=%q, want %q", i, got[i].ID, gotK, wantKind)
		}
	}
}

func assertEmptyList(t *testing.T, resp *Response) {
	t.Helper()
	assertNoError(t, resp)
	if len(resp.Sessions) != 0 {
		t.Fatalf("sessions = %v, want empty", sessionIDs(resp.Sessions))
	}
}

func assertNoSessionsFoundOutput(t *testing.T, resp *Response) {
	t.Helper()
	if strings.TrimSpace(resp.Output) != "No sessions found" {
		t.Fatalf("output = %q, want %q", resp.Output, "No sessions found")
	}
}

// assertHeaderKINDColumn checks first table line has SESSION ID, then KIND, then LAST ACTIVE.
func assertHeaderKINDColumn(t *testing.T, output string) {
	t.Helper()
	line := output
	if i := strings.IndexByte(output, '\n'); i >= 0 {
		line = output[:i]
	}
	sid := strings.Index(line, "SESSION ID")
	kind := strings.Index(line, "KIND")
	last := strings.Index(line, "LAST ACTIVE")
	if sid < 0 || kind < 0 || last < 0 {
		t.Fatalf("header missing columns: %q", line)
	}
	if !(sid < kind && kind < last) {
		t.Fatalf("header column order wrong (want SESSION ID then KIND then LAST ACTIVE): %q", line)
	}
}

// keep import used when types resolve
var _ = sessions.Session{}
```
