# Scenario

**Feature**: dual-signal Grok session status via filesystem + injectable procs

```
# fixture grok home + optional active_sessions.json + session dir
test harness -> writeGrokSession + writeActiveSessions
  -> inject ListProcs / Lsof (no live ps/lsof)

# primary path
sessions.Status(grokHome, id, checkPID, live)
  -> SessionStatus {State, FileActive, PIDs, PIDChecked}

# optional format
SessionStatus -> FormatStatusText | FormatStatusJSON | FormatActiveBlock
```

## Preconditions

- Package will export (RED until implementer):
  - `IsFileActive`, `LivePIDsForSession`, `Status`, `LivePID`, `LiveOptions`,
    `SessionStatus`
  - `FormatStatusText`, `FormatStatusJSON`, `FormatActiveBlock`
- `LiveOptions.ListProcs` returns `[]procresolve.Proc` (same shape as
  `pkgs/procresolve`).
- Open-file hard hits use path segment under `/.grok/sessions/…/<uuid>/…`
  (same as procresolve).
- Tests never shell out to real `ps` / `lsof` / real `~/.grok`.

## Steps

1. Root `Setup` allocates `req.TempDir` and `req.GrokHome = {temp}/.grok`.
2. Leaf `Setup` writes session fixtures, optional active list, injects procs
   and open files, sets `SessionID` / `CheckPID` / `Format` / `Op`.
3. Root `Run` maps fixtures into `LiveOptions` and calls the chosen API.
4. Leaf `Assert` checks State / PIDs / errors / formatted output.

## Context

- Canonical fixture session id: `019f283a-bbbb-7bbb-bbbb-bbbbbbbbbb01`
- Encoded cwd keys use `url.PathEscape(abs_cwd)` like other grok session trees.
- `writeActiveSessions` uses object form `{"sessions":[{sessionId,...}]}`.
- Helper `grokOpenPath(home, sessionID)` builds a realistic hard-hit path.

```go
import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/agent/grok/sessions"
)

const (
	fixtureStatusSessionID = "019f283a-bbbb-7bbb-bbbb-bbbbbbbbbb01"
	fixtureStatusCWD       = "/tmp/grok-status-fixture-project"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.TempDir = t.TempDir()
	req.GrokHome = filepath.Join(req.TempDir, ".grok")
	if err := os.MkdirAll(filepath.Join(req.GrokHome, "sessions"), 0o755); err != nil {
		return err
	}
	if req.OpenFiles == nil {
		req.OpenFiles = map[int][]string{}
	}
	return nil
}

func encodeCWD(t *testing.T, cwd string) string {
	t.Helper()
	abs, err := filepath.Abs(cwd)
	if err != nil {
		t.Fatalf("abs cwd %q: %v", cwd, err)
	}
	return url.PathEscape(abs)
}

func absPath(t *testing.T, p string) string {
	t.Helper()
	abs, err := filepath.Abs(p)
	if err != nil {
		t.Fatalf("abs %q: %v", p, err)
	}
	return abs
}

// writeStatusSession seeds summary.json under sessions/<encode(cwd)>/<id>/.
// Returns absolute session directory path.
func writeStatusSession(t *testing.T, grokHome, sessionID, cwd, title string) string {
	t.Helper()
	absCwd, err := filepath.Abs(cwd)
	if err != nil {
		t.Fatalf("abs cwd: %v", err)
	}
	dir := filepath.Join(grokHome, "sessions", url.PathEscape(absCwd), sessionID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir session: %v", err)
	}
	summary := map[string]any{
		"info": map[string]any{
			"id":  sessionID,
			"cwd": absCwd,
		},
		"generated_title": title,
		"created_at":      "2026-07-01T10:00:00.000Z",
		"updated_at":      "2026-07-01T11:00:00.000Z",
		"last_active_at":  "2026-07-01T11:00:00.000Z",
		"num_messages":    2,
		"num_chat_messages": 1,
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
			"cwd":       fixtureStatusCWD,
			"openedAt":  "2026-07-01T12:00:00Z",
		})
	}
	writeJSON(t, filepath.Join(grokHome, "active_sessions.json"), map[string]any{
		"sessions": entries,
	})
}

// grokOpenPath returns an absolute open-file path that hard-hits sessionID.
func grokOpenPath(sessionID string) string {
	// Marker /.grok/sessions/ is required by parseSessionFromPath; uuid segment
	// is a full path component.
	return "/Users/fixture/.grok/sessions/%2Ftmp%2Fproj/" + sessionID + "/events.jsonl"
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

func assertStatus(t *testing.T, resp *Response) *sessions.SessionStatus {
	t.Helper()
	if resp.Status == nil {
		t.Fatal("Status is nil")
	}
	return resp.Status
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
```
