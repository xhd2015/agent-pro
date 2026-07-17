# Scenario

**Feature**: agent-run CLI library extract (`pkgs/agentruncli.Handle`)

```
# library surface
import pkgs/agentruncli
  -> Handle(args)  (not package main)

# thin CLI
cmd/agent-run main
  -> agentruncli.Handle(os.Args[1:])

# smoke
Handle(["--help"]) -> stdout usage
Handle(["not-a-real-command"]) -> error
```

## Preconditions

- Package `github.com/xhd2015/agent-pro/pkgs/agentruncli` (P1 surface; RED until created).
- Public entry pinned: `func Handle(args []string) error`.
- `DOCTEST_ROOT` is `tests/agentruncli`; module root is `../..`.
- No real agent-run binary build, PATH LookPath, iTerm, web server, or network.
- Handle smoke captures stdout/stderr via temporary `os.Stdout`/`os.Stderr`
  redirect under a process-wide mutex (parallel leaf safe).
- Harness stores Handle errors on `Response.ErrString` (harness `error` is nil
  for expected Handle failures such as unknown command).
- Full CLI regression stays in `cmd/agent-run/tests/**` (not this tree).

## Steps

1. Root `Setup` validates Request and records module-relative context.
2. Grouping `Setup` sets `req.Mode`.
3. Leaf `Setup` fills Args / reinforces Mode.
4. `Run` calls package APIs or scans sources; leaf `Assert` checks outcomes.

## Context

- Import path: `github.com/xhd2015/agent-pro/pkgs/agentruncli`.
- Unknown-command probe token: `not-a-real-command` (matches existing
  `cmd/agent-run/tests/cli-edge/unknown-subcommand`).
- Help tokens required in stdout: `Usage:`, `web`, `run`, `sessions`, `status`,
  `--agent-runner`.

```go
import (
	"fmt"
	"strings"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	if req == nil {
		return fmt.Errorf("nil Request")
	}
	// Default empty Args means "no argv" only when Mode is set by descendants.
	// Root does not invent Mode; groupings own Mode assignment.
	if req.Mode != "" {
		req.Mode = strings.TrimSpace(req.Mode)
	}
	return nil
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected harness error: %v", err)
	}
}

func assertHandleError(t *testing.T, resp *Response) {
	t.Helper()
	if resp == nil || resp.ErrString == "" {
		t.Fatal("expected Handle error, got nil/empty")
	}
}

func assertNoHandleError(t *testing.T, resp *Response) {
	t.Helper()
	if resp != nil && resp.ErrString != "" {
		t.Fatalf("unexpected Handle error: %s", resp.ErrString)
	}
}

func assertEqual(t *testing.T, field string, got, want any) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: got %#v, want %#v", field, got, want)
	}
}

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("missing %q in %q", want, got)
	}
}

func assertContainsFold(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(strings.ToLower(got), strings.ToLower(want)) {
		t.Fatalf("missing %q (case-insensitive) in %q", want, got)
	}
}
```
