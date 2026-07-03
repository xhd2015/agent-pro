# grok-show-usage CLI Doctests

End-to-end doctests for the standalone `grok-show-usage` CLI
(`script/grok/show-usage`). The tool launches interactive **grok** in a PTY with
`--always-approve --permission-mode=bypassPermissions`, waits for the input
prompt, submits `/usage show`, captures usage lines, and prints exactly two
stdout lines:

```
Weekly limit: 1%
Next reset: July 9, 16:55 PT
```

# DSN (Domain Specific Notion)

**Participants**

- **Caller** — shell or user invoking `grok-show-usage` (or `go run ./script/grok/show-usage`).
- **grok-show-usage CLI** — spawns grok in a pseudo-terminal, polls scrollback for
  the prompt (`Grok ›` or `›`), writes `/usage show\r`, waits for usage output,
  parses `Weekly limit:` and `Next reset:` lines, prints them to stdout.
- **Grok TUI** — interactive grok process inside the PTY (production binary on PATH
  or a fake script).
- **Fake TUI hook** — `GROK_SHOW_USAGE_COMMAND` replaces grok argv for deterministic
  tests (same pattern as `AGENT_RUN_GROK_TTY_COMMAND`).
- **Test harness** — builds the CLI once per process, runs it with isolated env
  (`GROK_SHOW_USAGE_COMMAND`, optional `GROK_SHOW_USAGE_TIMEOUT`).

**Behaviors**

- Default argv: `{grokPath, --always-approve, --permission-mode=bypassPermissions}`.
- `GROK_SHOW_USAGE_COMMAND` set → parse as shell words and use instead of grok path.
- Wait for prompt in scrollback, submit `/usage show` with `\r`.
- Poll scrollback until `Weekly limit:` appears (idle debounce ~500ms).
- Parse with regex; print exactly two lines to stdout; stderr only on errors.
- Exit 0 on success; non-zero on grok not found, timeout, or parse failure.
- `GROK_SHOW_USAGE_TIMEOUT` overrides max wait seconds (default 30).

## Version

0.0.2

## Decision Tree

```
script/grok/show-usage/tests/
├── DOCTEST.md
├── SETUP.md                           # build grok-show-usage, fake TUI helpers
├── success/
│   ├── SETUP.md                       # default fake TUI hook
│   ├── prints-usage-lines/            # exit 0; stdout has both fixture lines exactly
│   └── parses-percent-and-date/       # fake with different values; stdout matches
├── errors/
│   ├── SETUP.md                       # error assertion helpers
│   ├── grok-not-found/                # no grok on PATH, no hook → stderr mentions grok
│   ├── timeout-no-usage/              # fake never prints Weekly limit → stderr timeout
│   └── malformed-output/              # fake prints garbage → stderr mentions parse
└── real-grok/                         # label: real-grok, slow (skipped by default)
    └── fetches-live-usage/            # live grok; pattern assertions, not exact date
```

Parameter ranking (most → least significant):

1. **Outcome** — success vs errors vs real-grok (live backend)
2. **Runner backend** — fake TUI (`GROK_SHOW_USAGE_COMMAND`) vs production grok
3. **Fake TUI variant** — default fixture, custom values, no usage, malformed output
4. **Timeout** — default (30s) vs shortened (`GROK_SHOW_USAGE_TIMEOUT` for busy fake)

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `success/prints-usage-lines` | Fake grok returns usage after `/usage show`; exit 0; stdout both lines exactly (RED) |
| 2 | `success/parses-percent-and-date` | Fake with different limit/reset values; stdout matches fixture (RED) |
| 3 | `errors/grok-not-found` | No grok on PATH, no hook; exit ≠0; stderr mentions grok (RED) |
| 4 | `errors/timeout-no-usage` | Fake never prints `Weekly limit:`; exit ≠0; stderr mentions timeout (RED) |
| 5 | `errors/malformed-output` | Fake prints garbage; exit ≠0; stderr mentions parse (RED) |
| 6 | `real-grok/fetches-live-usage` | Real grok on PATH; stdout matches usage line patterns (`label: real-grok, slow`) |

## How to Run

```sh
# CI / default — fake TUI only
doctest vet ./script/grok/show-usage/tests
doctest test ./script/grok/show-usage/tests
doctest test -v ./script/grok/show-usage/tests/success/prints-usage-lines

# Optional — real grok (requires grok on PATH)
doctest test --label real-grok ./script/grok/show-usage/tests/real-grok
doctest test --label slow ./script/grok/show-usage/tests/real-grok/fetches-live-usage
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
)

type Request struct {
	RepoRoot         string
	TempDir          string
	Bin              string   // path to built grok-show-usage binary
	Env              []string // extra env vars (KEY=value)
	ShowUsageCommand string   // GROK_SHOW_USAGE_COMMAND; empty = do not set
	SkipFakeCommand  bool     // true: omit GROK_SHOW_USAGE_COMMAND (production / grok-not-found)
	TimeoutSeconds   string   // GROK_SHOW_USAGE_TIMEOUT; empty = default in CLI
	MinimalPATH      bool     // true: PATH with no grok binary
}

type Response struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Err      error
}

var (
	builtBinOnce sync.Once
	builtBinPath string
	builtBinErr  error
)

func findModuleRoot() (string, error) {
	start := os.Getenv("DOCTEST_ROOT")
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

func buildShowUsage(t *testing.T) (string, error) {
	t.Helper()
	builtBinOnce.Do(func() {
		repoRoot, err := findModuleRoot()
		if err != nil {
			builtBinErr = err
			return
		}
		tmp, err := os.MkdirTemp("", "grok-show-usage-doctest-*")
		if err != nil {
			builtBinErr = err
			return
		}
		binPath := filepath.Join(tmp, "grok-show-usage")
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "go", "build", "-o", binPath, "./script/grok/show-usage")
		cmd.Dir = repoRoot
		var be bytes.Buffer
		cmd.Stderr = &be
		if err := cmd.Run(); err != nil {
			builtBinErr = fmt.Errorf("go build ./script/grok/show-usage: %w\n%s", err, be.String())
			return
		}
		builtBinPath = binPath
	})
	return builtBinPath, builtBinErr
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

func Run(t *testing.T, req *Request) (*Response, error) {
	if req.Bin == "" {
		t.Fatal("req.Bin not set; root Setup must build grok-show-usage")
	}

	timeout := 45 * time.Second
	if req.TimeoutSeconds != "" {
		if sec, err := time.ParseDuration(req.TimeoutSeconds + "s"); err == nil {
			timeout = sec + 15*time.Second
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
		env = append(env, "PATH="+filepath.Join(req.TempDir, "empty-path"))
	}
	env = mergeEnv(env, req.Env...)

	if !req.SkipFakeCommand && req.ShowUsageCommand != "" {
		env = withoutEnvKey(env, "GROK_SHOW_USAGE_COMMAND")
		env = append(env, "GROK_SHOW_USAGE_COMMAND="+req.ShowUsageCommand)
	}
	if req.SkipFakeCommand {
		env = withoutEnvKey(env, "GROK_SHOW_USAGE_COMMAND")
	}
	if req.TimeoutSeconds != "" {
		env = withoutEnvKey(env, "GROK_SHOW_USAGE_TIMEOUT")
		env = append(env, "GROK_SHOW_USAGE_TIMEOUT="+req.TimeoutSeconds)
	}
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