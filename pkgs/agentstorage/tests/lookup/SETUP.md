# Scenario

**Feature**: isolated FileStore home for runner_session_id lookup + lazy cache

```
t.TempDir/.agent-run
  -> NewFileStore(explicitHome)
  -> CreateSession seeds
  -> optional warm Find/List + mutate
  -> FindByGrokSessionID / ListByRunnerSessionID / IsGrokRunner
  -> Response metas, errors, CacheAfter
```

## Preconditions

- Nested root: no inheritance from `pkgs/agentstorage/tests/` parent.
- `NewFileStore(req.Home)` with explicit home — no `t.Setenv` / `os.Setenv` /
  `t.Chdir`.
- Product symbols (implementer): `IsGrokRunner`, `ListByRunnerSessionID`,
  `FindByGrokSessionID`; store gen bump + `index/by-runner-session/` lazy cache.
- Leaves are `t.Parallel()`-safe; harness does not mutate process globals.

## Steps

1. Root `Setup` creates `req.TempDir` and `req.Home` under `t.TempDir()`.
2. Leaf `Setup` sets `req.Op`, seeds, query id, optional warm/mutate.
3. `Run` opens the store, seeds, optionally warms/mutates, calls product API.
4. Leaf `Assert` checks metas, error strings, and cache filesystem facts.

## Context

- Default seed status is `finished` when omitted.
- Cache paths: `index/generation`, `index/by-runner-session/.gen`,
  `index/by-runner-session/<uuid>.json`.
- Grok Find filter is always runners `grok` and `grok-tty`.

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

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func assertNoRunError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Run harness error: %v", err)
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

func assertExactErr(t *testing.T, err error, want string) {
	t.Helper()
	got := errString(err)
	if got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func cacheHasUUID(snap CacheSnap, uuid string) bool {
	for _, u := range snap.UUIDFiles {
		if u == uuid {
			return true
		}
	}
	return false
}
```
