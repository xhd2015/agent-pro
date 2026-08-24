# `agent-pro grok session snapshot`

`agent-pro grok session snapshot (<session-id> | --tab SEL | --tab-index N) […]`
captures currently visible pane text for a live Grok host. Prefers a live
agent-run grok-tty snapshot when the Grok id is bound; otherwise uses
`iterm2.Contents`. Tab source uses shared `ResolveFromTab` /
`ResolveSessionSource`. No resume when no host — hard error.

L2 — injectable `SnapshotFake` (procs / iTerm / Contents / AgentRun). No live
AppleScript or live agent-run PTY.

# DSN (Domain Specific Notion)

## Participants

- **User** invokes `grok session snapshot` with a session id or tab selector.
- **Snapshot CLI** (`RunSnapshot`) validates args, prints help / pane text / JSON /
  dry-run plans, and returns errors without an `Error:` prefix.
- **Snapshot service** finds the Grok session, optionally prefers agent-run TTY
  snapshot, else discovers focus candidates and calls injectable Contents.
- **Injected process / iTerm / Contents / AgentRun** supplies deterministic boundaries.

## Behaviors

- Help documents `<session-id>`, `--tab`, `--tab-index`, `--index`, `--json`,
  `-o/--output`, `--dry-run`, `--iterm`, agent-run prefer.
- Parent session help names the snapshot command.
- Missing session source is a usage error; Contents is not invoked.
- Unknown Grok session returns `grok session not found` and never captures.
- One hosting tab dumps pane text to stdout (trailing newline); `source=iterm`.
- Agent-run prefer hit dumps sanitized text; Contents not called; `source=agent-run`.
- Agent-run soft miss / `--iterm` falls back to Contents.
- Ambiguous agent-run mapping warns on stderr and falls back to iTerm.
- Several tabs without `--index` list candidates and never capture.
- No live host is a hard error (`no hosting iTerm tab`); Contents not called.
- `--json` emits `session_id`, `iterm_session_id`, `app`, `source`, `contents`
  (+ `agent_run_session_id` when source is agent-run).
- `-o/--output` writes body to file; stdout empty.
- `--dry-run` prints the plan and never calls Contents.
- `--tab` / `--tab-index` resolve via `ResolveFromTab` then capture that pane.
- Session id cannot combine with `--tab` / `--tab-index`; `--index` cannot combine with tab source.

## Version

0.0.1

## Decision Tree

```text
snapshot/
├── help/
│   ├── parent-lists-snapshot/
│   └── snapshot-usage/
├── validation/
│   └── missing-session-id/
├── unknown-session/
├── agent-run/
│   ├── prefer/
│   ├── fallback-miss/
│   ├── force-iterm/
│   └── ambiguous-warn/
├── capture/
│   ├── exactly-one/
│   ├── multiple/
│   │   ├── requires-index/
│   │   └── selects-index/
│   └── no-live/
├── json/
│   └── shape/
├── output/
│   └── writes-file/
├── dry-run/
│   └── plan/
└── tab/
    ├── hit-capture/
    ├── id-and-tab/
    └── index-with-tab/
```

## Test Index

| Leaf | Contract |
|---|---|
| `help/parent-lists-snapshot/` | Parent session help line names `snapshot`. |
| `help/snapshot-usage/` | `snapshot --help` documents `--tab`, Contents, `--index`, `--json`, `--iterm`. |
| `validation/missing-session-id/` | No session source is a usage error without Contents. |
| `unknown-session/` | Unknown UUID → `grok session not found`; no Contents. |
| `agent-run/prefer/` | Live agent-run hit → text + `source=agent-run`; Contents not called. |
| `agent-run/fallback-miss/` | Soft miss → Contents path; AgentRun probed once. |
| `agent-run/force-iterm/` | `--iterm` skips AgentRun; Contents used. |
| `agent-run/ambiguous-warn/` | Ambiguous mapping → `warning:` stderr + Contents fallback. |
| `capture/exactly-one/` | One TTY match dumps fixture pane text; Contents once; ListITerm once. |
| `capture/multiple/requires-index/` | Two tabs list candidates with `snapshot` hint; never capture. |
| `capture/multiple/selects-index/` | `--index 1` captures candidate 1's pane. |
| `capture/no-live/` | Known session, no live host → hard error; Contents not called. |
| `json/shape/` | `--json` emits grok + iterm ids, app, `source=iterm`, contents. |
| `output/writes-file/` | `-o FILE` writes pane text; stdout empty. |
| `dry-run/plan/` | Would capture plan with `source: iterm`; Contents not called. |
| `tab/hit-capture/` | `--tab 2` captures resolved tab pane. |
| `tab/id-and-tab/` | Positional id + `--tab` errors. |
| `tab/index-with-tab/` | `--index` with `--tab` errors. |

## How to Run

```sh
doctest vet ./agent/grok/sessions/tests/snapshot
doctest test ./agent/grok/sessions/tests/snapshot
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
	Args             []string
	GrokHome         string
	TempDir          string
	ProjectDir       string
	SessionID        string
	Procs            []sessions.FocusProc
	OpenFiles        map[int][]string
	ITerm            []iterm2.SessionRef
	CurrentSessionID string
	ControllingTTY   string
	ContentsByID     map[string]iterm2.ContentsResult
	AgentRunByID     map[string]*sessions.AgentRunSnapshotResult
	AgentRunErr      error
	AgentRunEnabled  bool
	ParentHelp       bool
}

type Response struct {
	Stdout         string
	Stderr         string
	Err            error
	ContentsCalls  []string
	AgentRunCalls  []string
	ListITermCalls int
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d
	if req.ParentHelp {
		return &Response{Stdout: sessions.SnapshotCommandHelpLine + "\n"}, nil
	}
	fake := &sessions.SnapshotFake{
		FocusFake: sessions.FocusFake{
			Procs:     req.Procs,
			OpenFiles: req.OpenFiles,
			ITerm:     req.ITerm,
		},
		CurrentSessionID: req.CurrentSessionID,
		ControllingTTY:   req.ControllingTTY,
		ContentsByID:     req.ContentsByID,
		AgentRunByID:     req.AgentRunByID,
		AgentRunErr:      req.AgentRunErr,
		AgentRunEnabled:  req.AgentRunEnabled,
	}
	var stdout, stderr bytes.Buffer
	opts := fake.SnapshotOpts()
	err := sessions.RunSnapshot(req.Args, &stdout, &stderr, req.GrokHome, opts)
	return &Response{
		Stdout:         stdout.String(),
		Stderr:         stderr.String(),
		Err:            err,
		ContentsCalls:  append([]string(nil), fake.ContentsCalls...),
		AgentRunCalls:  append([]string(nil), fake.AgentRunCalls...),
		ListITermCalls: fake.ListITermCalls,
	}, nil
}
```
