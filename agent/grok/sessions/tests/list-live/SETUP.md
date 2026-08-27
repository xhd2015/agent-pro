# Scenario

**Feature**: list Grok session ids hosted in iTerm tabs

```
sessions.RunListLive(args, stdout, stderr, grokHome, fake.ListLiveOpts())
  -> withSharedListLiveProbes (prefetch ListITerm ∥ bulk/cached lsof)
  -> discoverLiveGrokSessions (sid + summary.json beside hard-hit dir)
  -> DiscoverFocusHosting (shared ps/lsof/iTerm) -> table/JSON
```

## Preconditions

- Package exports `ListLive`, `RunListLive`, `ListLiveFake`, `ListLiveHelp`,
  `ListLiveCommandHelpLine`.
- No live iTerm / ps / lsof.

## Steps

1. Leaf sets Args / Procs / OpenFiles / ITerm / PaneByTTY.
2. Root `Run` calls `RunListLive` against `ListLiveFake`.

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

const fixtureListLiveSID = "019f283a-aaaa-7aaa-aaaa-aaaaaaaaaa01"
const fixtureListLiveSID2 = "019f283a-bbbb-7bbb-bbbb-bbbbbbbbbb02"
const fixtureListLiveDiskCWD = "/tmp/list-live-disk-cwd"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if req.OpenFiles == nil {
		req.OpenFiles = map[int][]string{}
	}
	if req.PaneByTTY == nil {
		req.PaneByTTY = map[string]sessions.LivePaneInfo{}
	}
	if req.TempDir == "" {
		req.TempDir = t.TempDir()
	}
	if req.GrokHome == "" {
		req.GrokHome = filepath.Join(req.TempDir, ".grok")
	}
	if err := os.MkdirAll(filepath.Join(req.GrokHome, "sessions"), 0o755); err != nil {
		return err
	}
	return nil
}

func grokListLivePath(sessionID string) string {
	return "/Users/fixture/.grok/sessions/%2Ftmp%2Fproj/" + sessionID + "/events.jsonl"
}

func addLiveGrokHost(req *Request, pid int, ttyBare, sessionID, windowID string, tabIndex int) {
	req.Procs = append(req.Procs, sessions.FocusProc{
		PID:  pid,
		PPID: 1,
		TTY:  ttyBare,
		Cmd:  "/usr/local/bin/grok",
	})
	req.OpenFiles[pid] = []string{grokListLivePath(sessionID)}
	req.ITerm = append(req.ITerm, iterm2.SessionRef{
		WindowID:   windowID,
		WindowName: "work",
		TabIndex:   tabIndex,
		SessionID:  "iterm-" + ttyBare,
		TTY:        "/dev/" + ttyBare,
	})
	idle := true
	req.PaneByTTY["/dev/"+ttyBare] = sessions.LivePaneInfo{
		Idle: &idle,
		Cwd:  "/tmp/proj",
	}
}

func writeListLiveSession(t *testing.T, grokHome, sessionID, cwd, title string) string {
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
		"generated_title":   title,
		"created_at":        "2026-07-01T10:00:00.000Z",
		"updated_at":        "2026-07-01T11:00:00.000Z",
		"last_active_at":    "2026-07-01T11:00:00.000Z",
		"num_messages":      2,
		"num_chat_messages": 1,
	}
	data, err := json.Marshal(summary)
	if err != nil {
		t.Fatalf("marshal summary: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "summary.json"), append(data, '\n'), 0o644); err != nil {
		t.Fatalf("write summary: %v", err)
	}
	return dir
}

// pointOpenFileAtSession sets the PID open-files hard hit to events.jsonl under
// the real session dir (path-derived summary resolution).
func pointOpenFileAtSession(req *Request, pid int, sessionDir string) {
	req.OpenFiles[pid] = []string{filepath.Join(sessionDir, "events.jsonl")}
}

func assertNoHarnessErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected harness error: %v", err)
	}
}

func assertNoError(t *testing.T, resp *Response) {
	t.Helper()
	if resp.Err != nil {
		t.Fatalf("want nil err, got %v; stderr=%q stdout=%q", resp.Err, resp.Stderr, resp.Stdout)
	}
}

func assertContains(t *testing.T, got, want, label string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("%s must contain %q; got:\n%s", label, want, got)
	}
}
```
