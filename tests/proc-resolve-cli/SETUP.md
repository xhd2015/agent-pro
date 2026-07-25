# Scenario

**Feature**: agent-pro proc resolve CLI — build binary once, run with controlled env

```
# session-cached agent-pro binary
doctest -> go build ./cmd/agent-pro -> $TMPDIR/proc-resolve-cli-doctest-<session>/agent-pro
# leaves set Args (+ optional AGENT_PRO_PROCRESOLVE_TEST_SNAPSHOT)
caller: agent-pro proc resolve … -> stdout/stderr/exit
```

## Preconditions

- `go` is available in PATH (to build `./cmd/agent-pro`). Tests skip otherwise.
- Module root has `go.mod` and `cmd/agent-pro` (located from `d.DOCTEST_ROOT`).
- Session cache: `$TMPDIR/proc-resolve-cli-doctest-<d.DOCTEST_SESSION_ID>/`
  holds the shared binary (`binaries.ready` sentinel, `build.lock` flock).
- `cmd/agent-pro` gains a `proc` / `proc resolve` command (implementer); RED
  until then (help/unknown may fail or unknown-command).
- JSON hit leaf requires CLI to honor `AGENT_PRO_PROCRESOLVE_TEST_SNAPSHOT`
  when set (see root DSN).

## Steps

1. Skip if `go` missing.
2. Resolve module root from `d.DOCTEST_ROOT`.
3. Build `agent-pro` once per session into the session cache; set `req.Bin`.
4. Leaves set `req.Args` and optional `req.Snapshot`.
5. Root `Run` executes the binary, captures exit/stdout/stderr.

## Context

- User-facing stdout for successful human output should end with `\n` (JSON leaf
  also expects trailing newline on stdout).
- Parallel leaves share only the binary; each Run uses its own TempDir cwd.

```go
import (
	"os/exec"
	"strings"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skipf("skipping: go not found in PATH")
	}
	repoRoot, err := findModuleRoot(d.DOCTEST_ROOT)
	if err != nil {
		return err
	}
	req.RepoRoot = repoRoot

	bin, err := buildAgentProOnce(t, d.DOCTEST_SESSION_ID, repoRoot)
	if err != nil {
		return err
	}
	req.Bin = bin
	return nil
}

func assertExitCode(t *testing.T, resp *Response, want int) {
	t.Helper()
	if resp == nil {
		t.Fatal("response is nil")
	}
	if resp.ExitCode != want {
		t.Fatalf("exit code = %d, want %d\nstdout:\n%s\nstderr:\n%s",
			resp.ExitCode, want, resp.Stdout, resp.Stderr)
	}
}

func assertNonZeroExit(t *testing.T, resp *Response) {
	t.Helper()
	if resp == nil {
		t.Fatal("response is nil")
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0\nstdout:\n%s\nstderr:\n%s",
			resp.Stdout, resp.Stderr)
	}
}

func assertContainsFold(t *testing.T, haystack, needle string) {
	t.Helper()
	if needle == "" {
		t.Fatal("empty needle")
	}
	if !strings.Contains(strings.ToLower(haystack), strings.ToLower(needle)) {
		t.Fatalf("missing %q (case-insensitive) in:\n%s", needle, haystack)
	}
}

func assertContains(t *testing.T, haystack, needle string) {
	t.Helper()
	if needle == "" {
		t.Fatal("empty needle")
	}
	if !strings.Contains(haystack, needle) {
		t.Fatalf("missing %q in:\n%s", needle, haystack)
	}
}

func assertNotContains(t *testing.T, haystack, needle string) {
	t.Helper()
	if strings.Contains(haystack, needle) {
		t.Fatalf("unexpected %q in:\n%s", needle, haystack)
	}
}

func assertStdoutEndsWithNewline(t *testing.T, stdout string) {
	t.Helper()
	if stdout == "" || !strings.HasSuffix(stdout, "\n") {
		t.Fatalf("stdout must end with trailing newline; got %q", stdout)
	}
}
```
