# Scenario

**Feature**: capture visible pane text for a live Grok session

```
# fixture grok home + optional session summary
test harness -> writeSnapshotSession
  -> inject procs / open files / iTerm refs / Contents / AgentRun (no live ps/osascript)

# primary path (injected fake)
sessions.RunSnapshot(args, stdout, stderr, grokHome, fake.SnapshotOpts())
  -> Find -> prefer AgentRunSnapshot (optional) -> else DiscoverFocusHosting -> Contents

# production (no hooks): Find -> agent-run prefer -> else live TTY CaptureByTTY
```

## Preconditions

- Package exports `Snapshot`, `RunSnapshot`, `SnapshotFake`, `SnapshotHelp`,
  `SnapshotCommandHelpLine`, `DiscoverFocusHosting`.
- Tests never talk to live iTerm2, live `ps`/`lsof`, real `~/.grok`, or live agent-run PTY.

## Steps

1. Root `Setup` allocates `req.TempDir`, `req.GrokHome`, `req.ProjectDir`.
2. Leaf `Setup` writes session fixtures and sets `Args` / `Procs` / `OpenFiles` / `ITerm`.
3. Root `Run` calls `sessions.RunSnapshot` against `SnapshotFake`.
4. Leaf `Assert` checks stdout/stderr, error, ContentsCalls, ListITermCalls.

## Context

- Canonical fixture session id: `019f283a-eeee-7eee-eeee-eeeeeeeeee01`

```go
import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/agent/grok/sessions"
	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

const fixtureSnapshotSessionID = "019f283a-eeee-7eee-eeee-eeeeeeeeee01"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.TempDir = t.TempDir()
	req.GrokHome = filepath.Join(req.TempDir, ".grok")
	req.ProjectDir = filepath.Join(req.TempDir, "proj")
	if err := os.MkdirAll(filepath.Join(req.GrokHome, "sessions"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(req.ProjectDir, 0o755); err != nil {
		return err
	}
	if req.OpenFiles == nil {
		req.OpenFiles = map[int][]string{}
	}
	return nil
}

func writeSnapshotSession(t *testing.T, grokHome, sessionID, cwd, title string) {
	t.Helper()
	key := strings.TrimSpace(cwd)
	if key == "" {
		key = "/tmp/grok-snapshot-empty-cwd"
	}
	absKey, err := filepath.Abs(key)
	if err != nil {
		t.Fatalf("abs cwd key: %v", err)
	}
	dir := filepath.Join(grokHome, "sessions", url.PathEscape(absKey), sessionID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir session: %v", err)
	}
	summary := map[string]any{
		"info": map[string]any{
			"id":  sessionID,
			"cwd": cwd,
		},
		"generated_title":   title,
		"created_at":        "2026-07-01T10:00:00.000Z",
		"updated_at":        "2026-07-01T11:00:00.000Z",
		"last_active_at":    "2026-07-01T11:00:00.000Z",
		"num_messages":      2,
		"num_chat_messages": 1,
	}
	body, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		t.Fatalf("marshal summary: %v", err)
	}
	body = append(body, '\n')
	if err := os.WriteFile(filepath.Join(dir, "summary.json"), body, 0o644); err != nil {
		t.Fatalf("write summary: %v", err)
	}
}

func writeProjectSnapshotSession(t *testing.T, req *Request) {
	t.Helper()
	writeSnapshotSession(t, req.GrokHome, req.SessionID, req.ProjectDir, "snapshot fixture")
}

func grokSnapshotPath(sessionID string) string {
	return "/Users/fixture/.grok/sessions/%2Ftmp%2Fproj/" + sessionID + "/events.jsonl"
}

func grokProc(pid int, tty string) sessions.FocusProc {
	return sessions.FocusProc{PID: pid, PPID: 1, TTY: tty, Cmd: "/usr/local/bin/grok"}
}

func addLiveGrok(req *Request, pid int, tty string) {
	req.Procs = append(req.Procs, grokProc(pid, tty))
	if req.OpenFiles == nil {
		req.OpenFiles = map[int][]string{}
	}
	req.OpenFiles[pid] = []string{grokSnapshotPath(req.SessionID)}
}

func oneITermTab() []iterm2.SessionRef {
	return []iterm2.SessionRef{
		{WindowID: "3", WindowName: "worktrees", TabIndex: 1, SessionID: "w2t1p0", TTY: "/dev/ttys148"},
	}
}

func twoITermTabs() []iterm2.SessionRef {
	return []iterm2.SessionRef{
		{WindowID: "1", WindowName: "credit-pricing", TabIndex: 2, SessionID: "w0t2p0", TTY: "/dev/ttys148"},
		{WindowID: "3", WindowName: "worktrees", TabIndex: 1, SessionID: "w2t1p0", TTY: "/dev/ttys149"},
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

func assertNoContents(t *testing.T, resp *Response) {
	t.Helper()
	if len(resp.ContentsCalls) != 0 {
		t.Fatalf("ContentsCalls = %v, want none", resp.ContentsCalls)
	}
}

func assertErrorOutput(t *testing.T, resp *Response, template string) {
	t.Helper()
	assertError(t, resp)
	assert.Output(t, resp.Err.Error()+"\n", template)
}
```
