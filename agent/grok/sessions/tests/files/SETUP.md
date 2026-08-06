# Scenario

**Feature**: list all files in a Grok session directory from synthetic GROK_HOME

```
# harness builds session dir with multiple artifacts
writeFilesSession(summary, updates, signals)
  -> ListSessionFiles(grokHome, id)
  -> optional FormatSessionFilesTable | FormatSessionFilesJSON
```

## Preconditions

- Package will export (RED until implementer):
  - `ListSessionFiles`, `SessionFile`
  - `FormatSessionFilesTable`, `FormatSessionFilesJSON`
- Session location matches existing Find layout:
  `{grokHome}/sessions/{url.PathEscape(abs_cwd)}/{uuid}/`
- Tests never read real `~/.grok`.

## Steps

1. Root `Setup` allocates temp `.grok` home.
2. Leaf writes multi-file fixtures or leaves tree empty for unknown id.
3. `Run` calls `ListSessionFiles` and optional formatter.
4. Assert names / sizes / errors / table / json.

## Context

- Canonical fixture id: `019f283a-cccc-7ccc-cccc-cccccccccc01`
- Fixture cwd: `/tmp/grok-files-fixture-project`
- Helper seeds `summary.json`, optional `updates.jsonl`, `signals.json` with
  known non-empty bodies for size asserts.

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
	fixtureFilesSessionID = "019f283a-cccc-7ccc-cccc-cccccccccc01"
	fixtureFilesCWD       = "/tmp/grok-files-fixture-project"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.TempDir = t.TempDir()
	req.GrokHome = filepath.Join(req.TempDir, ".grok")
	if err := os.MkdirAll(filepath.Join(req.GrokHome, "sessions"), 0o755); err != nil {
		return err
	}
	return nil
}

// writeFilesSession seeds a session dir with named artifact files.
// bodies maps basename -> file content (must be non-empty for size checks).
// Returns absolute session directory.
func writeFilesSession(t *testing.T, grokHome, sessionID, cwd string, bodies map[string]string) string {
	t.Helper()
	absCwd, err := filepath.Abs(cwd)
	if err != nil {
		t.Fatalf("abs cwd: %v", err)
	}
	dir := filepath.Join(grokHome, "sessions", url.PathEscape(absCwd), sessionID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir session: %v", err)
	}
	// Always ensure summary.json is valid enough for Find if present in bodies.
	for name, body := range bodies {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}
	return dir
}

func defaultMultiFileBodies(sessionID string) map[string]string {
	summary, _ := json.Marshal(map[string]any{
		"info": map[string]any{
			"id":  sessionID,
			"cwd": fixtureFilesCWD,
		},
		"generated_title":   "files multi fixture",
		"created_at":        "2026-07-01T10:00:00.000Z",
		"updated_at":        "2026-07-01T11:00:00.000Z",
		"last_active_at":    "2026-07-01T11:00:00.000Z",
		"num_messages":      1,
		"num_chat_messages": 1,
	})
	return map[string]string{
		"summary.json":  string(summary) + "\n",
		"updates.jsonl": `{"type":"init","marker":"files-fixture"}` + "\n",
		"signals.json":  `{"contextTokensUsed":100,"contextWindowTokens":200000}` + "\n",
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

func fileNames(files []sessions.SessionFile) map[string]sessions.SessionFile {
	m := make(map[string]sessions.SessionFile, len(files))
	for _, f := range files {
		m[f.Name] = f
	}
	return m
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

// silence unused import if time only used in types elsewhere
var _ = time.Time{}
```
