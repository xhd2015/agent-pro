# `agent-pro grok session resolve`

Resolve a Grok session id by walking ancestors to the nearest grok runner
(open-file paths), or from a sibling iTerm2 tab (`--tab` / `--tab-index`).

Harness calls `sessions.RunResolve` (library), not a built `agent-pro` binary.

# DSN (Domain Specific Notion)

**Participants**

- **Caller** — doctest harness invoking `sessions.RunResolve(args, opts)`.
- **`RunResolve`** — parses argv, writes to `opts.Stdout` / `opts.Stderr`,
  returns `error`. Does **not** `os.Exit`. Does **not** print `agent-pro:` /
  `Error:` prefixes.
- **Ancestor resolve** — `FindAncestorGrok` + `ResolveFromAncestors` via
  `opts.ListProcs` / `opts.Lsof` / `opts.PID` (CLI `--pid` overrides).
  Session id from open files only (never cmdline `--resume` / `--session-id`).
- **Tab resolve** — `ResolveFromTab` via injected iTerm refs + FocusProc/TTY +
  Lsof. Selectors: `--tab` (1-based / next|left|right) or `--tab-index`
  (0-based). No wrap at edges.
- **Output shapes** — bare id (default); `-v` meta on stderr; `--dry-run`
  `[dry-run]` plan on stdout; `--json` encodes detail fields.

**Behaviors**

- Default success: one session id line on stdout.
- `--dry-run`: same discovery; plan lines with `[dry-run]` prefix; no bare-id
  success shape (go-best-practice dry-run: one pipeline, gate consumer effect).
- `-v`: bare id on stdout; detail fields on stderr.
- `--json`: indented JSON of detail fields; no bare-id line; `-v` meta skipped.
- Ancestor misses hard-fail: `no ancestor grok`, `session not resolved`, `pid not found`.
- Tab misses hard-fail: not in iTerm, edge/oob, no grok on tab, multiple
  unrelated grok. Parent + child subagent on the same tab resolves to parent.
- `--tab` and `--tab-index` mutually exclusive; `--pid` exclusive with tab flags.
- Extra positional args → error.
- Help: parent lists resolve; `resolve -h` prints ResolveHelp including tab flags.

## Locked contract

```text
func RunResolve(args []string, opts *ResolveOpts) error

# default stdout:
<session-id>\n

# ancestor dry-run stdout:
[dry-run] start pid:     <n>
[dry-run] ancestor pid:  <n>
[dry-run] runner pid:    <n>
[dry-run] would resolve: <session-id>
[dry-run] source:        open-files|open-files+tree
[dry-run] confidence:    hard

# tab dry-run stdout:
[dry-run] mode:          tab
[dry-run] window:        <id>
[dry-run] tab index:     <1-based>
[dry-run] tty:           <tty>
[dry-run] runner pid:    <n>
[dry-run] would resolve: <session-id>
[dry-run] source:        open-files
[dry-run] confidence:    hard
```

## Version

0.0.2

## Decision Tree

```text
resolve/
├── help/
│   ├── parent-lists-resolve/
│   └── resolve-usage/
├── hit/
│   ├── bare/
│   ├── verbose/
│   ├── dry-run/
│   ├── json/
│   ├── pid-select/
│   ├── from-grok-self/
│   └── nested-nearest/
├── miss/
│   ├── no-ancestor/
│   ├── session-unresolved/
│   └── unknown-pid/
├── flags/
│   └── unexpected-arg/
└── tab/
    ├── hit/
    │   ├── by-tab-1based/
    │   ├── by-tab-index/
    │   ├── next/
    │   ├── left/
    │   ├── wrapped-pty/
    │   ├── parent-subagent/
    │   └── dry-run/
    ├── miss/
    │   ├── no-grok/
    │   ├── multi-grok/
    │   ├── next-at-last/
    │   └── left-at-first/
    └── flags/
        ├── tab-and-tab-index/
        ├── pid-and-tab/
        └── tab-zero-invalid/
```

Parameter ranking (most → least significant):

1. **Outcome** — help | hit | miss | flags | tab
2. **Output / selector** — bare, verbose, dry-run, json, pid, self, nested, tab variants

## Test Index

