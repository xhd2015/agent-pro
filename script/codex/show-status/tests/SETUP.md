# Scenario

**Feature**: codex-show-status tty-watch ephemeral session, submits `/status`, prints usage lines

```
# build CLI + tty-watch once, run with CODEX_SHOW_STATUS_COMMAND fake TUI by default
caller -> codex-show-status -> tty-watch run/send/snapshot/kill -> parse status
doctest <- stdout: Monthly usage + Credits used + Next reset (exactly three lines on success)
```

## Preconditions

- `go` is available in PATH (to build `./script/codex/show-status` and `github.com/xhd2015/tty-watch/cmd/tty-watch`). Tests skip otherwise.
- `script/codex/show-status/main.go` and `agent/codex/tty/show_status.go` exist (added by implementer).
- Default-suite tests set `CODEX_SHOW_STATUS_COMMAND` to a fake interactive TUI script.
- Each test uses isolated `TTY_WATCH_HOME` under `req.TempDir`.

## Steps

1. Build `codex-show-status` and `tty-watch` once (cached across the process).
2. Create isolated `req.TempDir` and `req.TTYWatchHome` for each test run.
3. Grouping `Setup` narrows fake TUI variant, error profile, or real-codex backend.
4. Leaf `Setup` sets `ShowStatusCommand`, timeout, PATH overrides, or session id.
5. `Run` executes the CLI and captures stdout/stderr/exit code.
6. Leaf `Assert` checks output lines, error messages, or registry cleanup.

## Context

- Fake TUI must print `Codex ›` prompt marker, read stdin (the `/status` command),
  then print status fields before returning to prompt.
- Success leaves assert **exact** stdout fixture strings; real-codex leaf asserts patterns only.
- Error leaves assert non-zero exit and stderr substrings.
- Registry lives at `$TTY_WATCH_HOME/registry/<session-id>.json`.

```go
import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/tty-watch/ttywatchtest"
)

func Setup(t *testing.T, req *Request) error {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skipf("skipping: go not found in PATH")
	}
	repoRoot, err := findModuleRoot()
	if err != nil {
		return err
	}
	req.RepoRoot = repoRoot
	bin, err := buildShowStatus(t)
	if err != nil {
		return err
	}
	req.Bin = bin
	ttyBin, err := buildTTYWatch(t)
	if err != nil {
		return err
	}
	req.TTYWatchBin = ttyBin
	req.TempDir = t.TempDir()
	req.TTYWatchHome = filepath.Join(req.TempDir, ".tty-watch")
	if req.SessionID == "" {
		req.SessionID = "codex-status-usage"
	}
	return nil
}

// fakeTUIDefault mimics codex TUI: prompt, read /status, print canonical fixture status box.
func fakeTUIDefault() string {
	return `sh -c 'printf "Codex › "; read -r cmd; printf "Monthly credit limit: 42%% left (resets 08:00 on 1 Aug)\n6,519 of 11,250 credits used\n› "'`
}

// fakeTUICustom returns a fake TUI that prints the given percent-left, credits, and reset strings.
func fakeTUICustom(percentLeft, creditsUsed, creditsTotal, reset string) string {
	percentLeft = strings.ReplaceAll(percentLeft, "%", "%%")
	return fmt.Sprintf(
		`sh -c 'printf "Codex › "; read -r cmd; printf "Monthly credit limit: %s left (resets %s)\n%s of %s credits used\n› "'`,
		percentLeft, reset, creditsUsed, creditsTotal,
	)
}

// fakeTUIExtraNoise prints MCP warnings and tips before the status box.
func fakeTUIExtraNoise() string {
	return `sh -c 'printf "Codex › "; read -r cmd; printf "Starting MCP servers...\nWarning: MCP server demo failed to start\nTip: type /help for available commands\nMonthly credit limit: 42%% left (resets 08:00 on 1 Aug)\n6,519 of 11,250 credits used\n› "'`
}

// fakeTUINoStatus prints the prompt and reads input but never emits status fields.
func fakeTUINoStatus() string {
	return `sh -c 'printf "Codex › "; read -r cmd; while true; do sleep 1; done'`
}

// fakeTUIMalformed prints prompt then garbage without parseable status fields.
func fakeTUIMalformed() string {
	return `sh -c 'printf "Codex › "; read -r cmd; printf "not status data\n› "'`
}

// assertSuccessExit checks exit 0 and empty stderr for success leaves.
func assertSuccessExit(t *testing.T, resp *Response) {
	t.Helper()
	if resp.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0\nstdout:%s\nstderr:%s", resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	if resp.Stderr != "" {
		t.Fatalf("expected empty stderr on success, got:\n%s", resp.Stderr)
	}
}

// assertRegistrySessionGone fails when the session id still exists in tty-watch registry.
func assertRegistrySessionGone(t *testing.T, home, sessionID string) {
	t.Helper()
	if ttywatchtest.RegistryExists(home, sessionID) {
		t.Fatalf("registry still has %s under %s after fetch", sessionID, ttywatchtest.RegistryDir(home))
	}
}
```