# `agent-pro grok session list-live`

List Grok session ids that currently have a hosting iTerm2 tab.
L2 only — injectable `ListLiveFake` (procs / open files / iTerm / pane idle).

# DSN (Domain Specific Notion)

## Participants

- **Caller** — harness via `sessions.RunListLive` / `ListLive`.
- **ListLive** — discover live grok sids (lsof hard hit) → DiscoverFocusHosting → omit no-tab.
- **Injected probes** — FocusFake + PaneByTTY / CwdBySession.

## Behaviors

- Parent help names `list-live`.
- `--help` documents `--json` / `--limit`.
- Empty: no live grok or no iTerm → `0 sessions`.
- One host → one row with SESSION_ID + ITERM.
- Multi-tab same sid → `w=… t=…(+N)`.
- Live PID without iTerm tab → omitted.
- Codex runner ignored (grok-only).
- `--json` envelope with `sessions` + `summary.count`.
- `--limit N` caps rows.
- Multi-host: `ListProcs` / `Lsof` / `ListITerm` each run once per invocation (shared probes).
- Pane idle without cwd → WORKSPACE/TITLE from summary.json beside the lsof hard-hit session dir.
- Production path: one TTY-targeted FindSessionsByTTY AppleScript (skip full dump + Capture); one bulk `lsof` for grok PIDs; no WalkDir meta index.
- Columns: SESSION_ID, ITERM, TITLE, WORKSPACE (no SENDABLE).
- CaptureInventory (when set) still feeds hosting+pane from one load for enrich tests.

## Version

0.0.7

## Decision Tree

```text
list-live/
├── help/
│   ├── parent-lists/
│   └── usage/
├── empty/
├── one-host/
├── multi-tab/
├── omit-no-iterm/
├── omit-codex/
├── json/
├── limit/
├── probe-budget/
├── cwd-from-disk/
└── one-iterm-inventory/
```

## Test Index

| Leaf | Contract |
|---|---|
| `help/parent-lists/` | Help line names `list-live`. |
| `help/usage/` | `--help` documents `--json`, `--limit`. |
| `empty/` | Footer `0 sessions`. |
| `one-host/` | Row with sid + `w=3 t=1`; TITLE header; no SENDABLE. |
| `multi-tab/` | ITERM ends with `(+1)`. |
| `omit-no-iterm/` | Live sid with empty ITerm → omitted. |
| `omit-codex/` | Codex open-files ignored. |
| `json/` | JSON has session_id + title; no sendable. |
| `limit/` | `--limit 1` with two hosts → one row. |
| `probe-budget/` | Two hosts → 1× ListProcs, 1× Lsof per PID, 1× ListITerm. |
| `cwd-from-disk/` | Empty pane cwd → TITLE + WORKSPACE from path-derived summary.json. |
| `one-iterm-inventory/` | Unified CaptureInventory runs once for hosting+pane. |

## How to Run

```sh
doctest vet ./agent/grok/sessions/tests/list-live
doctest test ./agent/grok/sessions/tests/list-live
```

```go
import (
	"bytes"
	"testing"

	"github.com/xhd2015/agent-pro/agent/grok/sessions"
	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

type Request struct {
	Args         []string
	TempDir      string
	GrokHome     string
	Procs        []sessions.FocusProc
	OpenFiles    map[int][]string
	ITerm        []iterm2.SessionRef
	PaneByTTY    map[string]sessions.LivePaneInfo
	CwdBySession map[string]string
	// DiskCwd leaves FindSession nil so ListLive uses discoverSessions index.
	DiskCwd bool
	// UnifiedInventory leaves ListITerm+PaneByTTY nil and injects CaptureInventory.
	UnifiedInventory bool
	ParentHelp       bool
}

type Response struct {
	Stdout                string
	Stderr                string
	Err                   error
	ListProcsCalls        int
	LsofCalls             int
	ListITermCalls        int
	CaptureInventoryCalls int
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d
	if req.ParentHelp {
		return &Response{Stdout: sessions.ListLiveCommandHelpLine + "\n"}, nil
	}
	fake := &sessions.ListLiveFake{
		FocusFake: sessions.FocusFake{
			Procs:     req.Procs,
			OpenFiles: req.OpenFiles,
			ITerm:     req.ITerm,
		},
		PaneByTTY:    req.PaneByTTY,
		CwdBySession: req.CwdBySession,
	}
	opts := fake.ListLiveOpts()
	opts.GrokHome = req.GrokHome
	if req.DiskCwd {
		opts.FindSession = nil
	}
	var captureCalls int
	if req.UnifiedInventory {
		panes := req.PaneByTTY
		if panes == nil {
			panes = map[string]sessions.LivePaneInfo{}
		}
		refs := append([]iterm2.SessionRef(nil), req.ITerm...)
		opts.ListITerm = nil
		opts.PaneByTTY = nil
		opts.CaptureInventory = func() (map[string]sessions.LivePaneInfo, []iterm2.SessionRef, error) {
			captureCalls++
			outPanes := make(map[string]sessions.LivePaneInfo, len(panes))
			for k, v := range panes {
				outPanes[k] = v
			}
			return outPanes, append([]iterm2.SessionRef(nil), refs...), nil
		}
	}
	var stdout, stderr bytes.Buffer
	err := sessions.RunListLive(req.Args, &stdout, &stderr, req.GrokHome, opts)
	return &Response{
		Stdout:                stdout.String(),
		Stderr:                stderr.String(),
		Err:                   err,
		ListProcsCalls:        fake.ListProcsCalls,
		LsofCalls:             fake.LsofCalls,
		ListITermCalls:        fake.ListITermCalls,
		CaptureInventoryCalls: captureCalls,
	}, nil
}
```
