# grok-fork — two-mode CLI + launch (L2 library)

Classic TDD doctests for `github.com/xhd2015/agent-pro/agent/grok/fork`.
**RED** until the implementer lands `fork.Main` (and `fork.Run` / `fork.Fork`).

Harness calls the **library**, not `go build -C cmd ./grok-fork`. Thin
`cmd/grok-fork` glue (`Error: %v\n`, `os.Exit(1)`) is out of this tree.

# DSN (Domain Specific Notion)

**Participants**

- **Caller** — doctest harness invoking `fork.Main(args, opts)`.
- **`fork.Main`** — parses argv (no binary name), writes to `opts.Stdout` /
  `opts.Stderr`, returns `error`. Does **not** `os.Exit`. Does **not** print
  the `Error:` prefix (thin `cmd/grok-fork` does that later).
- **Ancestor resolve** — Mode A uses `procresolve.FindAncestorGrok` +
  `ResolveFromAncestors` via `opts.ListProcs` / `opts.Lsof` / `opts.PID`
  (CLI `--pid` overrides). Session id from open files only.
- **Session store** — `groksessions.Info(opts.GrokHome, id)` for cwd. Tests
  seed `$GROK_HOME/sessions/<url.PathEscape(absCWD)>/<uuid>/summary.json`.
  Main must **not** call `os.Getenv("GROK_HOME")` / `agenttty.GrokHome()`.
- **Mode A launcher** — `opts.OpenInNewTerminal(dir, followUp)` (production:
  `iterm2.OpenConfig` + `ModeForceNew`). Follow-up is
  `<opts.Executable()> --session-id <id>` plus optional `GROK_HOME=…` prefix
  (when that key is in `opts.Env`) and `--dir` (when the flag is set).
  Executable path is `os.Executable()` in production, injected here.
- **Mode B runner** — `opts.RunForeground(bin, argv, dir, env)` with
  `argv = --resume <id> --fork-session`. Bin is `opts.GrokBin` when set,
  else `LookPath("grok")`. Tests use basename `llm-mock-run-grok`.
- **Recorders / exec** — most leaves record Open/Foreground calls. The
  Mode B exec leaf actually runs a session-cached `llm-mock-run-grok` with
  child-only `LLM_MOCK_RUN_GROK_COMMAND` from `opts.Env`.

**Behaviors**

- Bare `grok-fork` (no `--session-id`) = Mode A: always a new iTerm2 window;
  never `grok --resume` in the follow-up.
- `grok-fork --session-id ID` = Mode B: current terminal; never
  `OpenInNewTerminal`.
- `--pid` + `--session-id` together → error.
- `--color` + `--no-color` → error `cannot be specified together`.
- Unknown flag → error mentioning `--help`.
- Extra positional args → error.
- No `-n` / `--new-terminal` on this binary.
- `--color` paints green `Opened`, gray dry-run labels; `--no-color` and
  auto-on-pipe (buffers) have no ANSI.
- Empty session cwd and no `--dir` → error containing `pass --dir`.
- Unknown Mode B id → error containing `grok session not found`.

## Locked contract

```text
func Main(args []string, opts *Options) error

type Options struct {
    Stdout, Stderr io.Writer
    PID            int    // default start pid when --pid omitted
    GrokHome       string // session lookup home (not process env)
    ListProcs      func() []procresolve.Proc
    Lsof           func(pid int) []string
    OpenInNewTerminal func(dir, followUp string) error
    RunForeground     func(bin string, argv []string, dir string, env []string) error
    GrokBin     string
    LookPath    func(file string) (string, error)
    Executable  func() (string, error)
    Env         []string // injected; GROK_HOME forwarded in Mode A follow-up
}

# Mode A dry-run stdout (trailing newline):
Would open new iTerm2 window
  ancestor pid: 4242
  grok id:      019f283b-aaaa-7aaa-aaaa-aaaaaaaaaaaa
  cwd:          <abs>
  command:      <executable> --session-id 019f283b-aaaa-7aaa-aaaa-aaaaaaaaaaaa

# Mode A success stdout:
Opened new window; launching grok-fork --session-id 019f283b-aaaa-7aaa-aaaa-aaaaaaaaaaaa

# Mode B dry-run stdout:
Would fork grok session
  grok id:   019f283b-aaaa-7aaa-aaaa-aaaaaaaaaaaa
  cwd:       <abs>
  command:   <grok-bin> --resume 019f283b-aaaa-7aaa-aaaa-aaaaaaaaaaaa --fork-session
  terminal:  current

# error fragments (returned error; no Error: prefix from Main)
no ancestor grok
session not resolved
pass --dir
--pid and --session-id cannot be specified together
--color and --no-color cannot be specified together
--help          (unknown flag)
unexpected argument | extra argument
grok session not found
```

