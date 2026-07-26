# codex-show-status CLI Doctests

End-to-end doctests for the standalone `codex-show-status` CLI
(`script/codex/show-status`). The tool launches an **ephemeral tty-watch session**,
waits for the Codex prompt, submits `/status`, snapshots scrollback, parses usage
fields, prints exactly three stdout lines, and **always kills** the session:

```
Monthly usage: 58%
Credits used: 6519 of 11250
Next reset: 08:00 on 1 Aug
```

# DSN (Domain Specific Notion)

**Participants**

- **Caller** — shell or user invoking `codex-show-status` (or `go run ./script/codex/show-status`).
- **codex-show-status CLI** — thin wrapper calling `agent/codex/tty.FetchStatusWithOptions`;
  prints canonical three-line stdout on success.
- **tty-watch** — embedded PTY session manager invoked as subprocess (`run`, `send`,
  `snapshot`, `kill`); registry lives under `$TTY_WATCH_HOME/registry/`.
- **Codex TUI** — interactive Codex process inside the tty-watch PTY (production binary
  on PATH or a fake script).
- **Fake TUI hook** — `CODEX_SHOW_STATUS_COMMAND` replaces Codex argv for deterministic
  tests (same pattern as `GROK_SHOW_USAGE_COMMAND` / `AGENT_RUN_CODEX_TTY_COMMAND`).
- **Test harness** — builds `codex-show-status` and `tty-watch` once per process,
  runs CLI with isolated `TTY_WATCH_HOME`, optional fake hook, and timeout overrides.

**Behaviors**

- Ephemeral session id defaults to `codex-status-usage` (`CODEX_SHOW_STATUS_SESSION_ID`).
- Flow: `tty-watch run --session-id <id> codex` (detached) → wait for prompt (~5s) →
  `tty-watch send <id> $'/status\n\r'` → sleep ~1s → `tty-watch snapshot <id>` →
  `tty-watch kill <id>` (even on error paths).
- `CODEX_SHOW_STATUS_COMMAND` set → parse as shell words and use instead of codex path.
- Parse snapshot: `Monthly credit limit: N% left` → usage `(100-N)%`; credits
  `X of Y credits used`; reset from `(resets …)` on monthly line.
- Exit 0 on success with exactly three stdout lines; stderr only on errors.
- `CODEX_SHOW_STATUS_TIMEOUT` overrides max wait seconds (default 60).

## Version

0.0.2

## Decision Tree

```
script/codex/show-status/tests/
├── DOCTEST.md
├── SETUP.md                           # build codex-show-status + tty-watch, fake TUI helpers
├── success/
│   ├── SETUP.md                       # default fake TUI hook
│   ├── prints-status-lines/           # exit 0; stdout has three canonical lines exactly
│   ├── parses-percent-left/           # 30% left → 70% monthly usage
│   ├── extra-noise/                   # MCP warnings + tips before status box
│   └── kills-session/                 # registry entry gone after fetch completes
├── errors/
│   ├── SETUP.md                       # error assertion helpers
│   ├── codex-not-found/               # no codex on PATH, no hook → stderr mentions codex
│   ├── timeout-no-status/             # fake never prints status fields → stderr timeout
│   └── malformed-output/              # fake prints garbage → stderr mentions parse
└── real-codex/                        # label: real-codex, slow (skipped by default)
    └── fetches-live-status/           # live codex; pattern assertions, not exact values
```

Parameter ranking (most → least significant):

1. **Outcome** — success vs errors vs real-codex (live backend)
2. **Runner backend** — fake TUI (`CODEX_SHOW_STATUS_COMMAND`) vs production codex
3. **Fake TUI variant** — default fixture, custom percent-left, extra noise, no status, malformed
4. **Session cleanup** — `kills-session` asserts tty-watch registry pruned after fetch
5. **Timeout** — default (60s) vs shortened (`CODEX_SHOW_STATUS_TIMEOUT` for busy fake)

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `success/prints-status-lines` | Fake codex returns status after `/status`; exit 0; stdout three lines exactly (RED) |
| 2 | `success/parses-percent-left` | Fake with `30% left`; stdout `70%` monthly usage (RED) |
| 3 | `success/extra-noise` | MCP warnings before status box; stdout still three canonical lines (RED) |
| 4 | `success/kills-session` | After successful fetch, `codex-status-usage` absent from tty-watch registry (RED) |
| 5 | `errors/codex-not-found` | No codex on PATH, no hook; exit ≠0; stderr mentions codex (RED) |
| 6 | `errors/timeout-no-status` | Fake never prints status fields; exit ≠0; stderr mentions timeout (RED) |
| 7 | `errors/malformed-output` | Fake prints garbage; exit ≠0; stderr mentions parse (RED) |
| 8 | `real-codex/fetches-live-status` | Real codex on PATH; stdout matches status line patterns (`label: real-codex, slow`) |

