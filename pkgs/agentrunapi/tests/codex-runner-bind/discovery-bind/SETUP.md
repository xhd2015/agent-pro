# Scenario

**Feature**: rollout discovery binds matching cwd + timestamp

```
unbound meta(workspace=W, created_at=T)
  + CODEX_HOME rollout session_meta cwd=W ts≥T id=X
  -> EnsureCodexRunnerBound -> runner_session_id=X (return + store)
```

## Preconditions

- No scrollback inject (discovery path only).
- Rollout timestamp is after meta.created_at.
- Workspace paths match exactly (absolute).

## Steps

1. Seed unbound codex-tty session with fixed created_at.
2. Write one matching rollout under injected CodexHome.
3. Call EnsureCodexRunnerBound; assert persist of Codex id X.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.CreatedAt = "2026-08-11T10:00:00Z"
	req.RunnerSessionID = ""
	req.ScrollbackText = ""
	req.Rollouts = []RolloutSeed{
		{
			CodexSessionID: fixtureCodexIDMatching,
			Cwd:            req.Workspace,
			Timestamp:      "2026-08-11T12:32:25Z",
		},
	}
	return nil
}
```