## Version

0.0.2

## Decision Tree

```
agent/grok/fork/tests/
├── DOCTEST.md
├── SETUP.md
├── help/
│   ├── long-flag/                       # --help
│   └── short-flag/                      # -h
├── mode-a/                              # bare grok-fork
│   ├── dry-run/
│   ├── launch/
│   │   ├── default/
│   │   ├── dir-override/
│   │   ├── grok-home-env/
│   │   ├── pid-select/
│   │   ├── nested-nearest/
│   │   └── quoted-executable/
│   └── error/
│       ├── no-ancestor/
│       ├── session-unresolved/
│       └── empty-cwd-no-dir/
├── mode-b/                              # --session-id
│   ├── dry-run/
│   ├── launch-record/
│   │   ├── default/
│   │   └── dir-override/
│   ├── launch-exec/                     # real llm-mock-run-grok + hook
│   └── error/
│       └── unknown-session/
├── flags/
│   ├── pid-and-session-id/
│   ├── color-and-no-color/
│   ├── unknown-flag/
│   └── extra-positional/
└── color/
    ├── force-on-opened/                 # --color green Opened
    ├── force-on-dry-run-labels/         # --color gray labels
    └── force-off/                       # --no-color, no ANSI
```

Parameter ranking (most → least significant):

1. **CLI mode** — help | Mode A | Mode B | parse-time flags | color
2. **Outcome** — dry-run | launch | error
3. **Launch details** — dir / GROK_HOME / pid / nested / quoting / exec

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `help/long-flag` | `--help` exit 0; mentions required flags; no `-n` / `--new-terminal` |
| 2 | `help/short-flag` | `-h` same usage contract |
| 3 | `mode-a/dry-run` | plan: ancestor 4242, grok id, cwd, executable `--session-id`; no open |
| 4 | `mode-a/launch/default` | one OpenInNewTerminal; dir=session cwd; success stdout; exit 0 |
| 5 | `mode-a/launch/dir-override` | open dir + follow-up `--dir` are the override |
| 6 | `mode-a/launch/grok-home-env` | follow-up prefixes `GROK_HOME=…` |
| 7 | `mode-a/launch/pid-select` | `--pid` selects which ancestor session |
| 8 | `mode-a/launch/nested-nearest` | forks nearest session id, not topmost |
| 9 | `mode-a/launch/quoted-executable` | path with spaces is shell-quoted |
| 10 | `mode-a/error/no-ancestor` | error `no ancestor grok`; no open; exit 1 |
| 11 | `mode-a/error/session-unresolved` | grok ancestor, no Lsof session; exit 1 |
| 12 | `mode-a/error/empty-cwd-no-dir` | empty info.cwd, no `--dir`; `pass --dir` |
| 13 | `mode-b/dry-run` | plan grok-bin `--resume` `--fork-session`; terminal current; no open |
| 14 | `mode-b/launch-record/default` | RunForeground argv + dir; basename `llm-mock-run-grok`; no open |
| 15 | `mode-b/launch-record/dir-override` | foreground dir is `--dir` |
| 16 | `mode-b/launch-exec` | exec built mock + hook; exit 0; marker file |
| 17 | `mode-b/error/unknown-session` | `grok session not found`; no open/foreground |
| 18 | `flags/pid-and-session-id` | `--pid` + `--session-id` error |
| 19 | `flags/color-and-no-color` | `--color` + `--no-color` error |
| 20 | `flags/unknown-flag` | error mentions `--help` |
| 21 | `flags/extra-positional` | leftover arg error |
| 22 | `color/force-on-opened` | `--color` success token has green ANSI |
| 23 | `color/force-on-dry-run-labels` | `--color --dry-run` gray labels |
| 24 | `color/force-off` | `--no-color` has no ANSI |

