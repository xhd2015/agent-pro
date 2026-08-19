# proc-resolve-cli — `agent-pro proc resolve` subprocess doctests

Classic TDD doctests for plan phase **P2** CLI surface:

```text
agent-pro proc resolve <pid> [--json] [--ascii-tree] [--no-enrich]
agent-pro proc resolve -h | --help
agent-pro proc --help
```

Library resolve / FormatTree / EnrichInfo are covered by `./tests/procresolve/`.
This tree only locks the **agent-pro CLI** wire: help text, JSON output, and
unknown-pid failure via a built `agent-pro` binary.

**Mode:** Classic TDD — RED until `cmd/agent-pro` grows a `proc` command that
calls `pkgs/procresolve`.

# DSN (Domain Specific Notion)

**Participants**

- **Caller** — shell or harness invoking `agent-pro` with argv after the binary
  name (e.g. `proc resolve 123 --json`).
- **agent-pro CLI** — `cmd/agent-pro` hand-rolled dispatch. New top-level
  `proc` command (or equivalent) routes `resolve` to the procresolve package.
- **`pkgs/procresolve`** — library used by the CLI: `ResolveFromPID`,
  `FormatTree`, optional enrich. CLI maps flags:
  - `--json` → machine-readable Result (no tree glyphs required)
  - `--ascii-tree` → FormatTree with ASCII connectors in human mode
  - `--no-enrich` → EnrichInfo false (default human path may enrich)
- **Process snapshot / Lsof** — production uses live listing; tests for the
  **JSON hit** path inject a deterministic snapshot via env
  `AGENT_PRO_PROCRESOLVE_TEST_SNAPSHOT` (JSON) so leaves never depend on live
  `ps`/`lsof`. Help and unknown-pid leaves do not need the env.
- **Test harness** — builds `agent-pro` once per `doctest test` session (file
  lock + session cache), runs subprocess with controlled env, captures
  stdout/stderr/exit code.

**Behaviors**

```
# help
agent-pro proc resolve -h | --help
agent-pro proc --help
  -> exit 0
  -> text mentions resolve and --json (and preferably --ascii-tree / --no-enrich)

# JSON resolve (fixture inject)
AGENT_PRO_PROCRESOLVE_TEST_SNAPSHOT='{"procs":[...],"open_files":{...}}'
agent-pro proc resolve <pid> --json
  -> exit 0
  -> stdout JSON includes kind + session id (or SessionID)
  -> stdout must NOT require Unicode tree glyphs ├── └── │

# unknown pid (live or empty snapshot)
agent-pro proc resolve 999999999
  -> exit != 0
  -> stderr mentions error (pid not found / similar)
```

**Test snapshot env contract (JSON hit leaf)**

```text
AGENT_PRO_PROCRESOLVE_TEST_SNAPSHOT
  {
    "procs": [{"pid":100,"ppid":1,"cmd":"/usr/local/bin/grok"}],
    "open_files": {
      "100": ["/tmp/fake-grok-home/.grok/sessions/2026-07/019fabcdef-1234-5678-9abc-def012345678/events.jsonl"]
    },
    "grok_home": "/tmp/fake-grok-home"
  }
```

When this env is set, CLI **must** use it for ListProcs/Lsof (and optional
homes) instead of live process inspection. Unset → production live path.

## Version

0.0.2

## Decision Tree

```
tests/proc-resolve-cli/
├── DOCTEST.md
├── SETUP.md
├── help/
│   ├── SETUP.md
│   ├── resolve-help/                    # proc resolve -h mentions resolve, --json
│   └── root-mentions-proc/              # agent-pro -h indexes proc + resolve (P5)
└── resolve/
    ├── SETUP.md
    ├── json-hit/                        # --json + fixture env → kind + session id
    └── unknown-pid/                     # missing pid → non-zero exit, stderr error
```

Parameter ranking (most → least significant):

1. **CLI outcome class** — help success | resolve JSON success | resolve error
2. **Flags** — `--json` on success path; help documents flags
3. **Data source** — fixture env (json-hit) vs live unknown pid

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `help/resolve-help` | `proc resolve -h` (or `proc --help`) exit 0; text mentions `resolve` and `--json` |
| 2 | `help/root-mentions-proc` | `agent-pro -h` exit 0; text mentions `proc` and `resolve` (P5 polish backfill) |
| 3 | `resolve/json-hit` | Fixture snapshot + `proc resolve 100 --json` → exit 0; stdout has kind grok + session uuid; no box-drawing tree glyphs required |
| 4 | `resolve/unknown-pid` | `proc resolve 999999999` → exit ≠ 0; stderr mentions not found / error |

## How to Run

```sh
doctest vet ./tests/proc-resolve-cli
doctest test ./tests/proc-resolve-cli

doctest test -v ./tests/proc-resolve-cli/help/resolve-help
doctest test -v ./tests/proc-resolve-cli/help/root-mentions-proc
doctest test -v ./tests/proc-resolve-cli/resolve/json-hit
doctest test -v ./tests/proc-resolve-cli/resolve/unknown-pid
```

