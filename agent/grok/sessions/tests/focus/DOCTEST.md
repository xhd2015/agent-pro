# `agent-pro grok session focus`

`agent-pro grok session focus <session-id> [--index N]` focuses the already-open
iTerm2 tab that hosts a live Grok CLI process for that session. It never
creates a window, tab, or session. Matching is by live PID → TTY, not by cwd.

# DSN (Domain Specific Notion)

## Participants

- **User** invokes `grok session focus` with a session id and optional candidate index.
- **Focus CLI** validates arguments, prints help or success, and returns errors.
- **Focus service** finds the Grok session, scans live grok PIDs, walks process TTYs, and selects one iTerm tab.
- **Injected process and iTerm snapshot** supplies deterministic PIDs, open files, and tabs.

## Behaviors

- User asks for help; CLI prints focus usage and does not scan processes or iTerm.
- Parent session help names the focus command.
- Missing session id or a non-integer `--index` is a usage error; iTerm is not listed.
- Unknown Grok session, no live grok PID, or no iTerm tab on the process TTY all return `not found` and never focus.
- One hosting tab is focused and reported as `focused: window W, tab T`.
- Two live PIDs that share one TTY still count as one tab.
- Several tabs that host the same session without `--index` list stable 0-based candidates; an explicit valid index focuses that row; an invalid index lists the candidates and does not focus.

## Version

0.0.2

## Decision Tree

```text
focus/
├── help/
│   ├── parent-lists-focus/
│   └── focus-usage/
├── validation/
│   ├── missing-session-id/
│   └── non-integer-index/
├── unknown-session/
└── matches/
    ├── no-live-pid/
    ├── none/
    ├── exactly-one/
    ├── same-tty-collapsed/
    └── multiple/
        ├── requires-index/
        ├── selects-index/
        └── index-oob/
```

## Test Index

| Leaf | Contract |
|---|---|
| `help/parent-lists-focus/` | Parent session help line names `focus` and `<session-id>`. |
| `help/focus-usage/` | `focus --help` succeeds and documents `<session-id>` plus `--index`. |
| `validation/missing-session-id/` | No session id is a usage error without listing iTerm. |
| `validation/non-integer-index/` | A non-numeric index is a parse error without listing iTerm. |
| `unknown-session/` | Unknown UUID → `not found`; iTerm is not listed. |
| `matches/no-live-pid/` | Known session, no grok open-file hit → `not found`; iTerm is not listed. |
| `matches/none/` | Live grok PID + TTY, zero iTerm matches → `not found`; Focus never called. |
| `matches/exactly-one/` | One TTY match is focused and reported. |
| `matches/same-tty-collapsed/` | Two grok PIDs on one TTY → one tab; no `--index` required. |
| `matches/multiple/requires-index/` | Two tabs for the same session list candidates and never focus. |
| `matches/multiple/selects-index/` | `--index 1` focuses exactly candidate 1. |
| `matches/multiple/index-oob/` | Invalid index is fatal, lists valid choices, and never focuses. |

## How to Run

```sh
doctest vet ./agent/grok/sessions/tests/focus
doctest test ./agent/grok/sessions/tests/focus
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
	Args       []string
	GrokHome   string
	TempDir    string
	ProjectDir string
	SessionID  string
	Procs      []sessions.FocusProc
	OpenFiles  map[int][]string
	ITerm      []iterm2.SessionRef
	ParentHelp bool
}

type Response struct {
	Stdout         string
	Err            error
	Focused        []string
	ListITermCalls int
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d
	if req.ParentHelp {
		return &Response{Stdout: sessions.FocusCommandHelpLine + "\n"}, nil
	}
	fake := &sessions.FocusFake{
		Procs:     req.Procs,
		OpenFiles: req.OpenFiles,
		ITerm:     req.ITerm,
	}
	var stdout bytes.Buffer
	err := sessions.RunFocus(req.Args, &stdout, req.GrokHome, fake.Opts())
	return &Response{
		Stdout:         stdout.String(),
		Err:            err,
		Focused:        fake.Focused,
		ListITermCalls: fake.ListITermCalls,
	}, nil
}
```