## How to Run

```sh
doctest vet ./agent/grok/fork/tests
doctest test ./agent/grok/fork/tests

doctest test -v ./agent/grok/fork/tests/help/long-flag
doctest test -v ./agent/grok/fork/tests/mode-a/dry-run
doctest test -v ./agent/grok/fork/tests/mode-a/launch/default
doctest test -v ./agent/grok/fork/tests/mode-b/launch-exec
```

Classic TDD: compile-fail or assert-fail until `agent/grok/fork` exists.

```go
import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/agent/grok/fork"
	"github.com/xhd2015/agent-pro/pkgs/procresolve"
	"github.com/xhd2015/doctest/session"
)

// FixtureProc is one process row in the injectable snapshot.
type FixtureProc struct {
	PID  int
	PPID int
	Cmd  string
}

// OpenCall is one recorded OpenInNewTerminal invocation.
type OpenCall struct {
	Dir      string
	FollowUp string
}

// ForegroundCall is one recorded RunForeground invocation.
type ForegroundCall struct {
	Bin  string
	Argv []string
	Dir  string
	Env  []string
}

// Request drives fork.Main through injectable opts.
type Request struct {
	Args       []string
	PID        int
	Procs      []FixtureProc
	OpenFiles  map[int][]string
	GrokHome   string
	Workspace  string
	OverrideDir string
	TempDir    string
	Executable string
	GrokBin    string
	Env        []string
	ExecMock   bool
	HookMarker string
}

// Response observes Main’s writers, returned error, and launch recorders.
type Response struct {
	ExitCode        int
	Stdout          string
	Stderr          string
	Err             error
	OpenCalls       []OpenCall
	ForegroundCalls []ForegroundCall
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d

	procs := make([]procresolve.Proc, 0, len(req.Procs))
	for _, p := range req.Procs {
		procs = append(procs, procresolve.Proc{PID: p.PID, PPID: p.PPID, Cmd: p.Cmd})
	}
	snap := procs
	files := req.OpenFiles
	if files == nil {
		files = map[int][]string{}
	}

	resp := &Response{}
	var stdout, stderr bytes.Buffer

	opts := &fork.Options{
		Stdout:   &stdout,
		Stderr:   &stderr,
		PID:      req.PID,
		GrokHome: req.GrokHome,
		ListProcs: func() []procresolve.Proc {
			return snap
		},
		Lsof: func(pid int) []string {
			return files[pid]
		},
		OpenInNewTerminal: func(dir, followUp string) error {
			resp.OpenCalls = append(resp.OpenCalls, OpenCall{Dir: dir, FollowUp: followUp})
			return nil
		},
		GrokBin: req.GrokBin,
		Executable: func() (string, error) {
			return req.Executable, nil
		},
		Env: append([]string(nil), req.Env...),
	}

	if req.ExecMock {
		opts.RunForeground = func(bin string, argv []string, dir string, env []string) error {
			resp.ForegroundCalls = append(resp.ForegroundCalls, ForegroundCall{
				Bin: bin, Argv: argv, Dir: dir, Env: append([]string(nil), env...),
			})
			return execGrokMock(t, bin, argv, dir, env)
		}
	} else {
		opts.RunForeground = func(bin string, argv []string, dir string, env []string) error {
			resp.ForegroundCalls = append(resp.ForegroundCalls, ForegroundCall{
				Bin: bin, Argv: argv, Dir: dir, Env: append([]string(nil), env...),
			})
			return nil
		}
	}

	err := fork.Main(req.Args, opts)
	resp.Stdout = stdout.String()
	resp.Stderr = stderr.String()
	resp.Err = err
	if err != nil {
		resp.ExitCode = 1
	}
	return resp, nil
}

func execGrokMock(t *testing.T, bin string, argv []string, dir string, env []string) error {
	t.Helper()
	if bin == "" {
		t.Fatal("ExecMock: empty grok bin")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, argv...)
	cmd.Dir = dir
	childEnv := append([]string{}, os.Environ()...)
	childEnv = append(childEnv, env...)
	cmd.Env = childEnv
	out, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("llm-mock-run-grok timed out: %v\n%s", ctx.Err(), out)
	}
	if err != nil {
		return err
	}
	return nil
}
```
