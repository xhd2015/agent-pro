# Scenario

**Feature**: type text into a live Grok iTerm host (optional --open resume)

```
# fixture grok home + optional session summary
test harness -> writeSendSession
  -> inject procs / open files / iTerm refs / SendText / open (no live ps or osascript)

# primary path
sessions.RunSend(args, stdout, stderr, grokHome, fake.SendOpts())
  -> Find/Info -> DiscoverFocusHosting -> [optional open+wait] -> SendText
```

## Preconditions

- Package exports `Send`, `RunSend`, `SendFake`, `SendHelp`,
  `SendCommandHelpLine`, `DiscoverFocusHosting`.
- Tests never talk to live iTerm2, live `ps`/`lsof`, or real `~/.grok`.

## Steps

1. Root `Setup` allocates `req.TempDir`, `req.GrokHome`, `req.ProjectDir`.
2. Leaf `Setup` writes session fixtures and sets `Args` / `Procs` / `OpenFiles` / `ITerm`.
3. Root `Run` calls `sessions.RunSend` against `SendFake`.
4. Leaf `Assert` checks stdout/stderr, error, SendCalls, Opened.

## Context

- Canonical fixture session id: `019f283a-aaaa-7aaa-aaaa-aaaaaaaaaa01`

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

const fixtureSendSessionID = "019f283a-aaaa-7aaa-aaaa-aaaaaaaaaa01"

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

func writeSendSession(t *testing.T, grokHome, sessionID, cwd, title string) {
	t.Helper()
	key := strings.TrimSpace(cwd)
	if key == "" {
		key = "/tmp/grok-send-empty-cwd"
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

func writeProjectSendSession(t *testing.T, req *Request) {
	t.Helper()
	writeSendSession(t, req.GrokHome, req.SessionID, req.ProjectDir, "send fixture")
}

func grokSendPath(sessionID string) string {
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
	req.OpenFiles[pid] = []string{grokSendPath(req.SessionID)}
}

func oneITermTab() []iterm2.SessionRef {
	return []iterm2.SessionRef{
		{WindowID: "3", WindowName: "worktrees", TabIndex: 1, SessionID: "w2t1p0", TTY: "/dev/ttys148"},
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

func assertNoSend(t *testing.T, resp *Response) {
	t.Helper()
	if len(resp.SendCalls) != 0 {
		t.Fatalf("SendCalls = %v, want none", resp.SendCalls)
	}
}

func assertNoOpen(t *testing.T, resp *Response) {
	t.Helper()
	if len(resp.Opened) != 0 {
		t.Fatalf("Opened = %v, want none", resp.Opened)
	}
}

func assertErrorContains(t *testing.T, resp *Response, want string) {
	t.Helper()
	assertError(t, resp)
	if !strings.Contains(resp.Err.Error(), want) {
		t.Fatalf("error %q must contain %q", resp.Err.Error(), want)
	}
}

func assertSentDefaults(t *testing.T, resp *Response, text string) {
	t.Helper()
	if len(resp.SendCalls) != 1 {
		t.Fatalf("SendCalls = %v, want 1", resp.SendCalls)
	}
	c := resp.SendCalls[0]
	if c.Text != text {
		t.Fatalf("SendText text = %q, want %q", c.Text, text)
	}
	if c.Opts.Focus || c.Opts.NoSubmit || c.Opts.NoCtrlU {
		t.Fatalf("default opts = %+v, want all false", c.Opts)
	}
	if c.SessionID != "w2t1p0" {
		t.Fatalf("SendText session = %q, want w2t1p0", c.SessionID)
	}
}
```