```go
import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

// Fixture session id locked with library tree.
const fixtureGrokSessionID = "019fabcdef-1234-5678-9abc-def012345678"

// SnapshotProc is one process row in AGENT_PRO_PROCRESOLVE_TEST_SNAPSHOT.
type SnapshotProc struct {
	PID  int    `json:"pid"`
	PPID int    `json:"ppid"`
	Cmd  string `json:"cmd"`
}

// TestSnapshot is the JSON body for AGENT_PRO_PROCRESOLVE_TEST_SNAPSHOT.
type TestSnapshot struct {
	Procs     []SnapshotProc      `json:"procs"`
	OpenFiles map[string][]string `json:"open_files"` // pid string → paths
	GrokHome  string              `json:"grok_home,omitempty"`
	CodexHome string              `json:"codex_home,omitempty"`
}

// Request drives one agent-pro invocation.
type Request struct {
	Bin      string   // built agent-pro path (root Setup)
	RepoRoot string   // module root
	Args     []string // argv after binary, e.g. ["proc","resolve","100","--json"]
	// Snapshot, when non-nil, is JSON-encoded into AGENT_PRO_PROCRESOLVE_TEST_SNAPSHOT.
	Snapshot *TestSnapshot
	EnvExtra []string // extra KEY=VAL
}

// Response is the subprocess observation.
type Response struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d
	if req.Bin == "" {
		t.Fatal("req.Bin not set; root Setup must build agent-pro")
	}
	if len(req.Args) == 0 {
		t.Fatal("req.Args empty; leaf Setup must set CLI args")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, req.Bin, req.Args...)
	// Neutral cwd under temp so HOME-relative paths stay isolated if used.
	cmd.Dir = t.TempDir()
	env := append([]string{}, os.Environ()...)
	// Strip parent color noise for stable output.
	env = filterEnv(env, "NO_COLOR", "FORCE_COLOR", "CLICOLOR_FORCE")
	if req.Snapshot != nil {
		raw, err := json.Marshal(req.Snapshot)
		if err != nil {
			return nil, fmt.Errorf("marshal test snapshot: %w", err)
		}
		env = append(env, "AGENT_PRO_PROCRESOLVE_TEST_SNAPSHOT="+string(raw))
	}
	env = append(env, req.EnvExtra...)
	cmd.Env = env

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	resp := &Response{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}
	if err != nil {
		if ctx.Err() != nil {
			return resp, fmt.Errorf("agent-pro timed out: %w\nstderr:\n%s", ctx.Err(), resp.Stderr)
		}
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			resp.ExitCode = exitErr.ExitCode()
			// Non-zero exit is an observation for Assert, not a harness failure.
			return resp, nil
		}
		return resp, err
	}
	resp.ExitCode = 0
	return resp, nil
}

func filterEnv(env []string, dropKeys ...string) []string {
	drop := map[string]bool{}
	for _, k := range dropKeys {
		drop[k] = true
	}
	out := make([]string, 0, len(env))
	for _, e := range env {
		key := e
		if i := strings.IndexByte(e, '='); i >= 0 {
			key = e[:i]
		}
		if drop[key] {
			continue
		}
		out = append(out, e)
	}
	return out
}

// sessionCacheDir is shared across parallel leaves in one doctest test run.
func sessionCacheDir(sessionID string) string {
	return filepath.Join(os.TempDir(), "proc-resolve-cli-doctest-"+sessionID)
}

func withFileLock(lockPath string, fn func() error) error {
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	return fn()
}

func ensureStubDist(distDir string) error {
	// DistComplete needs non-empty index.html and at least one assets/* file.
	// placeholder.txt alone is not enough; always ensure a minimal SPA shell.
	if err := os.MkdirAll(filepath.Join(distDir, "assets"), 0755); err != nil {
		return err
	}
	const shell = `<!doctype html>
<html lang="en">
<head><meta charset="UTF-8"><title>agent-run</title></head>
<body>
<div id="root"></div>
</body>
</html>
`
	indexPath := filepath.Join(distDir, "index.html")
	needIndex := true
	if data, err := os.ReadFile(indexPath); err == nil {
		s := string(data)
		if strings.Contains(s, `id="root"`) || strings.Contains(s, "id='root'") {
			needIndex = false
		}
	}
	if needIndex {
		if err := os.WriteFile(indexPath, []byte(shell), 0644); err != nil {
			return err
		}
	}
	assetPath := filepath.Join(distDir, "assets", "doctest-stub.js")
	if st, err := os.Stat(assetPath); err != nil || st.Size() == 0 {
		if err := os.WriteFile(assetPath, []byte("/* doctest stub */\n"), 0644); err != nil {
			return err
		}
	}
	return nil
}

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
			if _, err := os.Stat(filepath.Join(dir, "cmd", "agent-pro")); err == nil {
				return dir, nil
			}
		}
		if filepath.Dir(dir) == dir {
			return "", fmt.Errorf("could not find module root (go.mod + cmd/agent-pro) above %s", start)
		}
	}
}

// buildAgentProOnce builds ./cmd/agent-pro into the session cache (file-locked).
func buildAgentProOnce(t *testing.T, sessionID, repoRoot string) (string, error) {
	t.Helper()
	cache := sessionCacheDir(sessionID)
	binPath := filepath.Join(cache, "agent-pro")
	ready := filepath.Join(cache, "binaries.ready")
	lock := filepath.Join(cache, "build.lock")

	err := withFileLock(lock, func() error {
		if fileExists(ready) && fileExists(binPath) {
			return nil
		}
		if err := os.MkdirAll(cache, 0o755); err != nil {
			return err
		}
		distDir := filepath.Join(repoRoot, "frontend", "dist")
		if err := ensureStubDist(distDir); err != nil {
			return fmt.Errorf("ensure frontend/dist stub: %w", err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, runtime.GOROOT()+"/bin/go", "build", "-C", "cmd", "-o", binPath, "./agent-pro")
		cmd.Dir = repoRoot
		var be bytes.Buffer
		cmd.Stderr = &be
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("go build -C cmd ./agent-pro: %w\n%s", err, be.String())
		}
		return os.WriteFile(ready, []byte("ok"), 0o644)
	})
	if err != nil {
		return "", err
	}
	return binPath, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
```
