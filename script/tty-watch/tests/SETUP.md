# Scenario

**Feature**: tty-watch CLI manages embedded ptywrap sessions via isolated registry home

```
# build once, isolated TTY_WATCH_HOME per test
harness -> build tty-watch -> set TTY_WATCH_HOME -> phase dispatch via Run(req.Phase)
tty-watch -> embedded ptywrap + registry -> list/watch/snapshot/kill subcommands
doctest <- Response: exit codes, output, registry state, session reachability
```

## Preconditions

- `go` is available in PATH (to build `./script/tty-watch`). Tests skip otherwise.
- `script/tty-watch/main.go` exists with embedded ptywrap + registry logic (added by implementer).
- Each test uses an isolated `TTY_WATCH_HOME` temp directory.

## Steps

1. Build `tty-watch` once (cached across the process) via `buildTTYWatch`.
2. Set `req.Bin` and `req.TTYWatchHome` for every leaf.
3. Grouping `Setup` narrows `RunCommand`, phase-specific ids, or probe durations.
4. Leaf `Setup` sets `req.Phase` and scenario-specific fields.
5. `Run` dispatches to `ttywatchtest.Run` by phase key.
6. Leaf `Assert` checks output, registry state, or error messages.

## Context

- Default run requires a PTY; harness uses `creack/pty` for attach/detach/Ctrl-C tests.
- Detached sessions use Ctrl-] (`\x1d`) per tty-watch spec (not SIGINT).
- Registry lives at `$TTY_WATCH_HOME/registry/session-N.json`.
- All leaves are RED until implementer adds `script/tty-watch/main.go`.

```go
import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/script/tty-watch/ttywatchtest"
)

func Setup(t *testing.T, req *Request) error {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skipf("skipping: go not found in PATH")
	}
	req.Bin = buildTTYWatch(t)
	req.TTYWatchHome = isolatedHome(t)
	return nil
}

// assertRegistryHasSession fails when session id is missing from registry dir.
func assertRegistryHasSession(t *testing.T, home, sessionID string) {
	t.Helper()
	if !ttywatchtest.RegistryExists(home, sessionID) {
		t.Fatalf("registry missing %s under %s", sessionID, ttywatchtest.RegistryDir(home))
	}
}

// assertNoHostSessionID fails when combined host output leaks session-N id lines.
func assertNoHostSessionID(t *testing.T, combined string) {
	t.Helper()
	for _, line := range strings.Split(combined, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "session-") {
			t.Fatalf("host leaked session id on stdout/stderr: %q", trim)
		}
	}
}

// assertNonZeroExit fails when exit code is 0 for error-path leaves.
func assertNonZeroExit(t *testing.T, resp *Response) {
	t.Helper()
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0\nstdout:%s\nstderr:%s\ncombined:%s",
			resp.Stdout, resp.Stderr, resp.Combined)
	}
}

// processArgsLine returns the ps(1) args field for the registry owner PID.
func processArgsLine(t *testing.T, home, sessionID string) (string, error) {
	t.Helper()
	entry, err := ttywatchtest.ReadRegistryEntry(home, sessionID)
	if err != nil {
		return "", err
	}
	if entry.PID <= 0 {
		t.Fatalf("registry %s has invalid pid %d", sessionID, entry.PID)
	}
	out, err := exec.Command("ps", "-p", fmt.Sprintf("%d", entry.PID), "-o", "args=").Output()
	if err != nil {
		return "", fmt.Errorf("ps serve child pid %d: %w", entry.PID, err)
	}
	return strings.TrimSpace(string(out)), nil
}
```