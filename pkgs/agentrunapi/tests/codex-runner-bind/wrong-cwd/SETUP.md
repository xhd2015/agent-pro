# Scenario

**Feature**: rollout for another workspace must not bind

```
unbound meta(workspace=W)
  + rollout session_meta cwd=OTHER (≠W) id=X
  -> EnsureCodexRunnerBound -> stay unbound
```

## Preconditions

- Rollout timestamp is fresh (would match if cwd were equal).
- No scrollback inject (isolation of discovery cwd filter).

## Steps

1. Seed unbound codex-tty session for workspace W.
2. Seed rollout with cwd = W+"/other-project" and a valid timestamp.
3. Assert stay unbound (return + store).

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.CreatedAt = "2026-08-11T10:00:00Z"
	req.RunnerSessionID = ""
	req.ScrollbackText = ""
	other := filepath.Join(filepath.Dir(req.Workspace), "other-project")
	req.Rollouts = []RolloutSeed{
		{
			CodexSessionID: fixtureCodexIDMatching,
			Cwd:            other,
			Timestamp:      "2026-08-11T12:32:25Z",
		},
	}
	return nil
}
```
