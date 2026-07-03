# Scenario

**Feature**: grok-show-usage PTY-launches grok, submits `/usage show`, prints usage lines

```
# build CLI once, run with GROK_SHOW_USAGE_COMMAND fake TUI by default
caller -> grok-show-usage -> PTY grok (or fake) -> wait prompt -> /usage show -> parse usage
doctest <- stdout: Weekly limit + Next reset (exactly two lines on success)
```

## Preconditions

- `go` is available in PATH (to build `./script/grok/show-usage`). Tests skip otherwise.
- `script/grok/show-usage/main.go` exists with PTY usage-fetch logic (added by implementer).
- Default-suite tests set `GROK_SHOW_USAGE_COMMAND` to a fake interactive TUI script.

## Steps

1. Build `grok-show-usage` once (cached across the process) via `buildShowUsage`.
2. Create isolated `req.TempDir` for each test run.
3. Grouping `Setup` narrows fake TUI variant, error profile, or real-grok backend.
4. Leaf `Setup` sets `ShowUsageCommand`, timeout, or PATH overrides.
5. `Run` executes the CLI and captures stdout/stderr/exit code.
6. Leaf `Assert` checks output lines or error messages.

## Context

- Fake TUI must print `Grok ›` prompt marker, read stdin (the `/usage show` command),
  then print usage lines before returning to prompt.
- Success leaves assert **exact** stdout fixture strings; real-grok leaf asserts patterns only.
- Error leaves assert non-zero exit and stderr substrings.

```go
import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
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
	bin, err := buildShowUsage(t)
	if err != nil {
		return err
	}
	req.Bin = bin
	req.TempDir = t.TempDir()
	return nil
}

// fakeTUIDefault mimics grok TUI: prompt, read command, print fixture usage lines.
func fakeTUIDefault() string {
	return `sh -c 'printf "Grok › "; read -r cmd; printf "Weekly limit: 1%%\nNext reset: July 9, 16:55 PT\n› "'`
}

// fakeTUICustom returns a fake TUI that prints the given weekly limit and reset strings.
func fakeTUICustom(weeklyLimit, nextReset string) string {
	// Escape % for sh printf (e.g. "42%" -> "42%%").
	weeklyLimit = strings.ReplaceAll(weeklyLimit, "%", "%%")
	return fmt.Sprintf(
		`sh -c 'printf "Grok › "; read -r cmd; printf "Weekly limit: %s\nNext reset: %s\n› "'`,
		weeklyLimit, nextReset,
	)
}

// fakeTUINoUsage prints the prompt and reads input but never emits usage lines.
func fakeTUINoUsage() string {
	return `sh -c 'printf "Grok › "; read -r cmd; while true; do sleep 1; done'`
}

// fakeTUIMalformed prints prompt then garbage without parseable usage fields.
func fakeTUIMalformed() string {
	return `sh -c 'printf "Grok › "; read -r cmd; printf "not usage data\n› "'`
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
```