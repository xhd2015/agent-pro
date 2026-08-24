# `agent-pro grok session open`

`agent-pro grok session open (<session-id> | --tab SEL | --tab-index N) […]`
focuses the already-open iTerm2 tab that hosts a live Grok process for that
session when one exists. Otherwise it opens a new iTerm2 window and runs
`grok --resume <session-id>` (true resume, not fork). Tab source uses shared
`ResolveFromTab` / `ResolveSessionSource` (same as fork/resolve) and focuses
that tab without resume.

# DSN (Domain Specific Notion)

## Participants

- **User** invokes `grok session open` with a session id and optional flags.
- **Open CLI** (`RunOpen`) validates arguments, prints help / success / dry-run
  plans, and returns errors without an `Error:` prefix.
- **Open service** finds the Grok session, discovers focus candidates (shared
  with focus), focuses or resumes via injectable hooks.
- **Injected process / iTerm / OpenInNewWindow** supplies deterministic boundaries.

## Behaviors

- Help documents `<session-id>`, `--tab`, `--tab-index`, `--index`, `--dir`, `--no-agent-run`, `--dry-run`.
- Parent session help names the open command.
- Missing session source is a usage error; iTerm / open are not invoked.
- Unknown Grok session returns `grok session not found` and never focuses/opens.
- One hosting tab is focused and reported as `focused: window W, tab T`.
- Several tabs without `--index` list candidates and never focus or resume.
- No live host (or live but no iTerm match) resumes in a new window.
- Agent-run-managed live → focus via agent-run; exited → agent-run resume window.
- `--no-agent-run` forces bare grok --resume even when managed.
- Live PID with no iTerm match prints a `warning:` on stderr, then resumes.
- Empty session cwd without `--dir` is fatal.
- `--dry-run` prints the plan and never focuses or opens.
- `--tab` / `--tab-index` resolve via `ResolveFromTab` then focus that tab (no resume).
- Session id cannot combine with `--tab` / `--tab-index`; `--index` cannot combine with tab source.

## Version

0.0.1

## Decision Tree

```text
open/
├── help/
│   ├── parent-lists-open/
│   └── open-usage/
├── validation/
│   └── missing-session-id/
├── unknown-session/
├── focus/
│   ├── exactly-one/
│   └── multiple/
│       ├── requires-index/
│       └── selects-index/
├── resume/
│   ├── no-live-pid/
│   ├── live-no-iterm/
│   └── empty-cwd/
├── agent-run/
│   ├── focus-live/
│   ├── resume/
│   └── no-agent-run/
├── dry-run/
│   ├── focus/
│   └── resume/
└── tab/
    ├── hit-focus/
    ├── dry-run/
    ├── id-and-tab/
    ├── tab-and-tab-index/
    └── index-with-tab/
```

## Test Index

| Leaf | Contract |
|---|---|
| `help/parent-lists-open/` | Parent session help line names `open` and tab/id sources. |
| `help/open-usage/` | `open --help` documents `--tab`, resume-when-missing, `--index`. |
| `validation/missing-session-id/` | No session source is a usage error without listing iTerm / opening. |
| `unknown-session/` | Unknown UUID → `grok session not found`; no focus/open. |
| `focus/exactly-one/` | One TTY match is focused; OpenInNewWindow not called. |
| `focus/multiple/requires-index/` | Two tabs list candidates with `open` hint; never focus/open. |
| `focus/multiple/selects-index/` | `--index 1` focuses candidate 1. |
| `resume/no-live-pid/` | Known session, no live → OpenInNewWindow with `grok --resume`. |
| `resume/live-no-iterm/` | Live PID + zero iTerm → warning + resume. |
| `resume/empty-cwd/` | Empty cwd without `--dir` is fatal; no open. |
| `agent-run/focus-live/` | Managed live → focused via agent-run; no grok --resume. |
| `agent-run/resume/` | Managed exited → agent-run resume window. |
| `agent-run/no-agent-run/` | `--no-agent-run` forces bare grok --resume. |
| `dry-run/focus/` | Would focus; FocusITerm / OpenInNewWindow not called. |
| `dry-run/resume/` | Would open plan; OpenInNewWindow not called. |
| `tab/hit-focus/` | `--tab 2` focuses resolved tab; never resumes. |
| `tab/dry-run/` | `--tab 2 --dry-run` would focus; no side effects. |
| `tab/id-and-tab/` | Positional id + `--tab` errors. |
| `tab/tab-and-tab-index/` | Both tab flags error. |
| `tab/index-with-tab/` | `--index` with `--tab` errors. |

## How to Run

```sh
doctest vet ./agent/grok/sessions/tests/open
doctest test ./agent/grok/sessions/tests/open
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
	ParentHelp       bool
	NoAgentRun       bool
	AgentRunByID     map[string]*sessions.AgentRunOpenResult
	AgentRunErr      error
}

type Response struct {
	Stdout         string
	Stderr         string
	Err            error
	Focused        []string
	Opened         []string
	AgentRunCalls  []string
	ListITermCalls int
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d
	if req.ParentHelp {
		return &Response{Stdout: sessions.OpenCommandHelpLine + "\n"}, nil
	}
	fake := &sessions.OpenFake{
		FocusFake: sessions.FocusFake{
			Procs:     req.Procs,
			OpenFiles: req.OpenFiles,
			ITerm:     req.ITerm,
		},
		CurrentSessionID: req.CurrentSessionID,
		ControllingTTY:   req.ControllingTTY,
		AgentRunByID:     req.AgentRunByID,
		AgentRunErr:      req.AgentRunErr,
	}
	var stdout, stderr bytes.Buffer
	opts := fake.OpenOpts()
	opts.Stderr = &stderr
	opts.NoAgentRun = req.NoAgentRun
	err := sessions.RunOpen(req.Args, &stdout, &stderr, req.GrokHome, opts)
	return &Response{
		Stdout:         stdout.String(),
		Stderr:         stderr.String(),
		Err:            err,
		Focused:        fake.Focused,
		Opened:         fake.Opened,
		AgentRunCalls:  append([]string(nil), fake.AgentRunCalls...),
		ListITermCalls: fake.ListITermCalls,
	}, nil
}
```
