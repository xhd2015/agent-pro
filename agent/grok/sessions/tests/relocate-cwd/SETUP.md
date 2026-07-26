# Scenario

**Feature**: relocate a Grok session's cwd via filesystem-only fixtures

```
t.TempDir/.grok + sessions/<encode(oldCWD)>/<id>/{summary,prompt_context,updates}
  + session_search.sqlite + optional active_sessions.json
  -> sessions.RelocateCWD(sessionID, targetDir, &RelocateCWDOptions{GrokHome})
```

## Preconditions

- Package `github.com/xhd2015/agent-pro/agent/grok/sessions` will export
  `RelocateCWD`, `RelocateCWDOptions`, and `RelocateCWDResult` (RED until then).
- Tests never read real `~/.grok`.
- Encoding matches Grok: `url.PathEscape(filepath.Abs(cwd))` (`/` → `%2F`).
- Active detection uses `active_sessions.json` as documented in root DOCTEST DSN.

## Steps

1. Root `Setup` allocates `req.TempDir` and `req.GrokHome = {temp}/.grok`.
2. Leaf `Setup` creates workspace dirs, seeds session fixtures, sets SessionID /
   TargetDir / markers.
3. `Run` calls `sessions.RelocateCWD`.
4. Leaf `Assert` checks result, filesystem layout, and errors.

## Context

- Helpers write realistic `summary.json` / `prompt_context.json` / `updates.jsonl`.
- `writeActiveSessions` seeds object-form `active_sessions.json`.
- `encodeCWD` is the single encoding convention for session parent keys.
- SQLite fixture is a plain file with known marker bytes (not a real DB).

```go
import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.TempDir = t.TempDir()
	req.GrokHome = filepath.Join(req.TempDir, ".grok")
	if err := os.MkdirAll(filepath.Join(req.GrokHome, "sessions"), 0o755); err != nil {
		return err
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

// writeRelocateSession seeds a session under sessions/<encode(oldCWD)>/<id>/.
// Returns absolute session directory path.
func writeRelocateSession(t *testing.T, grokHome, sessionID, oldCWD string, opts relocateSessionOpts) string {
	t.Helper()
	absOld := absPath(t, oldCWD)
	encoded := encodeCWD(t, absOld)
	dir := filepath.Join(grokHome, "sessions", encoded, sessionID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir session: %v", err)
	}

	summary := map[string]any{
		"info": map[string]any{
			"id":  sessionID,
			"cwd": absOld,
		},
		"generated_title": opts.Title,
		"created_at":      "2026-07-01T10:00:00.000Z",
		"updated_at":      "2026-07-01T11:00:00.000Z",
		"last_active_at":  "2026-07-01T11:00:00.000Z",
		"num_messages":    2,
		"num_chat_messages": 1,
	}
	if opts.Title == "" {
		summary["generated_title"] = "relocate fixture"
	}
	if opts.GitRootDir != "" {
		summary["git_root_dir"] = opts.GitRootDir
	}
	if opts.GitRootEqualsOld {
		summary["git_root_dir"] = absOld
	}
	writeJSON(t, filepath.Join(dir, "summary.json"), summary)

	if opts.WritePromptContext {
		pc := map[string]any{
			"working_directory": absOld,
			"bootstrap":         true,
		}
		writeJSON(t, filepath.Join(dir, "prompt_context.json"), pc)
	}

	if opts.UpdatesBody != "" {
		path := filepath.Join(dir, "updates.jsonl")
		if err := os.WriteFile(path, []byte(opts.UpdatesBody), 0o644); err != nil {
			t.Fatalf("write updates: %v", err)
		}
	}

	return dir
}

type relocateSessionOpts struct {
	Title              string
	GitRootDir         string
	GitRootEqualsOld   bool
	WritePromptContext bool
	UpdatesBody        string
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
// With zero ids, writes {"sessions":[]} (inactive for all sessions).
func writeActiveSessions(t *testing.T, grokHome string, sessionIDs ...string) {
	t.Helper()
	entries := make([]map[string]any, 0, len(sessionIDs))
	for _, id := range sessionIDs {
		entries = append(entries, map[string]any{
			"sessionId": id,
			"cwd":       "/tmp/active-fixture",
			"openedAt":  "2026-07-01T12:00:00Z",
		})
	}
	doc := map[string]any{"sessions": entries}
	writeJSON(t, filepath.Join(grokHome, "active_sessions.json"), doc)
}

func writeSQLiteMarker(t *testing.T, grokHome, marker string) string {
	t.Helper()
	path := filepath.Join(grokHome, "sessions", "session_search.sqlite")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir sessions: %v", err)
	}
	if err := os.WriteFile(path, []byte(marker), 0o644); err != nil {
		t.Fatalf("write sqlite marker: %v", err)
	}
	return path
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
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

func readFileString(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

func readJSONMap(t *testing.T, path string) map[string]any {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal %s: %v", path, err)
	}
	return m
}

func summaryInfoCWD(t *testing.T, summaryPath string) string {
	t.Helper()
	m := readJSONMap(t, summaryPath)
	info, _ := m["info"].(map[string]any)
	if info == nil {
		t.Fatalf("summary missing info: %s", summaryPath)
	}
	cwd, _ := info["cwd"].(string)
	return cwd
}

func promptWorkingDirectory(t *testing.T, promptPath string) string {
	t.Helper()
	m := readJSONMap(t, promptPath)
	wd, _ := m["working_directory"].(string)
	return wd
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

func assertDirExists(t *testing.T, path string) {
	t.Helper()
	st, err := os.Stat(path)
	if err != nil || !st.IsDir() {
		t.Fatalf("expected directory %q: %v", path, err)
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

func assertFileEquals(t *testing.T, path, want string) {
	t.Helper()
	got := readFileString(t, path)
	if got != want {
		t.Fatalf("file %s content = %q, want %q", path, got, want)
	}
}

func sessionDirFor(t *testing.T, grokHome, cwd, sessionID string) string {
	t.Helper()
	return filepath.Join(grokHome, "sessions", encodeCWD(t, cwd), sessionID)
}
```
