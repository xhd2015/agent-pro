# Scenario

**Subcommand**: `sessions` — list stored sessions or print one session's events

```
agent-run sessions [--json] [--limit N] -> list under AGENT_RUN_HOME/sessions (flat)
# human UPDATED: FormatRelativeAge; --json: absolute timestamps
agent-run sessions <session_id> --print -> read meta + events.jsonl -> FormatState trace
# bare session id only; runner/id refs rejected (Q5)
```

## Preconditions

- `agent-run` binary is built (inherited from root `SETUP.md`).
- `AGENT_RUN_HOME` is isolated per test.
- Session layout is flat: `sessions/<session_id>/`.
- Print leaves seed data via `pkgs/agentstorage` without a full `run` subprocess.

## Steps

1. Grouping `list/` or `print/` `Setup` sets `req.Args` for that operation mode.
2. Leaf `Setup` seeds sessions or finalizes flags.
3. `Run` executes `agent-run` with accumulated `req.Args`.
4. `Assert` checks exit code, stdout/stderr shape.

```go
import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
)

const (
	printRunner    = "fake-codex"
	printSessionID = "web_test123"
)

func openAgentStore(t *testing.T, req *Request) agentstorage.Store {
	t.Helper()
	store, err := agentstorage.NewFileStore(req.Home)
	if err != nil {
		t.Fatalf("NewFileStore(%q): %v", req.Home, err)
	}
	return store
}

func seedSessionMeta(t *testing.T, store agentstorage.Store, sessionID, status string) {
	t.Helper()
	seedSessionMetaRunner(t, store, printRunner, sessionID, status)
}

func seedSessionMetaRunner(t *testing.T, store agentstorage.Store, runner, sessionID, status string) {
	t.Helper()
	meta := agentstorage.SessionMeta{
		Runner:    runner,
		SessionID: sessionID,
		Status:    status,
	}
	if err := store.CreateSession(sessionID, meta); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
}

func appendAgentMessage(t *testing.T, store agentstorage.Store, sessionID, text string) {
	t.Helper()
	ev := types.AgentEvent{
		Type: types.ActionMessage,
		Role: "assistant",
		Text: text,
	}
	if err := store.AppendEvent(sessionID, ev); err != nil {
		t.Fatalf("AppendEvent: %v", err)
	}
}

// seedFlatSessionMeta writes meta.json under sessions/<id>/ with explicit timestamps.
// Used for list sort/limit tests that need distinct updated_at values.
func seedFlatSessionMeta(t *testing.T, home, sessionID, runner, status, updatedAt string) {
	t.Helper()
	dir := filepath.Join(home, "sessions", sessionID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir session: %v", err)
	}
	created := updatedAt
	meta := map[string]any{
		"runner":     runner,
		"session_id": sessionID,
		"status":     status,
		"created_at": created,
		"updated_at": updatedAt,
	}
	data, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("marshal meta: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "meta.json"), data, 0o644); err != nil {
		t.Fatalf("write meta: %v", err)
	}
}

// seedNSessions creates n sessions with updated_at spaced so higher index is newer.
// IDs: sess_00 .. sess_{n-1}; sess_{n-1} is newest.
func seedNSessions(t *testing.T, home string, n int) {
	t.Helper()
	base := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < n; i++ {
		id := fmt.Sprintf("sess_%02d", i)
		runner := "fake-codex"
		if i%2 == 1 {
			runner = "fake-opencode"
		}
		updated := base.Add(time.Duration(i) * time.Minute).Format(time.RFC3339)
		seedFlatSessionMeta(t, home, id, runner, "finished", updated)
	}
}

func Setup(t *testing.T, req *Request) error {
	if req.SessionRunner == "" {
		req.SessionRunner = printRunner
	}
	return nil
}

// listDataRows returns whitespace-split fields per data line (skips header/footer).
func listDataRows(t *testing.T, stdout string) [][]string {
	t.Helper()
	var rows [][]string
	for _, line := range strings.Split(strings.TrimRight(stdout, "\n"), "\n") {
		trim := strings.TrimSpace(line)
		if trim == "" {
			continue
		}
		upper := strings.ToUpper(trim)
		if strings.HasPrefix(upper, "SESSION_ID") || strings.HasPrefix(trim, "(") {
			// header or truncation footer
			continue
		}
		rows = append(rows, strings.Fields(trim))
	}
	return rows
}

// listUpdatedCell returns the UPDATED column from a Fields-split data row.
// Relative ages include a space ("2s ago", "just now"), so fields[3:] are joined.
func listUpdatedCell(row []string) string {
	if len(row) < 4 {
		return ""
	}
	return strings.Join(row[3:], " ")
}
```