## How to Run

```sh
# Discovery skips labeled e2e/heavy/slow leaves by default.
# Run e2e / full suite explicitly when needed:
doctest test ./script/codex/show-status/tests                    # discovery (skips labeled e2e/heavy/slow)
doctest test --label e2e ./script/codex/show-status/tests
doctest test --label-all ./script/codex/show-status/tests

# CI / default — fake TUI only
doctest vet ./script/codex/show-status/tests
doctest test ./script/codex/show-status/tests
doctest test -v ./script/codex/show-status/tests/success/prints-status-lines

# Optional — real codex (requires codex on PATH)
doctest test --label real-codex ./script/codex/show-status/tests/real-codex
doctest test --label slow ./script/codex/show-status/tests/real-codex/fetches-live-status
```

```go
import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

type Request struct {
	RepoRoot          string
	TempDir           string
	Bin               string   // path to built codex-show-status binary
	TTYWatchBin       string   // path to built tty-watch binary (prepended to PATH)
	TTYWatchHome      string   // isolated TTY_WATCH_HOME for registry assertions
	SessionID         string   // CODEX_SHOW_STATUS_SESSION_ID; default codex-status-usage
	Env               []string // extra env vars (KEY=value)
	ShowStatusCommand string   // CODEX_SHOW_STATUS_COMMAND; empty = do not set
	SkipFakeCommand   bool     // true: omit CODEX_SHOW_STATUS_COMMAND (production / codex-not-found)
	TimeoutSeconds    string   // CODEX_SHOW_STATUS_TIMEOUT; empty = default in CLI
	MinimalPATH       bool     // true: PATH with no codex binary (tty-watch still reachable)
}

type Response struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Err      error
}

var (
	builtBinOnce      sync.Once
	builtBinPath      string
	builtBinErr       error
	builtTTYWatchOnce sync.Once
	builtTTYWatchPath string
	builtTTYWatchErr  error
)

func findModuleRoot(start string) (string, error) {
	if start == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		start = wd
	}
	for dir := start; ; dir = filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		if filepath.Dir(dir) == dir {
			return "", fmt.Errorf("could not find module root (go.mod) above %s", start)
		}
	}
}

func buildShowStatus(t *testing.T, d *session.Doctest) (string, error) {
	t.Helper()
	builtBinOnce.Do(func() {
		repoRoot, err := findModuleRoot(d.DOCTEST_ROOT)
		if err != nil {
			builtBinErr = err
			return
		}
		tmp, err := os.MkdirTemp("", "codex-show-status-doctest-*")
		if err != nil {
			builtBinErr = err
			return
		}
		binPath := filepath.Join(tmp, "codex-show-status")
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "go", "build", "-o", binPath, "./script/codex/show-status")
		cmd.Dir = repoRoot
		var be bytes.Buffer
		cmd.Stderr = &be
		if err := cmd.Run(); err != nil {
			builtBinErr = fmt.Errorf("go build ./script/codex/show-status: %w\n%s", err, be.String())
			return
		}
		builtBinPath = binPath
	})
	return builtBinPath, builtBinErr
}

func buildTTYWatch(t *testing.T, d *session.Doctest) (string, error) {
	t.Helper()
	builtTTYWatchOnce.Do(func() {
		repoRoot, err := findModuleRoot(d.DOCTEST_ROOT)
		if err != nil {
			builtTTYWatchErr = err
			return
		}
		tmp, err := os.MkdirTemp("", "tty-watch-doctest-*")
		if err != nil {
			builtTTYWatchErr = err
			return
		}
		binPath := filepath.Join(tmp, "tty-watch")
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "go", "build", "-o", binPath, "github.com/xhd2015/tty-watch/cmd/tty-watch")
		cmd.Dir = repoRoot
		var be bytes.Buffer
		cmd.Stderr = &be
		if err := cmd.Run(); err != nil {
			builtTTYWatchErr = fmt.Errorf("go build github.com/xhd2015/tty-watch/cmd/tty-watch: %w\n%s", err, be.String())
			return
		}
		builtTTYWatchPath = binPath
	})
	return builtTTYWatchPath, builtTTYWatchErr
}

func withoutEnvKey(env []string, key string) []string {
	prefix := key + "="
	out := make([]string, 0, len(env))
	for _, e := range env {
		if strings.HasPrefix(e, prefix) {
			continue
		}
		out = append(out, e)
	}
	return out
}

func mergeEnv(base []string, extra ...string) []string {
	keys := make(map[string]struct{})
	out := make([]string, 0, len(base)+len(extra))
	for _, e := range base {
		if i := strings.IndexByte(e, '='); i > 0 {
			keys[e[:i]] = struct{}{}
		}
		out = append(out, e)
	}
	for _, e := range extra {
		if i := strings.IndexByte(e, '='); i > 0 {
			k := e[:i]
			if _, ok := keys[k]; ok {
				continue
			}
			keys[k] = struct{}{}
		}
		out = append(out, e)
	}
	return out
}

func defaultSessionID(req *Request) string {
	if strings.TrimSpace(req.SessionID) != "" {
		return strings.TrimSpace(req.SessionID)
	}
	return "codex-status-usage"
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	if req.Bin == "" {
		t.Fatal("req.Bin not set; root Setup must build codex-show-status")
	}
	if req.TTYWatchBin == "" {
		t.Fatal("req.TTYWatchBin not set; root Setup must build tty-watch")
	}
	if req.TTYWatchHome == "" {
		t.Fatal("req.TTYWatchHome not set; root Setup must set isolated TTY_WATCH_HOME")
	}

	timeout := 75 * time.Second
	if req.TimeoutSeconds != "" {
		if sec, err := time.ParseDuration(req.TimeoutSeconds + "s"); err == nil {
			timeout = sec + 20*time.Second
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, req.Bin)
	cmd.Dir = req.TempDir
	if cmd.Dir == "" {
		cmd.Dir = t.TempDir()
	}

	env := os.Environ()
	if req.MinimalPATH {
		env = withoutEnvKey(env, "PATH")
		env = append(env, "PATH="+filepath.Join(req.TempDir, "empty-path")+string(os.PathListSeparator)+filepath.Dir(req.TTYWatchBin))
	} else {
		env = withoutEnvKey(env, "PATH")
		env = append(env, "PATH="+filepath.Dir(req.TTYWatchBin)+string(os.PathListSeparator)+os.Getenv("PATH"))
	}
	env = mergeEnv(env, req.Env...)

	env = withoutEnvKey(env, "TTY_WATCH_HOME")
	env = append(env, "TTY_WATCH_HOME="+req.TTYWatchHome)

	if !req.SkipFakeCommand && req.ShowStatusCommand != "" {
		env = withoutEnvKey(env, "CODEX_SHOW_STATUS_COMMAND")
		env = append(env, "CODEX_SHOW_STATUS_COMMAND="+req.ShowStatusCommand)
	}
	if req.SkipFakeCommand {
		env = withoutEnvKey(env, "CODEX_SHOW_STATUS_COMMAND")
	}
	if req.TimeoutSeconds != "" {
		env = withoutEnvKey(env, "CODEX_SHOW_STATUS_TIMEOUT")
		env = append(env, "CODEX_SHOW_STATUS_TIMEOUT="+req.TimeoutSeconds)
	}
	sid := defaultSessionID(req)
	env = withoutEnvKey(env, "CODEX_SHOW_STATUS_SESSION_ID")
	env = append(env, "CODEX_SHOW_STATUS_SESSION_ID="+sid)

	cmd.Env = env

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	resp := &Response{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
		Err:    err,
	}
	if err != nil {
		if ctx.Err() != nil {
			return resp, ctx.Err()
		}
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			resp.ExitCode = exitErr.ExitCode()
		}
	}
	return resp, nil
}
```