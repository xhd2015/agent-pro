# Scenario

**Feature**: fork grok session (explicit id or tab)

## Preconditions

- Package exports `RunFork`, `ForkOpts`, `ForkHelp`, `ForkCommandHelpLine`.
- Tests never talk to live iTerm2, live `ps`/`lsof`, or real PATH grok.
- Session fixtures are written under `req.GrokHome` for `Info`.

```go
import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/agent/grok/sessions"
	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

const (
	fixtureForkSessionID = "019f283b-aaaa-7aaa-aaaa-aaaaaaaaaaaa"
	fixtureTabSessionID  = "019f283b-dddd-7ddd-dddd-dddddddddddd"
	pidTabGrok2          = 8200
)

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
	if req.SessionID == "" {
		req.SessionID = fixtureForkSessionID
	}
	return nil
}

func writeForkSession(t *testing.T, grokHome, sessionID, cwd, title string) {
	t.Helper()
	key := strings.TrimSpace(cwd)
	if key == "" {
		key = "/tmp/grok-fork-empty-cwd"
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
		"generated_title": title,
		"created_at":      "2026-07-01T10:00:00.000Z",
		"updated_at":      "2026-07-01T11:00:00.000Z",
		"last_active_at":  "2026-07-01T11:00:00.000Z",
		"num_messages":    2,
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

func grokOpenPath(sessionID string) string {
	return "/Users/fixture/.grok/sessions/%2Ftmp%2Fproj/" + sessionID + "/events.jsonl"
}

func seedTabWindow(req *Request) {
	req.CurrentSessionID = "w0t1p0:CURRENT-TAB-UUID"
	req.ControllingTTY = "/dev/ttys101"
	req.ITerm = []iterm2.SessionRef{
		{WindowID: "100", WindowName: "work", TabIndex: 1, SessionID: "w0t1p0:CURRENT-TAB-UUID", TTY: "/dev/ttys101", Name: "current"},
		{WindowID: "100", WindowName: "work", TabIndex: 2, SessionID: "w0t2p0:TAB2-UUID", TTY: "/dev/ttys102", Name: "grok-tab"},
		{WindowID: "100", WindowName: "work", TabIndex: 3, SessionID: "w0t3p0:TAB3-UUID", TTY: "/dev/ttys103", Name: "bash-only"},
	}
	req.FocusProcs = []sessions.FocusProc{
		{PID: pidTabGrok2, PPID: 1, TTY: "/dev/ttys102", Cmd: "/usr/local/bin/grok"},
	}
	req.OpenFiles[pidTabGrok2] = []string{grokOpenPath(fixtureTabSessionID)}
}

func assertNoHarnessErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Run harness error: %v", err)
	}
}

func assertOK(t *testing.T, resp *Response) {
	t.Helper()
	if resp.Err != nil {
		t.Fatalf("unexpected error: %v", resp.Err)
	}
}

func assertErrContains(t *testing.T, resp *Response, sub string) {
	t.Helper()
	if resp.Err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(resp.Err.Error(), sub) {
		t.Fatalf("error %q does not contain %q", resp.Err.Error(), sub)
	}
}
```
