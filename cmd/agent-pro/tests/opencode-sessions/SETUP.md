# Scenario

**Feature**: agent-pro opencode sessions list and session info from synthetic data dir

```
# harness builds synthetic OpenCode storage with session + message JSON
test harness -> sessions package -> fixtures under DataDir/storage/

# list discovers sessions; table shows MSGS and relative last-active times
sessions package -> FormatListTable(now) -> formatted terminal output

# info locates one session by exact id and renders detail blocks
sessions.Find/Info -> FormatInfoText(now) -> key-value session detail
```

## Preconditions

- Package `agent/opencode/sessions` exposes List, FormatListTable, Find, Info, and
  FormatInfoText.
- Tests never read the real user `~/.local/share/opencode` directory.

## Steps

1. Root Setup allocates `req.DataDir` as `{temp}/.local/share/opencode`.
2. Root Setup sets `req.Now` to a fixed UTC instant for deterministic relative times.
3. Leaf Setup writes session and optional message JSON fixtures.

## Context

- Session path pattern:
  `{DataDir}/storage/session/{projectID}/ses_<id>.json`
- Message path pattern:
  `{DataDir}/storage/message/{sessionID}/msg_*.json`
- `time.updated` (epoch ms) drives sort order and LAST ACTIVE column.
- Message file count populates the MSGS column in list table output.
- Info tests require exact session ids (no prefix matching).

```go
import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const fixedNow = "2026-07-03T15:00:00.000Z"

type opencodeMessageOpts struct {
	InputTokens  int
	OutputTokens int
	Cost         float64
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.DataDir = filepath.Join(t.TempDir(), ".local", "share", "opencode")
	now, err := time.Parse(time.RFC3339, fixedNow)
	if err != nil {
		t.Fatalf("parse fixed now: %v", err)
	}
	req.Now = now.UTC()
	return nil
}

func epochMS(t time.Time) int64 {
	return t.UTC().UnixMilli()
}

func writeOpencodeSession(t *testing.T, dataDir, projectID, sessionID, title, directory string, updated time.Time) string {
	t.Helper()
	dir := filepath.Join(dataDir, "storage", "session", projectID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir session dir: %v", err)
	}

	payload := map[string]any{
		"id":        sessionID,
		"title":     title,
		"directory": directory,
		"time": map[string]any{
			"created": epochMS(updated.Add(-time.Hour)),
			"updated": epochMS(updated),
		},
	}
	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal session: %v", err)
	}
	path := filepath.Join(dir, sessionID+".json")
	if err := os.WriteFile(path, b, 0644); err != nil {
		t.Fatalf("write session %s: %v", path, err)
	}
	return path
}

func writeOpencodeMessage(t *testing.T, dataDir, sessionID, messageID string, opts opencodeMessageOpts) string {
	t.Helper()
	dir := filepath.Join(dataDir, "storage", "message", sessionID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir message dir: %v", err)
	}

	payload := map[string]any{
		"id":        messageID,
		"sessionID": sessionID,
		"role":      "assistant",
		"tokens": map[string]any{
			"input":  opts.InputTokens,
			"output": opts.OutputTokens,
			"total":  opts.InputTokens + opts.OutputTokens,
		},
		"cost": opts.Cost,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal message: %v", err)
	}
	path := filepath.Join(dir, messageID+".json")
	if err := os.WriteFile(path, b, 0644); err != nil {
		t.Fatalf("write message %s: %v", path, err)
	}
	return path
}

func assertSuccess(t *testing.T, resp *Response) {
	t.Helper()
	if resp.Err != nil {
		t.Fatalf("operation failed: %v", resp.Err)
	}
}

func assertError(t *testing.T, resp *Response) {
	t.Helper()
	if resp.Err == nil {
		t.Fatal("expected error but got nil")
	}
}

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("missing %q in:\n%s", want, got)
	}
}

func assertNotContains(t *testing.T, got, want string) {
	t.Helper()
	if strings.Contains(got, want) {
		t.Fatalf("unexpected %q in:\n%s", want, got)
	}
}
```