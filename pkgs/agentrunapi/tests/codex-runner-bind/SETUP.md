# Scenario

**Feature**: bind codex-tty `meta.runner_session_id` from scrollback or CODEX_HOME

```
# discovery
unbound meta(workspace=W, created_at=T) + CODEX_HOME rollout(cwd=W, ts≥T, id=X)
  -> EnsureCodexRunnerBound -> meta.runner_session_id=X (persisted)

# scrollback
unbound meta + SnapshotScrollback("… codex resume Y …")
  -> EnsureCodexRunnerBound -> meta.runner_session_id=Y (persisted)

# gates
already bound A + offer B -> stay A
non-codex runner + offer X -> stay unbound
wrong cwd rollout -> stay unbound
```

## Preconditions

- Package `github.com/xhd2015/agent-pro/pkgs/agentrunapi` gains
  `CodexRunnerBindOpts` and `EnsureCodexRunnerBound` locked in root `DOCTEST.md`.
- Classic TDD: leaves **RED** until those symbols exist and **persist**
  `runner_session_id` via `store.UpdateSessionRunnerSessionID`.
- Parallel-safe: no `t.Setenv` / `t.Chdir` / process globals. Inject store home
  and `CodexHome` under `t.TempDir()`. Scrollback via `SnapshotScrollback` inject.
- Fixtures only (fake rollouts + meta) — no live Codex / TTY registry.

## Steps

1. Root `Setup` creates isolated `req.Home`, `req.CodexHome`, `req.Workspace`.
2. Leaf `Setup` seeds session fields, optional scrollback, optional rollouts.
3. `Run` calls `EnsureCodexRunnerBound`; leaf `Assert` checks return + store re-read.

## Context

- Default session id: `sess-codex-bind-1`.
- Default workspace: `{temp}/workspace` (absolute).
- Default created_at for discovery leaves: `2026-08-11T10:00:00Z`.
- Matching rollout timestamps must be **≥** created_at (or opts.NotBefore).
- Resume footer shape:
  `To continue this session, run codex resume <uuid>`
- Codex ids in fixtures use real-evidence shape (UUID strings), not agent-run ids.

```go
import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

// Fixture Codex session ids (not agent-run session ids).
const (
	fixtureCodexIDMatching = "019fef17-ea39-7623-9ef6-b2376b1556c0"
	fixtureCodexIDOther    = "019fef18-0225-7d10-a08d-8250e433e045"
	fixtureCodexIDBound   = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	fixtureCodexIDOffer    = "11111111-2222-3333-4444-555555555555"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	base := t.TempDir()
	if req.Home == "" {
		req.Home = filepath.Join(base, ".agent-run")
	}
	if req.CodexHome == "" {
		req.CodexHome = filepath.Join(base, ".codex")
	}
	if req.Workspace == "" {
		req.Workspace = filepath.Join(base, "workspace")
	}
	if req.SessionID == "" {
		req.SessionID = "sess-codex-bind-1"
	}
	if req.Runner == "" {
		req.Runner = "codex-tty"
	}
	if req.Status == "" {
		req.Status = "running"
	}
	// Default seed for bind leaves; gates may override.
	if !req.SeedSession {
		req.SeedSession = true
	}
	return nil
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected harness error: %v", err)
	}
}

func assertEqual(t *testing.T, field string, got, want any) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: got %#v, want %#v", field, got, want)
	}
}

func assertBoundPersisted(t *testing.T, resp *Response, wantID string) {
	t.Helper()
	wantID = strings.TrimSpace(wantID)
	if resp == nil {
		t.Fatal("nil response")
	}
	if !resp.Bound {
		t.Fatalf("Bound: got false, want true (id %q)", wantID)
	}
	assertEqual(t, "Meta.RunnerSessionID", strings.TrimSpace(resp.Meta.RunnerSessionID), wantID)
	assertEqual(t, "StoredRunnerSessionID", strings.TrimSpace(resp.StoredRunnerSessionID), wantID)
}

func assertUnbound(t *testing.T, resp *Response) {
	t.Helper()
	if resp == nil {
		t.Fatal("nil response")
	}
	if resp.Bound {
		t.Fatalf("bound: got true, want false; meta.runner_session_id=%q stored=%q",
			resp.Meta.RunnerSessionID, resp.StoredRunnerSessionID)
	}
	if id := strings.TrimSpace(resp.Meta.RunnerSessionID); id != "" {
		t.Fatalf("Meta.RunnerSessionID: got %q, want empty", id)
	}
	if id := strings.TrimSpace(resp.StoredRunnerSessionID); id != "" {
		t.Fatalf("StoredRunnerSessionID: got %q, want empty", id)
	}
}
```
