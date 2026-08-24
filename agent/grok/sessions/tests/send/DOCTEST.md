# `agent-pro grok session send`

`agent-pro grok session send <text> (--session-id <id> | --tab SEL | --tab-index N)`
types text into the live iTerm host via `iterm2.SendText` (same as
`kool iterm2 session … send`). Optional `--open` resumes like `open` when no
host exists, waits up to 120s for a hosting tab, then sends.

L2 only — injectable `SendFake` (procs / iTerm / SendText / open / clock).
No live AppleScript.

# DSN (Domain Specific Notion)

## Participants

- **User** invokes `grok session send` with text + session source.
- **Send CLI** (`RunSend`) validates args, prints help / sent / open+sent /
  dry-run plans, and returns errors without an `Error:` prefix.
- **Send service** finds the Grok session, discovers focus candidates (shared
  with focus/open/snapshot), optionally resumes, then calls injectable SendText.
- **Injected process / iTerm / SendText / open / clock** supplies boundaries.

## Behaviors

- Help documents `--session-id`, `--tab`, `--tab-index`, `--index`, `--no-submit`,
  `--focus`, `--no-ctrl-u`, `--open`, `--dir`, `--dry-run`, and key/`--text` flags.
- Parent session help names the send command.
- Missing text-or-key / missing session source are usage errors; SendText not invoked.
- Key sequence flags (`--enter`/`--up`/`--ctrl-c`/`--text`/…) compose in CLI order;
  positional text is always last.
- `--open` with `--tab`/`--tab-index` is a usage error.
- Unknown Grok session → `grok session not found`; never sends.
- One hosting tab → `sent to session <id>`; SendText with default opts.
- No live host without `--open` → hard error; SendText not called.
- `--open` with no host → open resume, wait for host, then send (two stdout lines).
- `--open` timeout → hard error after wait.
- `--open` + agent-run-managed live → send via agent-run (no grok --resume / SendText).
- `--open` + agent-run-managed exited → agent-run resume window (prompt in child).
- `--open --no-agent-run` forces bare grok --resume even when managed.
- Ambiguous agent-run mapping → warning + fall back to grok --resume.
- `--tab` resolves via `ResolveFromTab` then sends.
- Flag opts (`--no-submit` / `--focus` / `--no-ctrl-u`) plumb into SendText.
- `--dry-run` prints plan; never opens or SendText.

## Version

0.0.1

## Decision Tree

```text
send/
├── help/
│   ├── parent-lists-send/
│   └── send-usage/
├── validation/
│   ├── missing-text/
│   ├── missing-session-source/
│   └── open-with-tab/
├── unknown-session/
├── send/
│   ├── exactly-one/
│   ├── opts-flags/
│   └── no-live/
├── open/
│   ├── resume-then-send/
│   ├── timeout/
│   └── agent-run/
│       ├── live-send/
│       ├── resume-send/
│       ├── no-agent-run/
│       └── ambiguous-warn/
├── tab/
│   └── hit-send/
├── keys/
│   ├── ctrl-c-only/
│   ├── up-up-enter/
│   └── interleave-text-enter/
└── dry-run/
    └── plan/
```

## Test Index

| Leaf | Contract |
|---|---|
| `help/parent-lists-send/` | Parent help line names `send`. |
| `help/send-usage/` | `send --help` documents `--session-id`, `--open`, keys, `--text`. |
| `validation/missing-text/` | Missing text-or-key → usage error; no SendText. |
| `validation/missing-session-source/` | No session source → usage error. |
| `validation/open-with-tab/` | `--open --tab` rejected. |
| `unknown-session/` | Unknown UUID → not found; no SendText. |
| `send/exactly-one/` | One host → sent line; SendText once with defaults. |
| `send/opts-flags/` | Focus/NoSubmit/NoCtrlU plumbed. |
| `send/no-live/` | Hard error without `--open`. |
| `open/resume-then-send/` | Open then send; two stdout lines. |
| `open/timeout/` | Wait expires → timeout error. |
| `open/agent-run/live-send/` | Prefer live agent-run send; no grok --resume / SendText. |
| `open/agent-run/resume-send/` | Prefer agent-run resume window; no SendText. |
| `open/agent-run/no-agent-run/` | `--no-agent-run` forces bare grok --resume. |
| `open/agent-run/ambiguous-warn/` | Ambiguous mapping warns and falls back. |
| `tab/hit-send/` | `--tab 1` sends to resolved pane. |
| `keys/ctrl-c-only/` | `--ctrl-c` → `\x03`; NoCtrlU+NoSubmit. |
| `keys/up-up-enter/` | two Ups + Enter; key-only opts. |
| `keys/interleave-text-enter/` | `--up --text --enter` + positional last. |
| `dry-run/plan/` | Plan only; no SendText/open. |

## How to Run

```sh
doctest vet ./agent/grok/sessions/tests/send
doctest test ./agent/grok/sessions/tests/send
```

```go
import (
	"bytes"
	"testing"
	"time"

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
	AfterOpenHost    bool // AfterOpen adds live host for --open tests
	OpenWait         time.Duration
	UseFakeClock     bool
	NoAgentRun       bool
	AgentRunByID     map[string]*sessions.AgentRunOpenResult
	AgentRunErr      error
}

type Response struct {
	Stdout         string
	Stderr         string
	Err            error
	SendCalls      []sessions.SendCall
	Opened         []string
	AgentRunCalls  []string
	ListITermCalls int
	SleepCalls     int
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d
	if req.ParentHelp {
		return &Response{Stdout: sessions.SendCommandHelpLine + "\n"}, nil
	}
	fake := &sessions.SendFake{
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
	if req.AfterOpenHost {
		sid := req.SessionID
		proj := req.ProjectDir
		fake.AfterOpen = func(f *sessions.SendFake) {
			f.Procs = []sessions.FocusProc{
				{PID: 9001, PPID: 1, TTY: "ttys148", Cmd: "/usr/local/bin/grok"},
			}
			f.OpenFiles = map[int][]string{
				9001: {"/Users/fixture/.grok/sessions/%2Ftmp%2Fproj/" + sid + "/events.jsonl"},
			}
			f.ITerm = []iterm2.SessionRef{
				{WindowID: "3", WindowName: "worktrees", TabIndex: 1, SessionID: "w2t1p0", TTY: "/dev/ttys148"},
			}
			_ = proj
		}
	}
	if req.UseFakeClock {
		fake.InitClock(time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC))
	}
	var stdout, stderr bytes.Buffer
	opts := fake.SendOpts()
	opts.NoAgentRun = req.NoAgentRun
	if req.OpenWait > 0 {
		opts.OpenWait = req.OpenWait
	}
	err := sessions.RunSend(req.Args, &stdout, &stderr, req.GrokHome, opts)
	return &Response{
		Stdout:         stdout.String(),
		Stderr:         stderr.String(),
		Err:            err,
		SendCalls:      append([]sessions.SendCall(nil), fake.SendCalls...),
		Opened:         append([]string(nil), fake.Opened...),
		AgentRunCalls:  append([]string(nil), fake.AgentRunCalls...),
		ListITermCalls: fake.ListITermCalls,
		SleepCalls:     fake.SleepCalls,
	}, nil
}
```
