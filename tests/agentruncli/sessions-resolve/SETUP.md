# Scenario

**Feature**: isolated FileStore + injected RunSessions for sessions resolve

```
t.TempDir/.agent-run
  -> NewFileStore(explicitHome)
  -> CreateSession seeds
  -> agentruncli.RunSessions(args, store, stdout, stderr)
  -> Response{Stdout, Stderr, Err}
```

## Preconditions

- Nested root: no inheritance from `tests/agentruncli/` Handle/stdio-mutex tree.
- Product entry (implementer): `func RunSessions(args []string, store agentstorage.Store, stdout, stderr io.Writer) error`.
- Args are **after** the `sessions` token (e.g. `resolve --grok-session-id UUID`).
- No `t.Setenv` / `os.Setenv` / `os.Stdout` reassignment — inject home and writers.
- Do not edit sealed P1 `pkgs/agentstorage/tests/lookup/`.

## Steps

1. Root `Setup` sets `req.TempDir` and `req.Home` under `t.TempDir()`.
2. Leaf `Setup` sets `req.Args` and optional `req.Seeds`.
3. `Run` opens the store, seeds, calls `RunSessions`, returns captured writers + Err.
4. Leaf `Assert` checks stdout (v3 where exact) and returned error text.

## Context

- Default seed status is `finished` when omitted.
- L2 asserts the **returned** error string (no `agent-run: ` main prefix).
- Resolve help locks a concrete Usage template; sessions help only requires mentioning `resolve`.

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.TempDir = t.TempDir()
	req.Home = filepath.Join(req.TempDir, ".agent-run")
	return nil
}

func assertNoRunError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Run harness error: %v", err)
	}
}

func assertExactErr(t *testing.T, err error, want string) {
	t.Helper()
	got := ""
	if err != nil {
		got = err.Error()
	}
	if got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func assertErrContains(t *testing.T, err error, substrs ...string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error containing %q, got nil", substrs)
	}
	msg := err.Error()
	for _, s := range substrs {
		if !strings.Contains(msg, s) {
			t.Fatalf("error %q missing %q", msg, s)
		}
	}
}

func assertStdoutEmpty(t *testing.T, stdout string) {
	t.Helper()
	if stdout != "" {
		t.Fatalf("expected empty stdout, got %q", stdout)
	}
}
```
