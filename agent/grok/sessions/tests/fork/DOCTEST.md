# `agent-pro grok session fork`

Fork a Grok session via `grok --resume … --fork-session`. Session source is
exactly one of: positional `<session-id>`, `--tab`, or `--tab-index`.

Harness calls `sessions.RunFork` (library), not a built `agent-pro` binary.

# DSN (Domain Specific Notion)

**Participants**

- **Caller** — doctest harness invoking `sessions.RunFork(args, opts)`.
- **`RunFork`** — parses argv; resolves session (explicit or tab); dry-run or
  launch via injectables. No `os.Exit`; no `Error:` prefix.
- **Tab resolve** — shared `ResolveFromTab` (same as `session resolve --tab`).
- **Launch** — default current terminal; `-n` / `--new-window` (alias
  `--new-terminal`) opens a new iTerm window.

## Version

0.0.1

## Decision Tree

```text
fork/
├── help/
│   ├── parent-lists-fork/
│   └── fork-usage/
├── id/
│   ├── dry-run-current/
│   └── dry-run-new-window/
├── tab/
│   ├── dry-run/
│   └── hit-new-window/
└── validation/
    ├── missing-source/
    ├── id-and-tab/
    └── tab-and-tab-index/
```

## Test Index

| Leaf | Contract |
|------|----------|
| `help/parent-lists-fork/` | Parent help includes ForkCommandHelpLine. |
| `help/fork-usage/` | `--help` documents `--tab`, `--tab-index`, `--new-window`. |
| `id/dry-run-current/` | Explicit id dry-run; terminal current. |
| `id/dry-run-new-window/` | `-n` dry-run; terminal new iTerm2 window. |
| `tab/dry-run/` | `--tab 2 --dry-run` includes window/tab/tty + grok id. |
| `tab/hit-new-window/` | `--tab 2 -n` opens new window with resolved id. |
| `validation/missing-source/` | No id/tab → error. |
| `validation/id-and-tab/` | Positional id + `--tab` → error. |
| `validation/tab-and-tab-index/` | Both tab flags → error. |

## How to Run

```sh
doctest vet ./agent/grok/sessions/tests/fork
doctest test ./agent/grok/sessions/tests/fork
```

```go
import (
	"bytes"
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

type Request struct {
	Args             []string
	GrokHome         string
	TempDir          string
	ProjectDir       string
	SessionID        string
	FocusProcs       []sessions.FocusProc
	OpenFiles        map[int][]string
	ITerm            []iterm2.SessionRef
	CurrentSessionID string
	ControllingTTY   string
	ParentHelp       bool
}

type Response struct {
	Stdout         string
	Stderr         string
	Err            error
	Opened         []string
	RunForegroundN int
	LastBin        string
	LastArgv       []string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d
	if req.ParentHelp {
		return &Response{Stdout: sessions.ForkCommandHelpLine + "\n"}, nil
	}
	var stdout, stderr bytes.Buffer
	resp := &Response{}
	files := req.OpenFiles
	if files == nil {
		files = map[int][]string{}
	}
	focusSnap := append([]sessions.FocusProc(nil), req.FocusProcs...)
	itermSnap := append([]iterm2.SessionRef(nil), req.ITerm...)
	err := sessions.RunFork(req.Args, &sessions.ForkOpts{
		Stdout:   &stdout,
		Stderr:   &stderr,
		GrokHome: req.GrokHome,
		Env:      []string{"PATH=/usr/bin"},
		GrokBin:  "/usr/local/bin/grok",
		ListFocusProcs: func() []sessions.FocusProc {
			return append([]sessions.FocusProc(nil), focusSnap...)
		},
		Lsof: func(pid int) []string { return files[pid] },
		ListITerm: func() ([]iterm2.SessionRef, error) {
			return append([]iterm2.SessionRef(nil), itermSnap...), nil
		},
		CurrentSessionID: func() string { return req.CurrentSessionID },
		ControllingTTY:   func() string { return req.ControllingTTY },
		AncestorTTYs:     func() []string { return nil },
		OpenInNewWindow: func(dir, followUp string) error {
			resp.Opened = append(resp.Opened, dir+"|"+followUp)
			return nil
		},
		RunForeground: func(bin string, argv []string, dir string, env []string) error {
			resp.RunForegroundN++
			resp.LastBin = bin
			resp.LastArgv = append([]string(nil), argv...)
			return nil
		},
		LookPath: func(string) (string, error) { return "/usr/local/bin/grok", nil },
	})
	resp.Stdout = stdout.String()
	resp.Stderr = stderr.String()
	resp.Err = err
	return resp, nil
}
```