| Leaf | Contract |
|------|----------|
| `help/parent-lists-resolve/` | Parent help line names `resolve`. |
| `help/resolve-usage/` | `resolve -h` documents `--pid`, `--tab`, `--tab-index`, `--dry-run`, `-v`, `--json`. |
| `hit/bare/` | Default stdout is bare session id. |
| `hit/verbose/` | Bare id on stdout; detail fields on stderr. |
| `hit/dry-run/` | `[dry-run]` plan; no bare-only stdout. |
| `hit/json/` | JSON includes session_id + verbose fields. |
| `hit/pid-select/` | `--pid` selects a different ancestor chain. |
| `hit/from-grok-self/` | Start at grok pid; ancestor is self. |
| `hit/nested-nearest/` | Nearer subagent grok wins. |
| `miss/no-ancestor/` | Error `no ancestor grok`. |
| `miss/session-unresolved/` | Ancestor grok but no Lsof hit → `session not resolved`. |
| `miss/unknown-pid/` | Error contains `pid not found`. |
| `flags/unexpected-arg/` | Extra positional → `unexpected argument`. |
| `tab/hit/by-tab-1based/` | `--tab 2` resolves sibling tab grok id. |
| `tab/hit/by-tab-index/` | `--tab-index 1` same as 1-based tab 2 in fixture. |
| `tab/hit/next/` | `--tab next` from first → second. |
| `tab/hit/left/` | `--tab left` from second → first. |
| `tab/hit/wrapped-pty/` | Grok on other PTY; match via ancestor shell TTY on tab. |
| `tab/hit/parent-subagent/` | Parent + child subagent on tab → parent id. |
| `tab/hit/dry-run/` | Tab dry-run plan includes mode/window/tab/tty. |
| `tab/miss/no-grok/` | Target tab has no grok → error. |
| `tab/miss/multi-grok/` | Two unrelated grok sessions on tab → refuse. |
| `tab/miss/next-at-last/` | `--tab next` on last tab → edge error. |
| `tab/miss/left-at-first/` | `--tab left` on first tab → edge error. |
| `tab/flags/tab-and-tab-index/` | Mutual exclusion error. |
| `tab/flags/pid-and-tab/` | `--pid` + `--tab` error. |
| `tab/flags/tab-zero-invalid/` | `--tab 0` invalid (1-based). |

## How to Run

```sh
doctest vet ./agent/grok/sessions/tests/resolve
doctest test ./agent/grok/sessions/tests/resolve
```

```go
import (
	"bytes"
	"testing"

	"github.com/xhd2015/agent-pro/agent/grok/sessions"
	"github.com/xhd2015/agent-pro/pkgs/procresolve"
	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

type Request struct {
	Args             []string
	PID              int
	Procs            []FixtureProc
	FocusProcs       []sessions.FocusProc
	OpenFiles        map[int][]string
	ITerm            []iterm2.SessionRef
	CurrentSessionID string
	ControllingTTY   string
	GrokHome         string
	TempDir          string
	ParentHelp       bool
	SessionMeta      map[string]sessions.TabSessionMeta
}

type FixtureProc struct {
	PID  int
	PPID int
	Cmd  string
}

type Response struct {
	Stdout string
	Stderr string
	Err    error
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d
	if req.ParentHelp {
		return &Response{Stdout: sessions.ResolveCommandHelpLine + "\n"}, nil
	}
	var stdout, stderr bytes.Buffer
	procs := make([]procresolve.Proc, 0, len(req.Procs))
	for _, p := range req.Procs {
		procs = append(procs, procresolve.Proc{PID: p.PID, PPID: p.PPID, Cmd: p.Cmd})
	}
	snap := procs
	files := req.OpenFiles
	if files == nil {
		files = map[int][]string{}
	}
	focusSnap := append([]sessions.FocusProc(nil), req.FocusProcs...)
	itermSnap := append([]iterm2.SessionRef(nil), req.ITerm...)
	resolveOpts := &sessions.ResolveOpts{
		Stdout: &stdout,
		Stderr: &stderr,
		PID:    req.PID,
		ListProcs: func() []procresolve.Proc {
			return append([]procresolve.Proc(nil), snap...)
		},
		Lsof: func(pid int) []string {
			return files[pid]
		},
		GrokHome: req.GrokHome,
		ListFocusProcs: func() []sessions.FocusProc {
			return append([]sessions.FocusProc(nil), focusSnap...)
		},
		ListITerm: func() ([]iterm2.SessionRef, error) {
			return append([]iterm2.SessionRef(nil), itermSnap...), nil
		},
		CurrentSessionID: func() string { return req.CurrentSessionID },
		ControllingTTY:   func() string { return req.ControllingTTY },
		AncestorTTYs:     func() []string { return nil },
	}
	if req.SessionMeta != nil {
		meta := req.SessionMeta
		resolveOpts.SessionMeta = func(sessionID string) (sessions.TabSessionMeta, bool) {
			m, ok := meta[sessionID]
			return m, ok
		}
	}
	err := sessions.RunResolve(req.Args, resolveOpts)
	return &Response{Stdout: stdout.String(), Stderr: stderr.String(), Err: err}, nil
}
```
