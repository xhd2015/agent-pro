# Scenario

**Feature**: focus the iTerm2 tab that hosts a live Grok session

```
# fixture grok home + optional session summary
test harness -> writeFocusSession
  -> inject procs / open files / iTerm refs (no live ps or osascript)

# primary path
sessions.RunFocus(args, stdout, grokHome, fake.Opts())
  -> Find session -> live grok PIDs -> TTY tree -> FindByTTY -> Focus
```

## Preconditions

- Package exports `Focus`, `RunFocus`, `FocusFake`, `FocusProc`,
  `FocusCandidate`, `ErrNotFound`, `FocusHelp`, and `FocusCommandHelpLine`.
- Tests never talk to live iTerm2, live `ps`/`lsof`, or real `~/.grok`.

## Steps

1. Root `Setup` allocates `req.TempDir`, `req.GrokHome = {temp}/.grok`, and
   `req.ProjectDir = {temp}/proj` (created).
2. Leaf `Setup` writes session fixtures and sets `Args` / `Procs` / `OpenFiles` / `ITerm`.
3. Root `Run` calls `sessions.RunFocus` against `FocusFake`.
4. Leaf `Assert` checks stdout, error, ListITermCalls, and Focused ids.

## Context

- Canonical fixture session id: `019f283a-dddd-7ddd-dddd-dddddddddd01`
- Encoded cwd keys use `url.PathEscape(abs_cwd)` like other grok session trees.
- Live PID hits use an open-file path under `/.grok/sessions/…/<uuid>/…`.

```go
import (
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/agent/grok/sessions"
	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

const fixtureFocusSessionID = "019f283a-dddd-7ddd-dddd-dddddddddd01"

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

func writeFocusSession(t *testing.T, grokHome, sessionID, cwd, title string) {
	t.Helper()
	key := strings.TrimSpace(cwd)
	if key == "" {
		key = "/tmp/grok-focus-empty-cwd"
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

func writeProjectFocusSession(t *testing.T, req *Request) {
	t.Helper()
	writeFocusSession(t, req.GrokHome, req.SessionID, req.ProjectDir, "focus fixture")
}

func grokOpenPath(sessionID string) string {
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
	req.OpenFiles[pid] = []string{grokOpenPath(req.SessionID)}
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

func assertNotFound(t *testing.T, resp *Response) {
	t.Helper()
	assertError(t, resp)
	if !errors.Is(resp.Err, sessions.ErrNotFound) && resp.Err.Error() != "not found" {
		t.Fatalf("error = %v, want not found", resp.Err)
	}
}

func assertNoITerm(t *testing.T, resp *Response) {
	t.Helper()
	if resp.ListITermCalls != 0 {
		t.Fatalf("ListITermCalls = %d, want 0", resp.ListITermCalls)
	}
	if len(resp.Focused) != 0 {
		t.Fatalf("Focused = %v, want none", resp.Focused)
	}
}

func assertErrorOutput(t *testing.T, resp *Response, template string) {
	t.Helper()
	assertError(t, resp)
	assert.Output(t, resp.Err.Error()+"\n", template)
}
```
